package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/sshutil"
	tanzuv2 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/tanzu/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/rest"
)

// TLSCertChain is the resolver for the tlsCertChain field.
func (r *clusterResolver) TLSCertChain(ctx context.Context, obj *model.Cluster) ([]*model.X509Certificate, error) {
	if obj.Endpoint == nil {
		return nil, nil
	}

	zctx, zcancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
	defer zcancel()

	chain, err := httputil.EZGetTLSCertificateChainContext(zctx, obj.Endpoint.String(), obj.InsecureSkipTLSVerify, []byte(obj.CaCert+"\n"))
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	var certs []*model.X509Certificate
	for _, each := range chain {
		certs = append(certs, convertX509Certificate(each))
	}
	return certs, nil
}

// SSHPrivateKey is the resolver for the sshPrivateKey field.
func (r *clusterResolver) SSHPrivateKey(ctx context.Context, obj *model.Cluster) (*string, error) {
	if obj.Type != model.ClusterTypeProvisioned {
		return nil, nil
	}

	var sshPrivateKey string
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		switch obj.Class {
		case model.ClusterClassTkgs:
			c := &model1.Cluster{ID: obj.ID}
			if err := tx.Take(c).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				panic(&ErrDBConnection{err})
			}
			spec := &model1.ClusterTKGSSpec{}
			if err := json.Unmarshal(c.Spec, spec); err != nil {
				panic(err)
			}

			cp := &model1.ClusterProvider{ID: c.ProviderID.UUID}
			if err := tx.Take(cp).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				panic(&ErrDBConnection{err})
			}
			cpSpec := &model1.ClusterProviderTKGSSpec{}
			if err := json.Unmarshal(cp.Spec, cpSpec); err != nil {
				panic(err)
			}
			block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
			if err != nil {
				panic(err)
			}
			sc, err := k8s.NewTanzuKubernetesGridService(cpSpec.Endpoint, cpSpec.AuthUser, block.Decrypt(cpSpec.AuthPassword), rest.TLSClientConfig{
				Insecure: cpSpec.InsecureSkipTLSVerify,
				CAData:   cpSpec.CACert,
			})
			if err != nil {
				panic(err)
			}

			privateKeyName := c.Name + "-ssh"
			ssecret, err := sc.GetSecret(spec.Namespace, privateKeyName)
			if err != nil {
				if k8s.IsErrObjectNotExists(err) {
					return nil
				}
				return errGatewayError(ctx, err)
			}
			if ssecret.Type != corev1.SecretTypeSSHAuth {
				return nil
			}
			privateKey, ok := ssecret.Data["ssh-privatekey"]
			if !ok {
				return nil
			}
			sshPrivateKey = string(privateKey)
		case model.ClusterClassTkgm:
			c := &model1.Cluster{ID: obj.ID}
			if err := tx.Take(c).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				panic(&ErrDBConnection{err})
			}
			spec := &model1.ClusterTKGMSpec{}
			if err := json.Unmarshal(c.Spec, spec); err != nil {
				panic(err)
			}
			sshPrivateKey = string(spec.SSHPrivateKey)
		case model.ClusterClassTkg2s:
			c := &model1.Cluster{ID: obj.ID}
			if err := tx.Take(c).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				panic(&ErrDBConnection{err})
			}
			spec := &model1.ClusterTKG2SSpec{}
			if err := json.Unmarshal(c.Spec, spec); err != nil {
				panic(err)
			}

			cp := &model1.ClusterProvider{ID: c.ProviderID.UUID}
			if err := tx.Take(cp).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				panic(&ErrDBConnection{err})
			}
			cpSpec := &model1.ClusterProviderTKG2SSpec{}
			if err := json.Unmarshal(cp.Spec, cpSpec); err != nil {
				panic(err)
			}
			block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
			if err != nil {
				panic(err)
			}
			sc, err := tanzuv2.NewVSphereWithTanzu(cpSpec.Endpoint, cpSpec.AuthUser, block.Decrypt(cpSpec.AuthPassword), rest.TLSClientConfig{
				Insecure: cpSpec.InsecureSkipTLSVerify,
				CAData:   cpSpec.CACert,
			})
			if err != nil {
				panic(err)
			}

			privateKeyName := c.Name + "-ssh"
			ssecret, err := sc.GetSecret(spec.Namespace, privateKeyName)
			if err != nil {
				if k8s.IsErrObjectNotExists(err) {
					return nil
				}
				return errGatewayError(ctx, err)
			}
			if ssecret.Type != corev1.SecretTypeSSHAuth {
				return nil
			}
			privateKey, ok := ssecret.Data["ssh-privatekey"]
			if !ok {
				return nil
			}
			sshPrivateKey = string(privateKey)
		case model.ClusterClassTkg2m:
			c := &model1.Cluster{ID: obj.ID}
			if err := tx.Take(c).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				panic(&ErrDBConnection{err})
			}
			spec := &model1.ClusterTKG2MSpec{}
			if err := json.Unmarshal(c.Spec, spec); err != nil {
				panic(err)
			}
			sshPrivateKey = string(spec.SSHPrivateKey)
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	if sshPrivateKey != "" {
		return &sshPrivateKey, nil
	}
	return nil, nil
}

// Namespaces is the resolver for the namespaces field.
func (r *clusterResolver) Namespaces(ctx context.Context, obj *model.Cluster, filters []*model.QueryFilter) ([]*model.Namespace, error) {
	return r.Resolver.Query().(*queryResolver).Namespaces(
		WithGuardSilenced(ctx),
		append(filters,
			&model.QueryFilter{Key: "cluster", Value: []string{obj.ID.String()}},
		),
	)
}

// Extensions is the resolver for the extensions field.
func (r *clusterResolver) Extensions(ctx context.Context, obj *model.Cluster, page *model.QueryPage, filters []*model.QueryFilter) (*model.ClusterExtensions, error) {
	return r.Resolver.Query().(*queryResolver).ClusterExtensions(
		WithGuardSilenced(ctx),
		page,
		append(filters,
			&model.QueryFilter{Key: "cluster", Value: []string{obj.ID.String()}},
		),
	)
}

// ImageRegistries is the resolver for the imageRegistries field.
func (r *clusterResolver) ImageRegistries(ctx context.Context, obj *model.Cluster, page *model.QueryPage, filters []*model.QueryFilter) (*model.ClusterImageRegistries, error) {
	result := &model.ClusterImageRegistries{Cluster: obj.ID, Page: &model.PageInfo{}}
	if page.Limit != nil {
		result.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		result.Page.Number = *page.Number
	}
	for _, each := range filters {
		result.Filters = append(result.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.ClusterImageRegistry{}).Where("`cluster_id` = ?", obj.ID)
	db = finalizeClusterImageRegistryQueryFilter(db, result.Filters...)
	if err := db.Count(&result.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(result.Page, result.TotalCount)
	return result, nil
}

// LoadBalancer is the resolver for the loadBalancer field.
func (r *clusterResolver) LoadBalancer(ctx context.Context, obj *model.Cluster) (*model.LoadBalancer, error) {
	if obj.LoadBalancer == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).LoadBalancer(WithGuardSilenced(ctx), obj.LoadBalancer.ID)
}

// Provider is the resolver for the provider field.
func (r *clusterResolver) Provider(ctx context.Context, obj *model.Cluster) (*model.ClusterProvider, error) {
	if obj.Provider == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).ClusterProvider(WithGuardSilenced(ctx), obj.Provider.ID)
}

