// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conductor

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/conductor/cluster"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	Registry.Register("cluster", DispatcherFactory(
		mapd.EventEmitterCluster,
		func() error {
			var clusters []model.Cluster
			if err := mapd.Database().Select("id").Find(&clusters).Error; err != nil {
				return fmt.Errorf("error obtaining existing clusters: %v", err)
			}
			go func() {
				for _, each := range clusters {
					mapd.EventEmitterCluster.Next(&mapd.Event{
						Type: mapd.EventTypeAdd,
						OID:  each.ID,
					})
				}
			}()
			return nil
		},
		syncCluster,
	))
}

// syncCluster handle Cluster transactions.
func syncCluster(id uuid.UUID, log *zap.SugaredLogger) time.Duration {
	defer log.Sync()
	next := time.Duration(-1)

	obj := &model.Cluster{ID: id}
	deleted := false
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		defer log.Sync()

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(obj).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				deleted = true
				return nil
			}
			return err
		}

		if obj.Status == model.ClusterStatusMaintenance {
			log.Debugf("cluster %s is in maintenance mode, skipping...", strconv.Quote(obj.Name))
			return nil
		}
		if obj.Status == model.ClusterStatusInitializing && obj.State == model.ClusterStateFailed {
			return nil
		}

		if obj.Provisioned < model.ClusterProvisionFinish {
			log = log.Named("phase").Named(strconv.FormatInt(int64(obj.ProvisionPhase), 10))
			defer log.Sync()
		}

		var registries []model.ImageRegistry
		if err := tx.Raw("SELECT `image_registry`.* FROM `cluster_image_registry` INNER JOIN `image_registry` ON `cluster_image_registry`.`image_registry_id` = `image_registry`.`id` WHERE `image_registry`.`deleted_at` IS NULL AND `cluster_image_registry`.`cluster_id` = ? ORDER BY `default` DESC", obj.ID).Scan(&registries).Error; err != nil {
			return err
		}
		if len(registries) == 0 {
			log.Errorf("no available image registry for cluster %s.", strconv.Quote(obj.Name))
			return nil
		}

		var handler cluster.Handler
		if obj.Class != model.ClusterClassUnknown {
			// provisioned cluster
			provider := &model.ClusterProvider{ID: obj.ProviderID.UUID}
			if err := tx.Take(provider).Error; err != nil {
				return err
			}

			switch provider.Class {
			case model.ClusterClassTKGS:
				handler = cluster.NewTKGSKubernetes(obj, registries, provider, log)
			case model.ClusterClassTKGM:
				handler = cluster.NewTKGMKubernetes(obj, registries, provider, log)
			case model.ClusterClassTKG2S:
				handler = cluster.NewTKG2SKubernetes(obj, registries, provider, log)
			case model.ClusterClassTKG2M:
				handler = cluster.NewTKG2MKubernetes(obj, registries, provider, log)
			default:
				return fmt.Errorf("unsupported cluster provider class: %s", provider.Class)
			}
		} else {
			// attached cluster
			handler = cluster.NewExternalKubernetes(obj, registries, log)
		}

		updates := make(map[string]interface{})
		switch obj.ProvisionPhase {
		case model.ClusterProvisionPrepare:
			updates["status"] = model.ClusterStatusInitializing
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}

			if err := handler.ProvisionPrepare(); err != nil {
				if cluster.IsCriticalError(err) {
					updates["state"] = model.ClusterStateFailed
					if err := tx.Model(obj).Updates(updates).Error; err != nil {
						return err
					}
					return nil
				}
				return nil
			}

			updates["provisioned"] = model.ClusterProvisionPrepare
			updates["provision_phase"] = model.ClusterProvisionKubernetes
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}
			log.Infof("ready for provisioning cluster %s. Pending next phase...", strconv.Quote(obj.Name))
			next = 0
			return nil

		case model.ClusterProvisionKubernetes:
			updates["status"] = model.ClusterStatusInitializing
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}

			endpoint, state, caCert, err := handler.ProvisionKubernetes()
			if err != nil {
				if cluster.IsCriticalError(err) {
					updates["state"] = model.ClusterStateFailed
					if err := tx.Model(obj).Updates(updates).Error; err != nil {
						return err
					}
					return nil
				}
				return nil
			}
			updates["state"] = state
			if state == model.ClusterStateRunning && endpoint != "" {
				defer log.Infof("cluster %s is spawned and running. Pending next phase...", strconv.Quote(obj.Name))

				updates["endpoint"] = endpoint
				if caCert != nil {
					updates["ca_cert"] = string(caCert)
				}
				updates["provisioned"] = model.ClusterProvisionKubernetes
				updates["provision_phase"] = model.ClusterProvisionNamespace
				next = 0
			}
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}
			return nil

		case model.ClusterProvisionNamespace:
			updates["status"] = model.ClusterStatusInitializing
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}

			if err := handler.ProvisionNamespace(); err != nil {
				if cluster.IsCriticalError(err) {
					updates["state"] = model.ClusterStateFailed
					if err := tx.Model(obj).Updates(updates).Error; err != nil {
						return err
					}
					return nil
				}
				return nil
			}

			updates["provisioned"] = model.ClusterProvisionNamespace
			updates["provision_phase"] = model.ClusterProvisionServiceAccount
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}
			next = 0
			return nil

		case model.ClusterProvisionServiceAccount:
			updates["status"] = model.ClusterStatusInitializing
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}

			if err := handler.ProvisionServiceAccount(); err != nil {
				if cluster.IsCriticalError(err) {
					updates["state"] = model.ClusterStateFailed
					if err := tx.Model(obj).Updates(updates).Error; err != nil {
						return err
					}
					return nil
				}
				return nil
			}

			updates["provisioned"] = model.ClusterProvisionServiceAccount
			updates["provision_phase"] = model.ClusterProvisionClusterRoleBinding
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}
			next = 0
			return nil

		case model.ClusterProvisionClusterRoleBinding:
			updates["status"] = model.ClusterStatusInitializing
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}

			token, err := handler.ProvisionClusterRoleBinding()
			if err != nil {
				if cluster.IsCriticalError(err) {
					updates["state"] = model.ClusterStateFailed
					if err := tx.Model(obj).Updates(updates).Error; err != nil {
						return err
					}
					return nil
				}
				return nil
			}

			if len(token) < 1 {
				return nil
			}

			c, err := cryptoutil.NewAESGCMCipher(obj.ID[:]) // AES-128
			if err != nil {
				panic(err)
			}
			updates["auth_token"] = string(c.Encrypt(token))
			updates["provisioned"] = model.ClusterProvisionClusterRoleBinding
			updates["provision_phase"] = model.ClusterProvisionValidation
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}
			next = 0
			return nil

		case model.ClusterProvisionValidation:
			updates["status"] = model.ClusterStatusInitializing
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}

			if err := handler.ProvisionValidation(); err != nil {
				if cluster.IsCriticalError(err) {
					updates["state"] = model.ClusterStateFailed
					if err := tx.Model(obj).Updates(updates).Error; err != nil {
						return err
					}
					return nil
				}
				return nil
			}

			updates["provisioned"] = model.ClusterProvisionValidation
			updates["provision_phase"] = model.ClusterProvisionPostScripts
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}
			next = 0
			return nil

		case model.ClusterProvisionPostScripts:
			updates["status"] = model.ClusterStatusInitializing
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}

			if err := handler.ProvisionPostScripts(); err != nil {
				if cluster.IsCriticalError(err) {
					updates["state"] = model.ClusterStateFailed
					if err := tx.Model(obj).Updates(updates).Error; err != nil {
						return err
					}
					return nil
				}
				return nil
			}

			updates["provisioned"] = model.ClusterProvisionPostScripts
			updates["provision_phase"] = model.ClusterProvisionFinish
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return fmt.Errorf("could not update cluster due to: %v", err)
			}
			next = 0
			return nil

		case model.ClusterProvisionFinish:
			if obj.Provisioned != model.ClusterProvisionFinish {
				// perform final phase
				updates["status"] = model.ClusterStatusInitializing
				if err := tx.Model(obj).Updates(updates).Error; err != nil {
					return err
				}

				if err := handler.ProvisionFinish(); err != nil {
					if cluster.IsCriticalError(err) {
						updates["state"] = model.ClusterStateFailed
						if err := tx.Model(obj).Updates(updates).Error; err != nil {
							return err
						}
						return nil
					}
					return nil
				}

				updates["provisioned"] = model.ClusterProvisionFinish
				if err := tx.Model(obj).Updates(updates).Error; err != nil {
					return err
				}
				log.Infof("finalized provision of cluster %s.", strconv.Quote(obj.Name))
				next = 0
				return nil
			}

			// cluster is fully provisioned.
			state, err := handler.CheckHealth()
			if err != nil {
				updates["status"] = model.ClusterStatusUnhealthy
				log.Warnf("cluster %s is unhealthy.", strconv.Quote(obj.Name))
			} else {
				updates["status"] = model.ClusterStatusHealthy
				if obj.Status == model.ClusterStatusUnhealthy {
					log.Infof("cluster %s is recovered and healthy.", strconv.Quote(obj.Name))
				}
			}
			updates["state"] = state
			if err := tx.Model(obj).Updates(updates).Error; err != nil {
				return err
			}
			return nil
		}
		return nil
	}); err != nil {
		deleted = false
		log.Warnf("error occurred: %v.", err)
		return time.Second
	}

	if deleted {
		mapd.EventEmitterCluster.Next(&mapd.Event{
			Type: mapd.EventTypeRemove,
			OID:  id,
		})
	}
	return next
}
