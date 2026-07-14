// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conductor

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"k8s.io/client-go/rest"
)

func init() {
	Registry.Register("clusterRestore", DispatcherFactory(
		mapd.EventEmitterClusterRestore,
		func() error {
			var restores []model.ClusterRestore
			if err := mapd.Database().Select("id").Find(&restores).Error; err != nil {
				return fmt.Errorf("error obtaining existing cluster restores: %v", err)
			}
			go func() {
				for _, each := range restores {
					mapd.EventEmitterClusterRestore.Next(&mapd.Event{
						Type: mapd.EventTypeAdd,
						OID:  each.ID,
					})
				}
			}()
			return nil
		},
		syncClusterRestore,
	))
}

// syncClusterRestore handle ClusterRestore transactions.
func syncClusterRestore(id uuid.UUID, log *zap.SugaredLogger) time.Duration {
	defer log.Sync()
	next := time.Duration(-1)

	obj := &model.ClusterRestore{ID: id}
	var federationsToUpdate []uuid.UUID
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

		c := &model.Cluster{ID: obj.ClusterID.UUID}
		if err := tx.Take(c).Error; err != nil {
			panic(err)
		}

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
		velero := k8s.NewVelero(wc)

		restore, err := velero.GetRestore(obj.Namespace, obj.Name)
		if err != nil {
			if !k8s.IsErrObjectNotExists(err) {
				log.Errorf("failed to retrieve state of restore %s/%s on cluster %s: %v.", obj.Namespace, obj.Name, strconv.Quote(c.Name), err)
				return nil
			}
			if err := tx.Delete(obj).Error; err != nil {
				return err
			}
			deleted = true
			return nil
		}

		switch restore.Status.Phase {
		case "", velerov1.RestorePhaseNew, velerov1.RestorePhaseInProgress:
			log.Infof("restore %s/%s on cluster %s is still running.", obj.Namespace, obj.Name, strconv.Quote(c.Name))
			return nil
		case velerov1.RestorePhaseFailedValidation, velerov1.RestorePhaseFailed, velerov1.RestorePhasePartiallyFailed:
			log.Errorf("restore %s/%s on cluster %s entered %s state. Instances will not be reprovisioned.", obj.Namespace, obj.Name, strconv.Quote(c.Name), strconv.Quote(string(restore.Status.Phase)))

			var insts []model.ClusterRestoreInstance
			if err := tx.Where("`restore_id` = ?", obj.ID).Find(&insts).Error; err != nil {
				return err
			}
			for _, each := range insts {
				if err := tx.Model(&model.Instance{}).Where("`id` = ?", each.InstanceID).Updates(map[string]interface{}{
					"state": model.InstanceStateFailure,
				}).Error; err != nil {
					return err
				}
			}

			if err := tx.Delete(obj).Error; err != nil {
				return err
			}
			deleted = true
			log.Infof("restore %s/%s on cluster %s is finished.", obj.Namespace, obj.Name, strconv.Quote(c.Name))
			return nil
		}
		log.Infof("restore %s/%s on cluster %s entered %s state. Notifying related instances to start provision...", obj.Namespace, obj.Name, strconv.Quote(c.Name), strconv.Quote(string(restore.Status.Phase)))

		var insts []model.ClusterRestoreInstance
		if err := tx.Where("`restore_id` = ?", obj.ID).Find(&insts).Error; err != nil {
			return err
		}
		fedMap := make(map[uuid.UUID]uuid.UUID)
		for _, each := range insts {
			fedMap[each.FederationID.UUID] = each.FederationID.UUID
		}
		for each := range fedMap {
			if err := tx.Model(&model.Instance{}).Where("`state` = ? AND `federation_id` = ?", model.InstanceStateSkipped, each).Updates(map[string]interface{}{
				"state": model.InstanceStatePending,
			}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.InstanceFederation{ID: each}).Updates(map[string]interface{}{
				"status": model.FederationStatusPending,
			}).Error; err != nil {
				return err
			}
			log.Infof("instances of federation %v requires provision due to restore %s/%s on cluster %s.", each, obj.Namespace, obj.Name, strconv.Quote(c.Name))
			federationsToUpdate = append(federationsToUpdate, each)
		}

		if err := tx.Delete(obj).Error; err != nil {
			return err
		}
		deleted = true
		log.Infof("restore %s/%s on cluster %s is finished.", obj.Namespace, obj.Name, strconv.Quote(c.Name))
		return nil
	}); err != nil {
		deleted = false
		log.Warnf("error occurred: %v.", err)
		return time.Second
	}

	if deleted {
		mapd.EventEmitterClusterRestore.Next(&mapd.Event{
			Type: mapd.EventTypeRemove,
			OID:  id,
		})
		for _, each := range federationsToUpdate {
			mapd.EventEmitterInstanceFederation.Next(&mapd.Event{
				Type: mapd.EventTypeModify,
				OID:  each,
			})
		}
	}
	return next
}
