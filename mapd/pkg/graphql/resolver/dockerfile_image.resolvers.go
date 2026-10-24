package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
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
	mapv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/mapd/v1alpha1"
	"go.uber.org/multierr"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// Cluster is the resolver for the cluster field.
func (r *dockerfileImageResolver) Cluster(ctx context.Context, obj *model.DockerfileImage) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// CreateDockerfileImage is the resolver for the CreateDockerfileImage field.
func (r *mutationResolver) CreateDockerfileImage(ctx context.Context, input model.CreateDockerfileImage) (*model.DockerfileImage, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}
	distImage, err := ParseImage(input.Spec.DistImage)
	if err != nil {
		return nil, errInvalidRequest(ctx, "input.spec.distImage", "invalid image", input.Spec.DistImage)
	}
	if distImage.Project == "" {
		distImage.Project = "library"
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
	cli := k8s.NewMap(wc)

	obj := &mapv1alpha1.DockerfileImage{
		ObjectMeta: metav1.ObjectMeta{
			Name:        input.Name,
			Namespace:   defNamespace,
			Annotations: input.Annotations,
			Labels: map[string]string{
				defLabelManaged: "true",
			},
		},
		Spec: mapv1alpha1.DockerfileImageSpec{
			Source:    *input.Spec.Source,
			DistImage: distImage.String(),
			Branch:    *input.Spec.Branch,
			SubDir:    *input.Spec.SubDir,
			FileName:  *input.Spec.FileName,
			URL:       input.Spec.URL.String(),
		},
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "create", "DockerfileImage", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = postDockerfileImage(cli, obj)
		if err != nil {
			if k8s.IsErrObjectAlreadyExists(err) {
				return errObjectDuplicated(ctx, "input.name", input.Name)
			}
			return errGatewayError(ctx, err)
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	image := &model.DockerfileImage{Cluster: &model.Cluster{ID: input.Cluster}}
	convertDockerfileImage(image, obj)
	return image, nil
}

// UpdateDockerfileImage is the resolver for the UpdateDockerfileImage field.
func (r *mutationResolver) UpdateDockerfileImage(ctx context.Context, input model.UpdateDockerfileImage) (*model.DockerfileImage, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if input.Name == "" {
		return nil, errInvalidRequest(ctx, "input.name", "value cannot be empty")
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
	cli := k8s.NewMap(wc)

	obj, err := getDockerfileImage(cli, defNamespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}

	if input.Annotations != nil {
		obj.Annotations = input.Annotations
	}
	if obj.Labels == nil {
		obj.Labels = make(map[string]string)
	}
	obj.Labels[defLabelManaged] = "true"
	if input.Spec != nil {
		if input.Spec.Source != nil {
			obj.Spec.Source = *input.Spec.Source
		}
		if input.Spec.DistImage != nil {
			distImage, err := ParseImage(*input.Spec.DistImage)
			if err != nil {
				return nil, errInvalidRequest(ctx, "input.spec.distImage", "invalid image", input.Spec.DistImage)
			}
			if distImage.Project == "" {
				distImage.Project = "library"
			}
			obj.Spec.DistImage = distImage.String()
		}
		if input.Spec.Branch != nil {
			obj.Spec.Branch = *input.Spec.Branch
		}
		if input.Spec.SubDir != nil {
			obj.Spec.SubDir = *input.Spec.SubDir
		}
		if input.Spec.FileName != nil {
			obj.Spec.FileName = *input.Spec.FileName
		}
		if input.Spec.URL != nil {
			obj.Spec.URL = input.Spec.URL.String()
		}
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "update", "DockerfileImage", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = putDockerfileImage(cli, obj)
		if err != nil {
			if k8s.IsErrObjectNotExists(err) {
				return errObjectNotFound(ctx, "input.name", input.Name)
			}
			return errGatewayError(ctx, err)
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	image := &model.DockerfileImage{Cluster: &model.Cluster{ID: input.Cluster}}
	convertDockerfileImage(image, obj)
	return image, nil
}

// DeleteDockerfileImage is the resolver for the DeleteDockerfileImage field.
func (r *mutationResolver) DeleteDockerfileImage(ctx context.Context, input model.DeleteDockerfileImage) (*model.DockerfileImage, error) {
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
	cli := k8s.NewMap(wc)

	obj, err := getDockerfileImage(cli, defNamespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}
	image := &model.DockerfileImage{Cluster: &model.Cluster{ID: input.Cluster}}
	convertDockerfileImage(image, obj)

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "delete", "DockerfileImage", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := wc.DeleteJobs(defNamespace, k8s.LabelEqual(defLabelController, "dockerfile_build"), k8s.LabelEqual(defLabelIssuer, image.Name)); err != nil && !k8s.IsErrObjectNotExists(err) {
			return errGatewayError(ctx, err)
		}
		if err := cli.DeleteDockerfileImage(defNamespace, input.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
			return errGatewayError(ctx, err)
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return image, nil
}

// DockerfileImage is the resolver for the DockerfileImage field.
func (r *queryResolver) DockerfileImage(ctx context.Context, cluster uuid.UUID, name string) (*model.DockerfileImage, error) {
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
	cli := k8s.NewMap(wc)

	obj, err := getDockerfileImage(cli, defNamespace, name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "name", name)
		}
		return nil, errGatewayError(ctx, err)
	}
	image := &model.DockerfileImage{Cluster: &model.Cluster{ID: cluster}}
	convertDockerfileImage(image, obj)
	return image, nil
}

// DockerfileImages is the resolver for the DockerfileImages field.
func (r *queryResolver) DockerfileImages(ctx context.Context, filters []*model.QueryFilter) ([]*model.DockerfileImage, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		clusters  []model1.Cluster
		images    []*model.DockerfileImage
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

	handle := func(outChan chan<- []*model.DockerfileImage, gqlErrChan chan<- *gqlerror.Error, errChan chan<- error, inChan <-chan *model1.Cluster) {
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
			cli := k8s.NewMap(wc)

			objs, err := getDockerfileImages(cli, defNamespace, selectors...)
			if err != nil {
				gqlErrChan <- errGatewayError(ctx, err)
				continue
			}

			var imgs []*model.DockerfileImage
			for i := range objs {
				image := &model.DockerfileImage{Cluster: &model.Cluster{ID: c.ID}}
				convertDockerfileImage(image, &objs[i])
				imgs = append(imgs, image)
			}
			outChan <- imgs
		}
	}

	inChan := make(chan *model1.Cluster, len(clusters))
	for i := range clusters {
		inChan <- &clusters[i]
	}
	close(inChan)
	outChan := make(chan []*model.DockerfileImage, 1)
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
			images = append(images, out...)
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
	return images, nil
}

// DockerfileImage returns graphql1.DockerfileImageResolver implementation.
func (r *Resolver) DockerfileImage() graphql1.DockerfileImageResolver {
	return &dockerfileImageResolver{r}
}

type dockerfileImageResolver struct{ *Resolver }
