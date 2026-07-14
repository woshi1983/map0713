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
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gorm.io/gorm"
	"k8s.io/client-go/rest"
)

// Health is the resolver for the health field.
func (r *instanceResolver) Health(ctx context.Context, obj *model.Instance) (*model.InstanceHealth, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	c := &model1.Cluster{ID: obj.Cluster.ID}
	if err := mapd.Database().Take(c).Error; err != nil {
		panic(&ErrDBConnection{err})
	}

	if c.Provisioned < model1.ClusterProvisionFinish {
		return nil, nil
	}

	block, err := cryptoutil.NewAESGCMCipher(c.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	wc, err := k8s.NewKubernetesFromToken(c.Endpoint, block.Decrypt(c.AuthToken), rest.TLSClientConfig{
		Insecure: c.InsecureTLS,
		CAData:   c.CACert,
	})
	if err != nil {
		panic(err)
	}
	workloads, err := getWorkloadHealthForInstance(wc, obj.Manifest)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	health := &model.InstanceHealth{Workloads: workloads}
	for _, each := range health.Workloads {
		if each.ReadyReplicas > 0 {
			health.Available++
		}
		if each.ReadyReplicas >= each.Replicas {
			health.Ready++
		}
	}
	return health, nil
}

// Services is the resolver for the services field.
func (r *instanceResolver) Services(ctx context.Context, obj *model.Instance, filters []*model.QueryFilter) ([]*model.Service, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Services(
		WithGuardSilenced(ctx),
		append(filters,
			&model.QueryFilter{Key: "cluster", Value: []string{obj.Cluster.ID.String()}},
			&model.QueryFilter{Key: "namespace", Value: []string{obj.Namespace}},
			&model.QueryFilter{Key: "instance", Value: []string{obj.ID.String()}},
		),
	)
}

// Cluster is the resolver for the cluster field.
func (r *instanceResolver) Cluster(ctx context.Context, obj *model.Instance) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// Release is the resolver for the release field.
func (r *instanceResolver) Release(ctx context.Context, obj *model.Instance) (*model.Release, error) {
	if obj.Release == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Release(WithGuardSilenced(ctx), obj.Release.ID)
}

// Federation is the resolver for the federation field.
func (r *instanceResolver) Federation(ctx context.Context, obj *model.Instance) (*model.InstanceFederation, error) {
	if obj.Federation == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).InstanceFederation(WithGuardSilenced(ctx), obj.Federation.ID)
}

// Component is the resolver for the component field.
func (r *instanceResolver) Component(ctx context.Context, obj *model.Instance) (*model.Component, error) {
	if obj.Component == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Component(WithGuardSilenced(ctx), obj.Component.ID)
}

// Application is the resolver for the application field.
func (r *instanceResolver) Application(ctx context.Context, obj *model.Instance) (*model.Application, error) {
	if obj.Application == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Application(WithGuardSilenced(ctx), obj.Application.ID)
}

// Organization is the resolver for the organization field.
func (r *instanceResolver) Organization(ctx context.Context, obj *model.Instance) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *instancesResolver) Edge(ctx context.Context, obj *model.Instances) (*model.InstanceNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.InstanceNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectInstance)
		}
		db = finalizeInstanceQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var instances []model1.Instance
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&instances).Error; err != nil {
				return err
			}
			for _, instance := range instances {
				nodes.Nodes = append(nodes.Nodes, convertInstanceFromSQLModel(&instance))
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

// Instance is the resolver for the Instance field.
func (r *queryResolver) Instance(ctx context.Context, id uuid.UUID) (*model.Instance, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectInstance)
	}

	instance := &model1.Instance{}
	if err := db.Where("`id` = ?", id).Take(instance).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", id.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertInstanceFromSQLModel(instance), nil
}

// Instances is the resolver for the Instances field.
func (r *queryResolver) Instances(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.Instances, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.Instances{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.Instance{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectInstance)
	}
	db = finalizeInstanceQueryFilter(db, obj.Filters...)
	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// Instance returns graphql1.InstanceResolver implementation.
func (r *Resolver) Instance() graphql1.InstanceResolver { return &instanceResolver{r} }

// Instances returns graphql1.InstancesResolver implementation.
func (r *Resolver) Instances() graphql1.InstancesResolver { return &instancesResolver{r} }

type instanceResolver struct{ *Resolver }
type instancesResolver struct{ *Resolver }
