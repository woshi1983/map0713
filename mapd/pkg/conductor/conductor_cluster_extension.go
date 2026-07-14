// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conductor

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/conductor/cluster"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/conductor/cluster/extension"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"k8s.io/client-go/rest"
)

func init() {
	Registry.Register("clusterExtension", DispatcherFactory(
		mapd.EventEmitterClusterExtension,
		func() error {
			var exts []model.ClusterExtension
			if err := mapd.Database().Select("id").Where("`state` IN (?, ?) AND `status` IN (?, ?)",
				model.ClusterExtensionStateEnabling,
				model.ClusterExtensionStateDisabling,
				model.ClusterExtensionStatusPending,
				model.ClusterExtensionStatusProvisioning,
			).Find(&exts).Error; err != nil {
				return fmt.Errorf("error obtaining existing cluster extensions: %v", err)
			}
			go func() {
				for _, each := range exts {
					mapd.EventEmitterClusterExtension.Next(&mapd.Event{
						Type: mapd.EventTypeAdd,
						OID:  each.ID,
					})
				}
			}()
			return nil
		},
		syncClusterExtension,
	))
}

// syncClusterExtension handle ClusterExtension transactions.
func syncClusterExtension(id uuid.UUID, log *zap.SugaredLogger) time.Duration {
	defer log.Sync()
	next := time.Duration(-1)

	obj := &model.ClusterExtension{ID: id}
	finished := false
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		defer log.Sync()

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(obj).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				finished = true
				return nil
			}
			return err
		}

		switch obj.Status {
		case model.ClusterExtensionStatusProvisioned, model.ClusterExtensionStatusFailed:
			finished = true
			return nil
		}

		c := &model.Cluster{ID: obj.ClusterID.UUID}
		if err := tx.Take(c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				finished = true
				return nil
			}
			panic(err)
		}

		if c.Provisioned < model.ClusterProvisionFinish {
			log.Debugf("pending provision of extension %s on cluster %s.", obj.Name, strconv.Quote(c.Name))
			return nil
		}

		var related []model.ClusterExtensionRelatedCluster
		if err := tx.Where("`extension_id` = ?", obj.ID).Find(&related).Error; err != nil {
			return err
		}
		for _, each := range related {
			rc := &model.Cluster{}
			if err := tx.Where("`id` = ?", each.GuestID).Take(rc).Error; err != nil {
				return err
			}
			if err := tx.Where("`cluster_id` = ?", rc.ID).Find(&rc.Extensions).Error; err != nil {
				return err
			}
			obj.Related = append(obj.Related, *rc)
		}

		var registries []model.ImageRegistry
		if err := tx.Raw("SELECT `image_registry`.* FROM `cluster_image_registry` INNER JOIN `image_registry` ON `cluster_image_registry`.`image_registry_id` = `image_registry`.`id` WHERE `image_registry`.`deleted_at` IS NULL AND `cluster_image_registry`.`cluster_id` = ? ORDER BY `default` DESC", c.ID).Scan(&registries).Error; err != nil {
			return err
		}
		if len(registries) == 0 {
			log.Errorf("no available image registry for cluster %s.", strconv.Quote(c.Name))
			return nil
		}
		imageRegistry := &registries[0]

		block, err := cryptoutil.NewAESGCMCipher(c.ID[:]) // AES-128
		if err != nil {
			panic(err)
		}
		wc, err := k8s.NewKubernetesFromToken(c.Endpoint, block.Decrypt(c.AuthToken), rest.TLSClientConfig{
			Insecure: c.InsecureTLS,
			CAData:   c.CACert,
		})
		if err != nil {
			panic(err)
		}

		updates := map[string]interface{}{"status": model.ClusterExtensionStatusProvisioning}

		// retrieve trusted root CA certificates
		var handler cluster.Handler
		if c.Class != model.ClusterClassUnknown {
			// provisioned cluster
			provider := &model.ClusterProvider{ID: c.ProviderID.UUID}
			if err := tx.Take(provider).Error; err != nil {
				return err
			}

			switch provider.Class {
			case model.ClusterClassTKGS:
				handler = cluster.NewTKGSKubernetes(c, registries, provider, log)
			case model.ClusterClassTKGM:
				handler = cluster.NewTKGMKubernetes(c, registries, provider, log)
			case model.ClusterClassTKG2S:
				handler = cluster.NewTKG2SKubernetes(c, registries, provider, log)
			case model.ClusterClassTKG2M:
				handler = cluster.NewTKG2MKubernetes(c, registries, provider, log)
			default:
				return fmt.Errorf("unsupported cluster provider class: %s", provider.Class)
			}
		} else {
			// attached cluster
			handler = cluster.NewExternalKubernetes(c, registries, log)
		}
		caCerts, err := handler.GetTrustedCACerts()
		if err != nil {
			log.Errorf("failed to retrieve trusted CA certificates for cluster %s due to: %v.", strconv.Quote(c.Name), err)
			return nil
		}

		installer, err := extension.Registry.NewInstaller(&extension.InstallerOptions{
			ClusterName:         c.Name,
			ClusterClass:        c.Class,
			Kubernetes:          wc,
			ImageRegistry:       imageRegistry,
			LoadBalancerEnabled: c.LoadBalancerID.Valid,
			CACerts:             caCerts,
			Namespace:           defNamespace,
			Extension:           obj,
			Logger:              log,
		})
		if err != nil {
			updates["status"] = model.ClusterExtensionStatusFailed
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}
			log.Errorf("extension %s is failed to be installed to cluster %s due to: %v.", obj.Name, strconv.Quote(c.Name), err)
			finished = true
			return nil
		}

		switch obj.State {
		case model.ClusterExtensionStateEnabling:
			if err := installer.Install(); err != nil {
				if extension.IsCriticalError(err) {
					updates["status"] = model.ClusterExtensionStatusFailed
					if err := tx.Model(obj).Updates(updates).Error; err != nil {
						return err
					}
					log.Errorf("extension %s is failed to be installed to cluster %s due to: %v.", obj.Name, strconv.Quote(c.Name), err)
					finished = true
					return nil
				}

				log.Infof("extension %s is still installing on cluster %s.", obj.Name, strconv.Quote(c.Name))
				if err := tx.Model(obj).Updates(updates).Error; err != nil {
					return err
				}
				return nil
			}

			// wait for another deadline threshold for endpoint
			if time.Since(obj.UpdatedAt) > defExtensionEndpointDeadlineSecs*time.Second {
				updates["status"] = model.ClusterExtensionStatusFailed
				if err := tx.Model(obj).Updates(updates).Error; err != nil {
					return err
				}
				log.Errorf("extension %s is failed to be installed to cluster %s due to: Endpoint could not be fulfilled: Timed out.", obj.Name, strconv.Quote(c.Name))
				finished = true
				return nil
			}
			endpoints, err := installer.GetEndpoints()
			if err != nil {
				log.Warnf("failed to get endpoints of extension %s from cluster %s due to: %v.", obj.Name, strconv.Quote(c.Name), err)
				return nil
			}
			b, err := json.Marshal(endpoints)
			if err != nil {
				panic(err)
			}
			updates["endpoints"] = string(b)
			updates["status"] = model.ClusterExtensionStatusProvisioned
			updates["state"] = model.ClusterExtensionStateEnabled
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}
			log.Infof("extension %s has been successfully installed to cluster %s.", obj.Name, strconv.Quote(c.Name))

		case model.ClusterExtensionStateDisabling:
			if err := installer.Uninstall(); err != nil {
				if extension.IsCriticalError(err) {
					updates["status"] = model.ClusterExtensionStatusFailed
					if err := tx.Model(obj).Updates(updates).Error; err != nil {
						return err
					}
					log.Errorf("extension %s is failed to be uninstalled from cluster %s due to: %v.", obj.Name, strconv.Quote(c.Name), err)
					finished = true
					return nil
				}

				log.Infof("extension %s is still uninstalling on cluster %s.", obj.Name, strconv.Quote(c.Name))
				if err := tx.Model(obj).Updates(updates).Error; err != nil {
					return err
				}
				return nil
			}
			if err := tx.Delete(obj).Error; err != nil {
				return err
			}
			log.Infof("extension %s has been successfully uninstalled from cluster %s.", obj.Name, strconv.Quote(c.Name))
		}

		finished = true
		return nil
	}); err != nil {
		finished = false
		log.Warnf("error occurred: %v.", err)
		return time.Second
	}

	if finished {
		mapd.EventEmitterClusterExtension.Next(&mapd.Event{
			Type: mapd.EventTypeRemove,
			OID:  id,
		})
	}
	return next
}
