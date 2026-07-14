// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
)

func getRestores(client *k8s.VeleroKubernetes, namespace string, selectors ...k8s.Selector) ([]velerov1.Restore, error) {
	list, err := client.GetRestores(namespace, selectors...)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func convertRestore(restore *model.Restore, obj *velerov1.Restore) {
	restore.Name = obj.ObjectMeta.Name
	restore.Labels = obj.ObjectMeta.Labels
	if obj.ObjectMeta.Labels != nil {
		loc, ok := obj.ObjectMeta.Labels["map.tanzu.vmware.com/location"]
		if ok {
			restore.Location = &model.BackupStorageLocation{
				Name: loc,
			}
		}
		backup, ok := obj.ObjectMeta.Labels["map.tanzu.vmware.com/backup"]
		if ok {
			restore.Backup = &model.Backup{
				Name: backup,
			}
		}
	}
	if !obj.ObjectMeta.CreationTimestamp.IsZero() {
		restore.CreatedAt = obj.ObjectMeta.CreationTimestamp.Time
	} else {
		restore.CreatedAt = time.Now()
	}

	restore.Spec = &model.RestoreSpec{
		IncludedNamespaces: obj.Spec.IncludedNamespaces,
		ExcludedNamespaces: obj.Spec.ExcludedNamespaces,
		NamespaceMapping:   obj.Spec.NamespaceMapping,
	}

	switch obj.Status.Phase {
	case velerov1.RestorePhaseNew:
		restore.Status = model.RestoreStatusPending
	case velerov1.RestorePhaseInProgress:
		restore.Status = model.RestoreStatusRunning
	case velerov1.RestorePhaseCompleted:
		restore.Status = model.RestoreStatusSuccess
	case velerov1.RestorePhaseFailedValidation, velerov1.RestorePhasePartiallyFailed, velerov1.RestorePhaseFailed:
		restore.Status = model.RestoreStatusFailure
	default:
		restore.Status = model.RestoreStatusUnknown
	}
	if !obj.Status.CompletionTimestamp.IsZero() {
		restore.LastModified = obj.Status.CompletionTimestamp.Time
	}

	if restore.CreatedAt.IsZero() {
		restore.CreatedAt = time.Unix(0, 0)
	}
	if restore.LastModified.IsZero() {
		restore.LastModified = restore.CreatedAt
	}
}

func getBackupAnnotation(backup *velerov1.Backup) *model1.BackupAnnotation {
	if backup.Annotations == nil {
		return nil
	}
	v, ok := backup.Annotations[model1.BackupAnnotationName]
	if !ok {
		return nil
	}
	obj := &model1.BackupAnnotation{}
	if err := json.Unmarshal([]byte(v), obj); err != nil {
		return nil
	}
	return obj
}

func restoreInstance(tx *gorm.DB, obj *model1.BackupAnnotationInstance, cluster *model1.Cluster, orgID uuid.UUID) (*model1.Instance, error) {
	block, err := cryptoutil.NewAESGCMCipher(cluster.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	wc, err := k8s.NewKubernetesFromToken(cluster.Endpoint, block.Decrypt(cluster.AuthToken), rest.TLSClientConfig{
		Insecure: cluster.InsecureTLS,
		CAData:   cluster.CACert,
	})
	if err != nil {
		panic(err)
	}

	appID, err := uuid.Parse(obj.ApplicationID)
	if err != nil {
		return nil, nil
	}
	comID, err := uuid.Parse(obj.ComponentID)
	if err != nil {
		return nil, nil
	}
	fedID, err := uuid.Parse(obj.FederationID)
	if err != nil {
		return nil, nil
	}
	relID, err := uuid.Parse(obj.ReleaseID)
	if err != nil {
		return nil, nil
	}

	// abort restore if any reference does not exist.
	origApp := &model1.Application{ID: appID}
	if err := tx.Unscoped().Take(origApp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}
	origCom := &model1.Component{ID: comID}
	if err := tx.Unscoped().Take(origCom).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}
	origFed := &model1.InstanceFederation{ID: fedID}
	if err := tx.Unscoped().Take(origFed).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}
	origRel := &model1.Release{ID: relID}
	if err := tx.Unscoped().Take(origRel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	app := &model1.Application{}
	if err := tx.Where("`name` = ? AND `organization_id` = ?", origApp.Name, orgID).Take(app).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			panic(&ErrDBConnection{err})
		}

		app = &model1.Application{
			ID:             uuid.New(),
			Name:           origApp.Name,
			Description:    origApp.Description,
			Avatar:         origApp.Avatar,
			OrganizationID: model1.NewNullUUID(orgID),
		}
		if err := tx.Create(app).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
	}
	com := &model1.Component{}
	if err := tx.Where("`name` = ? AND `organization_id` = ?", origCom.Name, orgID).Take(com).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			panic(&ErrDBConnection{err})
		}

		com = &model1.Component{
			ID:             uuid.New(),
			Name:           origCom.Name,
			Description:    origCom.Description,
			Type:           origCom.Type,
			Avatar:         origCom.Avatar,
			Template:       origCom.Template,
			ApplicationID:  model1.NewNullUUID(app.ID),
			OrganizationID: app.OrganizationID,
		}
		if err := tx.Create(com).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
	} else if com.Type != origCom.Type {
		return nil, fmt.Errorf("destination component type %v does not match its source type %v", com.Type, origCom.Type)
	}
	fed := &model1.InstanceFederation{}
	if err := tx.Where("`name` = ? AND `organization_id` = ?", origFed.Name, orgID).Take(fed).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			panic(&ErrDBConnection{err})
		}

		fed = &model1.InstanceFederation{
			ID:             uuid.New(),
			Name:           origFed.Name,
			Description:    origFed.Description,
			Status:         model1.FederationStatusProvisioned, // no need to provision, it's empty
			ComponentID:    model1.NewNullUUID(com.ID),
			ApplicationID:  com.ApplicationID,
			OrganizationID: app.OrganizationID,
		}
		if err := tx.Create(fed).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
	}

	var origReleases, releases []model1.Release
	if err := tx.Unscoped().Where("`component_id` = ?", origCom.ID).Find(&origReleases).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	sort.Slice(origReleases, func(i, j int) bool { return origReleases[i].Num < origReleases[j].Num })
	if err := tx.Where("`component_id` = ?", com.ID).Find(&releases).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	sort.Slice(releases, func(i, j int) bool { return releases[i].Num < releases[j].Num })

	var rel *model1.Release
	for i := 0; i < len(origReleases); i++ {
		if i >= len(releases) {
			origRelease := origReleases[i]
			release := &model1.Release{
				ID:             uuid.New(),
				Name:           origRelease.Name,
				Num:            origRelease.Num,
				CRDs:           origRelease.CRDs,
				Manifest:       origRelease.Manifest,
				ComponentID:    model1.NewNullUUID(com.ID),
				ApplicationID:  com.ApplicationID,
				OrganizationID: com.OrganizationID,
			}
			if err := tx.Create(release).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if release.Num == origRel.Num {
				rel = release
			}
			continue
		}
		origRelease, release := origReleases[i], releases[i]
		if string(origRelease.CRDs) != string(release.CRDs) || string(origRelease.Manifest) != string(release.Manifest) {
			if release.Num > origRel.Num {
				// those are currently not used, so errors are ignored.
				break
			}
			return nil, fmt.Errorf("release %s-%d is not the same as component %s's source release", release.Name, release.Num, com.Name)
		}
		if release.Num == origRel.Num {
			rel = &releases[i]
		}
	}

	// check if there is conflict in that namespace.
	var deployedInstances int64
	if err := tx.Model(&model1.Instance{}).Where("`name` = ? AND `namespace` = ? AND `cluster_id` = ? AND `federation_id` NOT IN (?)", rel.Name, obj.Namespace, cluster.ID, fed.ID).Count(&deployedInstances).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	if deployedInstances > 0 {
		return nil, fmt.Errorf("%s instance of other federation has been provisoned to namespace %s", rel.Name, obj.Namespace)
	}

	instance := &model1.Instance{}
	if err := tx.Where("`name` = ? AND `namespace` = ? AND `cluster_id` = ?", rel.Name, obj.Namespace, cluster.ID).Take(instance).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			panic(&ErrDBConnection{err})
		}

		instance = &model1.Instance{
			ID:             uuid.New(),
			Name:           rel.Name,
			Namespace:      obj.Namespace,
			CRDs:           rel.CRDs,
			State:          model1.InstanceStateSkipped,
			ClusterID:      model1.NewNullUUID(cluster.ID),
			ReleaseID:      model1.NewNullUUID(rel.ID),
			FederationID:   model1.NewNullUUID(fed.ID),
			ComponentID:    model1.NewNullUUID(com.ID),
			ApplicationID:  com.ApplicationID,
			OrganizationID: app.OrganizationID,
		}

		b, tasks, err := postRenderManifest(wc, rel.Manifest, instance.Namespace, map[string]string{
			defLabelApplication: instance.ApplicationID.String(),
			defLabelComponent:   instance.ComponentID.String(),
			defLabelFederation:  instance.FederationID.String(),
			defLabelInstance:    instance.ID.String(),
			defLabelRevision:    rel.ID.String(),
			defLabelManaged:     "true",
		})
		if err != nil {
			panic(err)
		}
		instance.Manifest = b
		b, err = json.Marshal(tasks)
		if err != nil {
			panic(err)
		}
		instance.Status = b

		if err := tx.Create(instance).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return instance, nil
	}

	b, tasks, err := postRenderManifest(wc, rel.Manifest, instance.Namespace, map[string]string{
		defLabelApplication: instance.ApplicationID.String(),
		defLabelComponent:   instance.ComponentID.String(),
		defLabelFederation:  instance.FederationID.String(),
		defLabelInstance:    instance.ID.String(),
		defLabelRevision:    rel.ID.String(),
		defLabelManaged:     "true",
	})
	if err != nil {
		panic(err)
	}

	// reject the request if the referenced service/ingress is changed.
	var backends []model1.RouteBackend
	if err := tx.Where("`instance_id` = ?", instance.ID).Find(&backends).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	if len(backends) > 0 {
		diff, _, err := compareManifestForBackendRef(instance.Manifest, b)
		if err != nil {
			panic(err)
		}
		diffM := make(map[string]*unstructured.Unstructured)
		for _, each := range diff {
			diffM[fmt.Sprintf("%s-%s", each.GetKind(), each.GetName())] = each
		}

		for _, backend := range backends {
			_, ok := diffM[fmt.Sprintf("%s-%s", backend.Type, backend.Name)]
			if ok {
				route := &model1.Route{ID: backend.RouteID.UUID}
				if err := tx.Take(route).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
				return nil, fmt.Errorf("must remove instance %s from route %s before restore", instance.Name, route.Name)
			}
		}
	}

	updates := make(map[string]interface{})
	updates["state"] = model1.InstanceStateSkipped
	updates["crds"] = string(rel.CRDs)
	updates["prev_manifest"] = string(instance.Manifest)
	updates["manifest"] = string(b)
	updates["release_id"] = rel.ID.String()
	b, err = json.Marshal(tasks)
	if err != nil {
		panic(err)
	}
	updates["status"] = string(b)

	if err := tx.Model(instance).Updates(updates).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	return instance, nil
}
