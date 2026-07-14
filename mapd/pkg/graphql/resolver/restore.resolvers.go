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
	velerov1 "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"go.uber.org/multierr"
	"gorm.io/gorm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

// CreateRestore is the resolver for the CreateRestore field.
func (r *mutationResolver) CreateRestore(ctx context.Context, input model.CreateRestore) (*model.Restore, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}
	if input.Backup == "" {
		return nil, errInvalidRequest(ctx, "input.backup", "value cannot be empty")
	}

	restore := &model.Restore{Cluster: &model.Cluster{ID: input.Cluster}}
	cr := &model1.ClusterRestore{
		ID:        uuid.New(),
		Name:      input.Name,
		Namespace: defNamespace,
		ClusterID: model1.NewNullUUID(input.Cluster),
	}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(tx, model1.PermReadWrite, model1.PolicyObjectClusterResource)
		}

		c := &model1.Cluster{ID: input.Cluster}
		if err := db.Take(c).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.cluster", c.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		cr.OrganizationID = c.OrganizationID

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

		backup, err := velero.GetBackup(defNamespace, input.Backup)
		if err != nil {
			if k8s.IsErrObjectNotExists(err) {
				return errObjectNotFound(ctx, "input.backup", input.Backup)
			}
			return errGatewayError(ctx, err)
		}
		if backup.Status.Phase != velerov1.BackupPhaseCompleted {
			return errInvalidRequest(ctx, "input.backup", "backup is not completed", input.Backup)
		}

		if err := tx.Create(cr).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		obj := &velerov1.Restore{
			ObjectMeta: metav1.ObjectMeta{
				Name:        cr.Name,
				Namespace:   cr.Namespace,
				Annotations: backup.Annotations,
				Labels: map[string]string{
					"map.tanzu.vmware.com/backup": input.Backup,
					defLabelManaged:               "true",
				},
			},
		}
		if backup.Spec.StorageLocation != "" {
			obj.Labels["map.tanzu.vmware.com/location"] = backup.Spec.StorageLocation
		}

		includedNamespacesM := make(map[string]bool)
		excludedNamespacesM := make(map[string]bool)
		for each := range defExcludedNamespaces {
			excludedNamespacesM[each] = true
		}
		for _, each := range input.Spec.ExcludedNamespaces {
			excludedNamespacesM[each] = true
		}
		for _, each := range input.Spec.IncludedNamespaces {
			if excludedNamespacesM[each] {
				continue
			}
			includedNamespacesM[each] = true
		}
		var (
			includedNamespaces, excludedNamespaces []string
		)
		for each := range includedNamespacesM {
			includedNamespaces = append(includedNamespaces, each)
		}
		for each := range excludedNamespacesM {
			excludedNamespaces = append(excludedNamespaces, each)
		}
		obj.Spec.IncludedNamespaces = includedNamespaces
		obj.Spec.ExcludedNamespaces = excludedNamespaces
		if input.Spec.NamespaceMapping != nil && len(input.Spec.NamespaceMapping) > 0 {
			obj.Spec.NamespaceMapping = input.Spec.NamespaceMapping
		}
		obj.Spec.BackupName = input.Backup

		anno := getBackupAnnotation(backup)
		for i := range anno.Instances {
			instance := &anno.Instances[i]
			if obj.Spec.NamespaceMapping != nil {
				v, ok := obj.Spec.NamespaceMapping[instance.Namespace]
				if ok && v != "" {
					instance.Namespace = v
				}
			}
			inst, err := restoreInstance(tx, instance, c, c.OrganizationID.UUID)
			if err != nil {
				e := newGQLErrorf(ctx, CodeInvalidRequestError, err.Error())
				return e
			}
			if inst != nil {
				cri := &model1.ClusterRestoreInstance{
					ID:           uuid.New(),
					RestoreID:    model1.NewNullUUID(cr.ID),
					InstanceID:   model1.NewNullUUID(inst.ID),
					FederationID: inst.FederationID,
				}
				if err := tx.Create(cri).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
			}
		}

		if err := appendAuditLog(tx, "create", "Restore", c.ID, obj.Name, sess.UserID, sess.UserName, c.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		obj, err = velero.CreateRestore(obj)
		if err != nil {
			if k8s.IsErrObjectAlreadyExists(err) {
				return errObjectDuplicated(ctx, "input.name", input.Name)
			}
			return errGatewayError(ctx, err)
		}
		convertRestore(restore, obj)
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	mapd.EventEmitterClusterRestore.Next(&mapd.Event{
		Type: mapd.EventTypeAdd,
		OID:  cr.ID,
	})
	return restore, nil
}