// Organization is the resolver for the organization field.
func (r *clusterResolver) Organization(ctx context.Context, obj *model.Cluster) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *clusterImageRegistriesResolver) Edge(ctx context.Context, obj *model.ClusterImageRegistries) (*model.ClusterImageRegistryNodes, error) {
	nodes := &model.ClusterImageRegistryNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Where("`cluster_id` = ?", obj.Cluster).Order("`id` ASC")
		db = finalizeClusterImageRegistryQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var registries []model1.ClusterImageRegistry
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&registries).Error; err != nil {
				return err
			}
			for _, registry := range registries {
				nodes.Nodes = append(nodes.Nodes, convertClusterImageRegistryFromSQLModel(&registry))
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

// ImageRegistry is the resolver for the imageRegistry field.
func (r *clusterImageRegistryResolver) ImageRegistry(ctx context.Context, obj *model.ClusterImageRegistry) (*model.ImageRegistry, error) {
	return r.Resolver.Query().(*queryResolver).ImageRegistry(WithGuardSilenced(ctx), obj.ID)
}

// Edge is the resolver for the edge field.
func (r *clustersResolver) Edge(ctx context.Context, obj *model.Clusters) (*model.ClusterNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.ClusterNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectCluster)
		}
		db = finalizeClusterQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var clusters []model1.Cluster
			if err := db.Limit(int(obj.Page.Count)).Offset(int(obj.Page.Offset)).Find(&clusters).Error; err != nil {
				return err
			}
			for _, cluster := range clusters {
				nodes.Nodes = append(nodes.Nodes, convertClusterFromSQLModel(&cluster))
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

// CreateCluster is the resolver for the CreateCluster field.
func (r *mutationResolver) CreateCluster(ctx context.Context, input model.CreateCluster) (*model.Cluster, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}

	cluster := &model1.Cluster{
		ID:             uuid.New(),
		Name:           input.Name,
		Type:           model1.ClusterTypeProvisioned,
		Version:        input.Version.String(),
		ProvisionPhase: model1.ClusterProvisionPrepare,
		State:          model1.ClusterStateCreating,
		Status:         model1.ClusterStatusPending,
		ProviderID:     model1.NewNullUUID(input.Provider),
	}
	if input.Description != nil {
		cluster.Description = *input.Description
	}
	var extensionsToProvision []*model1.ClusterExtension

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		cp := &model1.ClusterProvider{ID: cluster.ProviderID.UUID}
		if err := tx.Take(cp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.provider", cp.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectCluster, cp.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		cluster.Class = cp.Class
		cluster.LoadBalancerID = cp.LoadBalancerID
		cluster.OrganizationID = cp.OrganizationID

		var exists int64
		if err := tx.Model(&model1.Cluster{}).Where("`name` = ?", cluster.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", cluster.Name)
		}

		usage := &model1.OrganizationQuotaUsage{
			ID:             uuid.New(),
			OrganizationID: cluster.OrganizationID,
		}
		usageRef := &model1.ClusterQuotaUsage{
			ID:        uuid.New(),
			ClusterID: model1.NewNullUUID(cluster.ID),
			UsageID:   model1.NewNullUUID(usage.ID),
		}

		switch cluster.Class {
		case model1.ClusterClassTKGS:
			cpSpec := &model1.ClusterProviderTKGSSpec{}
			if err := json.Unmarshal(cp.Spec, cpSpec); err != nil {
				panic(err)
			}

			if input.Spec.Tkgs == nil {
				return errInvalidRequest(ctx, "input.spec.tkgs", "value cannot be empty")
			}
			if input.Spec.Tkgs.Namespace == nil || *input.Spec.Tkgs.Namespace == "" {
				input.Spec.Tkgs.Namespace = util.StringPtr(cpSpec.DefaultNS)
			}

			block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
			if err != nil {
				panic(err)
			}
			client, err := k8s.NewTanzuKubernetesGridService(cpSpec.Endpoint, cpSpec.AuthUser, block.Decrypt(cpSpec.AuthPassword), rest.TLSClientConfig{
				Insecure: cpSpec.InsecureSkipTLSVerify,
				CAData:   cpSpec.CACert,
			})
			if err != nil {
				panic(err)
			}
			ctlPlaneClass, err := client.GetVirtualMachineClass(input.Spec.Tkgs.CtlPlaneClass)
			if err != nil {
				if k8s.IsErrObjectNotExists(err) {
					return errObjectNotFound(ctx, "input.spec.tkgs.ctlPlaneClass", input.Spec.Tkgs.CtlPlaneClass)
				}
				return errGatewayError(ctx, err)
			}
			dataPlaneClass, err := client.GetVirtualMachineClass(input.Spec.Tkgs.DataPlaneClass)
			if err != nil {
				if k8s.IsErrObjectNotExists(err) {
					return errObjectNotFound(ctx, "input.spec.tkgs.dataPlaneClass", input.Spec.Tkgs.DataPlaneClass)
				}
				return errGatewayError(ctx, err)
			}

			spec := &model1.ClusterTKGSSpec{
				Namespace:             *input.Spec.Tkgs.Namespace,
				CNI:                   *input.Spec.Tkgs.Cni,
				PodCIDR:               input.Spec.Tkgs.PodCidr.String(),
				ServiceCIDR:           input.Spec.Tkgs.SvcCidr.String(),
				ServiceDomain:         *input.Spec.Tkgs.SvcDomain,
				DefStorage:            input.Spec.Tkgs.DefStorage,
				CtlPlaneClass:         input.Spec.Tkgs.CtlPlaneClass,
				CtlPlaneCount:         *input.Spec.Tkgs.CtlPlaneCount,
				CtlPlaneStorage:       input.Spec.Tkgs.CtlPlaneStorage,
				CtlPlaneStorageCapGi:  *input.Spec.Tkgs.CtlPlaneStorageCapGi,
				DataPlaneClass:        input.Spec.Tkgs.DataPlaneClass,
				DataPlaneCount:        *input.Spec.Tkgs.DataPlaneCount,
				DataPlaneStorage:      input.Spec.Tkgs.DataPlaneStorage,
				DataPlaneStorageCapGi: *input.Spec.Tkgs.DataPlaneStorageCapGi,
			}
			b, err := json.Marshal(spec)
			if err != nil {
				panic(err)
			}
			cluster.Spec = b

			usage.CPUs = spec.CtlPlaneCount*ctlPlaneClass.Spec.Hardware.Cpus + spec.DataPlaneCount*dataPlaneClass.Spec.Hardware.Cpus
			usage.MemoryMi = spec.CtlPlaneCount*ctlPlaneClass.Spec.Hardware.Memory.ScaledValue(resource.Mega) + spec.DataPlaneCount*dataPlaneClass.Spec.Hardware.Memory.ScaledValue(resource.Mega)
			usage.StorageGi = spec.CtlPlaneCount*spec.CtlPlaneStorageCapGi + spec.DataPlaneCount*spec.DataPlaneStorageCapGi
		case model1.ClusterClassTKGM:
			return errNoLongerSupport(ctx, "TKGm 1.X is no longer supported. Please consider to migrate your clusters to TKG 2.")
		case model1.ClusterClassTKG2S:
			cpSpec := &model1.ClusterProviderTKG2SSpec{}
			if err := json.Unmarshal(cp.Spec, cpSpec); err != nil {
				panic(err)
			}

			if input.Spec.Tkg2s == nil {
				return errInvalidRequest(ctx, "input.spec.tkg2s", "value cannot be empty")
			}
			if input.Spec.Tkg2s.Namespace == nil || *input.Spec.Tkg2s.Namespace == "" {
				input.Spec.Tkg2s.Namespace = util.StringPtr(cpSpec.DefaultNS)
			}

			block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
			if err != nil {
				panic(err)
			}
			client, err := tanzuv2.NewVSphereWithTanzu(cpSpec.Endpoint, cpSpec.AuthUser, block.Decrypt(cpSpec.AuthPassword), rest.TLSClientConfig{
				Insecure: cpSpec.InsecureSkipTLSVerify,
				CAData:   cpSpec.CACert,
			})
			if err != nil {
				panic(err)
			}
			ctlPlaneVmClass, err := client.GetVirtualMachineClass(input.Spec.Tkg2s.CtlPlaneVMClass)
			if err != nil {
				if k8s.IsErrObjectNotExists(err) {
					return errObjectNotFound(ctx, "input.spec.tkg2s.ctlPlaneVmClass", input.Spec.Tkg2s.CtlPlaneVMClass)
				}
				return errGatewayError(ctx, err)
			}
			workerVmClass, err := client.GetVirtualMachineClass(input.Spec.Tkg2s.WorkerVMClass)
			if err != nil {
				if k8s.IsErrObjectNotExists(err) {
					return errObjectNotFound(ctx, "input.spec.tkg2s.workerVmClass", input.Spec.Tkg2s.WorkerVMClass)
				}
				return errGatewayError(ctx, err)
			}

			spec := &model1.ClusterTKG2SSpec{
				Namespace:                    *input.Spec.Tkg2s.Namespace,
				OSName:                       *input.Spec.Tkg2s.OsName,
				CNI:                          *input.Spec.Tkg2s.Cni,
				ClusterCIDR:                  input.Spec.Tkg2s.ClusterCidr.String(),
				ServiceCIDR:                  input.Spec.Tkg2s.ServiceCidr.String(),
				StorageClass:                 input.Spec.Tkg2s.StorageClass,
				CtlPlaneReplicas:             *input.Spec.Tkg2s.CtlPlaneReplicas,
				CtlPlaneVmClass:              input.Spec.Tkg2s.CtlPlaneVMClass,
				CtlPlaneEtcdVolumeSizeGi:     *input.Spec.Tkg2s.CtlPlaneEtcdVolumeSizeGi,
				WorkerNodePool0Replicas:      *input.Spec.Tkg2s.WorkerNodePool0Replicas,
				WorkerVmClass:                input.Spec.Tkg2s.WorkerVMClass,
				WorkerContainerdVolumeSizeGi: *input.Spec.Tkg2s.WorkerContainerdVolumeSizeGi,
			}
			if input.Spec.Tkg2s.WorkerNodePool0Zone != nil {
				spec.WorkerNodePool0Zone = *input.Spec.Tkg2s.WorkerNodePool0Zone
			}
			if input.Spec.Tkg2s.WorkerNodePool1Replicas != nil {
				spec.WorkerNodePool1Replicas = *input.Spec.Tkg2s.WorkerNodePool1Replicas
			}
			if input.Spec.Tkg2s.WorkerNodePool1Zone != nil {
				spec.WorkerNodePool1Zone = *input.Spec.Tkg2s.WorkerNodePool1Zone
			}
			if input.Spec.Tkg2s.WorkerNodePool2Replicas != nil {
				spec.WorkerNodePool2Replicas = *input.Spec.Tkg2s.WorkerNodePool2Replicas
			}
			if input.Spec.Tkg2s.WorkerNodePool0Zone != nil {
				spec.WorkerNodePool0Zone = *input.Spec.Tkg2s.WorkerNodePool0Zone
			}
			if spec.CtlPlaneReplicas < 1 {
				return errInvalidRequest(ctx, "input.spec.tkg2s.ctlPlaneReplicas", "value must be greater than zero", spec.CtlPlaneReplicas)
			}
			if spec.WorkerNodePool0Replicas < 1 {
				return errInvalidRequest(ctx, "input.spec.tkg2s.workerNodePool0Replicas", "value must be greater than zero", spec.WorkerNodePool0Replicas)
			}
			if spec.WorkerNodePool1Replicas < 1 && spec.WorkerNodePool2Replicas > 0 {
				spec.WorkerNodePool1Replicas, spec.WorkerNodePool2Replicas = spec.WorkerNodePool2Replicas, spec.WorkerNodePool1Replicas
				spec.WorkerNodePool1Zone, spec.WorkerNodePool2Zone = spec.WorkerNodePool2Zone, spec.WorkerNodePool1Zone
			}
			for _, each := range cpSpec.TrustedCAs {
				spec.TrustedCAs = append(spec.TrustedCAs, each.Data)
			}

			b, err := json.Marshal(spec)
			if err != nil {
				panic(err)
			}
			cluster.Spec = b

			usage.CPUs = spec.CtlPlaneReplicas*ctlPlaneVmClass.Spec.Hardware.Cpus + (spec.WorkerNodePool0Replicas+spec.WorkerNodePool1Replicas+spec.WorkerNodePool2Replicas)*workerVmClass.Spec.Hardware.Cpus
			usage.MemoryMi = spec.CtlPlaneReplicas*ctlPlaneVmClass.Spec.Hardware.Memory.ScaledValue(resource.Mega) + (spec.WorkerNodePool0Replicas+spec.WorkerNodePool1Replicas+spec.WorkerNodePool2Replicas)*workerVmClass.Spec.Hardware.Memory.ScaledValue(resource.Mega)
			usage.StorageGi = spec.CtlPlaneReplicas*spec.CtlPlaneEtcdVolumeSizeGi + (spec.WorkerNodePool0Replicas+spec.WorkerNodePool1Replicas+spec.WorkerNodePool2Replicas)*spec.WorkerContainerdVolumeSizeGi
		case model1.ClusterClassTKG2M:
			cpSpec := &model1.ClusterProviderTKG2MSpec{}
			if err := json.Unmarshal(cp.Spec, cpSpec); err != nil {
				panic(err)
			}

			if input.Spec.Tkg2m == nil {
				return errInvalidRequest(ctx, "input.spec.tkg2m", "value cannot be empty")
			}
			if input.Spec.Tkg2m.Namespace == nil || *input.Spec.Tkg2m.Namespace == "" {
				input.Spec.Tkg2m.Namespace = util.StringPtr(cpSpec.DefaultNS)
			}

			var (
				sshPrivateKey crypto.PrivateKey
				err           error
			)
			if input.Spec.Tkg2m.SSHPrivateKey != nil && *input.Spec.Tkg2m.SSHPrivateKey != "" {
				sshPrivateKey, err = sshutil.LoadPrivateKey([]byte(*input.Spec.Tkg2m.SSHPrivateKey))
				if err != nil {
					return errInvalidRequest(ctx, "input.spec.tkg2m.sshPrivateKey", "private key is not in supported PEM layout", *input.Spec.Tkg2m.SSHPrivateKey)
				}
			} else {
				sshPrivateKey, err = sshutil.GeneratePrivateKey(4096)
				if err != nil {
					panic(err)
				}
			}
			pk, err := sshutil.EncodePrivateKey(sshPrivateKey)
			if err != nil {
				panic(err)
			}

			spec := &model1.ClusterTKG2MSpec{
				Namespace:               *input.Spec.Tkg2m.Namespace,
				OSName:                  *input.Spec.Tkg2m.OsName,
				CNI:                     *input.Spec.Tkg2m.Cni,
				ClusterCIDR:             input.Spec.Tkg2m.ClusterCidr.String(),
				ServiceCIDR:             input.Spec.Tkg2m.ServiceCidr.String(),
				CtlPlaneReplicas:        *input.Spec.Tkg2m.CtlPlaneReplicas,
				WorkerNodePool0Replicas: *input.Spec.Tkg2m.WorkerNodePool0Replicas,
				SSHPrivateKey:           pk,
				VsphereNetwork:          cpSpec.VsphereNetwork,
				VsphereResourcePool:     cpSpec.VsphereResourcePool,
				VsphereFolder:           cpSpec.VsphereFolder,
				VsphereDatastore:        cpSpec.VsphereDatastore,
				AviServiceEngineGroup:   cpSpec.AviServiceEngineGroup,
				AviDataNetwork:          cpSpec.AviDataNetwork,
				AviDataNetworkCidr:      cpSpec.AviDataNetworkCidr,
			}
			if spec.CtlPlaneReplicas < 1 {
				return errInvalidRequest(ctx, "input.spec.tkg2m.ctlPlaneReplicas", "value must be greater than zero", spec.CtlPlaneReplicas)
			}
			if spec.WorkerNodePool0Replicas < 1 {
				return errInvalidRequest(ctx, "input.spec.tkg2m.workerNodePool0Replicas", "value must be greater than zero", spec.WorkerNodePool0Replicas)
			}
			if input.Spec.Tkg2m.CtlPlaneNumCpus != nil {
				spec.CtlPlaneNumCpus = *input.Spec.Tkg2m.CtlPlaneNumCpus
			}
			if input.Spec.Tkg2m.CtlPlaneMemMiB != nil {
				spec.CtlPlaneMemMiB = *input.Spec.Tkg2m.CtlPlaneMemMiB
			}
			if input.Spec.Tkg2m.CtlPlaneDiskGiB != nil {
				spec.CtlPlaneDiskGiB = *input.Spec.Tkg2m.CtlPlaneDiskGiB
			}
			if input.Spec.Tkg2m.WorkerNodePool0Zone != nil {
				spec.WorkerNodePool0Zone = *input.Spec.Tkg2m.WorkerNodePool0Zone
			}
			if input.Spec.Tkg2m.WorkerNodePool1Replicas != nil {
				spec.WorkerNodePool1Replicas = *input.Spec.Tkg2m.WorkerNodePool1Replicas
			}
			if input.Spec.Tkg2m.WorkerNodePool1Zone != nil {
				spec.WorkerNodePool1Zone = *input.Spec.Tkg2m.WorkerNodePool1Zone
			}
			if input.Spec.Tkg2m.WorkerNodePool2Replicas != nil {
				spec.WorkerNodePool2Replicas = *input.Spec.Tkg2m.WorkerNodePool2Replicas
			}
			if input.Spec.Tkg2m.WorkerNodePool0Zone != nil {
				spec.WorkerNodePool0Zone = *input.Spec.Tkg2m.WorkerNodePool0Zone
			}
			if spec.WorkerNodePool1Replicas < 1 && spec.WorkerNodePool2Replicas > 0 {
				spec.WorkerNodePool1Replicas, spec.WorkerNodePool2Replicas = spec.WorkerNodePool2Replicas, spec.WorkerNodePool1Replicas
				spec.WorkerNodePool1Zone, spec.WorkerNodePool2Zone = spec.WorkerNodePool2Zone, spec.WorkerNodePool1Zone
			}
			if input.Spec.Tkg2m.WorkerNumCpus != nil {
				spec.WorkerNumCpus = *input.Spec.Tkg2m.WorkerNumCpus
			}
			if input.Spec.Tkg2m.WorkerMemMiB != nil {
				spec.WorkerMemMiB = *input.Spec.Tkg2m.WorkerMemMiB
			}
			if input.Spec.Tkg2m.WorkerDiskGiB != nil {
				spec.WorkerDiskGiB = *input.Spec.Tkg2m.WorkerDiskGiB
			}
			for _, each := range cpSpec.TrustedCAs {
				spec.TrustedCAs = append(spec.TrustedCAs, each.Data)
			}
			if input.Spec.Tkg2m.VsphereNetwork != nil {
				spec.VsphereNetwork = *input.Spec.Tkg2m.VsphereNetwork
			}
			if input.Spec.Tkg2m.VsphereResourcePool != nil {
				spec.VsphereResourcePool = *input.Spec.Tkg2m.VsphereResourcePool
			}
			if input.Spec.Tkg2m.VsphereFolder != nil {
				spec.VsphereFolder = *input.Spec.Tkg2m.VsphereFolder
			}
			if input.Spec.Tkg2m.VsphereDatastore != nil {
				spec.VsphereDatastore = *input.Spec.Tkg2m.VsphereDatastore
			}
			if input.Spec.Tkg2m.AviServiceEngineGroup != nil {
				spec.AviServiceEngineGroup = *input.Spec.Tkg2m.AviServiceEngineGroup
			}
			if input.Spec.Tkg2m.AviDataNetwork != nil {
				spec.AviDataNetwork = *input.Spec.Tkg2m.AviDataNetwork
			}
			if input.Spec.Tkg2m.AviDataNetworkCidr != nil {
				spec.AviDataNetworkCidr = *input.Spec.Tkg2m.AviDataNetworkCidr
			}

			b, err := json.Marshal(spec)
			if err != nil {
				panic(err)
			}
			cluster.Spec = b

			usage.CPUs = spec.CtlPlaneReplicas*spec.CtlPlaneNumCpus + (spec.WorkerNodePool0Replicas+spec.WorkerNodePool1Replicas+spec.WorkerNodePool2Replicas)*spec.WorkerNumCpus
			usage.MemoryMi = spec.CtlPlaneReplicas*spec.CtlPlaneMemMiB + (spec.WorkerNodePool0Replicas+spec.WorkerNodePool1Replicas+spec.WorkerNodePool2Replicas)*spec.WorkerMemMiB
			usage.StorageGi = spec.CtlPlaneReplicas*spec.CtlPlaneDiskGiB + (spec.WorkerNodePool0Replicas+spec.WorkerNodePool1Replicas+spec.WorkerNodePool2Replicas)*spec.WorkerDiskGiB
		}

		if err := appendAuditLog(tx, "create", "Cluster", cluster.ID, cluster.Name, sess.UserID, sess.UserName, cluster.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Create(cluster).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(cluster).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		var imageRegistries []model1.ClusterProviderImageRegistry
		if err := tx.Where("`provider_id` = ?", cp.ID).Find(&imageRegistries).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		for _, each := range imageRegistries {
			reg := &model1.ClusterImageRegistry{
				ID:              uuid.New(),
				Default:         each.Default,
				ClusterID:       model1.NewNullUUID(cluster.ID),
				ImageRegistryID: each.ImageRegistryID,
			}
			if err := tx.Create(reg).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}

		if err := tx.Create(usage).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Create(usageRef).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if !validateOrganizationQuotaUsage(tx, cluster.OrganizationID.UUID) {
			return errQuotaExceeded(ctx, "input.spec")
		}

		for i, extension := range input.Extensions {
			ext := &model1.ClusterExtension{
				ID:             uuid.New(),
				Name:           model1.ClusterExtensionName(extension.Name.String()),
				State:          model1.ClusterExtensionStateEnabling,
				ClusterID:      model1.NewNullUUID(cluster.ID),
				OrganizationID: cluster.OrganizationID,
			}
			if extension.Spec != nil {
				var err error
				ext.Spec, err = json.Marshal(extension.Spec)
				if err != nil {
					panic(err)
				}
			}

			var exists int64
			if err := tx.Model(&model1.ClusterExtension{}).Where("`cluster_id` = ? AND `name` = ?", ext.ClusterID, ext.Name).Count(&exists).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if exists > 0 {
				return errObjectDuplicated(ctx, fmt.Sprintf("input.extensions.%d.name", i), ext.Name.String())
			}

			for j, each := range extension.Related {
				c := &model1.Cluster{ID: each}
				if err := tx.Take(c).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return errObjectNotFound(ctx, fmt.Sprintf("input.extensions.%d.related.%d", i, j), each.String())
					}
					panic(&ErrDBConnection{err})
				}
				if c.Provisioned < model1.ClusterProvisionFinish {
					return errInvalidRequest(ctx, fmt.Sprintf("input.extensions.%d.related.%d", i, j), "cluster is not ready", each.String())
				}
				rel := &model1.ClusterExtensionRelatedCluster{
					ID:          uuid.New(),
					ExtensionID: model1.NewNullUUID(ext.ID),
					ClusterID:   ext.ClusterID,
					GuestID:     model1.NewNullUUID(each),
				}
				if err := tx.Create(rel).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
			}

			if err := appendAuditLog(tx, "enable", "ClusterExtension", ext.ID, ext.Name.String(), sess.UserID, sess.UserName, ext.OrganizationID.UUID); err != nil {
				panic(&ErrDBConnection{err})
			}

			if err := tx.Create(ext).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			extensionsToProvision = append(extensionsToProvision, ext)
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	mapd.EventEmitterCluster.Next(&mapd.Event{
		Type: mapd.EventTypeAdd,
		OID:  cluster.ID,
	})
	for _, each := range extensionsToProvision {
		mapd.EventEmitterClusterExtension.Next(&mapd.Event{
			Type: mapd.EventTypeAdd,
			OID:  each.ID,
		})
	}
	return convertClusterFromSQLModel(cluster), nil
}

// AttachCluster is the resolver for the AttachCluster field.
func (r *mutationResolver) AttachCluster(ctx context.Context, input model.AttachCluster) (*model.Cluster, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}

	cluster := &model1.Cluster{
		ID:             uuid.New(),
		Name:           input.Name,
		Type:           model1.ClusterTypeAttached,
		Endpoint:       input.Endpoint.String(),
		ProvisionPhase: model1.ClusterProvisionNamespace,
		Provisioned:    model1.ClusterProvisionKubernetes,
		Status:         model1.ClusterStatusInitializing,
		State:          model1.ClusterStateRunning,
	}
	if input.Description != nil {
		cluster.Description = *input.Description
	}
	if input.CaCert != nil {
		rootcas := x509.NewCertPool()
		if !rootcas.AppendCertsFromPEM([]byte(*input.CaCert)) {
			return nil, errInvalidRequest(ctx, "input.caCert", "invalid x509 certificate")
		}
		cluster.CACert = []byte(*input.CaCert)
	}
	if input.InsecureSkipTLSVerify != nil {
		cluster.InsecureTLS = *input.InsecureSkipTLSVerify
	}
	block, err := cryptoutil.NewAESGCMCipher(cluster.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	cluster.AuthToken = block.Encrypt([]byte(input.AuthToken))
	spec := &model1.ClusterExternalSpec{}
	if input.Spec != nil {
		for i, each := range input.Spec.RootCAs {
			rootcas := x509.NewCertPool()
			if !rootcas.AppendCertsFromPEM([]byte(each)) {
				return nil, errInvalidRequest(ctx, fmt.Sprintf("input.spec.%d", i), "invalid x509 certificate", each)
			}
			spec.RootCAs = append(spec.RootCAs, each)
		}
	}
	if cluster.Spec, err = json.Marshal(spec); err != nil {
		panic(err)
	}
	if len(input.ImageRegistries) == 0 {
		return nil, errInvalidRequest(ctx, "input.imageRegistries", "value cannot be empty")
	}
	var (
		defaultImageRegistryIndex int
		defaultImageRegistry      uuid.UUID
	)
	for i, each := range input.ImageRegistries {
		if each.Default != nil && *each.Default {
			if defaultImageRegistry == uuid.Nil {
				defaultImageRegistry = each.ID
				defaultImageRegistryIndex = i
			} else {
				each.Default = util.BoolPtr(false)
			}
		}
	}
	if defaultImageRegistry == uuid.Nil {
		input.ImageRegistries[0].Default = util.BoolPtr(true)
		defaultImageRegistry = input.ImageRegistries[0].ID
		defaultImageRegistryIndex = 0
	}
	if input.LoadBalancer != nil {
		cluster.LoadBalancerID = model1.NewNullUUID(*input.LoadBalancer)
	}
	var extensionsToProvision []*model1.ClusterExtension

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		registry := &model1.ImageRegistry{ID: defaultImageRegistry}
		if err := tx.Take(registry).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, fmt.Sprintf("input.imageRegistries.%d.id", defaultImageRegistryIndex), registry.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		if cluster.LoadBalancerID.Valid {
			loadBalancer := &model1.LoadBalancer{ID: cluster.LoadBalancerID.UUID}
			if err := tx.Take(loadBalancer).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, "input.loadBalancer", loadBalancer.ID.String())
				}
				panic(&ErrDBConnection{err})
			}
			if registry.OrganizationID.String() != loadBalancer.OrganizationID.String() {
				return errInvalidRequest(ctx, "input.loadBalancer", "different organization from default image registry", loadBalancer.ID.String())
			}
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectCluster, registry.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		cluster.OrganizationID = registry.OrganizationID

		var exists int64
		if err := tx.Model(&model1.Cluster{}).Where("`organization_id` = ? AND `name` = ?", cluster.OrganizationID, cluster.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", cluster.Name)
		}

		wc, err := k8s.NewKubernetesFromToken(cluster.Endpoint, block.Decrypt(cluster.AuthToken), rest.TLSClientConfig{
			Insecure: cluster.InsecureTLS,
			CAData:   cluster.CACert,
		})
		if err != nil {
			panic(err)
		}
		version, err := wc.Version()
		if err != nil {
			return errGatewayError(ctx, err)
		}
		v, err := abstract.MustParseVersion(version.GitVersion).SetPrerelease("")
		if err != nil {
			return errGatewayError(ctx, err)
		}
		v, err = v.SetMetadata("")
		if err != nil {
			return errGatewayError(ctx, err)
		}
		cluster.Version = v.String()

		if err := appendAuditLog(tx, "attach", "Cluster", cluster.ID, cluster.Name, sess.UserID, sess.UserName, cluster.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Create(cluster).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(cluster).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		for i, each := range input.ImageRegistries {
			if each.Default == nil || !*each.Default {
				r := &model1.ImageRegistry{ID: each.ID}
				if err := tx.Take(r).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return errObjectNotFound(ctx, fmt.Sprintf("input.imageRegistries.%d.id", i), r.ID.String())
					}
					panic(&ErrDBConnection{err})
				}
			}

			reg := &model1.ClusterImageRegistry{
				ID:              uuid.New(),
				ClusterID:       model1.NewNullUUID(cluster.ID),
				ImageRegistryID: model1.NewNullUUID(each.ID),
			}
			if each.Default != nil && *each.Default {
				reg.Default = true
			}
			if err := tx.Create(reg).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}

		for i, extension := range input.Extensions {
			ext := &model1.ClusterExtension{
				ID:             uuid.New(),
				Name:           model1.ClusterExtensionName(extension.Name.String()),
				State:          model1.ClusterExtensionStateEnabling,
				ClusterID:      model1.NewNullUUID(cluster.ID),
				OrganizationID: cluster.OrganizationID,
			}
			if extension.Spec != nil {
				var err error
				ext.Spec, err = json.Marshal(extension.Spec)
				if err != nil {
					panic(err)
				}
			}

			var exists int64
			if err := tx.Model(&model1.ClusterExtension{}).Where("`cluster_id` = ? AND `name` = ?", ext.ClusterID, ext.Name).Count(&exists).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if exists > 0 {
				return errObjectDuplicated(ctx, fmt.Sprintf("input.extensions.%d.name", i), ext.Name.String())
			}

			for j, each := range extension.Related {
				c := &model1.Cluster{ID: each}
				if err := tx.Take(c).Error; err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return errObjectNotFound(ctx, fmt.Sprintf("input.extensions.%d.related.%d", i, j), each.String())
					}
					panic(&ErrDBConnection{err})
				}
				if c.Provisioned < model1.ClusterProvisionFinish {
					return errInvalidRequest(ctx, fmt.Sprintf("input.extensions.%d.related.%d", i, j), "cluster is not ready", each.String())
				}
				rel := &model1.ClusterExtensionRelatedCluster{
					ID:          uuid.New(),
					ExtensionID: model1.NewNullUUID(ext.ID),
					ClusterID:   ext.ClusterID,
					GuestID:     model1.NewNullUUID(each),
				}
				if err := tx.Create(rel).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
			}

			if err := appendAuditLog(tx, "enable", "ClusterExtension", ext.ID, ext.Name.String(), sess.UserID, sess.UserName, ext.OrganizationID.UUID); err != nil {
				panic(&ErrDBConnection{err})
			}

			if err := tx.Create(ext).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			extensionsToProvision = append(extensionsToProvision, ext)
		}

		if input.Validate != nil && *input.Validate {
			if err := wc.PingContext(ctx); err != nil {
				return errGatewayError(ctx, err)
			}
		}

		if input.Dryrun != nil && *input.Dryrun {
			return errDryrun
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	mapd.EventEmitterCluster.Next(&mapd.Event{
		Type: mapd.EventTypeAdd,
		OID:  cluster.ID,
	})
	for _, each := range extensionsToProvision {
		mapd.EventEmitterClusterExtension.Next(&mapd.Event{
			Type: mapd.EventTypeAdd,
			OID:  each.ID,
		})
	}
	return convertClusterFromSQLModel(cluster), nil
}

