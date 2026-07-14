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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Related is the resolver for the related field.
func (r *clusterExtensionResolver) Related(ctx context.Context, obj *model.ClusterExtension, page *model.QueryPage, filters []*model.QueryFilter) (*model.Clusters, error) {
	return r.Resolver.Query().(*queryResolver).Clusters(WithGuardSilenced(ctx), page, append(filters, &model.QueryFilter{Key: "related", Value: []string{obj.ID.String()}}))
}

// Cluster is the resolver for the cluster field.
func (r *clusterExtensionResolver) Cluster(ctx context.Context, obj *model.ClusterExtension) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// Organization is the resolver for the organization field.
func (r *clusterExtensionResolver) Organization(ctx context.Context, obj *model.ClusterExtension) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *clusterExtensionsResolver) Edge(ctx context.Context, obj *model.ClusterExtensions) (*model.ClusterExtensionNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.ClusterExtensionNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectCluster)
		}
		db = finalizeClusterExtensionQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var exts []model1.ClusterExtension
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&exts).Error; err != nil {
				return err
			}
			for _, ext := range exts {
				nodes.Nodes = append(nodes.Nodes, convertClusterExtensionFromSQLModel(&ext))
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

// EnableClusterExtension is the resolver for the EnableClusterExtension field.
func (r *mutationResolver) EnableClusterExtension(ctx context.Context, input model.EnableClusterExtension) (*model.ClusterExtension, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	ext := &model1.ClusterExtension{
		ID:        uuid.New(),
		Name:      model1.ClusterExtensionName(input.Name.String()),
		State:     model1.ClusterExtensionStateEnabling,
		ClusterID: model1.NewNullUUID(input.Cluster),
	}
	if input.Spec != nil {
		var err error
		ext.Spec, err = json.Marshal(input.Spec)
		if err != nil {
			panic(err)
		}
	}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		c := &model1.Cluster{ID: ext.ClusterID.UUID}
		if err := tx.Take(c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.cluster", c.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		ext.OrganizationID = c.OrganizationID

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectCluster, c.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.ClusterExtension{}).Where("`cluster_id` = ? AND `name` = ?", ext.ClusterID, ext.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", ext.Name.String())
		}

		if err := appendAuditLog(tx, "enable", "ClusterExtension", ext.ID, ext.Name.String(), sess.UserID, sess.UserName, ext.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Create(ext).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(ext).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		for i, each := range input.Related {
			c := &model1.Cluster{ID: each}
			if err := tx.Take(c).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, fmt.Sprintf("input.related.%d", i), each.String())
				}
				panic(&ErrDBConnection{err})
			}
			if c.Provisioned < model1.ClusterProvisionFinish {
				return errInvalidRequest(ctx, fmt.Sprintf("input.related.%d", i), "cluster is not ready", each.String())
			}
			rel := &model1.ClusterExtensionRelatedCluster{
				ID:          uuid.New(),
				ExtensionID: model1.NewNullUUID(ext.ID),
				ClusterID:   ext.ClusterID,
				GuestID:     model1.NewNullUUID(each),
			}
			if err := tx.Create(rel).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	mapd.EventEmitterClusterExtension.Next(&mapd.Event{
		Type: mapd.EventTypeAdd,
		OID:  ext.ID,
	})
	return convertClusterExtensionFromSQLModel(ext), nil
}

