package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gorm.io/gorm"
)

// CreateRoleBinding is the resolver for the CreateRoleBinding field.
func (r *mutationResolver) CreateRoleBinding(ctx context.Context, input model.CreateRoleBinding) (*model.RoleBinding, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectOrganization, input.Organization) {
		return nil, errInsufficientPrivileges(ctx)
	}

	rb := &model1.RoleBinding{ID: uuid.New(), UserBindingID: model1.NewNullUUID(input.User), Role: model1.Role(input.Role.String()), OrganizationID: model1.NewNullUUID(input.Organization)}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		var exists int64
		if err := tx.Model(&model1.Organization{}).Where("`id` = ?", rb.OrganizationID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists < 1 {
			return errObjectNotFound(ctx, "input.organization", rb.OrganizationID.String())
		}
		if err := tx.Model(&model1.RoleBinding{}).Where("`user_binding_id` = ? AND `organization_id` = ?", rb.UserBindingID, rb.OrganizationID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.user", rb.UserBindingID.String())
		}

		user := &model1.UserBinding{ID: rb.UserBindingID.UUID}
		if err := tx.Take(user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.user", user.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if err := appendAuditLog(tx, "create", "RoleBinding", rb.ID, user.Name, sess.UserID, sess.UserName, rb.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
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
	return convertRoleBindingFromSQLModel(rb), nil
}

// UpdateRoleBinding is the resolver for the UpdateRoleBinding field.
func (r *mutationResolver) UpdateRoleBinding(ctx context.Context, input model.UpdateRoleBinding) (*model.RoleBinding, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	rb := &model1.RoleBinding{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(rb).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", input.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectOrganization, rb.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		user := &model1.UserBinding{ID: rb.UserBindingID.UUID}
		if err := tx.Take(user).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := appendAuditLog(tx, "update", "RoleBinding", rb.ID, user.Name, sess.UserID, sess.UserName, rb.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Model(&model1.RoleBinding{ID: rb.ID}).Updates(map[string]interface{}{"role": input.Role.String()}).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(rb).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertRoleBindingFromSQLModel(rb), nil
}

// DeleteRoleBinding is the resolver for the DeleteRoleBinding field.
func (r *mutationResolver) DeleteRoleBinding(ctx context.Context, input model.DeleteRoleBinding) (*model.RoleBinding, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	rb := &model1.RoleBinding{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(rb).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", input.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectOrganization, rb.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		user := &model1.UserBinding{ID: rb.UserBindingID.UUID}
		if err := tx.Take(user).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := appendAuditLog(tx, "delete", "RoleBinding", rb.ID, user.Name, sess.UserID, sess.UserName, rb.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Delete(rb).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertRoleBindingFromSQLModel(rb), nil
}

// RoleBinding is the resolver for the RoleBinding field.
func (r *queryResolver) RoleBinding(ctx context.Context, id uuid.UUID) (*model.RoleBinding, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	rb := &model1.RoleBinding{ID: id}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if !IsGuardSilenced(ctx) {
			tx = sess.FilterQueryByOrganization(tx, model1.PermReadOnly, model1.PolicyObjectOrganization)
		}
		if err := tx.Take(rb).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "id", id.String())
			}
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertRoleBindingFromSQLModel(rb), nil
}

// RoleBindings is the resolver for the RoleBindings field.
func (r *queryResolver) RoleBindings(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.RoleBindings, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj1 := &model.RoleBindings{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj1.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj1.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj1.Filters = append(obj1.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.RoleBinding{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectOrganization)
	}
	db = finalizeRoleBindingQueryFilter(db, obj1.Filters...)
	if err := db.Count(&obj1.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj1.Page, obj1.TotalCount)
	return obj1, nil
}

// User is the resolver for the user field.
func (r *roleBindingResolver) User(ctx context.Context, obj *model.RoleBinding) (*model.User, error) {
	if obj.User == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).User(WithGuardSilenced(ctx), obj.User.ID)
}

// Organization is the resolver for the organization field.
func (r *roleBindingResolver) Organization(ctx context.Context, obj *model.RoleBinding) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *roleBindingsResolver) Edge(ctx context.Context, obj *model.RoleBindings) (*model.RoleBindingNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.RoleBindingNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`user_binding_id` ASC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectOrganization)
		}
		db = finalizeRoleBindingQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var rbs []model1.RoleBinding
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&rbs).Error; err != nil {
				return err
			}
			for _, rb := range rbs {
				nodes.Nodes = append(nodes.Nodes, convertRoleBindingFromSQLModel(&rb))
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

// RoleBinding returns graphql1.RoleBindingResolver implementation.
func (r *Resolver) RoleBinding() graphql1.RoleBindingResolver { return &roleBindingResolver{r} }

// RoleBindings returns graphql1.RoleBindingsResolver implementation.
func (r *Resolver) RoleBindings() graphql1.RoleBindingsResolver { return &roleBindingsResolver{r} }

type roleBindingResolver struct{ *Resolver }
type roleBindingsResolver struct{ *Resolver }
