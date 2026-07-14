package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	cron "github.com/robfig/cron/v3"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Components is the resolver for the components field.
func (r *clusterBackupScheduleResolver) Components(ctx context.Context, obj *model.ClusterBackupSchedule, page *model.QueryPage, filters []*model.QueryFilter) (*model.Components, error) {
	return r.Resolver.Query().(*queryResolver).Components(WithGuardSilenced(ctx), page, append(filters, &model.QueryFilter{Key: "clusterBackupSchedule", Value: []string{obj.ID.String()}}))
}

// Cluster is the resolver for the cluster field.
func (r *clusterBackupScheduleResolver) Cluster(ctx context.Context, obj *model.ClusterBackupSchedule) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// Organization is the resolver for the organization field.
func (r *clusterBackupScheduleResolver) Organization(ctx context.Context, obj *model.ClusterBackupSchedule) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *clusterBackupSchedulesResolver) Edge(ctx context.Context, obj *model.ClusterBackupSchedules) (*model.ClusterBackupScheduleNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.ClusterBackupScheduleNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterResource)
		}
		db = finalizeClusterBackupScheduleQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var exts []model1.ClusterBackupSchedule
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&exts).Error; err != nil {
				return err
			}
			for _, ext := range exts {
				nodes.Nodes = append(nodes.Nodes, convertClusterBackupScheduleFromSQLModel(&ext))
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

// CreateClusterBackupSchedule is the resolver for the CreateClusterBackupSchedule field.
func (r *mutationResolver) CreateClusterBackupSchedule(ctx context.Context, input model.CreateClusterBackupSchedule) (*model.ClusterBackupSchedule, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}
	if input.Location == "" {
		return nil, errInvalidRequest(ctx, "input.location", "value cannot be empty", input.Location)
	}
	c, err := cron.ParseStandard(input.Schedule)
	if err != nil {
		return nil, errInvalidRequest(ctx, "input.schedule", "invalid cron expression", input.Schedule)
	}

	schedule := &model1.ClusterBackupSchedule{
		ID:        uuid.New(),
		Name:      input.Name,
		Namespace: defNamespace,
		Location:  input.Location,
		Schedule:  input.Schedule,
		NextFire:  c.Next(time.Now()),
		ClusterID: model1.NewNullUUID(input.Cluster),
	}
	spec := &model1.ClusterBackupScheduleSpec{}
	if input.Spec != nil {
		if input.Spec.UseRestic != nil && *input.Spec.UseRestic {
			spec.UseRestic = *input.Spec.UseRestic
		}
		spec.IncludedNamespaces = input.Spec.IncludedNamespaces
		spec.ExcludedNamespaces = input.Spec.ExcludedNamespaces
		if input.Spec.TTL != nil {
			spec.TTL = *input.Spec.TTL
		}
	}
	schedule.Spec, err = json.Marshal(input.Spec)
	if err != nil {
		panic(err)
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		c := &model1.Cluster{ID: schedule.ClusterID.UUID}
		if err := tx.Take(c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.cluster", c.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		schedule.OrganizationID = c.OrganizationID

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectClusterResource, c.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.ClusterBackupSchedule{}).Where("`cluster_id` = ? AND `name` = ?", schedule.ClusterID, schedule.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", schedule.Name)
		}

		if err := appendAuditLog(tx, "create", "ClusterBackupSchedule", schedule.ID, schedule.Name, sess.UserID, sess.UserName, schedule.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Create(schedule).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(schedule).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if input.Spec != nil {
			for i, each := range input.Spec.Components {
				if err := tx.Model(&model1.Component{}).Where("`id` = ?", each).Count(&exists).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
				if exists < 1 {
					return errObjectNotFound(ctx, fmt.Sprintf("input.spec.components.%d", i), each.String())
				}

				com := &model1.ClusterBackupScheduleComponent{
					ID:          uuid.New(),
					ComponentID: model1.NewNullUUID(each),
					ScheduleID:  model1.NewNullUUID(schedule.ID),
				}
				if err := tx.Create(com).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
			}
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	mapd.EventEmitterClusterBackupSchedule.Next(&mapd.Event{
		Type: mapd.EventTypeAdd,
		OID:  schedule.ID,
	})
	return convertClusterBackupScheduleFromSQLModel(schedule), nil
}

// UpdateClusterBackupSchedule is the resolver for the UpdateClusterBackupSchedule field.
func (r *mutationResolver) UpdateClusterBackupSchedule(ctx context.Context, input model.UpdateClusterBackupSchedule) (*model.ClusterBackupSchedule, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	schedule := &model1.ClusterBackupSchedule{ID: input.ID}
	updates := make(map[string]interface{})

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(schedule).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", schedule.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectClusterResource, schedule.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if input.Schedule != nil {
			c, err := cron.ParseStandard(*input.Schedule)
			if err != nil {
				return errInvalidRequest(ctx, "input.schedule", "invalid cron expression", *input.Schedule)
			}
			updates["schedule"] = *input.Schedule
			updates["next_fire"] = c.Next(time.Now())
		}

		if err := appendAuditLog(tx, "update", "ClusterBackupSchedule", schedule.ID, schedule.Name, sess.UserID, sess.UserName, schedule.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(updates) > 0 {
			if err := tx.Model(schedule).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(schedule).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	mapd.EventEmitterClusterBackupSchedule.Next(&mapd.Event{
		Type: mapd.EventTypeModify,
		OID:  schedule.ID,
	})
	return convertClusterBackupScheduleFromSQLModel(schedule), nil
}

// DeleteClusterBackupSchedule is the resolver for the DeleteClusterBackupSchedule field.
func (r *mutationResolver) DeleteClusterBackupSchedule(ctx context.Context, input model.DeleteClusterBackupSchedule) (*model.ClusterBackupSchedule, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	schedule := &model1.ClusterBackupSchedule{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(schedule).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", schedule.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermModify, model1.PolicyObjectClusterResource, schedule.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if err := appendAuditLog(tx, "delete", "ClusterBackupSchedule", schedule.ID, schedule.Name, sess.UserID, sess.UserName, schedule.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Delete(schedule).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	mapd.EventEmitterClusterBackupSchedule.Next(&mapd.Event{
		Type: mapd.EventTypeModify,
		OID:  schedule.ID,
	})
	return convertClusterBackupScheduleFromSQLModel(schedule), nil
}

// ClusterBackupSchedule is the resolver for the ClusterBackupSchedule field.
func (r *queryResolver) ClusterBackupSchedule(ctx context.Context, id uuid.UUID) (*model.ClusterBackupSchedule, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterResource)
	}

	schedule := &model1.ClusterBackupSchedule{ID: id}
	if err := db.Take(schedule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", id.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertClusterBackupScheduleFromSQLModel(schedule), nil
}

// ClusterBackupSchedules is the resolver for the ClusterBackupSchedules field.
func (r *queryResolver) ClusterBackupSchedules(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.ClusterBackupSchedules, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.ClusterBackupSchedules{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.ClusterBackupSchedule{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterResource)
	}
	db = finalizeClusterBackupScheduleQueryFilter(db, obj.Filters...)

	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// ClusterBackupSchedule returns graphql1.ClusterBackupScheduleResolver implementation.
func (r *Resolver) ClusterBackupSchedule() graphql1.ClusterBackupScheduleResolver {
	return &clusterBackupScheduleResolver{r}
}

// ClusterBackupSchedules returns graphql1.ClusterBackupSchedulesResolver implementation.
func (r *Resolver) ClusterBackupSchedules() graphql1.ClusterBackupSchedulesResolver {
	return &clusterBackupSchedulesResolver{r}
}

type clusterBackupScheduleResolver struct{ *Resolver }
type clusterBackupSchedulesResolver struct{ *Resolver }