// Restore is the resolver for the Restore field.
func (r *queryResolver) Restore(ctx context.Context, cluster uuid.UUID, name string) (*model.Restore, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
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
	velero := k8s.NewVelero(wc)

	obj, err := velero.GetRestore(defNamespace, name)
	if err != nil {
		if k8s.IsErrObjectNotExists(err) {
			if IsNotFoundOmitted(ctx) {
				return nil, nil
			}
			return nil, errObjectNotFound(ctx, "name", name)
		}
		return nil, errGatewayError(ctx, err)
	}
	restore := &model.Restore{Cluster: &model.Cluster{ID: cluster}}
	convertRestore(restore, obj)
	return restore, nil
}

// Restores is the resolver for the Restores field.
func (r *queryResolver) Restores(ctx context.Context, filters []*model.QueryFilter) ([]*model.Restore, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		clusters  []model1.Cluster
		restores  []*model.Restore
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
				selectors = append(selectors, k8s.LabelExists("map.tanzu.vmware.com/location"))
			case 1:
				selectors = append(selectors, k8s.LabelEqual("map.tanzu.vmware.com/location", each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn("map.tanzu.vmware.com/location", each.Value...))
			}
		case "backup":
			switch len(each.Value) {
			case 0:
				selectors = append(selectors, k8s.LabelExists("map.tanzu.vmware.com/backup"))
			case 1:
				selectors = append(selectors, k8s.LabelEqual("map.tanzu.vmware.com/backup", each.Value[0]))
			default:
				selectors = append(selectors, k8s.LabelIn("map.tanzu.vmware.com/backup", each.Value...))
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

	handle := func(outChan chan<- []*model.Restore, gqlErrChan chan<- *gqlerror.Error, errChan chan<- error, inChan <-chan *model1.Cluster) {
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

			objs, err := getRestores(velero, defNamespace, selectors...)
			if err != nil {
				gqlErrChan <- errGatewayError(ctx, err)
				continue
			}

			var restores []*model.Restore
			for i := range objs {
				restore := &model.Restore{Cluster: &model.Cluster{ID: c.ID}}
				convertRestore(restore, &objs[i])
				restores = append(restores, restore)
			}
			outChan <- restores
		}
	}

	inChan := make(chan *model1.Cluster, len(clusters))
	for i := range clusters {
		inChan <- &clusters[i]
	}
	close(inChan)
	outChan := make(chan []*model.Restore, 1)
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
			restores = append(restores, out...)
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
	return restores, nil
}

// Backup is the resolver for the backup field.
func (r *restoreResolver) Backup(ctx context.Context, obj *model.Restore) (*model.Backup, error) {
	if obj.Backup == nil || obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Backup(WithNotFoundOmitted(WithGuardSilenced(ctx)), obj.Cluster.ID, obj.Backup.Name)
}

// Location is the resolver for the location field.
func (r *restoreResolver) Location(ctx context.Context, obj *model.Restore) (*model.BackupStorageLocation, error) {
	if obj.Location == nil || obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).BackupStorageLocation(WithNotFoundOmitted(WithGuardSilenced(ctx)), obj.Cluster.ID, obj.Location.Name)
}

// Cluster is the resolver for the cluster field.
func (r *restoreResolver) Cluster(ctx context.Context, obj *model.Restore) (*model.Cluster, error) {
	if obj.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Cluster(WithGuardSilenced(ctx), obj.Cluster.ID)
}

// Restore returns graphql1.RestoreResolver implementation.
func (r *Resolver) Restore() graphql1.RestoreResolver { return &restoreResolver{r} }

type restoreResolver struct{ *Resolver }
