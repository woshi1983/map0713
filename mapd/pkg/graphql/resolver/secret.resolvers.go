package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// CreateSecret is the resolver for the CreateSecret field.
func (r *mutationResolver) CreateSecret(ctx context.Context, input model.CreateSecret) (*model.Secret, error) {
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

	obj := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Name,
			Namespace: input.Namespace,
			Labels: map[string]string{
				defLabelManaged: "true",
			},
		},
		Data: make(map[string][]byte),
		Type: corev1.SecretType(input.Type),
	}
	for k, v := range input.Data {
		obj.Data[k] = []byte(v)
	}

	switch obj.Type {
	case corev1.SecretTypeOpaque:
	case corev1.SecretTypeDockerConfigJson:
		b, ok := obj.Data[".dockerconfigjson"]
		if !ok || b == nil || len(b) == 0 {
			return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "value cannot be empty")
		}
		cred := &k8s.DockerConfigJSON{}
		if err := json.Unmarshal(b, cred); err != nil {
			return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "value cannot be decoded", string(b))
		}
		if cred.Auths == nil || len(cred.Auths) < 1 {
			return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "no available credential", string(b))
		}
		for k, v := range cred.Auths {
			_, err := url.Parse(k)
			if err != nil {
				return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "invalid registry server", string(b))
			}
			if v.Username == "" {
				return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "invalid registry username", string(b))
			}
			if v.Password == "" {
				return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "invalid registry password", string(b))
			}
			if v.Email != "" {
				_, err := mail.ParseAddress(v.Email)
				if err != nil {
					return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "invalid email address", string(b))
				}
			}
		}
	case corev1.SecretTypeBasicAuth:
		b, ok := obj.Data["username"]
		if !ok || b == nil || len(b) == 0 {
			return nil, errInvalidRequest(ctx, "input.data.username", "value cannot be empty")
		}
		b, ok = obj.Data["password"]
		if !ok || b == nil || len(b) == 0 {
			return nil, errInvalidRequest(ctx, "input.data.password", "value cannot be empty")
		}
	case corev1.SecretTypeSSHAuth:
		b, ok := obj.Data["ssh-privatekey"]
		if !ok || b == nil || len(b) == 0 {
			return nil, errInvalidRequest(ctx, "input.data[\"ssh-privatekey\"]", "value cannot be empty")
		}
	case corev1.SecretTypeTLS:
		b, ok := obj.Data["tls.key"]
		if !ok || b == nil || len(b) == 0 {
			return nil, errInvalidRequest(ctx, "input.data[\"tls.key\"]", "value cannot be empty")
		}
		b, ok = obj.Data["tls.crt"]
		if !ok || b == nil || len(b) == 0 {
			return nil, errInvalidRequest(ctx, "input.data[\"tls.crt\"]", "value cannot be empty")
		}
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "create", "Secret", c.ID, obj.Namespace+"/"+obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = wc.CreateSecret(obj)
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

	secret := &model.Secret{Cluster: &model.Cluster{ID: input.Cluster}}
	convertSecret(secret, obj)
	return secret, nil
}

// UpdateSecret is the resolver for the UpdateSecret field.
func (r *mutationResolver) UpdateSecret(ctx context.Context, input model.UpdateSecret) (*model.Secret, error) {
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

	obj, err := wc.GetSecret(input.Namespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}
	if input.Data != nil {
		obj.Data = make(map[string][]byte)
		for k, v := range input.Data {
			obj.Data[k] = []byte(v)
		}
	}
	switch obj.Type {
	case corev1.SecretTypeOpaque:
	case corev1.SecretTypeDockerConfigJson:
		b, ok := obj.Data[".dockerconfigjson"]
		if !ok || b == nil || len(b) == 0 {
			return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "value cannot be empty")
		}
		cred := &k8s.DockerConfigJSON{}
		if err := json.Unmarshal(b, cred); err != nil {
			return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "value cannot be decoded", string(b))
		}
		if cred.Auths == nil || len(cred.Auths) < 1 {
			return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "no available credential", string(b))
		}
		for k, v := range cred.Auths {
			_, err := url.Parse(k)
			if err != nil {
				return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "invalid registry server", string(b))
			}
			if v.Username == "" {
				return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "invalid registry username", string(b))
			}
			if v.Password == "" {
				return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "invalid registry password", string(b))
			}
			if v.Email != "" {
				_, err := mail.ParseAddress(v.Email)
				if err != nil {
					return nil, errInvalidRequest(ctx, "input.data[\".dockerconfigjson\"]", "invalid email address", string(b))
				}
			}
		}
	case corev1.SecretTypeBasicAuth:
		b, ok := obj.Data["username"]
		if !ok || b == nil || len(b) == 0 {
			return nil, errInvalidRequest(ctx, "input.data.username", "value cannot be empty")
		}
		b, ok = obj.Data["password"]
		if !ok || b == nil || len(b) == 0 {
			return nil, errInvalidRequest(ctx, "input.data.password", "value cannot be empty")
		}
	case corev1.SecretTypeSSHAuth:
		b, ok := obj.Data["ssh-privatekey"]
		if !ok || b == nil || len(b) == 0 {
			return nil, errInvalidRequest(ctx, "input.data[\"ssh-privatekey\"]", "value cannot be empty")
		}
	case corev1.SecretTypeTLS:
		b, ok := obj.Data["tls.key"]
		if !ok || b == nil || len(b) == 0 {
			return nil, errInvalidRequest(ctx, "input.data[\"tls.key\"]", "value cannot be empty")
		}
		b, ok = obj.Data["tls.crt"]
		if !ok || b == nil || len(b) == 0 {
			return nil, errInvalidRequest(ctx, "input.data[\"tls.crt\"]", "value cannot be empty")
		}
	}

	if obj.Labels == nil {
		obj.Labels = make(map[string]string)
	}
	obj.Labels[defLabelManaged] = "true"

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "update", "Secret", c.ID, obj.Namespace+"/"+obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = wc.UpdateSecret(obj)
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

	secret := &model.Secret{Cluster: &model.Cluster{ID: input.Cluster}}
	convertSecret(secret, obj)
	return secret, nil
}

