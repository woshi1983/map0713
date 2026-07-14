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
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// Instance is the resolver for the instance field.
func (r *ingressResolver) Instance(ctx context.Context, obj *model.Ingress) (*model.Instance, error) {
	if obj.Instance == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Instance(WithGuardSilenced(ctx), obj.Instance.ID)
}

// Component is the resolver for the component field.
func (r *ingressResolver) Component(ctx context.Context, obj *model.Ingress) (*model.Component, error) {
	if obj.Component == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Component(WithGuardSilenced(ctx), obj.Component.ID)
}

// Application is the resolver for the application field.
func (r *ingressResolver) Application(ctx context.Context, obj *model.Ingress) (*model.Application, error) {
	if obj.Application == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Application(WithGuardSilenced(ctx), obj.Application.ID)
}

// Cluster is the resolver for the cluster field.
func (r *ingressResolver) Cluster(ctx context.Context, obj *model.Ingress) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// CreateIngress is the resolver for the CreateIngress field.
func (r *mutationResolver) CreateIngress(ctx context.Context, input model.CreateIngress) (*model.Ingress, error) {
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

	obj := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        input.Name,
			Namespace:   input.Namespace,
			Annotations: input.Annotations,
			Labels: map[string]string{
				defLabelManaged: "true",
			},
		},
	}
	if input.Labels != nil {
		for k, v := range input.Labels {
			obj.Labels[k] = v
		}
	}
	if input.Spec != nil {
		obj.Spec.IngressClassName = input.Spec.IngressClass
		for _, each := range input.Spec.TLS {
			v := networkingv1.IngressTLS{Hosts: each.Hosts}
			if each.Secret != nil {
				v.SecretName = *each.Secret
			}
			obj.Spec.TLS = append(obj.Spec.TLS, v)
		}
		for _, each := range input.Spec.Rules {
			rule := networkingv1.IngressRule{}
			if each.Host != nil {
				rule.Host = *each.Host
			}
			if each.HTTP != nil {
				rule.HTTP = &networkingv1.HTTPIngressRuleValue{}
				for _, path := range each.HTTP.Paths {
					v := networkingv1.HTTPIngressPath{}
					if path.Path != nil {
						v.Path = *path.Path
					}
					if path.PathType != nil {
						t := networkingv1.PathType(*path.PathType)
						v.PathType = &t
					}
					if path.Backend != nil {
						if path.Backend.Service != nil {
							v.Backend.Service = &networkingv1.IngressServiceBackend{
								Name: path.Backend.Service.Name,
							}
							if path.Backend.Service.Port != nil {
								if path.Backend.Service.Port.Name != nil {
									v.Backend.Service.Port.Name = *path.Backend.Service.Port.Name
								}
								if path.Backend.Service.Port.Number != nil {
									v.Backend.Service.Port.Number = int32(*path.Backend.Service.Port.Number)
								}
							}
						}
					}
					rule.HTTP.Paths = append(rule.HTTP.Paths, v)
				}
			}
			obj.Spec.Rules = append(obj.Spec.Rules, rule)
		}
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "create", "Ingress", c.ID, obj.Namespace+"/"+obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = wc.CreateIngress(obj)
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
	ingress := &model.Ingress{Cluster: &model.Cluster{ID: input.Cluster}}
	convertIngress(ingress, obj)
	return ingress, nil
}

