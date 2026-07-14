package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/identicon"
	"gorm.io/gorm"
)

// CreateOrganization is the resolver for the CreateOrganization field.
func (r *mutationResolver) CreateOrganization(ctx context.Context, input model.CreateOrganization) (*model.Organization, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}

	organization := &model1.Organization{ID: uuid.New(), Name: input.Name}
	if input.Description != nil {
		organization.Description = *input.Description
	}
	if input.Avatar != nil {
		organization.Avatar = input.Avatar.String()
	}
	if organization.Avatar == "" {
		organization.Avatar = identicon.NewDataURL(organization.ID[:]).String()
	}
	if input.Link != nil {
		organization.Link = input.Link.String()
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		var exists int64
		if err := tx.Model(&model1.Organization{}).Where("`name` = ?", organization.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", organization.Name)
		}

		if err := appendAuditLog(tx, "create", "Organization", organization.ID, organization.Name, sess.UserID, sess.UserName, organization.ID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Create(organization).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(organization).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		quota := &model1.OrganizationQuota{ID: uuid.New(), OrganizationID: model1.NewNullUUID(organization.ID)}
		if input.Quota != nil {
			if input.Quota.Cpus != nil {
				quota.CPUs = *input.Quota.Cpus
			}
			if input.Quota.MemoryMi != nil {
				quota.MemoryMi = *input.Quota.MemoryMi
			}
			if input.Quota.StorageGi != nil {
				quota.StorageGi = *input.Quota.StorageGi
			}
		}
		if err := tx.Create(quota).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		// current user is the owner of the organization.
		rb := &model1.RoleBinding{
			ID:             uuid.New(),
			UserBindingID:  model1.NewNullUUID(sess.UserID),
			Role:           model1.RoleAdmin,
			OrganizationID: model1.NewNullUUID(organization.ID),
		}
		if err := tx.Create(rb).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertOrganizationFromSQLModel(organization), nil
}

// UpdateOrganization is the resolver for the UpdateOrganization field.
func (r *mutationResolver) UpdateOrganization(ctx context.Context, input model.UpdateOrganization) (*model.Organization, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	organization := &model1.Organization{ID: input.ID}
	updates := make(map[string]interface{})
	if input.Name != nil {
		if !validateName(*input.Name) {
			return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), *input.Name)
		}
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Avatar != nil {
		updates["avatar"] = input.Avatar.String()
	}
	if input.Link != nil {
		updates["link"] = input.Link.String()
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectOrganization, input.ID) {
			return errInsufficientPrivileges(ctx)
		}

		if name, ok := updates["name"]; ok {
			var exists int64
			if err := tx.Model(&model1.Organization{}).Where("`id` != ? AND `name` = ?", organization.ID, name).Count(&exists).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if exists > 0 {
				return errObjectDuplicated(ctx, "input.name", name)
			}
		}

		if err := appendAuditLog(tx, "update", "Organization", organization.ID, organization.Name, sess.UserID, sess.UserName, organization.ID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(updates) > 0 {
			if err := tx.Model(organization).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(organization).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", organization.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if input.Quota != nil {
			update := make(map[string]interface{})
			if input.Quota.Cpus != nil {
				update["cpus"] = *input.Quota.Cpus
			}
			if input.Quota.MemoryMi != nil {
				update["memory_mi"] = *input.Quota.MemoryMi
			}
			if input.Quota.StorageGi != nil {
				update["storage_gi"] = *input.Quota.StorageGi
			}
			if len(update) > 0 {
				if err := tx.Model(&model1.OrganizationQuota{}).Where("`organization_id` = ?", organization.ID).Updates(update).Error; err != nil {
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
	return convertOrganizationFromSQLModel(organization), nil
}

// DeleteOrganization is the resolver for the DeleteOrganization field.
func (r *mutationResolver) DeleteOrganization(ctx context.Context, input model.DeleteOrganization) (*model.Organization, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	organization := &model1.Organization{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(organization).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", organization.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermModify, model1.PolicyObjectOrganization, organization.ID) {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.Route{}).Where("`organization_id` = ?", organization.ID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectIsBusy(ctx, "input.id", "Route", nil, "", input.ID.String())
		}
		if err := tx.Model(&model1.Cluster{}).Where("`organization_id` = ?", organization.ID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectIsBusy(ctx, "input.id", "Cluster", nil, "", input.ID.String())
		}

		if err := appendAuditLog(tx, "delete", "Organization", organization.ID, organization.Name, sess.UserID, sess.UserName, organization.ID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Delete(organization).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertOrganizationFromSQLModel(organization), nil
}

// RoleBindings is the resolver for the roleBindings field.
func (r *organizationResolver) RoleBindings(ctx context.Context, obj *model.Organization, page *model.QueryPage, filters []*model.QueryFilter) (*model.RoleBindings, error) {
	return r.Resolver.Query().(*queryResolver).RoleBindings(WithGuardSilenced(ctx), page, append(filters, &model.QueryFilter{Key: "organization", Value: []string{obj.ID.String()}}))
}

// Used is the resolver for the used field.
func (r *organizationQuotaResolver) Used(ctx context.Context, obj *model.OrganizationQuota) (*model.OrganizationQuotaUsage, error) {
	var cpus, memoryMi, storageGi int64
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		cpus, memoryMi, storageGi = getOrganizationQuotaUsage(tx, obj.ID)
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return &model.OrganizationQuotaUsage{
		Cpus:      cpus,
		MemoryMi:  memoryMi,
		StorageGi: storageGi,
	}, nil
}

// Total is the resolver for the total field.
func (r *organizationQuotaResolver) Total(ctx context.Context, obj *model.OrganizationQuota) (*model.OrganizationQuotaUsage, error) {
	var cpus, memoryMi, storageGi int64
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		cpus, memoryMi, storageGi = getOrganizationQuota(tx, obj.ID)
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return &model.OrganizationQuotaUsage{
		Cpus:      cpus,
		MemoryMi:  memoryMi,
		StorageGi: storageGi,
	}, nil
}

// Edge is the resolver for the edge field.
func (r *organizationsResolver) Edge(ctx context.Context, obj *model.Organizations) (*model.OrganizationNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.OrganizationNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) && !sess.Admin {
			db = db.Where("`id` IN (SELECT `organization_id` FROM `role_binding` WHERE `user_binding_id` = ?)", sess.UserID)
		}
		db = finalizeOrganizationQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var organizations []model1.Organization
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&organizations).Error; err != nil {
				return err
			}
			for _, organization := range organizations {
				nodes.Nodes = append(nodes.Nodes, convertOrganizationFromSQLModel(&organization))
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

// Organization is the resolver for the Organization field.
func (r *queryResolver) Organization(ctx context.Context, id uuid.UUID) (*model.Organization, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) && !sess.Admin {
		db = db.Where("`id` IN (SELECT `organization_id` FROM `role_binding` WHERE `user_binding_id` = ?)", sess.UserID)
	}

	organization := &model1.Organization{}
	if err := db.Where("`id` = ?", id).Take(organization).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", id.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertOrganizationFromSQLModel(organization), nil
}

// Organizations is the resolver for the Organizations field.
func (r *queryResolver) Organizations(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.Organizations, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.Organizations{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.Organization{})
	if !IsGuardSilenced(ctx) && !sess.Admin {
		db = db.Where("`id` IN (SELECT `organization_id` FROM `role_binding` WHERE `user_binding_id` = ?)", sess.UserID)
	}
	db = finalizeOrganizationQueryFilter(db, obj.Filters...)
	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// Organization returns graphql1.OrganizationResolver implementation.
func (r *Resolver) Organization() graphql1.OrganizationResolver { return &organizationResolver{r} }

// OrganizationQuota returns graphql1.OrganizationQuotaResolver implementation.
func (r *Resolver) OrganizationQuota() graphql1.OrganizationQuotaResolver {
	return &organizationQuotaResolver{r}
}

// Organizations returns graphql1.OrganizationsResolver implementation.
func (r *Resolver) Organizations() graphql1.OrganizationsResolver { return &organizationsResolver{r} }

type organizationResolver struct{ *Resolver }
type organizationQuotaResolver struct{ *Resolver }
type organizationsResolver struct{ *Resolver }
