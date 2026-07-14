package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"

	api "github.com/dexidp/dex/api/v2"
	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/randutil"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateUser is the resolver for the CreateUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUser) (*model.CreatedUser, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}

	obj := &model.CreatedUser{
		ID:          uuid.New(),
		Name:        input.Name,
		Email:       input.Name + "@tanzu.local",
		DisplayName: input.DisplayName,
		Managed:     true,
	}
	if obj.DisplayName == "" {
		obj.DisplayName = obj.Name
	}
	if input.Admin != nil && *input.Admin {
		obj.Admin = true
	}
	if input.Password == nil || len(*input.Password) < 6 {
		obj.Password = randutil.StringN(8)
	} else {
		obj.Password = *input.Password
	}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if !IsGuardSilenced(ctx) && !sess.Admin {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.UserBinding{}).Where("`name` = ?", input.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", input.Name)
		}
		if err := tx.Model(&model1.UserBinding{}).Where("`email` = ?", obj.Email).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", input.Name)
		}

		user := &model1.UserBinding{
			ID:       obj.ID,
			Name:     obj.Name,
			FullName: obj.DisplayName,
			Email:    obj.Email,
			Admin:    obj.Admin,
			Managed:  true,
		}
		if err := tx.Create(user).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(user).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(obj.Password), bcrypt.DefaultCost)
		if err != nil {
			panic(fmt.Errorf("cannot generate password hash: %v", err))
		}
		pwd := &api.Password{
			Email:    user.Email,
			Hash:     hash,
			Username: user.Name,
			UserId:   user.ID.String(),
		}
		resp, err := r.DexClient.CreatePassword(ctx, &api.CreatePasswordReq{Password: pwd})
		if err != nil {
			if !resp.AlreadyExists {
				return errGatewayError(ctx, err)
			}

			if _, err := r.DexClient.UpdatePassword(ctx, &api.UpdatePasswordReq{
				Email:       pwd.Email,
				NewUsername: pwd.Username,
				NewHash:     hash,
			}); err != nil {
				return errGatewayError(ctx, err)
			}
		}

		if err := appendAuditLog(tx, "create", "User", user.ID, user.Name, sess.UserID, sess.UserName, uuid.Nil); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj.CreatedAt = user.CreatedAt
		obj.LastModified = user.UpdatedAt
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return obj, nil
}

