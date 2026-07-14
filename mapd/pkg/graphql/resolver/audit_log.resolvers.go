package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gorm.io/gorm"
)

// User is the resolver for the user field.
func (r *auditLogResolver) User(ctx context.Context, obj *model.AuditLog) (*model.User, error) {
	if obj.User == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).User(WithGuardSilenced(ctx), obj.User.ID)
}

// Organization is the resolver for the organization field.
func (r *auditLogResolver) Organization(ctx context.Context, obj *model.AuditLog) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *auditLogsResolver) Edge(ctx context.Context, obj *model.AuditLogs) (*model.AuditLogNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.AuditLogNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectAuditLog)
		}
		db = finalizeAuditLogQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var als []model1.AuditLog
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&als).Error; err != nil {
				return err
			}
			for _, al := range als {
				nodes.Nodes = append(nodes.Nodes, convertAuditLogFromSQLModel(&al))
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

// AuditLogs is the resolver for the AuditLogs field.
func (r *queryResolver) AuditLogs(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.AuditLogs, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.AuditLogs{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.AuditLog{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectAuditLog)
	}
	db = finalizeAuditLogQueryFilter(db, obj.Filters...)

	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// AuditLog returns graphql1.AuditLogResolver implementation.
func (r *Resolver) AuditLog() graphql1.AuditLogResolver { return &auditLogResolver{r} }

// AuditLogs returns graphql1.AuditLogsResolver implementation.
func (r *Resolver) AuditLogs() graphql1.AuditLogsResolver { return &auditLogsResolver{r} }

type auditLogResolver struct{ *Resolver }
type auditLogsResolver struct{ *Resolver }