// DeleteSecret is the resolver for the DeleteSecret field.
func (r *mutationResolver) DeleteSecret(ctx context.Context, input model.DeleteSecret) (*model.Secret, error) {
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

	obj, err := wc.GetSecret(input.Namespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}
	secret := &model.Secret{Cluster: &model.Cluster{ID: input.Cluster}}
	convertSecret(secret, obj)

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := appendAuditLog(tx, "delete", "Secret", c.ID, obj.Namespace+"/"+obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := wc.DeleteSecret(input.Namespace, input.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
			return errGatewayError(ctx, err)
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}

	return secret, nil
}

// Secret is the resolver for the Secret field.
func (r *queryResolver) Secret(ctx context.Context, cluster uuid.UUID, namespace string, name string) (*model.Secret, error) {
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

	obj, err := wc.GetSecret(namespace, name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "name", name)
		}
		return nil, errGatewayError(ctx, err)
	}
	secret := &model.Secret{Cluster: &model.Cluster{ID: cluster}}
	convertSecret(secret, obj)
	return secret, nil
}

// Secrets is the resolver for the Secrets field.
func (r *queryResolver) Secrets(ctx context.Context, filters []*model.QueryFilter) ([]*model.Secret, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		clusters  []model1.Cluster
		secrets   []*model.Secret
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
		case "type":
			if len(each.Value) > 0 {
				selectors = append(selectors, k8s.FieldEqual("type", each.Value[0]))
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

	handle := func(outChan chan<- []*model.Secret, gqlErrChan chan<- *gqlerror.Error, errChan chan<- error, inChan <-chan *model1.Cluster) {
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

			objs, err := getSecrets(wc, "", selectors...)
			if err != nil {
				gqlErrChan <- errGatewayError(ctx, err)
				continue
			}

			var secrets []*model.Secret
			for i := range objs {
				secret := &model.Secret{Cluster: &model.Cluster{ID: c.ID}}
				convertSecret(secret, &objs[i])
				secrets = append(secrets, secret)
			}
			outChan <- secrets
		}
	}

	inChan := make(chan *model1.Cluster, len(clusters))
	for i := range clusters {
		inChan <- &clusters[i]
	}
	close(inChan)
	outChan := make(chan []*model.Secret, 1)
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
			secrets = append(secrets, out...)
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
	return secrets, nil
}

// Instance is the resolver for the instance field.
func (r *secretResolver) Instance(ctx context.Context, obj *model.Secret) (*model.Instance, error) {
	if obj.Instance == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Instance(WithGuardSilenced(ctx), obj.Instance.ID)
}

// Component is the resolver for the component field.
func (r *secretResolver) Component(ctx context.Context, obj *model.Secret) (*model.Component, error) {
	if obj.Component == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Component(WithGuardSilenced(ctx), obj.Component.ID)
}

// Application is the resolver for the application field.
func (r *secretResolver) Application(ctx context.Context, obj *model.Secret) (*model.Application, error) {
	if obj.Application == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Application(WithGuardSilenced(ctx), obj.Application.ID)
}

// Cluster is the resolver for the cluster field.
func (r *secretResolver) Cluster(ctx context.Context, obj *model.Secret) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// Secret returns graphql1.SecretResolver implementation.
func (r *Resolver) Secret() graphql1.SecretResolver { return &secretResolver{r} }

type secretResolver struct{ *Resolver }