// UpdateUser is the resolver for the UpdateUser field.
func (r *mutationResolver) UpdateUser(ctx context.Context, input model.UpdateUser) (*model.User, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	user := &model1.UserBinding{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if sess.UserID.String() != input.ID.String() && !IsGuardSilenced(ctx) && !sess.Admin {
			return errInsufficientPrivileges(ctx)
		}

		if err := tx.Take(user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", user.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		var exists int64
		updates := make(map[string]interface{})
		if input.Name != nil && *input.Name != "" {
			if user.Name == "admin" && user.Managed {
				// admin user cannot be updated, unless with environment variables.
				return errInsufficientPrivileges(ctx)
			}

			if !validateName(*input.Name) {
				return errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), *input.Name)
			}

			if err := tx.Model(&model1.UserBinding{}).Where("`id` != ? AND `name` = ?", user.ID, *input.Name).Count(&exists).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if exists > 0 {
				return errObjectDuplicated(ctx, "input.name", *input.Name)
			}
			updates["name"] = *input.Name
		}
		if input.DisplayName != nil && *input.DisplayName != "" {
			if user.Name == "admin" && user.Managed {
				// admin user cannot be updated, unless with environment variables.
				return errInsufficientPrivileges(ctx)
			}

			updates["full_name"] = *input.DisplayName
		}
		if input.Admin != nil {
			if user.Name == "admin" && user.Managed {
				// admin user cannot be updated, unless with environment variables.
				return errInsufficientPrivileges(ctx)
			}

			updates["admin"] = *input.Admin
		}

		if err := appendAuditLog(tx, "update", "User", user.ID, user.Name, sess.UserID, sess.UserName, uuid.Nil); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(updates) > 0 {
			if err := tx.Model(user).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(user).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if input.Password != nil && *input.Password != "" {
			if sess.UserID.String() == input.ID.String() {
				if input.OldPassword == nil {
					return errInvalidRequest(ctx, "input.oldPassword", "value cannot be empty")
				}

				if resp, err := r.DexClient.VerifyPassword(ctx, &api.VerifyPasswordReq{
					Email:    user.Email,
					Password: *input.OldPassword,
				}); err != nil || !resp.Verified || resp.NotFound {
					return errInsufficientPrivileges(ctx)
				}
			}

			if len(*input.Password) < 6 {
				return errInvalidRequest(ctx, "input.password", "length must be greater than 5")
			}
			hash, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
			if err != nil {
				panic(fmt.Errorf("cannot generate password hash: %v", err))
			}

			if _, err := r.DexClient.UpdatePassword(ctx, &api.UpdatePasswordReq{
				Email:       user.Email,
				NewUsername: user.Name,
				NewHash:     hash,
			}); err != nil {
				return errGatewayError(ctx, err)
			}
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertUserFromSQLModel(user), nil
}

// DeleteUser is the resolver for the DeleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context, input model.DeleteUser) (*model.User, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	user := &model1.UserBinding{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if !IsGuardSilenced(ctx) && !sess.Admin {
			return errInsufficientPrivileges(ctx)
		}

		if err := tx.Take(user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", user.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if user.Name == "admin" {
			return errInsufficientPrivileges(ctx)
		}

		if err := tx.Delete(user).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		resp, err := r.DexClient.DeletePassword(ctx, &api.DeletePasswordReq{
			Email: user.Email,
		})
		if err != nil && !resp.NotFound {
			return errGatewayError(ctx, err)
		}

		if err := appendAuditLog(tx, "delete", "User", user.ID, user.Name, sess.UserID, sess.UserName, uuid.Nil); err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertUserFromSQLModel(user), nil
}

// Whoami is the resolver for the Whoami field.
func (r *queryResolver) Whoami(ctx context.Context) (*model.User, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}
	return r.Resolver.Query().(*queryResolver).User(WithGuardSilenced(ctx), sess.UserID)
}

// User is the resolver for the User field.
func (r *queryResolver) User(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user := &model1.UserBinding{ID: id}
	if err := mapd.Database().Take(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", id.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertUserFromSQLModel(user), nil
}

// Users is the resolver for the Users field.
func (r *queryResolver) Users(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.Users, error) {
	obj := &model.Users{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.UserBinding{})
	db = finalizeUserQueryFilter(db, obj.Filters...)

	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// RoleBindings is the resolver for the roleBindings field.
func (r *userResolver) RoleBindings(ctx context.Context, obj *model.User, page *model.QueryPage, filters []*model.QueryFilter) (*model.RoleBindings, error) {
	return r.Resolver.Query().(*queryResolver).RoleBindings(WithGuardSilenced(ctx), page, append(filters, &model.QueryFilter{Key: "user", Value: []string{obj.ID.String()}}))
}

// Edge is the resolver for the edge field.
func (r *usersResolver) Edge(ctx context.Context, obj *model.Users) (*model.UserNodes, error) {
	nodes := &model.UserNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		db = finalizeUserQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var users []model1.UserBinding
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&users).Error; err != nil {
				return err
			}
			for _, user := range users {
				nodes.Nodes = append(nodes.Nodes, convertUserFromSQLModel(&user))
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

// User returns graphql1.UserResolver implementation.
func (r *Resolver) User() graphql1.UserResolver { return &userResolver{r} }

// Users returns graphql1.UsersResolver implementation.
func (r *Resolver) Users() graphql1.UsersResolver { return &usersResolver{r} }

type userResolver struct{ *Resolver }
type usersResolver struct{ *Resolver }
