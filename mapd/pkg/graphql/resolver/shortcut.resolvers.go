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

// CreateShortcut is the resolver for the CreateShortcut field.
func (r *mutationResolver) CreateShortcut(ctx context.Context, input model.CreateShortcut) (*model.Shortcut, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if input.Name == "" {
		return nil, errInvalidRequest(ctx, "input.name", "value cannot be empty", input.Name)
	}

	shortcut := &model1.Shortcut{
		ID:             uuid.New(),
		Name:           input.Name,
		Endpoint:       input.Endpoint.String(),
		OrganizationID: model1.NewNullUUID(input.Organization),
	}
	if input.Description != nil {
		shortcut.Description = *input.Description
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectShortcut, shortcut.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.Organization{}).Where("`id` = ?", shortcut.OrganizationID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists < 1 {
			return errObjectNotFound(ctx, "input.organization", shortcut.OrganizationID.String())
		}

		if err := tx.Model(&model1.Shortcut{}).Where("`organization_id` = ? AND `name` = ?", shortcut.OrganizationID, shortcut.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", shortcut.Name)
		}

		if err := appendAuditLog(tx, "create", "Shortcut", shortcut.ID, shortcut.Name, sess.UserID, sess.UserName, shortcut.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Create(shortcut).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(shortcut).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertShortcutFromSQLModel(shortcut), nil
}

// UpdateShortcut is the resolver for the UpdateShortcut field.
func (r *mutationResolver) UpdateShortcut(ctx context.Context, input model.UpdateShortcut) (*model.Shortcut, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	shortcut := &model1.Shortcut{ID: input.ID}
	updates := make(map[string]interface{})
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Endpoint != nil {
		updates["endpoint"] = input.Endpoint.String()
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(shortcut).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", shortcut.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectShortcut, shortcut.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if err := appendAuditLog(tx, "update", "Shortcut", shortcut.ID, shortcut.Name, sess.UserID, sess.UserName, shortcut.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(updates) > 0 {
			if err := tx.Model(shortcut).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(shortcut).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertShortcutFromSQLModel(shortcut), nil
}

// DeleteShortcut is the resolver for the DeleteShortcut field.
func (r *mutationResolver) DeleteShortcut(ctx context.Context, input model.DeleteShortcut) (*model.Shortcut, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	shortcut := &model1.Shortcut{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(shortcut).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", shortcut.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermModify, model1.PolicyObjectShortcut, shortcut.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if err := appendAuditLog(tx, "delete", "Shortcut", shortcut.ID, shortcut.Name, sess.UserID, sess.UserName, shortcut.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if op := tx.Delete(shortcut); op.Error != nil {
			return &ErrDBConnection{op.Error}
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertShortcutFromSQLModel(shortcut), nil
}

// Shortcut is the resolver for the Shortcut field.
func (r *queryResolver) Shortcut(ctx context.Context, id uuid.UUID) (*model.Shortcut, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectShortcut)
	}

	shortcut := &model1.Shortcut{}
	if err := db.Where("`id` = ?", id).Take(shortcut).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", id.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertShortcutFromSQLModel(shortcut), nil
}

// Shortcuts is the resolver for the Shortcuts field.
func (r *queryResolver) Shortcuts(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.Shortcuts, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.Shortcuts{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.Shortcut{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectShortcut)
	}
	db = finalizeShortcutQueryFilter(db, obj.Filters...)

	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// Organization is the resolver for the organization field.
func (r *shortcutResolver) Organization(ctx context.Context, obj *model.Shortcut) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *shortcutsResolver) Edge(ctx context.Context, obj *model.Shortcuts) (*model.ShortcutNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.ShortcutNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectShortcut)
		}
		db = finalizeShortcutQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var shortcuts []model1.Shortcut
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&shortcuts).Error; err != nil {
				return err
			}
			for _, shortcut := range shortcuts {
				nodes.Nodes = append(nodes.Nodes, convertShortcutFromSQLModel(&shortcut))
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

// Shortcut returns graphql1.ShortcutResolver implementation.
func (r *Resolver) Shortcut() graphql1.ShortcutResolver { return &shortcutResolver{r} }

// Shortcuts returns graphql1.ShortcutsResolver implementation.
func (r *Resolver) Shortcuts() graphql1.ShortcutsResolver { return &shortcutsResolver{r} }

type shortcutResolver struct{ *Resolver }
type shortcutsResolver struct{ *Resolver }
