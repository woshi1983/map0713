package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	"github.com/vektah/gqlparser/v2/gqlerror"
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/multierr"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// Restores is the resolver for the restores field.
func (r *backupResolver) Restores(ctx context.Context, obj *model.Backup, filters []*model.QueryFilter) ([]*model.Restore, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Restores(
		WithGuardSilenced(ctx),
		append(filters,
			&model.QueryFilter{Key: "cluster", Value: []string{obj.Cluster.ID.String()}},
			&model.QueryFilter{Key: "backup", Value: []string{obj.Name}},
		),
	)
}

// Location is the resolver for the location field.
func (r *backupResolver) Location(ctx context.Context, obj *model.Backup) (*model.BackupStorageLocation, error) {
	if obj.Location == nil || obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).BackupStorageLocation(WithNotFoundOmitted(WithGuardSilenced(ctx)), obj.Cluster.ID, obj.Location.Name)
}

// Cluster is the resolver for the cluster field.
func (r *backupResolver) Cluster(ctx context.Context, obj *model.Backup) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// CreateBackup is the resolver for the CreateBackup field.
func (r *mutationResolver) CreateBackup(ctx context.Context, input model.CreateBackup) (*model.Backup, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}
	if input.Location == "" {
		return nil, errInvalidRequest(ctx, "input.location", "value cannot be empty")
	}

	backup := &model.Backup{Cluster: &model.Cluster{ID: input.Cluster}}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadWrite, model1.PolicyObjectClusterResource)
		}
		c := &model1.Cluster{ID: input.Cluster}
		if err := db.Take(c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.cluster", c.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if c.Provisioned < model1.ClusterProvisionFinish {
			return errInvalidRequest(ctx, "input.cluster", "cluster is not ready", input.Cluster.String())
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
		velero := k8s.NewVelero(wc)

		obj := &velerov1.Backup{
			ObjectMeta: metav1.ObjectMeta{
				Name:      input.Name,
				Namespace: defNamespace,
				Labels: map[string]string{
					defLabelManaged: "true",
				},
			},
			Spec: velerov1.BackupSpec{},
		}
		if input.Spec == nil {
			input.Spec = &model.CreateBackupSpec{}
		}
		if input.Spec.UseRestic != nil && *input.Spec.UseRestic {
			obj.Spec.DefaultVolumesToRestic = util.BoolPtr(true)
		}
		includedNamespacesM := make(map[string]bool)
		excludedNamespacesM := make(map[string]bool)
		for each := range defExcludedNamespaces {
			excludedNamespacesM[each] = true
		}
		for _, each := range input.Spec.ExcludedNamespaces {
			excludedNamespacesM[each] = true
		}
		var includedNamespaceNotSpecified bool
		if len(input.Spec.IncludedNamespaces) == 0 {
			namespaces, err := getNamespaces(wc)
			if err != nil {
				return errGatewayError(ctx, err)
			}
			for _, each := range namespaces {
				input.Spec.IncludedNamespaces = append(input.Spec.IncludedNamespaces, each.GetName())
			}
			includedNamespaceNotSpecified = true
		}
		for i, each := range input.Spec.IncludedNamespaces {
			if excludedNamespacesM[each] {
				continue
			}

			instancesNotReady := int64(0)
			if err := tx.Model(&model1.Instance{}).Where("`state` IN (?, ?, ?) AND `namespace` = ? AND `cluster_id` = ?", model1.InstanceStatePending, model1.InstanceStateProvisioning, model1.InstanceStateDeleting, each, c.ID).Count(&instancesNotReady).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if instancesNotReady > 0 {
				if includedNamespaceNotSpecified {
					return errObjectIsBusy(ctx, "input.spec.includedNamespaces", "Instance", nil, "")
				}
				return errObjectIsBusy(ctx, fmt.Sprintf("input.spec.includedNamespaces.%d", i), "Instance", nil, "", each)
			}
			includedNamespacesM[each] = true
		}

		db = tx
		var (
			includedNamespaces, excludedNamespaces []string
			annotation                             model1.BackupAnnotation
		)
		if len(input.Spec.Components) > 0 {
			selector := metav1.LabelSelectorRequirement{
				Key:      defLabelComponent,
				Operator: metav1.LabelSelectorOpIn,
			}

			var (
				ss   []string
				args []interface{}
			)
			for _, each := range input.Spec.Components {
				ss = append(ss, "?")
				args = append(args, each.String())
				selector.Values = append(selector.Values, each.String())
			}
			db = db.Where("`component_id` IN ("+strings.Join(ss, ", ")+")", args...)

			obj.Spec.LabelSelector = &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{selector},
			}
		}
		for each := range includedNamespacesM {
			instances := []model1.Instance{}
			if err := db.Where("`state` = ? AND `namespace` = ? AND `cluster_id` = ?", model1.InstanceStateSuccess, each, c.ID).Find(&instances).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			for _, instance := range instances {
				annotation.Instances = append(annotation.Instances, model1.BackupAnnotationInstance{
					Name:          instance.Name,
					Namespace:     instance.Namespace,
					ReleaseID:     instance.ReleaseID.String(),
					FederationID:  instance.FederationID.String(),
					ComponentID:   instance.ComponentID.String(),
					ApplicationID: instance.ApplicationID.String(),
				})
			}
			includedNamespaces = append(includedNamespaces, each)
		}
		for each := range excludedNamespacesM {
			excludedNamespaces = append(excludedNamespaces, each)
		}
		b, err := json.Marshal(annotation)
		if err != nil {
			panic(err)
		}
		obj.Annotations = map[string]string{
			model1.BackupAnnotationName: string(b),
		}
		obj.Spec.IncludedNamespaces = includedNamespaces
		obj.Spec.ExcludedNamespaces = excludedNamespaces
		obj.Spec.IncludedResources = input.Spec.IncludedResources
		obj.Spec.ExcludedResources = input.Spec.ExcludedResources
		obj.Spec.StorageLocation = input.Location
		if input.Spec.TTL != nil {
			obj.Spec.TTL = metav1.Duration{Duration: *input.Spec.TTL}
		}
		obj.Spec.IncludeClusterResources = input.Spec.IncludeClusterResources

		if err := appendAuditLog(tx, "create", "Backup", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = velero.CreateBackup(obj)
		if err != nil {
			if k8s.IsErrObjectAlreadyExists(err) {
				return errObjectDuplicated(ctx, "input.name", input.Name)
			}
			return errGatewayError(ctx, err)
		}
		convertBackup(backup, obj)
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return backup, nil
}

