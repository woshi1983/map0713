package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"runtime"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/multierr"
	"gorm.io/gorm"
	"k8s.io/client-go/rest"
)

// Instance is the resolver for the instance field.
func (r *jobResolver) Instance(ctx context.Context, obj *model.Job) (*model.Instance, error) {
	if obj.Instance == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Instance(WithGuardSilenced(ctx), obj.Instance.ID)
}

// Component is the resolver for the component field.
func (r *jobResolver) Component(ctx context.Context, obj *model.Job) (*model.Component, error) {
	if obj.Component == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Component(WithGuardSilenced(ctx), obj.Component.ID)
}

// Application is the resolver for the application field.
func (r *jobResolver) Application(ctx context.Context, obj *model.Job) (*model.Application, error) {
	if obj.Application == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Application(WithGuardSilenced(ctx), obj.Application.ID)
}

// Cluster is the resolver for the cluster field.
func (r *jobResolver) Cluster(ctx context.Context, obj *model.Job) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// Job is the resolver for the Job field.
func (r *queryResolver) Job(ctx context.Context, cluster uuid.UUID, namespace string, name string) (*model.Job, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if name == "" {
		return nil, errInvalidRequest(ctx, "name", "value cannot be empty")
	}
	if namespace == "" {
		return nil, errInvalidRequest(ctx, "namespace", "value cannot be empty")
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterResource)
	}

	c := &model1.Cluster{ID: cluster}
	if err := db.Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "cluster", c.ID.String())
		}
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

	obj, err := wc.GetJob(namespace, name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "name", name)
		}
		return nil, errGatewayError(ctx, err)
	}
	job := &model.Job{Cluster: &model.Cluster{ID: cluster}}
	convertJob(job, obj)
	return job, nil
}

// Jobs is the resolver for the Jobs field.
func (r *queryResolver) Jobs(ctx context.Context, filters []*model.QueryFilter) ([]*model.Job, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		clusters  []model1.Cluster
		jobs      []*model.Job
		selectors []k8s.Selector
	)

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterResource)
	}
	for _, each := range filters {
		switch each.Key {
		case "name":
			if len(each.Value) > 0 {
				selectors = append(selectors, k8s.FieldEqual("metadata.name", each.Value[0]))
			}
		case "namespace":
			if len(each.Value) > 0 {
				selectors = append(selectors, k8s.FieldEqual("metadata.namespace", each.Value[0]))
			}
		case "binaryimage":
			selectors = append(selectors, k8s.LabelEqual(defLabelController, "binary_build"))
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists(defLabelImage))
			case 1:
				selectors = append(selectors, k8s.LabelEqual(defLabelImage, each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn(defLabelImage, each.Value...))
			}
		case "binarybuild":
			selectors = append(selectors, k8s.LabelEqual(defLabelController, "binary_build"))
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists(defLabelIssuer))
			case 1:
				selectors = append(selectors, k8s.LabelEqual(defLabelIssuer, each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn(defLabelIssuer, each.Value...))
			}
		case "dockerfileimage":
			selectors = append(selectors, k8s.LabelEqual(defLabelController, "dockerfile_build"))
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists(defLabelImage))
			case 1:
				selectors = append(selectors, k8s.LabelEqual(defLabelImage, each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn(defLabelImage, each.Value...))
			}
		case "dockerfilebuild":
			selectors = append(selectors, k8s.LabelEqual(defLabelController, "dockerfile_build"))
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists(defLabelIssuer))
			case 1:
				selectors = append(selectors, k8s.LabelEqual(defLabelIssuer, each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn(defLabelIssuer, each.Value...))
			}
		case "instance":
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists(defLabelInstance))
			case 1:
				selectors = append(selectors, k8s.LabelEqual(defLabelInstance, each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn(defLabelInstance, each.Value...))
			}
		case "component":
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists(defLabelComponent))
			case 1:
				selectors = append(selectors, k8s.LabelEqual(defLabelComponent, each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn(defLabelComponent, each.Value...))
			}
		case "application":
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists(defLabelApplication))
			case 1:
				selectors = append(selectors, k8s.LabelEqual(defLabelApplication, each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn(defLabelApplication, each.Value...))
			}
		case "cluster":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "organization":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`organization_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		}
	}
	if err := db.Where("`provisioned` >= ?", model1.ClusterProvisionFinish).Find(&clusters).Error; err != nil {
		panic(&ErrDBConnection{err})
	}

	handle := func(outChan chan<- []*model.Job, gqlErrChan chan<- *gqlerror.Error, errChan chan<- error, inChan <-chan *model1.Cluster) {
		defer func() {
			if r := recover(); r != nil {
				if e, ok := r.(error); ok {
					errChan <- e
				}
			}
		}()

		for c := range inChan {
			if c == nil {
				continue
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

			objs, err := getJobs(wc, "", selectors...)
			if err != nil {
				gqlErrChan <- errGatewayError(ctx, err)
				continue
			}

			var jobs []*model.Job
			for i := range objs {
				job := &model.Job{Cluster: &model.Cluster{ID: c.ID}}
				convertJob(job, &objs[i])
				jobs = append(jobs, job)
			}
			outChan <- jobs
		}
	}

	inChan := make(chan *model1.Cluster, len(clusters))
	for i := range clusters {
		inChan <- &clusters[i]
	}
	close(inChan)
	outChan := make(chan []*model.Job, 1)
	defer close(outChan)
	gqlErrChan := make(chan *gqlerror.Error, 1)
	defer close(gqlErrChan)
	errChan := make(chan error, 1)
	defer close(errChan)

	for i := 0; i < runtime.NumCPU(); i++ {
		go handle(outChan, gqlErrChan, errChan, inChan)
	}

	var err error
	for i := 0; i < len(clusters); i++ {
		select {
		case out := <-outChan:
			jobs = append(jobs, out...)
		case e := <-gqlErrChan:
			graphql.AddError(ctx, e)
		case e := <-errChan:
			if e != nil {
				err = multierr.Append(err, e)
			}
		}
	}
	if err != nil {
		panic(err)
	}
	return jobs, nil
}

// Job returns graphql1.JobResolver implementation.
func (r *Resolver) Job() graphql1.JobResolver { return &jobResolver{r} }

type jobResolver struct{ *Resolver }
