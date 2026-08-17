package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
)

// Instances is the resolver for the instances field.
func (r *instanceFederationResolver) Instances(ctx context.Context, obj *model.InstanceFederation) ([]*model.Instance, error) {
	var instances []model1.Instance
	if err := mapd.Database().Where("`federation_id` = ?", obj.ID).Find(&instances).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	var result []*model.Instance
	for _, instance := range instances {
		result = append(result, convertInstanceFromSQLModel(&instance))
	}
	return result, nil
}

// Routes is the resolver for the routes field.
func (r *instanceFederationResolver) Routes(ctx context.Context, obj *model.InstanceFederation) ([]*model.Route, error) {
	var routes []model1.Route
	if err := mapd.Database().Where("`federation_id` = ?", obj.ID).Find(&routes).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	var result []*model.Route
	for _, route := range routes {
		result = append(result, convertRouteFromSQLModel(&route))
	}
	return result, nil
}

// Component is the resolver for the component field.
func (r *instanceFederationResolver) Component(ctx context.Context, obj *model.InstanceFederation) (*model.Component, error) {
	if obj.Component == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Component(WithGuardSilenced(ctx), obj.Component.ID)
}

// Application is the resolver for the application field.
func (r *instanceFederationResolver) Application(ctx context.Context, obj *model.InstanceFederation) (*model.Application, error) {
	if obj.Application == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Application(WithGuardSilenced(ctx), obj.Application.ID)
}

// Organization is the resolver for the organization field.
func (r *instanceFederationResolver) Organization(ctx context.Context, obj *model.InstanceFederation) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *instanceFederationsResolver) Edge(ctx context.Context, obj *model.InstanceFederations) (*model.InstanceFederationNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.InstanceFederationNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectInstance)
		}
		db = finalizeInstanceFederationQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var federations []model1.InstanceFederation
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&federations).Error; err != nil {
				return err
			}
			for _, federation := range federations {
				nodes.Nodes = append(nodes.Nodes, convertInstanceFederationFromSQLModel(&federation))
			}
			return nil
		}); err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return nodes, nil
}

// CreateInstanceFederation is the resolver for the CreateInstanceFederation field.
func (r *mutationResolver) CreateInstanceFederation(ctx context.Context, input model.CreateInstanceFederation) (*model.InstanceFederation, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}

	federation := &model1.InstanceFederation{ID: uuid.New(), Name: input.Name, ComponentID: model1.NewNullUUID(input.Component)}
	if input.Description != nil {
		federation.Description = *input.Description
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		com := &model1.Component{ID: federation.ComponentID.UUID}
		if err := tx.Take(com).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.component", com.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectInstance, com.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		federation.ApplicationID = com.ApplicationID
		federation.OrganizationID = com.OrganizationID

		var exists int64
		if err := tx.Model(&model1.InstanceFederation{}).Where("`organization_id` = ? AND `name` = ?", federation.OrganizationID, federation.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", federation.Name)
		}

		if err := appendAuditLog(tx, "create", "InstanceFederation", federation.ID, federation.Name, sess.UserID, sess.UserName, federation.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Create(federation).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		for i, each := range input.Instances {
			release := &model1.Release{ID: each.Release}
			if err := tx.Where("`component_id` = ? AND `organization_id` = ?", federation.ComponentID, federation.OrganizationID).Take(release).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, fmt.Sprintf("input.instances.%d.release", i), each.Release.String())
				}
				panic(&ErrDBConnection{err})
			}
			cluster := &model1.Cluster{ID: each.Cluster}
			if err := tx.Where("`organization_id` = ?", federation.OrganizationID).Take(cluster).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, fmt.Sprintf("input.instances.%d.cluster", i), each.Cluster.String())
				}
				panic(&ErrDBConnection{err})
			}

			if *each.Namespace == "" {
				*each.Namespace = "default"
			}
			if defExcludedNamespaces[*each.Namespace] {
				return errInvalidRequest(ctx, fmt.Sprintf("input.instances.%d.namespace", i), "provision on internal namespace is not allowed", *each.Namespace)
			}

			instance := &model1.Instance{
				ID:             uuid.New(),
				Name:           release.Name,
				Namespace:      *each.Namespace,
				CRDs:           release.CRDs,
				ClusterID:      model1.NewNullUUID(each.Cluster),
				ReleaseID:      model1.NewNullUUID(release.ID),
				FederationID:   model1.NewNullUUID(federation.ID),
				ComponentID:    federation.ComponentID,
				ApplicationID:  federation.ApplicationID,
				OrganizationID: federation.OrganizationID,
			}
			if err := tx.Model(&model1.Instance{}).Where("`cluster_id` = ? AND `namespace` = ? AND `name` = ?", instance.ClusterID, instance.Namespace, instance.Name).Count(&exists).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if exists > 0 {
				return errObjectDuplicated(ctx, fmt.Sprintf("input.instances.%d.namespace", i), *each.Namespace)
			}

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
			b, tasks, err := postRenderManifest(wc, release.Manifest, instance.Namespace, map[string]string{
				defLabelApplication: instance.ApplicationID.String(),
				defLabelComponent:   instance.ComponentID.String(),
				defLabelFederation:  instance.FederationID.String(),
				defLabelInstance:    instance.ID.String(),
				defLabelRevision:    release.ID.String(),
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
		}

		if err := tx.Take(federation).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	mapd.EventEmitterInstanceFederation.Next(&mapd.Event{
		Type: mapd.EventTypeAdd,
		OID:  federation.ID,
	})
	return convertInstanceFederationFromSQLModel(federation), nil
}