// UpdateCluster is the resolver for the UpdateCluster field.
func (r *mutationResolver) UpdateCluster(ctx context.Context, input model.UpdateCluster) (*model.Cluster, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	cluster := &model1.Cluster{ID: input.ID}
	updates := make(map[string]interface{})
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	reprovision := false

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(cluster).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", cluster.ID.String())
			}
			if model1.IsDeadlockError(err) {
				return newGQLErrorf(ctx, CodeObjectIsBusy, "object is busy")
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectCluster, cluster.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if input.Recover != nil && *input.Recover && cluster.Status == model1.ClusterStatusInitializing && cluster.State == model1.ClusterStateFailed {
			if cluster.ProvisionPhase == model1.ClusterProvisionPostScripts {
				block, err := cryptoutil.NewAESGCMCipher(cluster.ID[:]) // AES-128
				if err != nil {
					panic(err)
				}
				wc, err := k8s.NewKubernetesFromToken(cluster.Endpoint, block.Decrypt(cluster.AuthToken), rest.TLSClientConfig{
					Insecure: cluster.InsecureTLS,
					CAData:   cluster.CACert,
				})
				if err != nil {
					panic(err)
				}
				tkgm := k8s.NewKappController(wc)
				if err := tkgm.DeletePackageInstall(defNamespace, defInitPackageName); err != nil && !k8s.IsErrObjectNotExists(err) {
					return errGatewayError(ctx, err)
				}
				if err := wc.DeleteJob(defNamespace, defInitJobName); err != nil && !k8s.IsErrObjectNotExists(err) {
					return errGatewayError(ctx, err)
				}
			}
			updates["state"] = model1.ClusterStateRunning
			reprovision = true
		}

		if cluster.Type == model1.ClusterTypeProvisioned {
			var reconfig int

			switch cluster.Class {
			case model1.ClusterClassTKGS:
				spec := &model1.ClusterTKGSSpec{}
				if err := json.Unmarshal(cluster.Spec, spec); err != nil {
					panic(err)
				}

				if input.Spec != nil && input.Spec.Tkgs != nil {
					if input.Spec.Tkgs.CtlPlaneCount != nil {
						if *input.Spec.Tkgs.CtlPlaneCount < spec.CtlPlaneCount {
							return errInvalidRequest(ctx, "input.spec.tkgs.ctlPlaneCount", "scale down is not allowed", *input.Spec.Tkgs.CtlPlaneCount)
						}

						if *input.Spec.Tkgs.CtlPlaneCount != spec.CtlPlaneCount {
							spec.CtlPlaneCount = *input.Spec.Tkgs.CtlPlaneCount
							reconfig++
						}
					}
					if input.Spec.Tkgs.CtlPlaneClass != nil && spec.CtlPlaneClass != *input.Spec.Tkgs.CtlPlaneClass {
						spec.CtlPlaneClass = *input.Spec.Tkgs.CtlPlaneClass
						reconfig++
					}
					if input.Spec.Tkgs.DataPlaneCount != nil {
						if *input.Spec.Tkgs.DataPlaneCount < 1 {
							return errInvalidRequest(ctx, "input.spec.tkgs.dataPlaneCount", "must be greater than zero", *input.Spec.Tkgs.DataPlaneCount)
						}

						if spec.DataPlaneCount != *input.Spec.Tkgs.DataPlaneCount {
							spec.DataPlaneCount = *input.Spec.Tkgs.DataPlaneCount
							reconfig++
						}
					}
					if input.Spec.Tkgs.DataPlaneClass != nil && spec.DataPlaneClass != *input.Spec.Tkgs.DataPlaneClass {
						spec.DataPlaneClass = *input.Spec.Tkgs.DataPlaneClass
						reconfig++
					}
				}
				if reconfig > 0 {
					b, err := json.Marshal(spec)
					if err != nil {
						panic(err)
					}
					updates["spec"] = string(b)
					reprovision = true

					cp := &model1.ClusterProvider{ID: cluster.ProviderID.UUID}
					if err := tx.Take(cp).Error; err != nil {
						panic(&ErrDBConnection{err})
					}
					cpSpec := &model1.ClusterProviderTKGSSpec{}
					if err := json.Unmarshal(cp.Spec, cpSpec); err != nil {
						panic(err)
					}
					block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
					if err != nil {
						panic(err)
					}
					client, err := k8s.NewTanzuKubernetesGridService(cpSpec.Endpoint, cpSpec.AuthUser, block.Decrypt(cpSpec.AuthPassword), rest.TLSClientConfig{
						Insecure: cpSpec.InsecureSkipTLSVerify,
						CAData:   cpSpec.CACert,
					})
					if err != nil {
						panic(err)
					}
					ctlPlaneClass, err := client.GetVirtualMachineClass(spec.CtlPlaneClass)
					if err != nil {
						if k8s.IsErrObjectNotExists(err) {
							return errObjectNotFound(ctx, "input.spec.tkgs.ctlPlaneClass", spec.CtlPlaneClass)
						}
						return errGatewayError(ctx, err)
					}
					dataPlaneClass, err := client.GetVirtualMachineClass(spec.DataPlaneClass)
					if err != nil {
						if k8s.IsErrObjectNotExists(err) {
							return errObjectNotFound(ctx, "input.spec.tkgs.dataPlaneClass", spec.DataPlaneClass)
						}
						return errGatewayError(ctx, err)
					}

					updateUsage := map[string]interface{}{
						"cpus":       spec.CtlPlaneCount*ctlPlaneClass.Spec.Hardware.Cpus + spec.DataPlaneCount*dataPlaneClass.Spec.Hardware.Cpus,
						"memory_mi":  spec.CtlPlaneCount*ctlPlaneClass.Spec.Hardware.Memory.ScaledValue(resource.Mega) + spec.DataPlaneCount*dataPlaneClass.Spec.Hardware.Memory.ScaledValue(resource.Mega),
						"storage_gi": spec.CtlPlaneCount*spec.CtlPlaneStorageCapGi + spec.DataPlaneCount*spec.DataPlaneStorageCapGi,
					}
					if err := tx.Model(&model1.OrganizationQuotaUsage{}).Where("`id` IN (SELECT `usage_id` FROM `cluster_quota_usage` WHERE `cluster_id` = ?)", cluster.ID).Updates(updateUsage).Error; err != nil {
						panic(&ErrDBConnection{err})
					}
					if !validateOrganizationQuotaUsage(tx, cluster.OrganizationID.UUID) {
						return errQuotaExceeded(ctx, "input.spec")
					}
				}
			case model1.ClusterClassTKGM:
				return errNoLongerSupport(ctx, "TKGm 1.X is no longer supported. Please consider to migrate your clusters to TKG 2.")
			case model1.ClusterClassTKG2S:
				spec := &model1.ClusterTKG2SSpec{}
				if err := json.Unmarshal(cluster.Spec, spec); err != nil {
					panic(err)
				}

				if input.Spec != nil && input.Spec.Tkg2s != nil {
					if input.Spec.Tkg2s.CtlPlaneReplicas != nil {
						if *input.Spec.Tkg2s.CtlPlaneReplicas < 1 {
							return errInvalidRequest(ctx, "input.spec.tkg2s.ctlPlaneReplicas", "scale down to zero is not allowed", *input.Spec.Tkgs.CtlPlaneCount)
						}

						if *input.Spec.Tkg2s.CtlPlaneReplicas != spec.CtlPlaneReplicas {
							spec.CtlPlaneReplicas = *input.Spec.Tkg2s.CtlPlaneReplicas
							reconfig++
						}
					}
					if input.Spec.Tkg2s.CtlPlaneVMClass != nil && spec.CtlPlaneVmClass != *input.Spec.Tkg2s.CtlPlaneVMClass {
						spec.CtlPlaneVmClass = *input.Spec.Tkg2s.CtlPlaneVMClass
						reconfig++
					}

					if input.Spec.Tkg2s.WorkerNodePool0Replicas != nil {
						if *input.Spec.Tkg2s.WorkerNodePool0Replicas < 1 {
							return errInvalidRequest(ctx, "input.spec.tkg2s.workerNodePool0Replicas", "must be greater than zero", *input.Spec.Tkg2s.WorkerNodePool0Replicas)
						}

						if spec.WorkerNodePool0Replicas != *input.Spec.Tkg2s.WorkerNodePool0Replicas {
							spec.WorkerNodePool0Replicas = *input.Spec.Tkg2s.WorkerNodePool0Replicas
							reconfig++
						}
					}
					if input.Spec.Tkg2s.WorkerNodePool0Zone != nil {
						spec.WorkerNodePool0Zone = *input.Spec.Tkg2s.WorkerNodePool0Zone
						reconfig++
					}
					if input.Spec.Tkg2s.WorkerNodePool1Replicas != nil {
						if *input.Spec.Tkg2s.WorkerNodePool1Replicas < 0 {
							*input.Spec.Tkg2s.WorkerNodePool1Replicas = 0
						}

						if spec.WorkerNodePool1Replicas != *input.Spec.Tkg2s.WorkerNodePool1Replicas {
							spec.WorkerNodePool1Replicas = *input.Spec.Tkg2s.WorkerNodePool1Replicas
							reconfig++
						}
					}
					if input.Spec.Tkg2s.WorkerNodePool1Zone != nil {
						spec.WorkerNodePool1Zone = *input.Spec.Tkg2s.WorkerNodePool1Zone
						reconfig++
					}
					if input.Spec.Tkg2s.WorkerNodePool2Replicas != nil {
						if *input.Spec.Tkg2s.WorkerNodePool2Replicas < 0 {
							*input.Spec.Tkg2s.WorkerNodePool2Replicas = 0
						}

						if spec.WorkerNodePool2Replicas != *input.Spec.Tkg2s.WorkerNodePool2Replicas {
							spec.WorkerNodePool2Replicas = *input.Spec.Tkg2s.WorkerNodePool2Replicas
							reconfig++
						}
					}
					if input.Spec.Tkg2s.WorkerNodePool2Zone != nil {
						spec.WorkerNodePool2Zone = *input.Spec.Tkg2s.WorkerNodePool2Zone
						reconfig++
					}
					if input.Spec.Tkg2s.WorkerVMClass != nil && spec.WorkerVmClass != *input.Spec.Tkg2s.WorkerVMClass {
						spec.WorkerVmClass = *input.Spec.Tkg2s.WorkerVMClass
						reconfig++
					}

				}
				if reconfig > 0 {
					b, err := json.Marshal(spec)
					if err != nil {
						panic(err)
					}
					updates["spec"] = string(b)
					reprovision = true

					cp := &model1.ClusterProvider{ID: cluster.ProviderID.UUID}
					if err := tx.Take(cp).Error; err != nil {
						panic(&ErrDBConnection{err})
					}
					cpSpec := &model1.ClusterProviderTKGSSpec{}
					if err := json.Unmarshal(cp.Spec, cpSpec); err != nil {
						panic(err)
					}
					block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
					if err != nil {
						panic(err)
					}
					client, err := k8s.NewTanzuKubernetesGridService(cpSpec.Endpoint, cpSpec.AuthUser, block.Decrypt(cpSpec.AuthPassword), rest.TLSClientConfig{
						Insecure: cpSpec.InsecureSkipTLSVerify,
						CAData:   cpSpec.CACert,
					})
					if err != nil {
						panic(err)
					}
					ctlPlaneVmClass, err := client.GetVirtualMachineClass(spec.CtlPlaneVmClass)
					if err != nil {
						if k8s.IsErrObjectNotExists(err) {
							return errObjectNotFound(ctx, "input.spec.tkg2s.ctlPlaneVmClass", spec.CtlPlaneVmClass)
						}
						return errGatewayError(ctx, err)
					}
					workerVmClass, err := client.GetVirtualMachineClass(spec.WorkerVmClass)
					if err != nil {
						if k8s.IsErrObjectNotExists(err) {
							return errObjectNotFound(ctx, "input.spec.tkg2s.workerVmClass", spec.WorkerVmClass)
						}
						return errGatewayError(ctx, err)
					}

					updateUsage := map[string]interface{}{
						"cpus":       spec.CtlPlaneReplicas*ctlPlaneVmClass.Spec.Hardware.Cpus + (spec.WorkerNodePool0Replicas+spec.WorkerNodePool1Replicas+spec.WorkerNodePool2Replicas)*workerVmClass.Spec.Hardware.Cpus,
						"memory_mi":  spec.CtlPlaneReplicas*ctlPlaneVmClass.Spec.Hardware.Memory.ScaledValue(resource.Mega) + (spec.WorkerNodePool0Replicas+spec.WorkerNodePool1Replicas+spec.WorkerNodePool2Replicas)*workerVmClass.Spec.Hardware.Memory.ScaledValue(resource.Mega),
						"storage_gi": spec.CtlPlaneReplicas*spec.CtlPlaneEtcdVolumeSizeGi + (spec.WorkerNodePool0Replicas+spec.WorkerNodePool1Replicas+spec.WorkerNodePool2Replicas)*spec.WorkerContainerdVolumeSizeGi,
					}
					if err := tx.Model(&model1.OrganizationQuotaUsage{}).Where("`id` IN (SELECT `usage_id` FROM `cluster_quota_usage` WHERE `cluster_id` = ?)", cluster.ID).Updates(updateUsage).Error; err != nil {
						panic(&ErrDBConnection{err})
					}
					if !validateOrganizationQuotaUsage(tx, cluster.OrganizationID.UUID) {
						return errQuotaExceeded(ctx, "input.spec")
					}
				}
			case model1.ClusterClassTKG2M:
				spec := &model1.ClusterTKG2MSpec{}
				if err := json.Unmarshal(cluster.Spec, spec); err != nil {
					panic(err)
				}

				if input.Spec != nil && input.Spec.Tkg2m != nil {
					reconfig++
					if input.Spec.Tkg2m.CtlPlaneReplicas != nil {
						if *input.Spec.Tkg2m.CtlPlaneReplicas < 1 {
							return errInvalidRequest(ctx, "input.spec.tkg2m.ctlPlaneReplicas", "must be greater than zero", *input.Spec.Tkg2m.CtlPlaneReplicas)
						}

						if *input.Spec.Tkg2m.CtlPlaneReplicas != spec.CtlPlaneReplicas {
							spec.CtlPlaneReplicas = *input.Spec.Tkg2m.CtlPlaneReplicas
						}
					}
					if input.Spec.Tkg2m.CtlPlaneNumCpus != nil {
						if *input.Spec.Tkg2m.CtlPlaneNumCpus < 1 {
							return errInvalidRequest(ctx, "input.spec.tkg2m.ctlPlaneNumCpus", "must be greater than zero", *input.Spec.Tkg2m.CtlPlaneNumCpus)
						}

						if *input.Spec.Tkg2m.CtlPlaneNumCpus != spec.CtlPlaneNumCpus {
							spec.CtlPlaneNumCpus = *input.Spec.Tkg2m.CtlPlaneNumCpus
						}
					}
					if input.Spec.Tkg2m.CtlPlaneMemMiB != nil {
						if *input.Spec.Tkg2m.CtlPlaneMemMiB < 512 {
							return errInvalidRequest(ctx, "input.spec.tkg2m.ctlPlaneMemMiB", "must be greater than 512", *input.Spec.Tkg2m.CtlPlaneMemMiB)
						}

						if *input.Spec.Tkg2m.CtlPlaneMemMiB != spec.CtlPlaneMemMiB {
							spec.CtlPlaneMemMiB = *input.Spec.Tkg2m.CtlPlaneMemMiB
						}
					}
					if input.Spec.Tkg2m.CtlPlaneDiskGiB != nil {
						if *input.Spec.Tkg2m.CtlPlaneDiskGiB < 8 {
							return errInvalidRequest(ctx, "input.spec.tkg2m.ctlPlaneDiskGiB", "must be greater than 8", *input.Spec.Tkg2m.CtlPlaneDiskGiB)
						}

						if *input.Spec.Tkg2m.CtlPlaneDiskGiB != spec.CtlPlaneDiskGiB {
							spec.CtlPlaneDiskGiB = *input.Spec.Tkg2m.CtlPlaneDiskGiB
						}
					}
					if input.Spec.Tkg2m.WorkerNodePool0Replicas != nil {
						if *input.Spec.Tkg2m.WorkerNodePool0Replicas < 1 {
							return errInvalidRequest(ctx, "input.spec.tkg2m.workerNodePool0Replicas", "must be greater than zero", *input.Spec.Tkg2m.WorkerNodePool0Replicas)
						}

						if spec.WorkerNodePool0Replicas != *input.Spec.Tkg2m.WorkerNodePool0Replicas {
							spec.WorkerNodePool0Replicas = *input.Spec.Tkg2m.WorkerNodePool0Replicas
							reconfig++
						}
					}
					if input.Spec.Tkg2m.WorkerNodePool0Zone != nil {
						spec.WorkerNodePool0Zone = *input.Spec.Tkg2m.WorkerNodePool0Zone
						reconfig++
					}
					if input.Spec.Tkg2m.WorkerNodePool1Replicas != nil {
						if *input.Spec.Tkg2m.WorkerNodePool1Replicas < 0 {
							*input.Spec.Tkg2m.WorkerNodePool1Replicas = 0
						}

						if spec.WorkerNodePool1Replicas != *input.Spec.Tkg2m.WorkerNodePool1Replicas {
							spec.WorkerNodePool1Replicas = *input.Spec.Tkg2m.WorkerNodePool1Replicas
							reconfig++
						}
					}
					if input.Spec.Tkg2m.WorkerNodePool1Zone != nil {
						spec.WorkerNodePool1Zone = *input.Spec.Tkg2m.WorkerNodePool1Zone
						reconfig++
					}
					if input.Spec.Tkg2m.WorkerNodePool2Replicas != nil {
						if *input.Spec.Tkg2m.WorkerNodePool2Replicas < 0 {
							*input.Spec.Tkg2m.WorkerNodePool2Replicas = 0
						}

						if spec.WorkerNodePool2Replicas != *input.Spec.Tkg2m.WorkerNodePool2Replicas {
							spec.WorkerNodePool2Replicas = *input.Spec.Tkg2m.WorkerNodePool2Replicas
							reconfig++
						}
					}
					if input.Spec.Tkg2m.WorkerNodePool2Zone != nil {
						spec.WorkerNodePool2Zone = *input.Spec.Tkg2m.WorkerNodePool2Zone
						reconfig++
					}
					if input.Spec.Tkg2m.WorkerNumCpus != nil {
						if *input.Spec.Tkg2m.WorkerNumCpus < 1 {
							return errInvalidRequest(ctx, "input.spec.tkg2m.workerNumCpus", "must be greater than zero", *input.Spec.Tkg2m.WorkerNumCpus)
						}

						if *input.Spec.Tkg2m.WorkerNumCpus != spec.WorkerNumCpus {
							spec.WorkerNumCpus = *input.Spec.Tkg2m.WorkerNumCpus
						}
					}
					if input.Spec.Tkg2m.WorkerMemMiB != nil {
						if *input.Spec.Tkg2m.WorkerMemMiB < 512 {
							return errInvalidRequest(ctx, "input.spec.tkg2m.workerMemMiB", "must be greater than 512", *input.Spec.Tkg2m.WorkerMemMiB)
						}

						if *input.Spec.Tkg2m.WorkerMemMiB != spec.WorkerMemMiB {
							spec.WorkerMemMiB = *input.Spec.Tkg2m.WorkerMemMiB
						}
					}
					if input.Spec.Tkg2m.WorkerDiskGiB != nil {
						if *input.Spec.Tkg2m.WorkerDiskGiB < 8 {
							return errInvalidRequest(ctx, "input.spec.tkg2m.workerDiskGiB", "must be greater than 8", *input.Spec.Tkg2m.WorkerDiskGiB)
						}

						if *input.Spec.Tkg2m.WorkerDiskGiB != spec.WorkerDiskGiB {
							spec.WorkerDiskGiB = *input.Spec.Tkg2m.WorkerDiskGiB
						}
					}
				}
				if reconfig > 0 {
					b, err := json.Marshal(spec)
					if err != nil {
						panic(err)
					}
					updates["spec"] = string(b)
					reprovision = true

					updateUsage := map[string]interface{}{
						"cpus":       spec.CtlPlaneReplicas*spec.CtlPlaneNumCpus + (spec.WorkerNodePool0Replicas+spec.WorkerNodePool1Replicas+spec.WorkerNodePool2Replicas)*spec.WorkerNumCpus,
						"memory_mi":  spec.CtlPlaneReplicas*spec.CtlPlaneMemMiB + (spec.WorkerNodePool0Replicas+spec.WorkerNodePool1Replicas+spec.WorkerNodePool2Replicas)*spec.WorkerMemMiB,
						"storage_gi": spec.CtlPlaneReplicas*spec.CtlPlaneDiskGiB + (spec.WorkerNodePool0Replicas+spec.WorkerNodePool1Replicas+spec.WorkerNodePool2Replicas)*spec.WorkerDiskGiB,
					}
					if err := tx.Model(&model1.OrganizationQuotaUsage{}).Where("`id` IN (SELECT `usage_id` FROM `cluster_quota_usage` WHERE `cluster_id` = ?)", cluster.ID).Updates(updateUsage).Error; err != nil {
						panic(&ErrDBConnection{err})
					}
					if !validateOrganizationQuotaUsage(tx, cluster.OrganizationID.UUID) {
						return errQuotaExceeded(ctx, "input.spec")
					}
				}
			}
			if input.Version != nil {
				switch cluster.Class {
				case model1.ClusterClassTKGS, model1.ClusterClassTKGM:
					if reconfig > 0 {
						return errInvalidRequest(ctx, "input.version", "cannot update version on cluster reconfig", input.Version.String())
					}
				}
				oldVersion := abstract.MustParseVersion(cluster.Version)
				if input.Version.LessThan(oldVersion) {
					return errInvalidRequest(ctx, "input.version", fmt.Sprintf("must be greater than %v", oldVersion), input.Version.String())
				}
				if !oldVersion.Equal(input.Version) {
					updates["version"] = input.Version.String()
					reprovision = true
				}
			}
		}

		if reprovision {
			updates["provisioned"] = model1.ClusterProvisionPrepare
			updates["provision_phase"] = model1.ClusterProvisionKubernetes
			updates["status"] = model1.ClusterStatusInitializing
		}

		if input.Maintenance != nil {
			if cluster.Provisioned < model1.ClusterProvisionFinish {
				if *input.Maintenance {
					return errInvalidRequest(ctx, "input.maintenance", "cluster status is locked", *input.Maintenance)
				}
			} else {
				if *input.Maintenance {
					updates["status"] = model1.ClusterStatusMaintenance
				} else {
					updates["status"] = model1.ClusterStatusPending
				}
			}
		}

		if err := appendAuditLog(tx, "update", "Cluster", cluster.ID, cluster.Name, sess.UserID, sess.UserName, cluster.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(updates) > 0 {
			if err := tx.Model(cluster).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(cluster).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", cluster.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	if reprovision {
		mapd.EventEmitterCluster.Next(&mapd.Event{
			Type: mapd.EventTypeModify,
			OID:  cluster.ID,
		})
	}
	return convertClusterFromSQLModel(cluster), nil
}

// DeleteCluster is the resolver for the DeleteCluster field.
func (r *mutationResolver) DeleteCluster(ctx context.Context, input model.DeleteCluster) (*model.Cluster, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	force := input.Force != nil && *input.Force
	if force && !sess.Admin {
		return nil, errInsufficientPrivileges(ctx)
	}

	cluster := &model1.Cluster{ID: input.ID}
	var (
		removedClusterExtension      []model1.ClusterExtension
		removedClusterBackupSchedule []model1.ClusterBackupSchedule
	)
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(cluster).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", cluster.ID.String())
			}
			if model1.IsDeadlockError(err) {
				return newGQLErrorf(ctx, CodeObjectIsBusy, "object is busy")
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermModify, model1.PolicyObjectCluster, cluster.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.Instance{}).Where("`cluster_id` = ?", cluster.ID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectIsBusy(ctx, "input.id", "Instance", nil, "", input.ID.String())
		}

		if cluster.Type == model1.ClusterTypeProvisioned && !force {
			cp := &model1.ClusterProvider{ID: cluster.ProviderID.UUID}
			if err := tx.Take(cp).Error; err != nil {
				panic(&ErrDBConnection{err})
			}

			switch cp.Class {
			case model1.ClusterClassTKGS:
				cpSpec := &model1.ClusterProviderTKGSSpec{}
				if err := json.Unmarshal(cp.Spec, cpSpec); err != nil {
					panic(err)
				}
				block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
				if err != nil {
					panic(err)
				}
				client, err := k8s.NewTanzuKubernetesGridService(cpSpec.Endpoint, cpSpec.AuthUser, block.Decrypt(cpSpec.AuthPassword), rest.TLSClientConfig{
					Insecure: cpSpec.InsecureSkipTLSVerify,
					CAData:   cpSpec.CACert,
				})
				if err != nil {
					return errGatewayError(ctx, err)
				}

				spec := &model1.ClusterTKGSSpec{}
				if err := json.Unmarshal(cluster.Spec, spec); err != nil {
					panic(err)
				}

				if err := client.DeleteTanzuKubernetesCluster(spec.Namespace, cluster.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
					return errGatewayError(ctx, err)
				}
			case model1.ClusterClassTKGM:
				return errNoLongerSupport(ctx, "TKGm 1.X is no longer supported. Please consider to migrate your clusters to TKG 2.")
			case model1.ClusterClassTKG2S:
				cpSpec := &model1.ClusterProviderTKG2SSpec{}
				if err := json.Unmarshal(cp.Spec, cpSpec); err != nil {
					panic(err)
				}
				block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
				if err != nil {
					panic(err)
				}
				sc, err := tanzuv2.NewVSphereWithTanzu(cpSpec.Endpoint, cpSpec.AuthUser, block.Decrypt(cpSpec.AuthPassword), rest.TLSClientConfig{
					Insecure: cpSpec.InsecureSkipTLSVerify,
					CAData:   cpSpec.CACert,
				})
				if err != nil {
					return errGatewayError(ctx, err)
				}

				spec := &model1.ClusterTKG2SSpec{}
				if err := json.Unmarshal(cluster.Spec, spec); err != nil {
					panic(err)
				}

				if err := sc.DeleteCluster(spec.Namespace, cluster.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
					return errGatewayError(ctx, err)
				}
			case model1.ClusterClassTKG2M:
				cpSpec := &model1.ClusterProviderTKG2MSpec{}
				if err := json.Unmarshal(cp.Spec, cpSpec); err != nil {
					panic(err)
				}
				block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
				if err != nil {
					panic(err)
				}
				mgmt, err := tanzuv2.NewTanzuKubernetesGrid(block.Decrypt(cpSpec.KubeConfig))
				if err != nil {
					return errGatewayError(ctx, err)
				}

				spec := &model1.ClusterTKG2MSpec{}
				if err := json.Unmarshal(cluster.Spec, spec); err != nil {
					panic(err)
				}

				if err := mgmt.DeleteCluster(spec.Namespace, cluster.Name); err != nil && !k8s.IsErrObjectNotExists(err) {
					return errGatewayError(ctx, err)
				}
			}
		}

		// notify shared clusters in the same supervisor namespace to update octant configuration.
		rupdates := make(map[string]interface{})
		rupdates["provisioned"] = model1.ClusterProvisionValidation
		rupdates["provision_phase"] = model1.ClusterProvisionPostScripts
		if err := tx.Model(&model1.Cluster{}).Where("`id` IN (SELECT `cluster_id` FROM `cluster_extension_related_cluster` WHERE `guest_id` = ?) AND `id` != ? AND `provisioned` > ?", cluster.ID, cluster.ID, model1.ClusterProvisionValidation).Updates(rupdates).Error; err != nil {
			if model1.IsDeadlockError(err) {
				return newGQLErrorf(ctx, CodeObjectIsBusy, "object is busy")
			}
			panic(&ErrDBConnection{err})
		}
		if err := tx.Where("`guest_id` = ?", cluster.ID).Delete(&model1.ClusterExtensionRelatedCluster{}).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Where("`cluster_id` = ?", cluster.ID).Find(&removedClusterExtension).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Where("`cluster_id` = ?", cluster.ID).Delete(&model1.ClusterExtension{}).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Where("`cluster_id` = ?", cluster.ID).Find(&removedClusterBackupSchedule).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Where("`cluster_id` = ?", cluster.ID).Delete(&model1.ClusterBackupSchedule{}).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Where("`id` IN (SELECT `usage_id` FROM `cluster_quota_usage` WHERE `cluster_id` = ?)", cluster.ID).Delete(&model1.OrganizationQuotaUsage{}).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := appendAuditLog(tx, "delete", "Cluster", cluster.ID, cluster.Name, sess.UserID, sess.UserName, cluster.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Where("`cluster_id` = ?", cluster.ID).Delete(&model1.ClusterImageRegistry{}).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Delete(cluster).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	for _, each := range removedClusterExtension {
		mapd.EventEmitterClusterExtension.Next(&mapd.Event{
			Type: mapd.EventTypeRemove,
			OID:  each.ID,
		})
	}
	for _, each := range removedClusterBackupSchedule {
		mapd.EventEmitterClusterBackupSchedule.Next(&mapd.Event{
			Type: mapd.EventTypeRemove,
			OID:  each.ID,
		})
	}
	mapd.EventEmitterCluster.Next(&mapd.Event{
		Type: mapd.EventTypeRemove,
		OID:  cluster.ID,
	})
	return convertClusterFromSQLModel(cluster), nil
}

// Cluster is the resolver for the Cluster field.
func (r *queryResolver) Cluster(ctx context.Context, id uuid.UUID) (*model.Cluster, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectCluster)
	}

	cluster := &model1.Cluster{ID: id}
	if err := db.Take(cluster).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", id.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertClusterFromSQLModel(cluster), nil
}

// Clusters is the resolver for the Clusters field.
func (r *queryResolver) Clusters(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.Clusters, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.Clusters{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.Cluster{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectCluster)
	}
	db = finalizeClusterQueryFilter(db, obj.Filters...)

	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// Cluster returns graphql1.ClusterResolver implementation.
func (r *Resolver) Cluster() graphql1.ClusterResolver { return &clusterResolver{r} }

// ClusterImageRegistries returns graphql1.ClusterImageRegistriesResolver implementation.
func (r *Resolver) ClusterImageRegistries() graphql1.ClusterImageRegistriesResolver {
	return &clusterImageRegistriesResolver{r}
}

// ClusterImageRegistry returns graphql1.ClusterImageRegistryResolver implementation.
func (r *Resolver) ClusterImageRegistry() graphql1.ClusterImageRegistryResolver {
	return &clusterImageRegistryResolver{r}
}

// Clusters returns graphql1.ClustersResolver implementation.
func (r *Resolver) Clusters() graphql1.ClustersResolver { return &clustersResolver{r} }

type clusterResolver struct{ *Resolver }
type clusterImageRegistriesResolver struct{ *Resolver }
type clusterImageRegistryResolver struct{ *Resolver }
type clustersResolver struct{ *Resolver }
