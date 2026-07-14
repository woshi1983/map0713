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
	"go.uber.org/multierr"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// CreatePersistentVolumeClaim is the resolver for the CreatePersistentVolumeClaim field.
func (r *mutationResolver) CreatePersistentVolumeClaim(ctx context.Context, input model.CreatePersistentVolumeClaim) (*model.PersistentVolumeClaim, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}
	if input.Namespace == "" {
		return nil, errInvalidRequest(ctx, "input.namespace", "value cannot be empty")
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

	obj := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Name,
			Namespace: input.Namespace,
			Labels: map[string]string{
				defLabelManaged: "true",
			},
		},
	}
	for _, each := range input.Spec.AccessModes {
		obj.Spec.AccessModes = append(obj.Spec.AccessModes, corev1.PersistentVolumeAccessMode(each))
	}
	if len(obj.Spec.AccessModes) == 0 {
		obj.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	}
	if input.Spec.Resources != nil {
		if input.Spec.Resources.Limits != nil {
			obj.Spec.Resources.Limits = make(corev1.ResourceList)
			if input.Spec.Resources.Limits.Storage != nil {
				v, err := resource.ParseQuantity(*input.Spec.Resources.Limits.Storage)
				if err != nil {
					return nil, errInvalidRequest(ctx, "input.spec.resources.limits.storage", err.Error(), *input.Spec.Resources.Limits.Storage)
				}
				obj.Spec.Resources.Limits[corev1.ResourceStorage] = v
			}
		}
		if input.Spec.Resources.Requests != nil {
			obj.Spec.Resources.Requests = make(corev1.ResourceList)
			v, err := resource.ParseQuantity(input.Spec.Resources.Requests.Storage)
			if err != nil {
				return nil, errInvalidRequest(ctx, "input.spec.resources.requests.storage", err.Error(), input.Spec.Resources.Requests.Storage)
			}
			obj.Spec.Resources.Requests[corev1.ResourceStorage] = v
		}
	}
	if input.Spec.StorageClass != nil {
		obj.Spec.StorageClassName = input.Spec.StorageClass
	}
	if input.Spec.VolumeMode != nil {
		v := corev1.PersistentVolumeMode(*input.Spec.VolumeMode)
		obj.Spec.VolumeMode = &v
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "create", "PersistentVolumeClaim", c.ID, obj.Namespace+"/"+obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = wc.CreatePersistentVolumeClaim(obj)
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

	pvc := &model.PersistentVolumeClaim{Cluster: &model.Cluster{ID: input.Cluster}}
	convertPersistentVolumeClaim(pvc, obj)
	return pvc, nil
}