// UpdateInstanceFederation is the resolver for the UpdateInstanceFederation field.
func (r *mutationResolver) UpdateInstanceFederation(ctx context.Context, input model.UpdateInstanceFederation) (*model.InstanceFederation, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var deletedInstances []*model1.Instance

	federation := &model1.InstanceFederation{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(federation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", federation.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectInstance, federation.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		for i, each := range input.RemoveInstances {
			instance := &model1.Instance{ID: each.ID}
			if err := tx.Where("`federation_id` = ?", federation.ID).Take(instance).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, fmt.Sprintf("input.removeInstances.%d.id", i), each.ID.String())
				}
				panic(&ErrDBConnection{err})
			}

			var count int64
			if err := tx.Model(&model1.RouteBackend{}).Where("`instance_id` = ?", instance.ID).Count(&count).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if count > 0 {
				return errObjectIsBusy(ctx, fmt.Sprintf("input.removeInstances.%d.id", i), "Route", nil, "", each.ID.String())
			}

			if each.Force != nil && *each.Force {
				if !sess.Admin {
					return errInsufficientPrivileges(ctx)
				}

				if err := tx.Delete(instance).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
				deletedInstances = append(deletedInstances, instance)
				continue
			}

			if federation.Status != model1.FederationStatusProvisioned {
				return errInvalidRequest(ctx, "input.id", "object is provisioning or deleting", input.ID.String())
			}

			updates := make(map[string]interface{})
			updates["state"] = model1.InstanceStateDeleting

			if err := tx.Model(instance).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}

		if len(input.ModifyInstances)+len(input.AddInstances) > 0 && federation.Status != model1.FederationStatusProvisioned {
			return errInvalidRequest(ctx, "input.id", "object is provisioning or deleting", input.ID.String())
		}

		for i, each := range input.ModifyInstances {
			instance := &model1.Instance{ID: each.ID}
			if err := tx.Where("`federation_id` = ?", federation.ID).Take(instance).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, fmt.Sprintf("input.modifyInstances.%d.id", i), each.ID.String())
				}
				panic(&ErrDBConnection{err})
			}

			updates := make(map[string]interface{})
			if each.AutoRestart != nil {
				updates["auto_restart"] = *each.AutoRestart
			}
			if each.Release != nil {
				release := &model1.Release{ID: *each.Release}
				if err := tx.Where("`component_id` = ? AND `organization_id` = ?", federation.ComponentID, federation.OrganizationID).Take(release).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return errObjectNotFound(ctx, fmt.Sprintf("input.modifyInstances.%d.release", i), each.Release.String())
					}
					panic(&ErrDBConnection{err})
				}

				cluster := &model1.Cluster{ID: instance.ClusterID.UUID}
				if err := tx.Take(cluster).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
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
				b, tasks, err := postRenderManifest(wc, release.Manifest, instance.Namespace, map[string]string{
					defLabelApplication: instance.ApplicationID.String(),
					defLabelComponent:   instance.ComponentID.String(),
					defLabelFederation:  instance.FederationID.String(),
					defLabelInstance:    instance.ID.String(),
					defLabelRevision:    release.ID.String(),
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
							return errObjectIsBusy(ctx, fmt.Sprintf("input.modifyInstances.%d.id", i), "Route", nil, "", each.ID.String())
						}
					}
				}

				updates["crds"] = string(release.CRDs)
				updates["prev_manifest"] = string(instance.Manifest)
				updates["manifest"] = string(b)
				updates["state"] = model1.InstanceStatePending
				updates["release_id"] = release.ID.String()
				b, err = json.Marshal(tasks)
				if err != nil {
					panic(err)
				}
				updates["status"] = string(b)
			}
			if len(updates) > 0 {
				if err := tx.Model(instance).Updates(updates).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
			}
		}

		for i, each := range input.AddInstances {
			release := &model1.Release{ID: each.Release}
			if err := tx.Where("`component_id` = ? AND `organization_id` = ?", federation.ComponentID, federation.OrganizationID).Take(release).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, fmt.Sprintf("input.addInstances.%d.release", i), each.Release.String())
				}
				panic(&ErrDBConnection{err})
			}
			cluster := &model1.Cluster{ID: each.Cluster}
			if err := tx.Where("`organization_id` = ?", federation.OrganizationID).Take(cluster).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, fmt.Sprintf("input.addInstances.%d.cluster", i), each.Cluster.String())
				}
				panic(&ErrDBConnection{err})
			}

			if *each.Namespace == "" {
				*each.Namespace = "default"
			}
			if defExcludedNamespaces[*each.Namespace] {
				return errInvalidRequest(ctx, fmt.Sprintf("input.addInstances.%d.namespace", i), "provision on internal namespace is not allowed", *each.Namespace)
			}

			instance := &model1.Instance{
				ID:             uuid.New(),
				Name:           release.Name,
				Namespace:      *each.Namespace,
				CRDs:           release.CRDs,
				ClusterID:      model1.NewNullUUID(each.Cluster),
				ReleaseID:      model1.NewNullUUID(release.ID),
				FederationID:   model1.NewNullUUID(federation.ID),
				ComponentID:    federation.ComponentID,
				ApplicationID:  federation.ApplicationID,
				OrganizationID: federation.OrganizationID,
			}
			if each.AutoRestart != nil {
				instance.AutoRestart = *each.AutoRestart
			}
			var exists int64
			if err := tx.Model(&model1.Instance{}).Where("`cluster_id` = ? AND `namespace` = ? AND `name` = ?", instance.ClusterID, instance.Namespace, instance.Name).Count(&exists).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if exists > 0 {
				return errObjectDuplicated(ctx, fmt.Sprintf("input.addInstances.%d.namespace", i), *each.Namespace)
			}

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
			b, tasks, err := postRenderManifest(wc, release.Manifest, instance.Namespace, map[string]string{
				defLabelApplication: instance.ApplicationID.String(),
				defLabelComponent:   instance.ComponentID.String(),
				defLabelFederation:  instance.FederationID.String(),
				defLabelInstance:    instance.ID.String(),
				defLabelRevision:    release.ID.String(),
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
		}

		updates := make(map[string]interface{})
		if input.Description != nil {
			updates["description"] = *input.Description
		}
		if len(input.RemoveInstances)+len(input.ModifyInstances)+len(input.AddInstances) > 0 {
			updates["status"] = model1.FederationStatusPending
		}

		if err := appendAuditLog(tx, "update", "InstanceFederation", federation.ID, federation.Name, sess.UserID, sess.UserName, federation.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(updates) > 0 {
			if err := tx.Model(federation).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(federation).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	for _, each := range deletedInstances {
		mapd.EventEmitterInstance.Next(&mapd.Event{
			Type: mapd.EventTypeRemove,
			OID:  each.ID,
		})
	}
	mapd.EventEmitterInstanceFederation.Next(&mapd.Event{
		Type: mapd.EventTypeModify,
		OID:  federation.ID,
	})
	return convertInstanceFederationFromSQLModel(federation), nil
}

// DeleteInstanceFederation is the resolver for the DeleteInstanceFederation field.
func (r *mutationResolver) DeleteInstanceFederation(ctx context.Context, input model.DeleteInstanceFederation) (*model.InstanceFederation, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		deletedInstances []model1.Instance
		force            = input.Force != nil && *input.Force
	)
	if force && !sess.Admin {
		return nil, errInsufficientPrivileges(ctx)
	}

	federation := &model1.InstanceFederation{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(federation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", input.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermModify, model1.PolicyObjectInstance, federation.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if err := appendAuditLog(tx, "delete", "InstanceFederation", federation.ID, federation.Name, sess.UserID, sess.UserName, federation.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if force {
			if err := tx.Where("`federation_id` = ?", federation.ID).Find(&deletedInstances).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if err := tx.Where("`federation_id` = ?", federation.ID).Delete(&model1.Instance{}).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if err := tx.Delete(federation).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			return nil
		}

		if federation.Status == model1.FederationStatusDeleting {
			return errInvalidRequest(ctx, "input.id", "object is deleting", input.ID.String())
		}

		var exists int64
		if err := tx.Model(&model1.Route{}).
			Where("EXISTS (SELECT 1 FROM `route_backend` INNER JOIN `instance` ON `route_backend`.`instance_id` = `instance`.`id` WHERE `route_backend`.`route_id` = `route`.`id` AND `instance`.`federation_id` = ?)", federation.ID).
			Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectIsBusy(ctx, "input.id", "Route", nil, "", input.ID.String())
		}

		updates := make(map[string]interface{})
		updates["status"] = model1.FederationStatusDeleting
		if err := tx.Model(federation).Updates(updates).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(federation).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	if force {
		for _, each := range deletedInstances {
			mapd.EventEmitterInstance.Next(&mapd.Event{
				Type: mapd.EventTypeRemove,
				OID:  each.ID,
			})
		}
		mapd.EventEmitterInstanceFederation.Next(&mapd.Event{
			Type: mapd.EventTypeRemove,
			OID:  federation.ID,
		})
	} else {
		mapd.EventEmitterInstanceFederation.Next(&mapd.Event{
			Type: mapd.EventTypeModify,
			OID:  federation.ID,
		})
	}
	return convertInstanceFederationFromSQLModel(federation), nil
}

// InstanceFederation is the resolver for the InstanceFederation field.
func (r *queryResolver) InstanceFederation(ctx context.Context, id uuid.UUID) (*model.InstanceFederation, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectInstance)
	}

	federation := &model1.InstanceFederation{}
	if err := db.Where("`id` = ?", id).Take(federation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", id.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertInstanceFederationFromSQLModel(federation), nil
}

// InstanceFederations is the resolver for the InstanceFederations field.
func (r *queryResolver) InstanceFederations(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.InstanceFederations, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.InstanceFederations{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.InstanceFederation{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectInstance)
	}
	db = finalizeInstanceFederationQueryFilter(db, obj.Filters...)
	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// InstanceFederation returns graphql1.InstanceFederationResolver implementation.
func (r *Resolver) InstanceFederation() graphql1.InstanceFederationResolver {
	return &instanceFederationResolver{r}
}

// InstanceFederations returns graphql1.InstanceFederationsResolver implementation.
func (r *Resolver) InstanceFederations() graphql1.InstanceFederationsResolver {
	return &instanceFederationsResolver{r}
}

type instanceFederationResolver struct{ *Resolver }
type instanceFederationsResolver struct{ *Resolver }
