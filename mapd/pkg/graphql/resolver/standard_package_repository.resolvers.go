package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	model2 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gorm.io/gorm"
)

// CreateStandardPackageRepository is the resolver for the CreateStandardPackageRepository field.
func (r *mutationResolver) CreateStandardPackageRepository(ctx context.Context, input model2.CreateStandardPackageRepository) (*model2.StandardPackageRepository, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}
	if !canStandardPackageRepository(sess, model1.PermReadWrite) {
		return nil, errInsufficientPrivileges(ctx)
	}
	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}
	if err := validateStandardPackageRepositoryImage(input.Image); err != nil {
		return nil, errInvalidRequest(ctx, "input.image", err.Error(), input.Image)
	}

	repo := &model1.StandardPackageRepository{
		ID:      uuid.New(),
		Name:    input.Name,
		Image:   input.Image,
		Default: input.Default != nil && *input.Default,
	}
	if input.Description != nil {
		repo.Description = *input.Description
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		var exists int64
		if err := tx.Model(&model1.StandardPackageRepository{}).Where("`name` = ?", repo.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", repo.Name)
		}
		if err := tx.Model(&model1.StandardPackageRepository{}).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists == 0 {
			repo.Default = true
		}
		if err := tx.Create(repo).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if repo.Default {
			if err := resetDefaultStandardPackageRepository(tx, repo); err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(repo).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if input.Dryrun != nil && *input.Dryrun {
			return errDryrun
		}
		if err := appendAuditLog(tx, "create", "StandardPackageRepository", repo.ID, repo.Name, sess.UserID, sess.UserName, uuid.Nil); err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertStandardPackageRepositoryFromSQLModel(repo), nil
}

// UpdateStandardPackageRepository is the resolver for the UpdateStandardPackageRepository field.
func (r *mutationResolver) UpdateStandardPackageRepository(ctx context.Context, input model2.UpdateStandardPackageRepository) (*model2.StandardPackageRepository, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}
	if !canStandardPackageRepository(sess, model1.PermReadWrite) {
		return nil, errInsufficientPrivileges(ctx)
	}

	repo := &model1.StandardPackageRepository{ID: input.ID}
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
	if input.Image != nil {
		if err := validateStandardPackageRepositoryImage(*input.Image); err != nil {
			return nil, errInvalidRequest(ctx, "input.image", err.Error(), *input.Image)
		}
		updates["image"] = *input.Image
	}
	if input.Default != nil {
		updates["default"] = *input.Default
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(repo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", input.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		if input.Name != nil {
			var exists int64
			if err := tx.Model(&model1.StandardPackageRepository{}).Where("`id` != ? AND `name` = ?", repo.ID, *input.Name).Count(&exists).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if exists > 0 {
				return errObjectDuplicated(ctx, "input.name", *input.Name)
			}
		}
		if len(updates) > 0 {
			if err := tx.Model(repo).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if input.Default != nil && *input.Default {
			if err := resetDefaultStandardPackageRepository(tx, repo); err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(repo).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if input.Dryrun != nil && *input.Dryrun {
			return errDryrun
		}
		if err := appendAuditLog(tx, "update", "StandardPackageRepository", repo.ID, repo.Name, sess.UserID, sess.UserName, uuid.Nil); err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertStandardPackageRepositoryFromSQLModel(repo), nil
}

// SetDefaultStandardPackageRepository is the resolver for the SetDefaultStandardPackageRepository field.
func (r *mutationResolver) SetDefaultStandardPackageRepository(ctx context.Context, input model2.SetDefaultStandardPackageRepository) (*model2.StandardPackageRepository, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}
	if !canStandardPackageRepository(sess, model1.PermReadWrite) {
		return nil, errInsufficientPrivileges(ctx)
	}

	repo := &model1.StandardPackageRepository{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(repo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", input.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		if err := resetDefaultStandardPackageRepository(tx, repo); err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(repo).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := appendAuditLog(tx, "set_default", "StandardPackageRepository", repo.ID, repo.Name, sess.UserID, sess.UserName, uuid.Nil); err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertStandardPackageRepositoryFromSQLModel(repo), nil
}

// DeleteStandardPackageRepository is the resolver for the DeleteStandardPackageRepository field.
func (r *mutationResolver) DeleteStandardPackageRepository(ctx context.Context, input model2.DeleteStandardPackageRepository) (*model2.StandardPackageRepository, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}
	if !canStandardPackageRepository(sess, model1.PermModify) {
		return nil, errInsufficientPrivileges(ctx)
	}

	repo := &model1.StandardPackageRepository{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(repo).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", input.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		wasDefault := repo.Default
		if err := appendAuditLog(tx, "delete", "StandardPackageRepository", repo.ID, repo.Name, sess.UserID, sess.UserName, uuid.Nil); err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Delete(repo).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if wasDefault {
			nextDefault := &model1.StandardPackageRepository{}
			if err := tx.Order("updated_at DESC").Take(nextDefault).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				panic(&ErrDBConnection{err})
			}
			if err := resetDefaultStandardPackageRepository(tx, nextDefault); err != nil {
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
	return convertStandardPackageRepositoryFromSQLModel(repo), nil
}

// StandardPackageRepository is the resolver for the StandardPackageRepository field.
func (r *queryResolver) StandardPackageRepository(ctx context.Context, id uuid.UUID) (*model2.StandardPackageRepository, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}
	if !canStandardPackageRepository(sess, model1.PermReadOnly) {
		return nil, errInsufficientPrivileges(ctx)
	}

	repo := &model1.StandardPackageRepository{ID: id}
	if err := mapd.Database().Take(repo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", id.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertStandardPackageRepositoryFromSQLModel(repo), nil
}

// StandardPackageRepositories is the resolver for the StandardPackageRepositories field.
func (r *queryResolver) StandardPackageRepositories(ctx context.Context, page *model2.QueryPage, filters []*model2.QueryFilter) (*model2.StandardPackageRepositories, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}
	if !canStandardPackageRepository(sess, model1.PermReadOnly) {
		return nil, errInsufficientPrivileges(ctx)
	}

	obj := &model2.StandardPackageRepositories{Page: &model2.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model2.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.StandardPackageRepository{})
	db = finalizeStandardPackageRepositoryQueryFilter(db, obj.Filters...)
	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// Edge is the resolver for the edge field.
func (r *standardPackageRepositoriesResolver) Edge(ctx context.Context, obj *model2.StandardPackageRepositories) (*model2.StandardPackageRepositoryNodes, error) {
	if obj.Edge != nil {
		return obj.Edge, nil
	}

	nodes := &model2.StandardPackageRepositoryNodes{}
	db := mapd.Database().Model(&model1.StandardPackageRepository{})
	db = finalizeStandardPackageRepositoryQueryFilter(db, obj.Filters...)
	if obj.Page.Limit > 0 {
		var repos []model1.StandardPackageRepository
		if err := db.Limit(int(obj.Page.Limit)).Offset(int(obj.Page.Offset)).Order("`default` DESC, `updated_at` DESC").Find(&repos).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		for _, repo := range repos {
			nodes.Nodes = append(nodes.Nodes, convertStandardPackageRepositoryFromSQLModel(&repo))
		}
	}
	obj.Edge = nodes
	return nodes, nil
}

// StandardPackageRepositories returns graphql1.StandardPackageRepositoriesResolver implementation.
func (r *Resolver) StandardPackageRepositories() graphql1.StandardPackageRepositoriesResolver {
	return &standardPackageRepositoriesResolver{r}
}

type standardPackageRepositoriesResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func validateStandardPackageRepositoryImage(image string) error {
	if strings.TrimSpace(image) != image || image == "" {
		return fmt.Errorf("value cannot be empty")
	}
	parsed, err := ParseImage(image)
	if err != nil {
		return err
	}
	if !parsed.IsQualified() {
		return fmt.Errorf("image must include registry, project and repository")
	}
	if parsed.Tag == "" {
		return fmt.Errorf("image tag is required")
	}
	if strings.EqualFold(parsed.Tag, "latest") {
		return fmt.Errorf("latest tag is not allowed")
	}
	return nil
}
func resetDefaultStandardPackageRepository(tx *gorm.DB, repo *model1.StandardPackageRepository) error {
	if err := tx.Exec("UPDATE `standard_package_repository` SET `default` = ? WHERE `id` != ? AND `deleted_at` IS NULL", false, repo.ID).Error; err != nil {
		return err
	}
	return tx.Exec("UPDATE `standard_package_repository` SET `default` = ? WHERE `id` = ? AND `deleted_at` IS NULL", true, repo.ID).Error
}
