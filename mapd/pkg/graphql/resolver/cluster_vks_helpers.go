// Copyright (c) 2026 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"context"
	"errors"
	"fmt"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/vks"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var errVKSClusterOwnershipMismatch = errors.New("Supervisor cluster ownership or UID mismatch")

func refreshVKSClusterReplicaCounts(ctx context.Context, cluster *model1.Cluster, result *model.Cluster, fetch func(context.Context, string, string) (*unstructured.Unstructured, error)) error {
	spec, ok := result.Spec.(*model.VKSClusterSpec)
	if !ok {
		return fmt.Errorf("VKS cluster spec is unavailable")
	}
	remote, err := fetch(ctx, spec.Namespace, cluster.Name)
	if err != nil {
		return err
	}
	if remote == nil {
		return fmt.Errorf("Supervisor returned an empty VKS cluster")
	}
	if remote.GetLabels()[vks.OwnerLabel] != cluster.ID.String() || spec.RemoteUID == "" || string(remote.GetUID()) != spec.RemoteUID {
		return errVKSClusterOwnershipMismatch
	}
	cp, workers, err := vks.VKSReplicas(remote)
	if err != nil {
		return err
	}
	spec.CtlPlaneReplicas, spec.WorkerReplicas = cp, workers
	return nil
}
