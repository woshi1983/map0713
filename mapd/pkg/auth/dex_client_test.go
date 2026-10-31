// Copyright (c) 2026 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/dexidp/dex/api/v2"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
)

type recordingDexClient struct {
	api.DexClient
	deleted   int
	created   int
	updated   int
	exists    bool
	deleteErr error
	updateErr error
}

func (d *recordingDexClient) DeleteClient(context.Context, *api.DeleteClientReq, ...grpc.CallOption) (*api.DeleteClientResp, error) {
	d.deleted++
	if d.deleteErr != nil {
		return nil, d.deleteErr
	}
	return &api.DeleteClientResp{}, nil
}
func (d *recordingDexClient) CreateClient(_ context.Context, req *api.CreateClientReq, _ ...grpc.CallOption) (*api.CreateClientResp, error) {
	d.created++
	if req.Client.Secret != "persisted-secret" {
		panic("unexpected secret")
	}
	return &api.CreateClientResp{AlreadyExists: d.exists}, nil
}
func (d *recordingDexClient) UpdateClient(context.Context, *api.UpdateClientReq, ...grpc.CallOption) (*api.UpdateClientResp, error) {
	d.updated++
	if d.updateErr != nil {
		return nil, d.updateErr
	}
	return &api.UpdateClientResp{}, nil
}

func TestInitDexClientRequiresExplicitSecretForStaticClient(t *testing.T) {
	cfg := &oauth2.Config{ClientID: clientID, ClientSecret: "persisted-secret", RedirectURL: "https://map.example/callback"}
	withoutSecret := &recordingDexClient{exists: true, deleteErr: errors.New("client is read-only")}
	if err := initDexClient(withoutSecret, cfg, true, false); err == nil {
		t.Fatal("a static client without its configured secret must be rejected")
	}
	if withoutSecret.created != 0 {
		t.Fatal("must not register an unusable client")
	}

	withSecret := &recordingDexClient{exists: true, deleteErr: errors.New("client is read-only"), updateErr: errors.New("client is read-only")}
	if err := initDexClient(withSecret, cfg, true, true); err != nil {
		t.Fatal(err)
	}
	if withSecret.deleted != 1 || withSecret.created != 1 || withSecret.updated != 1 {
		t.Fatalf("unexpected static client calls: %+v", withSecret)
	}
}

func TestInitDexClientOnlyReplacesOnFirstMigration(t *testing.T) {
	cfg := &oauth2.Config{ClientID: clientID, ClientSecret: "persisted-secret", RedirectURL: "https://map.example/api/auth/.well-known/callback"}
	first := &recordingDexClient{}
	if err := initDexClient(first, cfg, true, false); err != nil {
		t.Fatal(err)
	}
	if first.deleted != 1 || first.created != 1 {
		t.Fatalf("first migration should replace once: %+v", first)
	}

	repeated := &recordingDexClient{exists: true}
	if err := initDexClient(repeated, cfg, false, false); err != nil {
		t.Fatal(err)
	}
	if repeated.deleted != 0 || repeated.created != 1 || repeated.updated != 1 {
		t.Fatalf("subsequent start must reuse client: %+v", repeated)
	}
}
