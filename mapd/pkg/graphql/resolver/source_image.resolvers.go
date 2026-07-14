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
	buildv1alpha1 "github.com/pivotal/kpack/pkg/apis/build/v1alpha1"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// CreateSourceImage is the resolver for the CreateSourceImage field.
func (r *mutationResolver) CreateSourceImage(ctx context.Context, input model.CreateSourceImage) (*model.SourceImage, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
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

	obj := &buildv1alpha1.Image{
		ObjectMeta: metav1.ObjectMeta{
			Name:        input.Name,
			Namespace:   defNamespace,
			Annotations: input.Annotations,
			Labels: map[string]string{
				defLabelManaged: "true",
			},
		},
		Spec: buildv1alpha1.ImageSpec{
			Builder: corev1.ObjectReference{
				Kind: "ClusterBuilder",
				Name: "default",
			},
			ServiceAccount: "default",
		},
	}
	if input.Spec.Source != nil {
		switch {
		case input.Spec.Source.Git != nil:
			obj.Spec.Source.Git = &buildv1alpha1.Git{
				Revision: *input.Spec.Source.Git.Revision,
				URL:      input.Spec.Source.Git.URL.String(),
			}
		case input.Spec.Source.Blob != nil:
			obj.Spec.Source.Blob = &buildv1alpha1.Blob{
				URL: input.Spec.Source.Blob.URL.String(),
			}
		case input.Spec.Source.Registry != nil:
			image, err := ParseImage(input.Spec.Source.Registry.Image)
			if err != nil {
				return nil, errInvalidRequest(ctx, "input.spec.source.registry.image", "invalid image", input.Spec.Source.Registry.Image)
			}
			if image.Project == "" {
				image.Project = "library"
			}
			if image.Registry == "" {
				registry := &model1.ImageRegistry{}
				if err := mapd.Database().Where("`id` IN (SELECT `image_registry_id` FROM `cluster_image_registry` WHERE `cluster_id` = ? AND `default` = ?)", c.ID, true).Take(registry).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
				u := abstract.MustParseURL(registry.Endpoint)
				image.Registry = u.Host
			}
			obj.Spec.Source.Registry = &buildv1alpha1.Registry{
				Image:            image.String(),
				ImagePullSecrets: []corev1.LocalObjectReference{{Name: defRegistrySecret}},
			}
		default:
			return nil, errInvalidRequest(ctx, "input.spec.source", "at least one source must be specified")
		}
		if input.Spec.Source.SubPath != nil {
			obj.Spec.Source.SubPath = *input.Spec.Source.SubPath
		}
	}
	if input.Spec.Env != nil {
		obj.Spec.Build = &buildv1alpha1.ImageBuild{}
		for k, v := range input.Spec.Env {
			obj.Spec.Build.Env = append(obj.Spec.Build.Env, corev1.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}
	img, err := ParseImage(input.Spec.DistImage)
	if err != nil {
		return nil, errInvalidRequest(ctx, "input.spec.distImage", "invalid image", input.Spec.DistImage)
	}
	if img.Project == "" {
		img.Project = "library"
	}
	if img.Registry == "" {
		registry := &model1.ImageRegistry{}
		if err := mapd.Database().Where("`id` IN (SELECT `image_registry_id` FROM `cluster_image_registry` WHERE `cluster_id` = ? AND `default` = ?)", c.ID, true).Take(registry).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		u := abstract.MustParseURL(registry.Endpoint)
		img.Registry = u.Host
	}
	img.Tag = ""
	obj.Spec.Tag = img.String()

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "create", "SourceImage", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = kpack.CreateImage(obj)
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

	image := &model.SourceImage{Cluster: &model.Cluster{ID: input.Cluster}}
	convertImage(image, obj)
	return image, nil
}

// UpdateSourceImage is the resolver for the UpdateSourceImage field.
func (r *mutationResolver) UpdateSourceImage(ctx context.Context, input model.UpdateSourceImage) (*model.SourceImage, error) {
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
	kpack := k8s.NewKpack(wc)

	obj, err := kpack.GetImage(defNamespace, input.Name)
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
			switch {
			case input.Spec.Source.Git != nil:
				if obj.Spec.Source.Git == nil {
					if input.Spec.Source.Git.Revision == nil || *input.Spec.Source.Git.Revision == "" {
						return nil, errInvalidRequest(ctx, "input.spec.source.git.revision", "value cannot be empty")
					}
					if input.Spec.Source.Git.URL == nil {
						return nil, errInvalidRequest(ctx, "input.spec.source.git.url", "value cannot be empty")
					}
					obj.Spec.Source.Git = &buildv1alpha1.Git{}
				}
				if input.Spec.Source.Git.Revision != nil {
					obj.Spec.Source.Git.Revision = *input.Spec.Source.Git.Revision
				}
				if input.Spec.Source.Git.URL != nil {
					obj.Spec.Source.Git.URL = input.Spec.Source.Git.URL.String()
				}
				obj.Spec.Source.Blob = nil
				obj.Spec.Source.Registry = nil
			case input.Spec.Source.Blob != nil:
				if obj.Spec.Source.Blob == nil {
					if input.Spec.Source.Blob.URL == nil {
						return nil, errInvalidRequest(ctx, "input.spec.source.blob.url", "value cannot be empty")
					}
					obj.Spec.Source.Blob = &buildv1alpha1.Blob{}
				}
				if input.Spec.Source.Blob.URL != nil {
					obj.Spec.Source.Blob.URL = input.Spec.Source.Blob.URL.String()
				}
				obj.Spec.Source.Git = nil
				obj.Spec.Source.Registry = nil
			case input.Spec.Source.Registry != nil:
				if obj.Spec.Source.Registry == nil {
					if input.Spec.Source.Registry.Image == nil || *input.Spec.Source.Registry.Image == "" {
						return nil, errInvalidRequest(ctx, "input.spec.source.registry.image", "value cannot be empty")
					}
					obj.Spec.Source.Registry = &buildv1alpha1.Registry{ImagePullSecrets: []corev1.LocalObjectReference{{Name: defRegistrySecret}}}
				}
				if input.Spec.Source.Registry.Image != nil && *input.Spec.Source.Registry.Image != "" {
					image, err := ParseImage(*input.Spec.Source.Registry.Image)
					if err != nil {
						return nil, errInvalidRequest(ctx, "input.spec.source.registry.image", "invalid image", *input.Spec.Source.Registry.Image)
					}
					if image.Project == "" {
						image.Project = "library"
					}
					if image.Registry == "" {
						registry := &model1.ImageRegistry{}
						if err := mapd.Database().Where("`id` IN (SELECT `image_registry_id` FROM `cluster_image_registry` WHERE `cluster_id` = ? AND `default` = ?)", c.ID, true).Take(registry).Error; err != nil {
							panic(&ErrDBConnection{err})
						}
						u := abstract.MustParseURL(registry.Endpoint)
						image.Registry = u.Host
					}
					obj.Spec.Source.Registry.Image = image.String()
				}
				obj.Spec.Source.Git = nil
				obj.Spec.Source.Blob = nil
			}
			if input.Spec.Source.SubPath != nil {
				obj.Spec.Source.SubPath = *input.Spec.Source.SubPath
			}
		}
		if input.Spec.Env != nil {
			if obj.Spec.Build == nil {
				obj.Spec.Build = &buildv1alpha1.ImageBuild{}
			}
			obj.Spec.Build.Env = obj.Spec.Build.Env[:0]
			for k, v := range input.Spec.Env {
				obj.Spec.Build.Env = append(obj.Spec.Build.Env, corev1.EnvVar{Name: k, Value: v})
			}
		}
		if input.Spec.DistImage != nil {
			img, err := ParseImage(*input.Spec.DistImage)
			if err != nil {
				return nil, errInvalidRequest(ctx, "input.spec.distImage", "invalid image", *input.Spec.DistImage)
			}
			if img.Project == "" {
				img.Project = "library"
			}
			if img.Registry == "" {
				registry := &model1.ImageRegistry{}
				if err := mapd.Database().Where("`id` IN (SELECT `image_registry_id` FROM `cluster_image_registry` WHERE `cluster_id` = ? AND `default` = ?)", c.ID, true).Take(registry).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
				u := abstract.MustParseURL(registry.Endpoint)
				img.Registry = u.Host
			}
			img.Tag = ""
			obj.Spec.Tag = img.String()
		}
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "update", "SourceImage", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = kpack.UpdateImage(obj)
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
	image := &model.SourceImage{Cluster: &model.Cluster{ID: input.Cluster}}
	convertImage(image, obj)
	return image, nil
}

// DeleteSourceImage is the resolver for the DeleteSourceImage field.
func (r *mutationResolver) DeleteSourceImage(ctx context.Context, input model.DeleteSourceImage) (*model.SourceImage, error) {
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

	obj, err := kpack.GetImage(defNamespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}
	image := &model.SourceImage{Cluster: &model.Cluster{ID: input.Cluster}}
	convertImage(image, obj)

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "delete", "SourceImage", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := kpack.DeleteImage(defNamespace, input.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
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

// SourceImage is the resolver for the SourceImage field.
func (r *queryResolver) SourceImage(ctx context.Context, cluster uuid.UUID, name string) (*model.SourceImage, error) {
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
	kpack := k8s.NewKpack(wc)

	obj, err := kpack.GetImage(defNamespace, name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "name", name)
		}
		return nil, errGatewayError(ctx, err)
	}
	image := &model.SourceImage{Cluster: &model.Cluster{ID: cluster}}
	convertImage(image, obj)
	return image, nil
}

// SourceImages is the resolver for the SourceImages field.
func (r *queryResolver) SourceImages(ctx context.Context, filters []*model.QueryFilter) ([]*model.SourceImage, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		clusters  []model1.Cluster
		images    []*model.SourceImage
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
				exprs = append(exprs, "`cluster_id` = ?")
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
	if err := db.Where("`id` IN (SELECT `cluster_id` FROM `cluster_extension` WHERE `name` = ?)", model1.ClusterExtensionNameTBS).Where("`provisioned` >= ?", model1.ClusterProvisionFinish).Find(&clusters).Error; err != nil {
		panic(&ErrDBConnection{err})
	}

	handle := func(outChan chan<- []*model.SourceImage, gqlErrChan chan<- *gqlerror.Error, errChan chan<- error, inChan <-chan *model1.Cluster) {
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

			objs, err := getImages(kpack, defNamespace, selectors...)
			if err != nil {
				gqlErrChan <- errGatewayError(ctx, err)
				continue
			}

			var imgs []*model.SourceImage
			for i := range objs {
				image := &model.SourceImage{Cluster: &model.Cluster{ID: c.ID}}
				convertImage(image, &objs[i])
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
	outChan := make(chan []*model.SourceImage, 1)
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

// Cluster is the resolver for the cluster field.
func (r *sourceImageResolver) Cluster(ctx context.Context, obj *model.SourceImage) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// SourceImage returns graphql1.SourceImageResolver implementation.
func (r *Resolver) SourceImage() graphql1.SourceImageResolver { return &sourceImageResolver{r} }

type sourceImageResolver struct{ *Resolver }