// ModifyClusterExtension is the resolver for the ModifyClusterExtension field.
func (r *mutationResolver) ModifyClusterExtension(ctx context.Context, input model.ModifyClusterExtension) (*model.ClusterExtension, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	ext := &model1.ClusterExtension{ID: input.ID}
	updates := make(map[string]interface{})
	var reprovision bool

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(ext).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", ext.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectCluster, ext.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if input.Spec != nil {
			b, err := json.Marshal(input.Spec)
			if err != nil {
				panic(err)
			}
			if spec := string(b); spec != string(ext.Spec) {
				updates["spec"] = spec
				updates["state"] = model1.ClusterExtensionStateEnabling
				updates["status"] = model1.ClusterExtensionStatusPending
			}
		}
		if len(input.RemoveRelated)+len(input.AddRelated) > 0 {
			updates["state"] = model1.ClusterExtensionStateEnabling
			updates["status"] = model1.ClusterExtensionStatusPending
		}

		if err := appendAuditLog(tx, "modify", "ClusterExtension", ext.ID, ext.Name.String(), sess.UserID, sess.UserName, ext.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(updates) > 0 {
			if err := tx.Model(ext).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			reprovision = true
		}
		if err := tx.Take(ext).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		for _, each := range input.RemoveRelated {
			if err := tx.Where("`extension_id` = ? AND `guest_id` = ?", ext.ID, each).Delete(&model1.ClusterExtensionRelatedCluster{}).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		for i, each := range input.AddRelated {
			c := &model1.Cluster{ID: each}
			if err := tx.Take(c).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, fmt.Sprintf("input.addRelated.%d", i), each.String())
				}
				panic(&ErrDBConnection{err})
			}
			if c.Provisioned < model1.ClusterProvisionFinish {
				return errInvalidRequest(ctx, fmt.Sprintf("input.addRelated.%d", i), "cluster is not ready", each.String())
			}
			rel := &model1.ClusterExtensionRelatedCluster{
				ID:          uuid.New(),
				ExtensionID: model1.NewNullUUID(ext.ID),
				ClusterID:   ext.ClusterID,
				GuestID:     model1.NewNullUUID(each),
			}
			if err := tx.Create(rel).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	if reprovision {
		mapd.EventEmitterClusterExtension.Next(&mapd.Event{
			Type: mapd.EventTypeModify,
			OID:  ext.ID,
		})
	}
	return convertClusterExtensionFromSQLModel(ext), nil
}

// DisableClusterExtension is the resolver for the DisableClusterExtension field.
func (r *mutationResolver) DisableClusterExtension(ctx context.Context, input model.DisableClusterExtension) (*model.ClusterExtension, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	ext := &model1.ClusterExtension{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(ext).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", ext.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectCluster, ext.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if err := appendAuditLog(tx, "disable", "ClusterExtension", ext.ID, ext.Name.String(), sess.UserID, sess.UserName, ext.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Model(ext).Updates(map[string]interface{}{
			"state":  model1.ClusterExtensionStateDisabling,
			"status": model1.ClusterExtensionStatusPending,
		}).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(ext).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	mapd.EventEmitterClusterExtension.Next(&mapd.Event{
		Type: mapd.EventTypeModify,
		OID:  ext.ID,
	})
	return convertClusterExtensionFromSQLModel(ext), nil
}

// ClusterExtension is the resolver for the ClusterExtension field.
func (r *queryResolver) ClusterExtension(ctx context.Context, id uuid.UUID) (*model.ClusterExtension, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectCluster)
	}

	ext := &model1.ClusterExtension{ID: id}
	if err := db.Take(ext).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", id.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertClusterExtensionFromSQLModel(ext), nil
}

// ClusterExtensions is the resolver for the ClusterExtensions field.
func (r *queryResolver) ClusterExtensions(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.ClusterExtensions, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.ClusterExtensions{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.ClusterExtension{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectCluster)
	}
	db = finalizeClusterExtensionQueryFilter(db, obj.Filters...)

	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// ClusterExtension returns graphql1.ClusterExtensionResolver implementation.
func (r *Resolver) ClusterExtension() graphql1.ClusterExtensionResolver {
	return &clusterExtensionResolver{r}
}

// ClusterExtensions returns graphql1.ClusterExtensionsResolver implementation.
func (r *Resolver) ClusterExtensions() graphql1.ClusterExtensionsResolver {
	return &clusterExtensionsResolver{r}
}

type clusterExtensionResolver struct{ *Resolver }
type clusterExtensionsResolver struct{ *Resolver }
