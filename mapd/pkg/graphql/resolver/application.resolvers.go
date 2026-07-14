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

// Organization is the resolver for the organization field.
func (r *applicationResolver) Organization(ctx context.Context, obj *model.Application) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *applicationsResolver) Edge(ctx context.Context, obj *model.Applications) (*model.ApplicationNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.ApplicationNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectApplication)
		}
		db = finalizeApplicationQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var apps []model1.Application
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&apps).Error; err != nil {
				return err
			}
			for _, app := range apps {
				nodes.Nodes = append(nodes.Nodes, convertApplicationFromSQLModel(&app))
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

// CreateApplication is the resolver for the CreateApplication field.
func (r *mutationResolver) CreateApplication(ctx context.Context, input model.CreateApplication) (*model.Application, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}

	app := &model1.Application{ID: uuid.New(), Name: input.Name, OrganizationID: model1.NewNullUUID(input.Organization)}
	if input.Description != nil {
		app.Description = *input.Description
	}
	if input.Avatar != nil {
		app.Avatar = input.Avatar.String()
	}
	if app.Avatar == "" {
		app.Avatar = identicon.NewDataURL(app.ID[:]).String()
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectApplication, app.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.Organization{}).Where("`id` = ?", app.OrganizationID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists < 1 {
			return errObjectNotFound(ctx, "input.organization", app.OrganizationID.String())
		}

		if err := tx.Model(&model1.Application{}).Where("`organization_id` = ? AND `name` = ?", app.OrganizationID, app.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", app.Name)
		}

		if err := appendAuditLog(tx, "create", "Application", app.ID, app.Name, sess.UserID, sess.UserName, app.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Create(app).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(app).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertApplicationFromSQLModel(app), nil
}

// UpdateApplication is the resolver for the UpdateApplication field.
func (r *mutationResolver) UpdateApplication(ctx context.Context, input model.UpdateApplication) (*model.Application, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	app := &model1.Application{ID: input.ID}
	updates := make(map[string]interface{})
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Avatar != nil {
		updates["avatar"] = input.Avatar.String()
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(app).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", app.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectApplication, app.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if err := appendAuditLog(tx, "update", "Application", app.ID, app.Name, sess.UserID, sess.UserName, app.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(updates) > 0 {
			if err := tx.Model(app).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(app).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertApplicationFromSQLModel(app), nil
}

// DeleteApplication is the resolver for the DeleteApplication field.
func (r *mutationResolver) DeleteApplication(ctx context.Context, input model.DeleteApplication) (*model.Application, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	app := &model1.Application{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(app).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", app.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermModify, model1.PolicyObjectApplication, app.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.Component{}).Where("`application_id` = ?", app.ID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectIsBusy(ctx, "input.id", "Component", nil, "", input.ID.String())
		}

		if err := appendAuditLog(tx, "delete", "Application", app.ID, app.Name, sess.UserID, sess.UserName, app.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if op := tx.Delete(app); op.Error != nil {
			return &ErrDBConnection{op.Error}
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertApplicationFromSQLModel(app), nil
}

// Application is the resolver for the Application field.
func (r *queryResolver) Application(ctx context.Context, id uuid.UUID) (*model.Application, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectApplication)
	}

	app := &model1.Application{}
	if err := db.Where("`id` = ?", id).Take(app).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", id.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertApplicationFromSQLModel(app), nil
}

// Applications is the resolver for the Applications field.
func (r *queryResolver) Applications(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.Applications, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.Applications{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.Application{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectApplication)
	}
	db = finalizeApplicationQueryFilter(db, obj.Filters...)

	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// Application returns graphql1.ApplicationResolver implementation.
func (r *Resolver) Application() graphql1.ApplicationResolver { return &applicationResolver{r} }

// Applications returns graphql1.ApplicationsResolver implementation.
func (r *Resolver) Applications() graphql1.ApplicationsResolver { return &applicationsResolver{r} }

type applicationResolver struct{ *Resolver }
type applicationsResolver struct{ *Resolver }