// UpdatePersistentVolumeClaim is the resolver for the UpdatePersistentVolumeClaim field.
func (r *mutationResolver) UpdatePersistentVolumeClaim(ctx context.Context, input model.UpdatePersistentVolumeClaim) (*model.PersistentVolumeClaim, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if input.Name == "" {
		return nil, errInvalidRequest(ctx, "input.name", "value cannot be empty")
	}
	if input.Namespace == "" {
		return nil, errInvalidRequest(ctx, "input.namespace", "value cannot be empty")
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

	obj, err := wc.GetPersistentVolumeClaim(input.Namespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}
	if input.Spec != nil {
		if len(input.Spec.AccessModes) > 0 {
			obj.Spec.AccessModes = obj.Spec.AccessModes[:0]
			for _, each := range input.Spec.AccessModes {
				obj.Spec.AccessModes = append(obj.Spec.AccessModes, corev1.PersistentVolumeAccessMode(each))
			}
		}
		if input.Spec.Resources != nil {
			if input.Spec.Resources.Limits != nil {
				if obj.Spec.Resources.Limits == nil {
					obj.Spec.Resources.Limits = make(corev1.ResourceList)
				}
				if input.Spec.Resources.Limits.Storage != nil {
					v, err := resource.ParseQuantity(*input.Spec.Resources.Limits.Storage)
					if err != nil {
						return nil, errInvalidRequest(ctx, "input.spec.resources.limits.storage", err.Error(), *input.Spec.Resources.Limits.Storage)
					}
					obj.Spec.Resources.Limits[corev1.ResourceStorage] = v
				}
			}
			if input.Spec.Resources.Requests != nil {
				if obj.Spec.Resources.Requests == nil {
					obj.Spec.Resources.Requests = make(corev1.ResourceList)
				}
				if input.Spec.Resources.Requests.Storage != nil {
					v, err := resource.ParseQuantity(*input.Spec.Resources.Requests.Storage)
					if err != nil {
						return nil, errInvalidRequest(ctx, "input.spec.resources.requests.storage", err.Error(), *input.Spec.Resources.Requests.Storage)
					}
					obj.Spec.Resources.Requests[corev1.ResourceStorage] = v
				}
			}
		}
		if input.Spec.StorageClass != nil {
			obj.Spec.StorageClassName = input.Spec.StorageClass
		}
		if input.Spec.VolumeMode != nil {
			v := corev1.PersistentVolumeMode(*input.Spec.VolumeMode)
			obj.Spec.VolumeMode = &v
		}
	}

	if obj.Labels == nil {
		obj.Labels = make(map[string]string)
	}
	obj.Labels[defLabelManaged] = "true"

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "update", "PersistentVolumeClaim", c.ID, obj.Namespace+"/"+obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = wc.UpdatePersistentVolumeClaim(obj)
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

	pvc := &model.PersistentVolumeClaim{Cluster: &model.Cluster{ID: input.Cluster}}
	convertPersistentVolumeClaim(pvc, obj)
	return pvc, nil
}

// DeletePersistentVolumeClaim is the resolver for the DeletePersistentVolumeClaim field.
func (r *mutationResolver) DeletePersistentVolumeClaim(ctx context.Context, input model.DeletePersistentVolumeClaim) (*model.PersistentVolumeClaim, error) {
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

	obj, err := wc.GetPersistentVolumeClaim(input.Namespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}
	pvc := &model.PersistentVolumeClaim{Cluster: &model.Cluster{ID: input.Cluster}}
	convertPersistentVolumeClaim(pvc, obj)

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "delete", "PersistentVolumeClaim", c.ID, obj.Namespace+"/"+obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := wc.DeletePersistentVolumeClaim(input.Namespace, input.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
			return errGatewayError(ctx, err)
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return pvc, nil
}

// Instance is the resolver for the instance field.
func (r *persistentVolumeClaimResolver) Instance(ctx context.Context, obj *model.PersistentVolumeClaim) (*model.Instance, error) {
	if obj.Instance == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Instance(WithGuardSilenced(ctx), obj.Instance.ID)
}

// Component is the resolver for the component field.
func (r *persistentVolumeClaimResolver) Component(ctx context.Context, obj *model.PersistentVolumeClaim) (*model.Component, error) {
	if obj.Component == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Component(WithGuardSilenced(ctx), obj.Component.ID)
}

// Application is the resolver for the application field.
func (r *persistentVolumeClaimResolver) Application(ctx context.Context, obj *model.PersistentVolumeClaim) (*model.Application, error) {
	if obj.Application == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Application(WithGuardSilenced(ctx), obj.Application.ID)
}

// Cluster is the resolver for the cluster field.
func (r *persistentVolumeClaimResolver) Cluster(ctx context.Context, obj *model.PersistentVolumeClaim) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// PersistentVolumeClaim is the resolver for the PersistentVolumeClaim field.
func (r *queryResolver) PersistentVolumeClaim(ctx context.Context, cluster uuid.UUID, namespace string, name string) (*model.PersistentVolumeClaim, error) {
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

	obj, err := wc.GetPersistentVolumeClaim(namespace, name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "name", name)
		}
		return nil, errGatewayError(ctx, err)
	}
	pvc := &model.PersistentVolumeClaim{Cluster: &model.Cluster{ID: cluster}}
	convertPersistentVolumeClaim(pvc, obj)
	return pvc, nil
}

// PersistentVolumeClaims is the resolver for the PersistentVolumeClaims field.
func (r *queryResolver) PersistentVolumeClaims(ctx context.Context, filters []*model.QueryFilter) ([]*model.PersistentVolumeClaim, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		clusters  []model1.Cluster
		pvcs      []*model.PersistentVolumeClaim
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

	handle := func(outChan chan<- []*model.PersistentVolumeClaim, gqlErrChan chan<- *gqlerror.Error, errChan chan<- error, inChan <-chan *model1.Cluster) {
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

			objs, err := getPersistentVolumeClaims(wc, "", selectors...)
			if err != nil {
				gqlErrChan <- errGatewayError(ctx, err)
				continue
			}

			var pvcs []*model.PersistentVolumeClaim
			for i := range objs {
				pvc := &model.PersistentVolumeClaim{Cluster: &model.Cluster{ID: c.ID}}
				convertPersistentVolumeClaim(pvc, &objs[i])
				pvcs = append(pvcs, pvc)
			}
			outChan <- pvcs
		}
	}

	inChan := make(chan *model1.Cluster, len(clusters))
	for i := range clusters {
		inChan <- &clusters[i]
	}
	close(inChan)
	outChan := make(chan []*model.PersistentVolumeClaim, 1)
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
			pvcs = append(pvcs, out...)
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
	return pvcs, nil
}

// PersistentVolumeClaim returns graphql1.PersistentVolumeClaimResolver implementation.
func (r *Resolver) PersistentVolumeClaim() graphql1.PersistentVolumeClaimResolver {
	return &persistentVolumeClaimResolver{r}
}

type persistentVolumeClaimResolver struct{ *Resolver }