// UpdateIngress is the resolver for the UpdateIngress field.
func (r *mutationResolver) UpdateIngress(ctx context.Context, input model.UpdateIngress) (*model.Ingress, error) {
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

	obj, err := wc.GetIngress(input.Namespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}

	if input.Annotations != nil {
		if len(input.Annotations) == 0 {
			obj.Annotations = nil
		} else {
			obj.Annotations = input.Annotations
		}
	}
	if input.Labels != nil {
		obj.Labels = input.Labels
	}
	if obj.Labels == nil {
		obj.Labels = make(map[string]string)
	}
	obj.Labels[defLabelManaged] = "true"
	if input.Spec != nil {
		obj.Spec.IngressClassName = input.Spec.IngressClass
		obj.Spec.TLS = obj.Spec.TLS[:0]
		for _, each := range input.Spec.TLS {
			v := networkingv1.IngressTLS{Hosts: each.Hosts}
			if each.Secret != nil {
				v.SecretName = *each.Secret
			}
			obj.Spec.TLS = append(obj.Spec.TLS, v)
		}
		obj.Spec.Rules = obj.Spec.Rules[:0]
		for _, each := range input.Spec.Rules {
			rule := networkingv1.IngressRule{}
			if each.Host != nil {
				rule.Host = *each.Host
			}
			if each.HTTP != nil {
				rule.HTTP = &networkingv1.HTTPIngressRuleValue{}
				for _, path := range each.HTTP.Paths {
					v := networkingv1.HTTPIngressPath{}
					if path.Path != nil {
						v.Path = *path.Path
					}
					if path.PathType != nil {
						t := networkingv1.PathType(*path.PathType)
						v.PathType = &t
					}
					if path.Backend != nil {
						if path.Backend.Service != nil {
							v.Backend.Service = &networkingv1.IngressServiceBackend{
								Name: path.Backend.Service.Name,
							}
							if path.Backend.Service.Port != nil {
								if path.Backend.Service.Port.Name != nil {
									v.Backend.Service.Port.Name = *path.Backend.Service.Port.Name
								}
								if path.Backend.Service.Port.Number != nil {
									v.Backend.Service.Port.Number = int32(*path.Backend.Service.Port.Number)
								}
							}
						}
					}
					rule.HTTP.Paths = append(rule.HTTP.Paths, v)
				}
			}
			obj.Spec.Rules = append(obj.Spec.Rules, rule)
		}
	}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "update", "Ingress", c.ID, obj.Namespace+"/"+obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = wc.UpdateIngress(obj)
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
	ingress := &model.Ingress{Cluster: &model.Cluster{ID: input.Cluster}}
	convertIngress(ingress, obj)
	return ingress, nil
}

// DeleteIngress is the resolver for the DeleteIngress field.
func (r *mutationResolver) DeleteIngress(ctx context.Context, input model.DeleteIngress) (*model.Ingress, error) {
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

	obj, err := wc.GetIngress(input.Namespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}
	ingress := &model.Ingress{Cluster: &model.Cluster{ID: input.Cluster}}
	convertIngress(ingress, obj)

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "delete", "Ingress", c.ID, obj.Namespace+"/"+obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := wc.DeleteIngress(input.Namespace, input.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
			return errGatewayError(ctx, err)
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return ingress, nil
}

// Ingress is the resolver for the Ingress field.
func (r *queryResolver) Ingress(ctx context.Context, cluster uuid.UUID, namespace string, name string) (*model.Ingress, error) {
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

	obj, err := wc.GetIngress(namespace, name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "name", name)
		}
		return nil, errGatewayError(ctx, err)
	}
	ingress := &model.Ingress{Cluster: &model.Cluster{ID: cluster}}
	convertIngress(ingress, obj)
	return ingress, nil
}

// Ingresses is the resolver for the Ingresses field.
func (r *queryResolver) Ingresses(ctx context.Context, filters []*model.QueryFilter) ([]*model.Ingress, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		clusters  []model1.Cluster
		ingresses []*model.Ingress
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

	handle := func(outChan chan<- []*model.Ingress, gqlErrChan chan<- *gqlerror.Error, errChan chan<- error, inChan <-chan *model1.Cluster) {
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

			objs, err := getIngresses(wc, "", selectors...)
			if err != nil {
				gqlErrChan <- errGatewayError(ctx, err)
				continue
			}

			var ings []*model.Ingress
			for i := range objs {
				ing := &model.Ingress{Cluster: &model.Cluster{ID: c.ID}}
				convertIngress(ing, &objs[i])
				ings = append(ings, ing)
			}
			outChan <- ings
		}
	}

	inChan := make(chan *model1.Cluster, len(clusters))
	for i := range clusters {
		inChan <- &clusters[i]
	}
	close(inChan)
	outChan := make(chan []*model.Ingress, 1)
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
			ingresses = append(ingresses, out...)
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
	return ingresses, nil
}

// Ingress returns graphql1.IngressResolver implementation.
func (r *Resolver) Ingress() graphql1.IngressResolver { return &ingressResolver{r} }

type ingressResolver struct{ *Resolver }