// DeleteBackup is the resolver for the DeleteBackup field.
func (r *mutationResolver) DeleteBackup(ctx context.Context, input model.DeleteBackup) (*model.Backup, error) {
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
	velero := k8s.NewVelero(wc)

	obj, err := velero.GetBackup(defNamespace, input.Name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			return nil, errObjectNotFound(ctx, "input.name", input.Name)
		}
		return nil, errGatewayError(ctx, err)
	}
	backup := &model.Backup{Cluster: &model.Cluster{ID: input.Cluster}}
	convertBackup(backup, obj)

	if backup.Location != nil {
		_, err := velero.GetBackupStorageLocation(defNamespace, backup.Location.Name)
		if err != nil && !k8s.IsErrObjectNotExists(err) {
			return nil, errGatewayError(ctx, err)
		}
		if err == nil {
			if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
				if err := appendAuditLog(tx, "delete", "Backup", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
					panic(&ErrDBConnection{err})
				}

				if err := deleteBackupFromStorageLocation(velero, defNamespace, input.Name); err != nil && !k8s.IsErrObjectAlreadyExists(err) {
					return errGatewayError(ctx, err)
				}
				return nil
			}); err != nil {
				if !IsGQLError(err) {
					panic(err)
				}
				return nil, err
			}
			return backup, nil
		}
	}
	if err := velero.DeleteBackup(defNamespace, input.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
		return nil, errGatewayError(ctx, err)
	}
	return backup, nil
}

// Backup is the resolver for the Backup field.
func (r *queryResolver) Backup(ctx context.Context, cluster uuid.UUID, name string) (*model.Backup, error) {
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
	velero := k8s.NewVelero(wc)

	obj, err := velero.GetBackup(defNamespace, name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "name", name)
		}
		return nil, errGatewayError(ctx, err)
	}
	backup := &model.Backup{Cluster: &model.Cluster{ID: cluster}}
	convertBackup(backup, obj)
	return backup, nil
}

// Backups is the resolver for the Backups field.
func (r *queryResolver) Backups(ctx context.Context, filters []*model.QueryFilter) ([]*model.Backup, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		clusters  []model1.Cluster
		backups   []*model.Backup
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
		case "location":
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists("velero.io/storage-location"))
			case 1:
				selectors = append(selectors, k8s.LabelEqual("velero.io/storage-location", each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn("velero.io/storage-location", each.Value...))
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

	handle := func(outChan chan<- []*model.Backup, gqlErrChan chan<- *gqlerror.Error, errChan chan<- error, inChan <-chan *model1.Cluster) {
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
			velero := k8s.NewVelero(wc)

			list, err := velero.GetBackups(defNamespace, selectors...)
			if err != nil {
				gqlErrChan <- errGatewayError(ctx, err)
				continue
			}

			var backups []*model.Backup
			for i := range list.Items {
				secret := &model.Backup{Cluster: &model.Cluster{ID: c.ID}}
				convertBackup(secret, &list.Items[i])
				backups = append(backups, secret)
			}
			outChan <- backups
		}
	}

	inChan := make(chan *model1.Cluster, len(clusters))
	for i := range clusters {
		inChan <- &clusters[i]
	}
	close(inChan)
	outChan := make(chan []*model.Backup, 1)
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
			backups = append(backups, out...)
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
	return backups, nil
}

// Backup returns graphql1.BackupResolver implementation.
func (r *Resolver) Backup() graphql1.BackupResolver { return &backupResolver{r} }

type backupResolver struct{ *Resolver }
