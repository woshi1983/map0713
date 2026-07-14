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

// Release is the resolver for the Release field.
func (r *queryResolver) Release(ctx context.Context, id uuid.UUID) (*model.Release, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectComponent)
	}

	release := &model1.Release{ID: id}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(release).Error; err != nil {
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
	return convertReleaseFromSQLModel(release), nil
}

// Releases is the resolver for the Releases field.
func (r *queryResolver) Releases(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.Releases, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.Releases{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.Release{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectComponent)
	}
	db = finalizeReleaseQueryFilter(db, obj.Filters...)

	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// Component is the resolver for the component field.
func (r *releaseResolver) Component(ctx context.Context, obj *model.Release) (*model.Component, error) {
	if obj.Component == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Component(WithGuardSilenced(ctx), obj.Component.ID)
}

// Application is the resolver for the application field.
func (r *releaseResolver) Application(ctx context.Context, obj *model.Release) (*model.Application, error) {
	if obj.Application == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Application(WithGuardSilenced(ctx), obj.Application.ID)
}

// Organization is the resolver for the organization field.
func (r *releaseResolver) Organization(ctx context.Context, obj *model.Release) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *releasesResolver) Edge(ctx context.Context, obj *model.Releases) (*model.ReleaseNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.ReleaseNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectComponent)
		}
		db = finalizeReleaseQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var releases []model1.Release
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&releases).Error; err != nil {
				return err
			}
			for _, release := range releases {
				nodes.Nodes = append(nodes.Nodes, convertReleaseFromSQLModel(&release))
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

// Release returns graphql1.ReleaseResolver implementation.
func (r *Resolver) Release() graphql1.ReleaseResolver { return &releaseResolver{r} }

// Releases returns graphql1.ReleasesResolver implementation.
func (r *Resolver) Releases() graphql1.ReleasesResolver { return &releasesResolver{r} }

type releaseResolver struct{ *Resolver }
type releasesResolver struct{ *Resolver }
