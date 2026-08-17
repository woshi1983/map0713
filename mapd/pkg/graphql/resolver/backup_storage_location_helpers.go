// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gorm.io/gorm"
	"k8s.io/client-go/rest"
)

func getBackupStorageLocations(client *k8s.VeleroKubernetes, namespace string, selectors ...k8s.Selector) ([]velerov1.BackupStorageLocation, error) {
	list, err := client.GetBackupStorageLocations(namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertBackupStorageLocation(location *model.BackupStorageLocation, obj *velerov1.BackupStorageLocation) {
	location.Name = obj.ObjectMeta.Name
	location.Labels = obj.ObjectMeta.Labels
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		location.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		location.CreatedAt = time.Now()
	}

	location.Spec = &model.BackupStorageLocationSpec{}
	location.Spec.Provider = obj.Spec.Provider
	if obj.Spec.Config != nil {
		location.Spec.Config = &model.BackupStorageLocationSpecConfig{}
		if v, ok := obj.Spec.Config["region"]; ok {
			location.Spec.Config.Region = &v
		}
		if v, ok := obj.Spec.Config["s3ForcePathStyle"]; ok {
			v, err := strconv.ParseBool(v)
			if err == nil {
				location.Spec.Config.S3ForcePathStyle = &v
			}
		}
		if v, ok := obj.Spec.Config["s3Url"]; ok {
			location.Spec.Config.S3Url = abstract.MustParseURL(v)
		}
	}
	if obj.Spec.ObjectStorage != nil {
		objStorage := &model.BackupStorageLocationSpecObjectStorage{
			Bucket: obj.Spec.ObjectStorage.Bucket,
		}
		if obj.Spec.ObjectStorage.Prefix != "" {
			objStorage.Prefix = util.StringPtr(obj.Spec.ObjectStorage.Prefix)
		}
		if obj.Spec.ObjectStorage.CACert != nil {
			objStorage.CaCert = util.StringPtr(string(obj.Spec.ObjectStorage.CACert))
		}
		location.Spec.ObjectStorage = objStorage
	}
	switch obj.Spec.AccessMode {
	case velerov1.BackupStorageLocationAccessModeReadWrite:
		location.Spec.AccessMode = model.BackupStorageLocationAccessModeReadWrite
	case velerov1.BackupStorageLocationAccessModeReadOnly:
		location.Spec.AccessMode = model.BackupStorageLocationAccessModeReadOnly
	}

	if obj.Status.Phase == velerov1.BackupStorageLocationPhaseAvailable {
		location.Status = model.BackupStorageLocationStatusAvailable
	} else {
		location.Status = model.BackupStorageLocationStatusUnavailable
	}
	if !obj.Status.LastSyncedTime.IsZero() {
		location.LastModified = obj.Status.LastSyncedTime.Time
	}

	if location.CreatedAt.IsZero() {
		location.CreatedAt = time.Unix(0, 0)
	}
	if location.LastModified.IsZero() {
		location.LastModified = location.CreatedAt
	}
}

func getClusterWithTrustedCACertsForCreateBackupStorageLocation(ctx context.Context, id uuid.UUID) (*model1.Cluster, []byte, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	c := &model1.Cluster{ID: id}
	var caCerts []byte
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadWrite, model1.PolicyObjectClusterResource)
		}
		if err := db.Take(c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.cluster", c.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if c.Class != model1.ClusterClassUnknown {
			// provisioned cluster
			provider := &model1.ClusterProvider{ID: c.ProviderID.UUID}
			if err := tx.Take(provider).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			providerSpec := &model1.ClusterProviderTKGSSpec{}
			if err := json.Unmarshal(provider.Spec, providerSpec); err != nil {
				panic(err)
			}

			switch provider.Class {
			case model1.ClusterClassTKGS:
				spec := &model1.ClusterTKGSSpec{}
				if err := json.Unmarshal(c.Spec, spec); err != nil {
					panic(err)
				}

				block, err := cryptoutil.NewAESGCMCipher(provider.ID[:]) // AES-128
				if err != nil {
					panic(err)
				}

				sc, err := k8s.NewTanzuKubernetesGridService(providerSpec.Endpoint, providerSpec.AuthUser, block.Decrypt(providerSpec.AuthPassword), rest.TLSClientConfig{
					Insecure: providerSpec.InsecureSkipTLSVerify,
					CAData:   providerSpec.CACert,
				})
				if err != nil {
					return errGatewayError(ctx, err)
				}
				tkc, err := sc.GetTanzuKubernetesCluster(spec.Namespace, c.Name)
				if err != nil {
					return errGatewayError(ctx, err)
				}

				var buf bytes.Buffer
				if tkc.Spec.Settings != nil && tkc.Spec.Settings.Network != nil && tkc.Spec.Settings.Network.Trust != nil {
					for _, each := range tkc.Spec.Settings.Network.Trust.AdditionalTrustedCAs {
						b, err := base64.StdEncoding.DecodeString(each.Data)
						if err != nil {
							continue
						}
						buf.Write(bytes.TrimSpace(b))
						buf.WriteString("\n")
					}
				}
				caCerts = buf.Bytes()

			case model1.ClusterClassTKGM:
				spec := &model1.ClusterTKGMSpec{}
				if err := json.Unmarshal(c.Spec, spec); err != nil {
					panic(err)
				}

				var buf bytes.Buffer
				for _, each := range spec.TrustedCAs {
					buf.Write(bytes.TrimSpace(each))
					buf.WriteString("\n")
				}
				caCerts = buf.Bytes()

			case model1.ClusterClassTKG2S:
				spec := &model1.ClusterTKG2SSpec{}
				if err := json.Unmarshal(c.Spec, spec); err != nil {
					panic(err)
				}

				var buf bytes.Buffer
				for _, each := range spec.TrustedCAs {
					buf.Write(bytes.TrimSpace(each))
					buf.WriteString("\n")
				}
				caCerts = buf.Bytes()

			case model1.ClusterClassTKG2M:
				spec := &model1.ClusterTKG2MSpec{}
				if err := json.Unmarshal(c.Spec, spec); err != nil {
					panic(err)
				}

				var buf bytes.Buffer
				for _, each := range spec.TrustedCAs {
					buf.Write(bytes.TrimSpace(each))
					buf.WriteString("\n")
				}
				caCerts = buf.Bytes()

			default:
				panic(fmt.Errorf("unsupported cluster provider class: %s", provider.Class))
			}
		} else {
			spec := &model1.ClusterExternalSpec{}
			if err := json.Unmarshal(c.Spec, spec); err != nil {
				panic(err)
			}
			var buf bytes.Buffer
			for _, each := range spec.RootCAs {
				buf.WriteString(strings.TrimSpace(each))
				buf.WriteString("\n")
			}
			caCerts = buf.Bytes()
		}

		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, nil, err
	}
	return c, caCerts, nil
}
