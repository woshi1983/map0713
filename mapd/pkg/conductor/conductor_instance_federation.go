// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conductor

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	Registry.Register("instanceFederation", DispatcherFactory(
		mapd.EventEmitterInstanceFederation,
		func() error {
			var federations []model.InstanceFederation
			if err := mapd.Database().Select("id").Where("`status` != ?", model.FederationStatusProvisioned).Find(&federations).Error; err != nil {
				return fmt.Errorf("error obtaining existing instance federations: %v", err)
			}
			go func() {
				for _, each := range federations {
					mapd.EventEmitterInstanceFederation.Next(&mapd.Event{
						Type: mapd.EventTypeAdd,
						OID:  each.ID,
					})
				}
			}()
			return nil
		},
		syncInstanceFederation,
	))
}

// syncInstanceFederation handle InstanceFederation transactions.
func syncInstanceFederation(id uuid.UUID, log *zap.SugaredLogger) time.Duration {
	defer log.Sync()
	next := time.Duration(-1)

	federation := &model.InstanceFederation{ID: id}
	var instances []model.Instance
	finished := false
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(federation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				finished = true
				return nil
			}
			return err
		}

		if federation.Status == model.FederationStatusProvisioned {
			finished = true
			return nil
		}

		if federation.Status == model.FederationStatusDeleting {
			if err := tx.Where("`federation_id` = ?", federation.ID).Find(&instances).Error; err != nil {
				return err
			}
			log.Infof("deleting federation %s... [%d instance(s) left]", federation.Name, len(instances))

			if len(instances) == 0 {
				if err := tx.Delete(federation).Error; err != nil {
					return err
				}
				log.Infof("federation %s has been deleted.", federation.Name)
				finished = true
				return nil
			}

			instances = instances[:0]
			if err := tx.Where("`federation_id` = ? AND `state` != ?", federation.ID, model.InstanceStateDeleting).Clauses(clause.Locking{Strength: "UPDATE"}).Find(&instances).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Instance{}).Where("`federation_id` = ? AND `state` != ?", federation.ID, model.InstanceStateDeleting).Updates(map[string]interface{}{"state": model.InstanceStateDeleting}).Error; err != nil {
				return err
			}
			finished = true
			return nil
		}

		if err := tx.Where("`federation_id` = ? AND `state` NOT IN (?, ?, ?)", federation.ID, model.InstanceStateSkipped, model.InstanceStateSuccess, model.InstanceStateFailure).Find(&instances).Error; err != nil {
			return err
		}
		log.Infof("federation %s is provisioning... [%d instance(s) left]", federation.Name, len(instances))

		updates := make(map[string]interface{})
		if len(instances) > 0 {
			updates["status"] = model.FederationStatusProvisioning
			if err := tx.Model(&model.InstanceFederation{ID: federation.ID}).Updates(updates).Error; err != nil {
				return err
			}
			finished = true
			return nil
		}

		updates["status"] = model.FederationStatusProvisioned
		if err := tx.Model(federation).Updates(updates).Error; err != nil {
			return err
		}
		log.Infof("federation %s has been provisioned.", federation.Name)
		finished = true
		return nil
	}); err != nil {
		finished = false
		log.Warnf("error occurred: %v.", err)
		return time.Second
	}

	if finished {
		mapd.EventEmitterInstanceFederation.Next(&mapd.Event{
			Type: mapd.EventTypeRemove,
			OID:  id,
		})
		switch federation.Status {
		case model.FederationStatusPending:
			for _, each := range instances {
				mapd.EventEmitterInstance.Next(&mapd.Event{
					Type: mapd.EventTypeAdd,
					OID:  each.ID,
				})
			}
		case model.FederationStatusDeleting:
			for _, each := range instances {
				mapd.EventEmitterInstance.Next(&mapd.Event{
					Type: mapd.EventTypeModify,
					OID:  each.ID,
				})
			}
		}
	}
	return next
}
