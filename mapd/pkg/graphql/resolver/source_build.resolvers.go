package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/multierr"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// CreateSourceBuild is the resolver for the CreateSourceBuild field.
func (r *mutationResolver) CreateSourceBuild(ctx context.Context, input model.CreateSourceBuild) (*model.SourceBuild, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadWrite, model1.PolicyObjectClusterResource)
	}

	c := &model1.Cluster{ID: input.Cluster}
	if err := db.Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "input.cluster", c.ID.String())
		}
		panic(&ErrDBConnection{err})
	}

	if c.Provisioned < model1.ClusterProvisionFinish {
		return nil, errInvalidRequest(ctx, "input.cluster", "cluster is not ready", input.Cluster.String())
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
	kpack := k8s.NewKpack(wc)

	image, err := kpack.GetImage(defNamespace, input.Image)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "input.image", input.Image)
		}
		return nil, errGatewayError(ctx, err)
	}
	ready := false
	if image.Status.Conditions != nil {
		for _, each := range image.Status.Conditions {
			if each.Type == "Ready" && each.Status == corev1.ConditionTrue {
				ready = true
				break
			}
		}
	}
	if !ready {
		return nil, errInvalidRequest(ctx, "input.image", "cannot build unless the image is ready", input.Image)
	}
	builds, err := getBuilds(kpack, defNamespace, k8s.LabelEqual("image.kpack.io/image", input.Image), k8s.LabelEqual("image.kpack.io/buildNumber", strconv.FormatInt(image.Status.BuildCounter, 10)))
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	if len(builds) < 1 {
		return nil, errInvalidRequest(ctx, "input.image", "cannot find any existing build", input.Image)
	}
	obj := &builds[0]
	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}
	for k, v := range image.Annotations {
		obj.Annotations[k] = v
	}
	obj.Annotations["image.kpack.io/additionalBuildNeeded"] = time.Now().Format(abstract.ISO8601)

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "create", "SourceBuild", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = kpack.UpdateBuild(obj)
		if err != nil {
			return errGatewayError(ctx, err)
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}

	build := &model.SourceBuild{Cluster: &model.Cluster{ID: input.Cluster}}
	convertBuild(build, obj)
	return build, nil
}

// DeleteSourceBuild is the resolver for the DeleteSourceBuild field.
func (r *mutationResolver) DeleteSourceBuild(ctx context.Context, input model.DeleteSourceBuild) (*model.SourceBuild, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermModify, model1.PolicyObjectClusterResource)
	}

	c := &model1.Cluster{ID: input.Cluster}
	if err := db.Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "input.cluster", c.ID.String())
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
	kpack := k8s.NewKpack(wc)

	obj, err := kpack.GetBuild(defNamespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}
	build := &model.SourceBuild{Cluster: &model.Cluster{ID: input.Cluster}}
	convertBuild(build, obj)

	if obj.Labels != nil {
		name := obj.Labels["image.kpack.io/image"]
		if name != "" {
			image, err := kpack.GetImage(defNamespace, name)
			if err != nil && !k8s.IsErrObjectNotExists(err) {
				return nil, errGatewayError(ctx, err)
			}
			if build.Status != nil && build.Status.Num != nil && image.Status.BuildCounter == *build.Status.Num {
				return nil, errInvalidRequest(ctx, "input.name", "cannot delete the latest build", input.Name)
			}
		}
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "delete", "SourceBuild", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := kpack.DeleteBuild(defNamespace, input.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
			return errGatewayError(ctx, err)
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}

	return build, nil
}

// SourceBuild is the resolver for the SourceBuild field.
func (r *queryResolver) SourceBuild(ctx context.Context, cluster uuid.UUID, name string) (*model.SourceBuild, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if name == "" {
		return nil, errInvalidRequest(ctx, "name", "value cannot be empty")
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterResource)
	}

	c := &model1.Cluster{ID: cluster}
	if err := db.Take(c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "cluster", cluster.String())
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
	kpack := k8s.NewKpack(wc)

	obj, err := kpack.GetBuild(defNamespace, name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "name", name)
		}
		return nil, errGatewayError(ctx, err)
	}
	build := &model.SourceBuild{Cluster: &model.Cluster{ID: cluster}}
	convertBuild(build, obj)
	return build, nil
}

// SourceBuilds is the resolver for the SourceBuilds field.
func (r *queryResolver) SourceBuilds(ctx context.Context, filters []*model.QueryFilter) ([]*model.SourceBuild, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		clusters  []model1.Cluster
		builds    []*model.SourceBuild
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
		case "image":
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists("image.kpack.io/image"))
			case 1:
				selectors = append(selectors, k8s.LabelEqual("image.kpack.io/image", each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn("image.kpack.io/image", each.Value...))
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

	handle := func(outChan chan<- []*model.SourceBuild, gqlErrChan chan<- *gqlerror.Error, errChan chan<- error, inChan <-chan *model1.Cluster) {
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
			kpack := k8s.NewKpack(wc)

			objs, err := getBuilds(kpack, "", selectors...)
			if err != nil {
				gqlErrChan <- errGatewayError(ctx, err)
				continue
			}

			var builds []*model.SourceBuild
			for i := range objs {
				build := &model.SourceBuild{Cluster: &model.Cluster{ID: c.ID}}
				convertBuild(build, &objs[i])
				builds = append(builds, build)
			}
			outChan <- builds
		}
	}

	inChan := make(chan *model1.Cluster, len(clusters))
	for i := range clusters {
		inChan <- &clusters[i]
	}
	close(inChan)
	outChan := make(chan []*model.SourceBuild, 1)
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
			builds = append(builds, out...)
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
	return builds, nil
}

// Cluster is the resolver for the cluster field.
func (r *sourceBuildResolver) Cluster(ctx context.Context, obj *model.SourceBuild) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// SourceBuild returns graphql1.SourceBuildResolver implementation.
func (r *Resolver) SourceBuild() graphql1.SourceBuildResolver { return &sourceBuildResolver{r} }

type sourceBuildResolver struct{ *Resolver }
