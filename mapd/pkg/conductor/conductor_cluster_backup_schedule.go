// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conductor

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func init() {
	Registry.Register("clusterBackupSchedule", DispatcherFactory(
		mapd.EventEmitterClusterBackupSchedule,
		func() error {
			var schedules []model.ClusterBackupSchedule
			if err := mapd.Database().Select("id").Find(&schedules).Error; err != nil {
				return fmt.Errorf("error obtaining existing cluster restores: %v", err)
			}
			go func() {
				for _, each := range schedules {
					mapd.EventEmitterClusterBackupSchedule.Next(&mapd.Event{
						Type: mapd.EventTypeAdd,
						OID:  each.ID,
					})
				}
			}()
			return nil
		},
		syncClusterBackupSchedule,
	))
}

// syncClusterBackupSchedule handle ClusterBackupSchedule transactions.
func syncClusterBackupSchedule(id uuid.UUID, log *zap.SugaredLogger) time.Duration {
	defer log.Sync()
	next := time.Duration(-1)

	obj := &model.ClusterBackupSchedule{ID: id}
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

		log.Infof("preparing schedule %s for launching backup...", strconv.Quote(obj.Name))
		if due := time.Until(obj.NextFire); due > 0 {
			log.Infof("pending backup for schedule %s. [ETA %v].", strconv.Quote(obj.Name), due)
			next = due
			return nil
		}

		c := &model.Cluster{ID: obj.ClusterID.UUID}
		if err := tx.Take(c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				deleted = true
				return nil
			}
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

		_, err = velero.GetBackupStorageLocation(obj.Namespace, obj.Location)
		if err != nil {
			if !k8s.IsErrObjectNotExists(err) {
				log.Errorf("failed to retrieve state of backup storage location %s/%s on cluster %s: %v.", obj.Namespace, obj.Location, strconv.Quote(c.Name), err)
				return nil
			}
			log.Warnf("backup storage location %s/%s no longer exist on cluster %s. Cleanup schedule %s...", obj.Namespace, obj.Location, strconv.Quote(c.Name), strconv.Quote(obj.Name))
			if err := tx.Delete(obj).Error; err != nil {
				return err
			}
			deleted = true
			return nil
		}

		spec := &model.ClusterBackupScheduleSpec{}
		if err := json.Unmarshal(obj.Spec, spec); err != nil {
			panic(err)
		}
		var relatedComponents []model.Component
		if err := tx.Where("`id` IN (SELECT `component_id` FROM `cluster_backup_schedule_component` WHERE `schedule_id` = ?)", obj.ID).Find(&relatedComponents).Error; err != nil {
			return err
		}

		backup := &velerov1.Backup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%d", obj.Name, obj.Fired+1),
				Namespace: defNamespace,
				Labels: map[string]string{
					defLabelManaged: "true",
				},
			},
			Spec: velerov1.BackupSpec{
				DefaultVolumesToRestic: util.BoolPtr(spec.UseRestic),
				StorageLocation:        obj.Location,
			},
		}

		includedNamespacesM := make(map[string]bool)
		excludedNamespacesM := make(map[string]bool)
		for each := range defExcludedNamespaces {
			excludedNamespacesM[each] = true
		}
		for _, each := range spec.ExcludedNamespaces {
			excludedNamespacesM[each] = true
		}
		if len(spec.IncludedNamespaces) == 0 {
			namespaceList, err := wc.GetNamespaces()
			if err != nil {
				log.Errorf("failed to retrieve list of namespaces of cluster %s: %v.", strconv.Quote(c.Name), err)
				return nil
			}
			for _, each := range namespaceList.Items {
				spec.IncludedNamespaces = append(spec.IncludedNamespaces, each.GetName())
			}
		}
		for _, each := range spec.IncludedNamespaces {
			if excludedNamespacesM[each] {
				continue
			}

			instancesNotReady := int64(0)
			if err := tx.Model(&model.Instance{}).Where("`state` IN (?, ?, ?) AND `namespace` = ? AND `cluster_id` = ?", model.InstanceStatePending, model.InstanceStateProvisioning, model.InstanceStateDeleting, each, c.ID).Count(&instancesNotReady).Error; err != nil {
				return err
			}
			if instancesNotReady > 0 {
				log.Warnf("at least %d instance(s) are not ready. Pending backup for schedule %s until they are ready...", strconv.Quote(obj.Name))
				return nil
			}
			includedNamespacesM[each] = true
		}

		var (
			includedNamespaces, excludedNamespaces []string
			annotation                             model.BackupAnnotation
			db                                     = tx
		)
		if len(relatedComponents) > 0 {
			selector := metav1.LabelSelectorRequirement{
				Key:      defLabelComponent,
				Operator: metav1.LabelSelectorOpIn,
			}

			var (
				ss   []string
				args []interface{}
			)
			for _, each := range relatedComponents {
				ss = append(ss, "?")
				args = append(args, each.ID.String())
				selector.Values = append(selector.Values, each.ID.String())
			}
			db = db.Where("`component_id` IN ("+strings.Join(ss, ", ")+")", args...)

			backup.Spec.LabelSelector = &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{selector},
			}
		}
		var instanceCount int
		for each := range includedNamespacesM {
			instances := []model.Instance{}
			if err := db.Where("`state` = ? AND `namespace` = ? AND `cluster_id` = ?", model.InstanceStateSuccess, each, c.ID).Find(&instances).Error; err != nil {
				return err
			}
			for _, instance := range instances {
				annotation.Instances = append(annotation.Instances, model.BackupAnnotationInstance{
					Name:          instance.Name,
					Namespace:     instance.Namespace,
					ReleaseID:     instance.ReleaseID.String(),
					FederationID:  instance.FederationID.String(),
					ComponentID:   instance.ComponentID.String(),
					ApplicationID: instance.ApplicationID.String(),
				})
			}
			includedNamespaces = append(includedNamespaces, each)
			instanceCount += len(instances)
		}
		for each := range excludedNamespacesM {
			excludedNamespaces = append(excludedNamespaces, each)
		}
		b, err := json.Marshal(annotation)
		if err != nil {
			panic(err)
		}
		backup.Annotations = map[string]string{
			model.BackupAnnotationName: string(b),
		}
		backup.Spec.IncludedNamespaces = includedNamespaces
		backup.Spec.ExcludedNamespaces = excludedNamespaces

		log.Infof("launching backup %s for schedule %s on cluster %s. [%d component(s), %d instance(s)]", strconv.Quote(backup.Name), strconv.Quote(obj.Name), strconv.Quote(c.Name), len(relatedComponents), instanceCount)
		updates := make(map[string]interface{})
		_, err = velero.CreateBackup(backup)
		if err != nil {
			if k8s.IsErrObjectAlreadyExists(err) {
				log.Warnf("failed to create backup for schedule %s on cluster %s due to: %s already exist. It will be retried with another name later.", strconv.Quote(obj.Name), strconv.Quote(c.Name), backup.Name)
				updates["fired"] = obj.Fired + 1
				if err := tx.Model(obj).Updates(updates).Error; err != nil {
					return err
				}
				return nil
			}
			log.Errorf("failed to create backup for schedule %s on cluster %s due to: %v.", strconv.Quote(obj.Name), strconv.Quote(c.Name), err)
			return nil
		}

		sche, err := cron.ParseStandard(obj.Schedule)
		if err != nil {
			panic(err)
		}
		now := time.Now()
		updates["last_fire"] = now
		nextDue := sche.Next(now)
		next = time.Until(nextDue)
		updates["next_fire"] = nextDue
		updates["fired"] = obj.Fired + 1
		if err := tx.Model(obj).Updates(updates).Error; err != nil {
			return err
		}
		log.Infof("backup has been successfully launched for schedule %s on cluster %s. Next launch will be: %s.", strconv.Quote(obj.Name), strconv.Quote(c.Name), nextDue.Format(time.RFC3339))
		return nil
	}); err != nil {
		deleted = false
		log.Warnf("error occurred: %v.", err)
		return time.Second
	}

	if deleted {
		mapd.EventEmitterClusterBackupSchedule.Next(&mapd.Event{
			Type: mapd.EventTypeRemove,
			OID:  id,
		})
	}
	return next
}

var (
	defExcludedNamespaces = map[string]bool{
		"avi-system":                   true,
		"build-service":                true,
		"cert-manager":                 true,
		"default":                      true,
		"kpack":                        true,
		"kube-node-lease":              true,
		"kube-public":                  true,
		"kube-system":                  true,
		"stacks-operator-system":       true,
		"tanzu-package-repo-global":    true,
		"tanzu-system-dashboards":      true,
		"tanzu-system-ingress":         true,
		"tanzu-system-logging":         true,
		"tanzu-system-monitoring":      true,
		"tkg-system":                   true,
		"tkg-system-public":            true,
		"vmware-system-auth":           true,
		"vmware-system-cloud-provider": true,
		"vmware-system-csi":            true,
		"vmware-system-elastic":        true,
		"vmware-system-map":            true,
		"vmware-system-octant":         true,
	}
)
