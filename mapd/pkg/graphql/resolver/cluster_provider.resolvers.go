package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	semver "github.com/Masterminds/semver/v3"
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
	runv1alpha1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s/apis/run.tanzu/v1alpha1"
	tanzuv2 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/tanzu/v2"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// TLSCertChain is the resolver for the tlsCertChain field.
func (r *clusterProviderResolver) TLSCertChain(ctx context.Context, obj *model.ClusterProvider) ([]*model.X509Certificate, error) {
	var (
		uri                   string
		caCert                []byte
		insecureSkipTLSVerify bool
	)
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		cp := &model1.ClusterProvider{ID: obj.ID}
		if err := tx.Take(cp).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		switch cp.Class {
		case model1.ClusterClassTKGS:
			v := &model1.ClusterProviderTKGSSpec{}
			if err := json.Unmarshal(cp.Spec, v); err != nil {
				panic(err)
			}
			uri, caCert, insecureSkipTLSVerify = v.Endpoint, v.CACert, v.InsecureSkipTLSVerify
		case model1.ClusterClassTKGM:
			v := &model1.ClusterProviderTKGMSpec{}
			if err := json.Unmarshal(cp.Spec, v); err != nil {
				panic(err)
			}

			block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
			if err != nil {
				panic(err)
			}
			cli, err := k8s.NewKubernetesFromKubeConfig(block.Decrypt(v.KubeConfig))
			if err != nil {
				panic(err)
			}

			if cli.Config.CAData != nil && len(cli.Config.CAData) > 0 {
				caCert = cli.Config.CAData
			} else if cli.Config.CAFile != "" {
				p, err := os.ReadFile(cli.Config.CAFile)
				if err == nil {
					caCert = p
				}
			}
			uri, insecureSkipTLSVerify = cli.Config.Host, cli.Config.Insecure
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}

	zctx, zcancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
	defer zcancel()

	if uri == "" {
		return nil, nil
	}
	chain, err := httputil.EZGetTLSCertificateChainContext(zctx, uri, insecureSkipTLSVerify, append(caCert, '\n'))
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	var certs []*model.X509Certificate
	for _, each := range chain {
		certs = append(certs, convertX509Certificate(each))
	}
	return certs, nil
}

// ImageRegistries is the resolver for the imageRegistries field.
func (r *clusterProviderResolver) ImageRegistries(ctx context.Context, obj *model.ClusterProvider, page *model.QueryPage, filters []*model.QueryFilter) (*model.ClusterProviderImageRegistries, error) {
	result := &model.ClusterProviderImageRegistries{Provider: obj.ID, Page: &model.PageInfo{}}
	if page.Limit != nil {
		result.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		result.Page.Number = *page.Number
	}
	for _, each := range filters {
		result.Filters = append(result.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.ClusterProviderImageRegistry{}).Where("`provider_id` = ?", obj.ID)
	db = finalizeClusterProviderImageRegistryQueryFilter(db, result.Filters...)
	if err := db.Count(&result.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(result.Page, result.TotalCount)
	return result, nil
}

// LoadBalancer is the resolver for the loadBalancer field.
func (r *clusterProviderResolver) LoadBalancer(ctx context.Context, obj *model.ClusterProvider) (*model.LoadBalancer, error) {
	if obj.LoadBalancer == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).LoadBalancer(WithGuardSilenced(ctx), obj.LoadBalancer.ID)
}

// Organization is the resolver for the organization field.
func (r *clusterProviderResolver) Organization(ctx context.Context, obj *model.ClusterProvider) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *clusterProviderImageRegistriesResolver) Edge(ctx context.Context, obj *model.ClusterProviderImageRegistries) (*model.ClusterProviderImageRegistryNodes, error) {
	nodes := &model.ClusterProviderImageRegistryNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Where("`provider_id` = ?", obj.Provider).Order("`id` ASC")
		db = finalizeClusterProviderImageRegistryQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var registries []model1.ClusterProviderImageRegistry
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&registries).Error; err != nil {
				return err
			}
			for _, registry := range registries {
				nodes.Nodes = append(nodes.Nodes, convertClusterProviderImageRegistryFromSQLModel(&registry))
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
func (r *clusterProviderImageRegistryResolver) ImageRegistry(ctx context.Context, obj *model.ClusterProviderImageRegistry) (*model.ImageRegistry, error) {
	return r.Resolver.Query().(*queryResolver).ImageRegistry(WithGuardSilenced(ctx), obj.ID)
}

// Edge is the resolver for the edge field.
func (r *clusterProvidersResolver) Edge(ctx context.Context, obj *model.ClusterProviders) (*model.ClusterProviderNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.ClusterProviderNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterProvider)
		db = finalizeClusterProviderQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var cps []model1.ClusterProvider
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&cps).Error; err != nil {
				return err
			}
			for _, cp := range cps {
				spec := interface{}(nil)
				switch cp.Class {
				case model1.ClusterClassTKGS:
					v := &model1.ClusterProviderTKGSSpec{}
					if err := json.Unmarshal(cp.Spec, v); err != nil {
						panic(err)
					}
					spec = v
				case model1.ClusterClassTKGM:
					v := &model1.ClusterProviderTKGMSpec{}
					if err := json.Unmarshal(cp.Spec, v); err != nil {
						panic(err)
					}
					spec = v
				case model1.ClusterClassTKG2S:
					v := &model1.ClusterProviderTKG2SSpec{}
					if err := json.Unmarshal(cp.Spec, v); err != nil {
						panic(err)
					}
					spec = v
				case model1.ClusterClassTKG2M:
					v := &model1.ClusterProviderTKG2MSpec{}
					if err := json.Unmarshal(cp.Spec, v); err != nil {
						panic(err)
					}
					spec = v
				}
				nodes.Nodes = append(nodes.Nodes, convertClusterProviderFromSQLModel(&cp, spec))
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

// CreateTKGSClusterProvider is the resolver for the CreateTKGSClusterProvider field.
func (r *mutationResolver) CreateTKGSClusterProvider(ctx context.Context, input model.CreateTKGSClusterProvider) (*model.ClusterProvider, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}
	for i, each := range input.Spec.TrustedCAs {
		if !validateName(each.Name) {
			return nil, errInvalidRequest(ctx, fmt.Sprintf("input.spec.trustedCAs.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
		}
		rootcas := x509.NewCertPool()
		if !rootcas.AppendCertsFromPEM([]byte(each.Cert)) {
			return nil, errInvalidRequest(ctx, fmt.Sprintf("input.spec.trustedCAs.%d.cert", i), "invalid x509 certificate", each.Name)
		}
	}
	cp := &model1.ClusterProvider{
		ID:    uuid.New(),
		Name:  input.Name,
		Class: model1.ClusterClassTKGS,
	}
	if input.Description != nil {
		cp.Description = *input.Description
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
		cp.LoadBalancerID = model1.NewNullUUID(*input.LoadBalancer)
	}
	if input.Spec.DefaultNamespace == "" {
		return nil, errInvalidRequest(ctx, "input.spec.defaultNamespace", "value cannot be empty")
	}
	spec := &model1.ClusterProviderTKGSSpec{
		Endpoint:  input.Spec.Endpoint.String(),
		AuthUser:  input.Spec.AuthUser,
		DefaultNS: input.Spec.DefaultNamespace,
	}
	if input.Spec.CaCert != nil {
		spec.CACert = []byte(*input.Spec.CaCert)
	}
	if input.Spec.InsecureSkipTLSVerify != nil {
		spec.InsecureSkipTLSVerify = *input.Spec.InsecureSkipTLSVerify
	}
	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	spec.AuthPassword = block.Encrypt([]byte(input.Spec.AuthPassword))
	cp.Spec, err = json.Marshal(spec)
	if err != nil {
		panic(err)
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		registry := &model1.ImageRegistry{ID: defaultImageRegistry}
		if err := tx.Take(registry).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, fmt.Sprintf("input.imageRegistries.%d.id", defaultImageRegistryIndex), registry.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		if cp.LoadBalancerID.Valid {
			loadBalancer := &model1.LoadBalancer{ID: cp.LoadBalancerID.UUID}
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

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectClusterProvider, registry.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		cp.OrganizationID = registry.OrganizationID

		var exists int64
		if err := tx.Model(&model1.ClusterProvider{}).Where("`organization_id` = ? AND `name` = ?", cp.OrganizationID, cp.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", cp.Name)
		}

		if input.Validate != nil && *input.Validate {
			client, err := k8s.NewTanzuKubernetesGridService(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
				Insecure: spec.InsecureSkipTLSVerify,
				CAData:   spec.CACert,
			})
			if err != nil {
				panic(err)
			}

			if err := client.PingContext(ctx); err != nil {
				return errGatewayError(ctx, err)
			}
		}

		if err := tx.Create(cp).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(cp).Error; err != nil {
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

			reg := &model1.ClusterProviderImageRegistry{
				ID:              uuid.New(),
				ProviderID:      model1.NewNullUUID(cp.ID),
				ImageRegistryID: model1.NewNullUUID(each.ID),
			}
			if each.Default != nil && *each.Default {
				reg.Default = true
			}
			if err := tx.Create(reg).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}

		if input.Dryrun != nil && *input.Dryrun {
			return errDryrun
		}

		if err := appendAuditLog(tx, "create", "ClusterProvider", cp.ID, cp.Name, sess.UserID, sess.UserName, cp.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(input.Spec.TrustedCAs) > 0 {
			client, err := k8s.NewTanzuKubernetesGridService(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
				Insecure: spec.InsecureSkipTLSVerify,
				CAData:   spec.CACert,
			})
			if err != nil {
				panic(err)
			}

			cfg, err := client.GetTkgServiceConfiguration(defTkgServiceConfiguration)
			if err != nil {
				return errGatewayError(ctx, err)
			}
			if cfg.Spec.Trust == nil {
				cfg.Spec.Trust = &runv1alpha1.TrustConfiguration{}
			}
			cert := make(map[string]string)
			for _, each := range cfg.Spec.Trust.AdditionalTrustedCAs {
				cert[each.Name] = each.Data
			}
			for _, each := range input.Spec.TrustedCAs {
				cert[each.Name] = base64.StdEncoding.EncodeToString([]byte(each.Cert))
			}
			cfg.Spec.Trust.AdditionalTrustedCAs = cfg.Spec.Trust.AdditionalTrustedCAs[:0]
			for name, data := range cert {
				cfg.Spec.Trust.AdditionalTrustedCAs = append(cfg.Spec.Trust.AdditionalTrustedCAs, runv1alpha1.TLSCertificate{
					Name: name,
					Data: data,
				})
			}
			if _, err := patchTkgServiceConfigurationAdditionalTrustedCAs(client, defTkgServiceConfiguration, cfg.Spec.Trust.AdditionalTrustedCAs); err != nil {
				return errGatewayError(ctx, err)
			}
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	result := convertClusterProviderFromSQLModel(cp, spec)
	if input.Dryrun != nil && *input.Dryrun {
		fields := graphql.CollectFieldsCtx(ctx, []string{"ClusterProvider"})
	LOOP:
		for _, field := range fields {
			if field.Name == "spec" {
				oc := graphql.GetOperationContext(ctx)
				fs := graphql.CollectFields(oc, field.Selections, []string{"TKGSClusterProviderSpec", "ClusterProviderSpec"})
				for _, f := range fs {
					if f.Name == "namespaces" {
						ns, err := getTKGSClusterProviderNamespaces(ctx, cp.ID, spec.Endpoint, spec.InsecureSkipTLSVerify, spec.CACert, spec.AuthUser, block.Decrypt(spec.AuthPassword), []*model.QueryFilter{})
						if err != nil {
							return nil, err
						}
						result.Spec.(*model.TKGSClusterProviderSpec).Namespaces = ns
						break LOOP
					}
				}
				break
			}
		}
	}
	return result, nil
}

// CreateTKGMClusterProvider is the resolver for the CreateTKGMClusterProvider field.
func (r *mutationResolver) CreateTKGMClusterProvider(ctx context.Context, input model.CreateTKGMClusterProvider) (*model.ClusterProvider, error) {
	return nil, errNoLongerSupport(ctx, "TKGm 1.X is no longer supported. Please consider to migrate your clusters to TKG 2.")
}

// CreateTKG2SClusterProvider is the resolver for the CreateTKG2SClusterProvider field.
func (r *mutationResolver) CreateTKG2SClusterProvider(ctx context.Context, input model.CreateTKG2SClusterProvider) (*model.ClusterProvider, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}
	for i, each := range input.Spec.TrustedCAs {
		if !validateName(each.Name) {
			return nil, errInvalidRequest(ctx, fmt.Sprintf("input.spec.trustedCAs.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
		}
		rootcas := x509.NewCertPool()
		if !rootcas.AppendCertsFromPEM([]byte(each.Cert)) {
			return nil, errInvalidRequest(ctx, fmt.Sprintf("input.spec.trustedCAs.%d.cert", i), "invalid x509 certificate", each.Name)
		}
	}
	cp := &model1.ClusterProvider{
		ID:    uuid.New(),
		Name:  input.Name,
		Class: model1.ClusterClassTKG2S,
	}
	if input.Description != nil {
		cp.Description = *input.Description
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
		cp.LoadBalancerID = model1.NewNullUUID(*input.LoadBalancer)
	}
	if tkg2sProviderProtectedNamespaces[input.Spec.DefaultNamespace] {
		return nil, errInvalidRequest(ctx, "input.spec.defaultNamespace", "cannot be a system namespace", input.Spec.DefaultNamespace)
	}
	if input.Spec.DefaultNamespace == "" {
		return nil, errInvalidRequest(ctx, "input.spec.defaultNamespace", "value cannot be empty")
	}
	spec := &model1.ClusterProviderTKG2SSpec{
		Endpoint:  input.Spec.Endpoint.String(),
		AuthUser:  input.Spec.AuthUser,
		DefaultNS: input.Spec.DefaultNamespace,
	}
	if input.Spec.CaCert != nil {
		spec.CACert = []byte(*input.Spec.CaCert)
	}
	if input.Spec.InsecureSkipTLSVerify != nil {
		spec.InsecureSkipTLSVerify = *input.Spec.InsecureSkipTLSVerify
	}
	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	spec.AuthPassword = block.Encrypt([]byte(input.Spec.AuthPassword))
	for i, each := range input.Spec.TrustedCAs {
		if !validateName(each.Name) {
			return nil, errInvalidRequest(ctx, fmt.Sprintf("input.spec.trustedCAs.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
		}
		rootcas := x509.NewCertPool()
		if !rootcas.AppendCertsFromPEM([]byte(each.Cert)) {
			return nil, errInvalidRequest(ctx, fmt.Sprintf("input.spec.trustedCAs.%d.cert", i), "invalid x509 certificate", each.Name)
		}
		spec.TrustedCAs = append(spec.TrustedCAs, &model1.CACert{
			Name: each.Name,
			Data: []byte(each.Cert),
		})
	}
	cp.Spec, err = json.Marshal(spec)
	if err != nil {
		panic(err)
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		registry := &model1.ImageRegistry{ID: defaultImageRegistry}
		if err := tx.Take(registry).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, fmt.Sprintf("input.imageRegistries.%d.id", defaultImageRegistryIndex), registry.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		if cp.LoadBalancerID.Valid {
			loadBalancer := &model1.LoadBalancer{ID: cp.LoadBalancerID.UUID}
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

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectClusterProvider, registry.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		cp.OrganizationID = registry.OrganizationID

		var exists int64
		if err := tx.Model(&model1.ClusterProvider{}).Where("`organization_id` = ? AND `name` = ?", cp.OrganizationID, cp.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", cp.Name)
		}

		if input.Validate != nil && *input.Validate {
			client, err := tanzuv2.NewVSphereWithTanzu(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
				Insecure: spec.InsecureSkipTLSVerify,
				CAData:   spec.CACert,
			})
			if err != nil {
				panic(err)
			}

			ctx, cancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
			defer cancel()

			if err := client.PingContext(ctx); err != nil {
				return errGatewayError(ctx, err)
			}
		}

		if err := tx.Create(cp).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(cp).Error; err != nil {
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

			reg := &model1.ClusterProviderImageRegistry{
				ID:              uuid.New(),
				ProviderID:      model1.NewNullUUID(cp.ID),
				ImageRegistryID: model1.NewNullUUID(each.ID),
			}
			if each.Default != nil && *each.Default {
				reg.Default = true
			}
			if err := tx.Create(reg).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}

		if input.Dryrun != nil && *input.Dryrun {
			return errDryrun
		}

		if err := appendAuditLog(tx, "create", "ClusterProvider", cp.ID, cp.Name, sess.UserID, sess.UserName, cp.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	result := convertClusterProviderFromSQLModel(cp, spec)
	if input.Dryrun != nil && *input.Dryrun {
		fields := graphql.CollectFieldsCtx(ctx, []string{"ClusterProvider"})
	LOOP:
		for _, field := range fields {
			if field.Name == "spec" {
				oc := graphql.GetOperationContext(ctx)
				fs := graphql.CollectFields(oc, field.Selections, []string{"TKG2SClusterProviderSpec", "ClusterProviderSpec"})
				for _, f := range fs {
					if f.Name == "namespaces" {
						ns, err := getTKG2SClusterProviderNamespaces(ctx, cp.ID, spec.Endpoint, spec.InsecureSkipTLSVerify, spec.CACert, spec.AuthUser, block.Decrypt(spec.AuthPassword), []*model.QueryFilter{})
						if err != nil {
							return nil, err
						}
						result.Spec.(*model.TKG2SClusterProviderSpec).Namespaces = ns
						break LOOP
					}
				}
				break
			}
		}
	}
	return result, nil
}

// CreateTKG2MClusterProvider is the resolver for the CreateTKG2MClusterProvider field.
func (r *mutationResolver) CreateTKG2MClusterProvider(ctx context.Context, input model.CreateTKG2MClusterProvider) (*model.ClusterProvider, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}
	cp := &model1.ClusterProvider{
		ID:    uuid.New(),
		Name:  input.Name,
		Class: model1.ClusterClassTKG2M,
	}
	if input.Description != nil {
		cp.Description = *input.Description
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
		cp.LoadBalancerID = model1.NewNullUUID(*input.LoadBalancer)
	}

	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	if tkg2mProviderProtectedNamespaces[input.Spec.DefaultNamespace] {
		return nil, errInvalidRequest(ctx, "input.spec.defaultNamespace", "cannot be a system namespace", input.Spec.DefaultNamespace)
	}
	if input.Spec.DefaultNamespace == "" {
		return nil, errInvalidRequest(ctx, "input.spec.defaultNamespace", "value cannot be empty")
	}
	spec := &model1.ClusterProviderTKG2MSpec{
		KubeConfig: block.Encrypt([]byte(input.Spec.KubeConfig)),
		DefaultNS:  input.Spec.DefaultNamespace,
	}
	for i, each := range input.Spec.TrustedCAs {
		if !validateName(each.Name) {
			return nil, errInvalidRequest(ctx, fmt.Sprintf("input.spec.trustedCAs.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
		}
		rootcas := x509.NewCertPool()
		if !rootcas.AppendCertsFromPEM([]byte(each.Cert)) {
			return nil, errInvalidRequest(ctx, fmt.Sprintf("input.spec.trustedCAs.%d.cert", i), "invalid x509 certificate", each.Name)
		}
		spec.TrustedCAs = append(spec.TrustedCAs, &model1.CACert{
			Name: each.Name,
			Data: []byte(each.Cert),
		})
	}
	if input.Spec.VsphereNetwork != nil {
		spec.VsphereNetwork = *input.Spec.VsphereNetwork
	}
	if input.Spec.VsphereResourcePool != nil {
		spec.VsphereResourcePool = *input.Spec.VsphereResourcePool
	}
	if input.Spec.VsphereFolder != nil {
		spec.VsphereFolder = *input.Spec.VsphereFolder
	}
	if input.Spec.VsphereDatastore != nil {
		spec.VsphereDatastore = *input.Spec.VsphereDatastore
	}
	if input.Spec.AviServiceEngineGroup != nil {
		spec.AviServiceEngineGroup = *input.Spec.AviServiceEngineGroup
	}
	if input.Spec.AviDataNetwork != nil {
		spec.AviDataNetwork = *input.Spec.AviDataNetwork
	}
	if input.Spec.AviDataNetworkCidr != nil {
		spec.AviDataNetworkCidr = *input.Spec.AviDataNetworkCidr
	}

	cp.Spec, err = json.Marshal(spec)
	if err != nil {
		panic(err)
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		registry := &model1.ImageRegistry{ID: defaultImageRegistry}
		if err := tx.Take(registry).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, fmt.Sprintf("input.imageRegistries.%d.id", defaultImageRegistryIndex), registry.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		if cp.LoadBalancerID.Valid {
			loadBalancer := &model1.LoadBalancer{ID: cp.LoadBalancerID.UUID}
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

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectClusterProvider, registry.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		cp.OrganizationID = registry.OrganizationID

		var exists int64
		if err := tx.Model(&model1.ClusterProvider{}).Where("`organization_id` = ? AND `name` = ?", cp.OrganizationID, cp.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", cp.Name)
		}

		if input.Validate != nil && *input.Validate {
			kubeConfig := block.Decrypt(spec.KubeConfig)

			if err := validateTKG2ManagementCluster(ctx, kubeConfig); err != nil {
				return err
			}
		}

		if err := tx.Create(cp).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(cp).Error; err != nil {
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

			reg := &model1.ClusterProviderImageRegistry{
				ID:              uuid.New(),
				ProviderID:      model1.NewNullUUID(cp.ID),
				ImageRegistryID: model1.NewNullUUID(each.ID),
			}
			if each.Default != nil && *each.Default {
				reg.Default = true
			}
			if err := tx.Create(reg).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}

		if input.Dryrun != nil && *input.Dryrun {
			return errDryrun
		}

		if err := appendAuditLog(tx, "create", "ClusterProvider", cp.ID, cp.Name, sess.UserID, sess.UserName, cp.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	result := convertClusterProviderFromSQLModel(cp, spec)
	if input.Dryrun != nil && *input.Dryrun {
		fields := graphql.CollectFieldsCtx(ctx, []string{"ClusterProvider"})
	LOOP:
		for _, field := range fields {
			if field.Name == "spec" {
				oc := graphql.GetOperationContext(ctx)
				fs := graphql.CollectFields(oc, field.Selections, []string{"TKG2MClusterProviderSpec", "ClusterProviderSpec"})
				for _, f := range fs {
					if f.Name == "namespaces" {
						kubeConfig := block.Decrypt(spec.KubeConfig)
						ns, err := getTKG2MClusterProviderNamespaces(ctx, cp.ID, kubeConfig, []*model.QueryFilter{})
						if err != nil {
							return nil, err
						}
						result.Spec.(*model.TKG2MClusterProviderSpec).Namespaces = ns
						break LOOP
					}
				}
				break
			}
		}
	}
	return result, nil
}

// UpdateTKGSClusterProvider is the resolver for the UpdateTKGSClusterProvider field.
func (r *mutationResolver) UpdateTKGSClusterProvider(ctx context.Context, input model.UpdateTKGSClusterProvider) (*model.ClusterProvider, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	cp := &model1.ClusterProvider{ID: input.ID}
	updates := make(map[string]interface{})
	if input.Description != nil {
		updates["description"] = *input.Description
	}

	var spec *model1.ClusterProviderTKGSSpec
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(cp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", cp.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if cp.Class != model1.ClusterClassTKGS {
			return errInvalidRequest(ctx, "input.id", "must be a TKGS ClusterProvider", cp.ID.String())
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectClusterProvider, cp.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		spec = &model1.ClusterProviderTKGSSpec{}
		if err := json.Unmarshal(cp.Spec, spec); err != nil {
			panic(err)
		}
		block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
		if err != nil {
			panic(err)
		}
		if input.Spec != nil {
			if input.Spec.Endpoint != nil {
				spec.Endpoint = input.Spec.Endpoint.String()
			}
			if input.Spec.CaCert != nil {
				rootcas := x509.NewCertPool()
				if !rootcas.AppendCertsFromPEM([]byte(*input.Spec.CaCert)) {
					return errInvalidRequest(ctx, "input.spec.caCert", "invalid x509 certificate")
				}
				spec.CACert = []byte(*input.Spec.CaCert)
			}
			if input.Spec.AuthUser != nil && *input.Spec.AuthUser != "" {
				spec.AuthUser = *input.Spec.AuthUser
			}
			if input.Spec.AuthPassword != nil && *input.Spec.AuthPassword != "" {
				spec.AuthPassword = block.Encrypt([]byte(*input.Spec.AuthPassword))
			}
			if input.Spec.DefaultNamespace != nil && *input.Spec.DefaultNamespace != "" {
				var count int64
				if err := tx.Model(&model1.Cluster{}).Where("`provider_id` = ?", cp.ID).Count(&count).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
				if count > 0 && spec.DefaultNS != *input.Spec.DefaultNamespace {
					return errObjectIsBusy(ctx, "input.spec.defaultNamespace", "Cluster", nil, "", *input.Spec.DefaultNamespace)
				}
				spec.DefaultNS = *input.Spec.DefaultNamespace
			}
			b, err := json.Marshal(spec)
			if err != nil {
				panic(err)
			}
			updates["spec"] = string(b)
		}

		for _, each := range input.RemoveImageRegistries {
			if err := tx.Where("`provider_id` = ? AND `image_registry_id` = ?", cp.ID, each.ID).Delete(&model1.ClusterProviderImageRegistry{}).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		for _, each := range input.AddImageRegistries {
			reg := &model1.ClusterProviderImageRegistry{}
			if err := tx.Where("`provider_id` = ? AND `image_registry_id` = ?", cp.ID, each.ID).Take(reg).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					panic(&ErrDBConnection{err})
				}

				reg = &model1.ClusterProviderImageRegistry{
					ID:              uuid.New(),
					ProviderID:      model1.NewNullUUID(cp.ID),
					ImageRegistryID: model1.NewNullUUID(each.ID),
				}
				if each.Default != nil && *each.Default {
					reg.Default = true
				}
				if err := tx.Create(reg).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
				continue
			}

			update := make(map[string]interface{})
			if each.Default != nil {
				update["default"] = *each.Default
			}
			if len(update) > 0 {
				if err := tx.Model(&model1.ClusterProviderImageRegistry{}).Where("`provider_id` = ? AND `image_registry_id` = ?", cp.ID, each.ID).Updates(update).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
			}
		}
		if len(input.RemoveImageRegistries)+len(input.AddImageRegistries) > 0 {
			var defaultImageRegistries int64
			if err := tx.Model(&model1.ClusterProviderImageRegistry{}).Where("`provider_id` = ? AND `default` = ?", cp.ID, true).Count(&defaultImageRegistries).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if defaultImageRegistries > 1 {
				return errInvalidRequest(ctx, "input.addImageRegistries", "multiple image registries are specified as default")
			} else if defaultImageRegistries < 1 {
				return errInvalidRequest(ctx, "input.removeImageRegistries", "default image registry is required")
			}
		}
		if input.LoadBalancer != nil {
			lb := &model1.LoadBalancer{ID: *input.LoadBalancer}
			if err := tx.Take(lb).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, "input.loadBalancer", *input.LoadBalancer)
				}
				panic(&ErrDBConnection{err})
			}
			if lb.OrganizationID.String() != cp.OrganizationID.String() {
				return errInvalidRequest(ctx, "input.loadBalancer", "different organization", lb.ID.String())
			}
			updates["load_balancer_id"] = input.LoadBalancer.String()
		}

		if input.Validate != nil && *input.Validate {
			client, err := k8s.NewTanzuKubernetesGridService(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
				Insecure: spec.InsecureSkipTLSVerify,
				CAData:   spec.CACert,
			})
			if err != nil {
				panic(err)
			}

			if err := client.PingContext(ctx); err != nil {
				return errGatewayError(ctx, err)
			}
		}

		if len(updates) > 0 {
			if err := tx.Model(cp).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(cp).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if input.Dryrun != nil && *input.Dryrun {
			return errDryrun
		}

		if err := appendAuditLog(tx, "update", "ClusterProvider", cp.ID, cp.Name, sess.UserID, sess.UserName, cp.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(input.Spec.AddTrustedCAs) > 0 || len(input.Spec.RemoveTrustedCAs) > 0 {
			client, err := k8s.NewTanzuKubernetesGridService(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
				Insecure: spec.InsecureSkipTLSVerify,
				CAData:   spec.CACert,
			})
			if err != nil {
				panic(err)
			}

			cfg, err := client.GetTkgServiceConfiguration(defTkgServiceConfiguration)
			if err != nil {
				return errGatewayError(ctx, err)
			}
			if cfg.Spec.Trust == nil {
				cfg.Spec.Trust = &runv1alpha1.TrustConfiguration{}
			}
			cert := make(map[string]string)
			for _, each := range cfg.Spec.Trust.AdditionalTrustedCAs {
				cert[each.Name] = each.Data
			}
			for _, each := range input.Spec.RemoveTrustedCAs {
				delete(cert, each.Name)
			}
			for _, each := range input.Spec.AddTrustedCAs {
				cert[each.Name] = base64.StdEncoding.EncodeToString([]byte(each.Cert))
			}
			cfg.Spec.Trust.AdditionalTrustedCAs = cfg.Spec.Trust.AdditionalTrustedCAs[:0]
			for name, data := range cert {
				cfg.Spec.Trust.AdditionalTrustedCAs = append(cfg.Spec.Trust.AdditionalTrustedCAs, runv1alpha1.TLSCertificate{
					Name: name,
					Data: data,
				})
			}
			if _, err := patchTkgServiceConfigurationAdditionalTrustedCAs(client, defTkgServiceConfiguration, cfg.Spec.Trust.AdditionalTrustedCAs); err != nil {
				return errGatewayError(ctx, err)
			}
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	result := convertClusterProviderFromSQLModel(cp, spec)
	if input.Dryrun != nil && *input.Dryrun && input.Validate != nil && *input.Validate {
		fields := graphql.CollectFieldsCtx(ctx, []string{"ClusterProvider"})
	LOOP:
		for _, field := range fields {
			if field.Name == "spec" {
				oc := graphql.GetOperationContext(ctx)
				fs := graphql.CollectFields(oc, field.Selections, []string{"TKGSClusterProviderSpec", "ClusterProviderSpec"})
				for _, f := range fs {
					if f.Name == "namespaces" {
						block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
						if err != nil {
							panic(err)
						}
						ns, err := getTKGSClusterProviderNamespaces(ctx, cp.ID, spec.Endpoint, spec.InsecureSkipTLSVerify, spec.CACert, spec.AuthUser, block.Decrypt(spec.AuthPassword), []*model.QueryFilter{})
						if err != nil {
							return nil, err
						}
						result.Spec.(*model.TKGSClusterProviderSpec).Namespaces = ns
						break LOOP
					}
				}
				break
			}
		}
	}
	return result, nil
}

// UpdateTKGMClusterProvider is the resolver for the UpdateTKGMClusterProvider field.
func (r *mutationResolver) UpdateTKGMClusterProvider(ctx context.Context, input model.UpdateTKGMClusterProvider) (*model.ClusterProvider, error) {
	return nil, errNoLongerSupport(ctx, "TKGm 1.X is no longer supported. Please consider to migrate your clusters to TKG 2.")
}

// UpdateTKG2SClusterProvider is the resolver for the UpdateTKG2SClusterProvider field.
func (r *mutationResolver) UpdateTKG2SClusterProvider(ctx context.Context, input model.UpdateTKG2SClusterProvider) (*model.ClusterProvider, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	cp := &model1.ClusterProvider{ID: input.ID}
	updates := make(map[string]interface{})
	if input.Description != nil {
		updates["description"] = *input.Description
	}

	var spec *model1.ClusterProviderTKG2SSpec
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(cp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", cp.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if cp.Class != model1.ClusterClassTKG2S {
			return errInvalidRequest(ctx, "input.id", "must be a TKG2S ClusterProvider", cp.ID.String())
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectClusterProvider, cp.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		spec = &model1.ClusterProviderTKG2SSpec{}
		if err := json.Unmarshal(cp.Spec, spec); err != nil {
			panic(err)
		}
		block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
		if err != nil {
			panic(err)
		}
		if input.Spec != nil {
			if input.Spec.Endpoint != nil {
				spec.Endpoint = input.Spec.Endpoint.String()
			}
			if input.Spec.CaCert != nil {
				rootcas := x509.NewCertPool()
				if !rootcas.AppendCertsFromPEM([]byte(*input.Spec.CaCert)) {
					return errInvalidRequest(ctx, "input.spec.caCert", "invalid x509 certificate")
				}
				spec.CACert = []byte(*input.Spec.CaCert)
			}
			if input.Spec.AuthUser != nil && *input.Spec.AuthUser != "" {
				spec.AuthUser = *input.Spec.AuthUser
			}
			if input.Spec.AuthPassword != nil && *input.Spec.AuthPassword != "" {
				spec.AuthPassword = block.Encrypt([]byte(*input.Spec.AuthPassword))
			}
			if input.Spec.DefaultNamespace != nil && *input.Spec.DefaultNamespace != "" {
				var count int64
				if err := tx.Model(&model1.Cluster{}).Where("`provider_id` = ?", cp.ID).Count(&count).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
				if count > 0 && spec.DefaultNS != *input.Spec.DefaultNamespace {
					return errObjectIsBusy(ctx, "input.spec.defaultNamespace", "Cluster", nil, "", *input.Spec.DefaultNamespace)
				}
				if tkg2sProviderProtectedNamespaces[*input.Spec.DefaultNamespace] {
					return errInvalidRequest(ctx, "input.spec.defaultNamespace", "cannot be a system namespace", *input.Spec.DefaultNamespace)
				}
				spec.DefaultNS = *input.Spec.DefaultNamespace
			}
			if len(input.Spec.AddTrustedCAs)+len(input.Spec.RemoveTrustedCAs) > 0 {
				m := make(map[string][]byte)
				for _, each := range spec.TrustedCAs {
					m[each.Name] = each.Data
				}
				for _, each := range input.Spec.RemoveTrustedCAs {
					delete(m, each.Name)
				}
				for i, each := range input.Spec.AddTrustedCAs {
					if !validateName(each.Name) {
						return errInvalidRequest(ctx, fmt.Sprintf("input.spec.addTrustedCAs.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
					}
					rootcas := x509.NewCertPool()
					if !rootcas.AppendCertsFromPEM([]byte(each.Cert)) {
						return errInvalidRequest(ctx, fmt.Sprintf("input.spec.addTrustedCAs.%d.cert", i), "invalid x509 certificate", each.Name)
					}
					m[each.Name] = []byte(each.Cert)
				}

				spec.TrustedCAs = spec.TrustedCAs[:0]
				for name, data := range m {
					spec.TrustedCAs = append(spec.TrustedCAs, &model1.CACert{
						Name: name,
						Data: data,
					})
				}
			}
			b, err := json.Marshal(spec)
			if err != nil {
				panic(err)
			}
			updates["spec"] = string(b)
		}

		for _, each := range input.RemoveImageRegistries {
			if err := tx.Where("`provider_id` = ? AND `image_registry_id` = ?", cp.ID, each.ID).Delete(&model1.ClusterProviderImageRegistry{}).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		for _, each := range input.AddImageRegistries {
			reg := &model1.ClusterProviderImageRegistry{}
			if err := tx.Where("`provider_id` = ? AND `image_registry_id` = ?", cp.ID, each.ID).Take(reg).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					panic(&ErrDBConnection{err})
				}

				reg = &model1.ClusterProviderImageRegistry{
					ID:              uuid.New(),
					ProviderID:      model1.NewNullUUID(cp.ID),
					ImageRegistryID: model1.NewNullUUID(each.ID),
				}
				if each.Default != nil && *each.Default {
					reg.Default = true
				}
				if err := tx.Create(reg).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
				continue
			}

			update := make(map[string]interface{})
			if each.Default != nil {
				update["default"] = *each.Default
			}
			if len(update) > 0 {
				if err := tx.Model(&model1.ClusterProviderImageRegistry{}).Where("`provider_id` = ? AND `image_registry_id` = ?", cp.ID, each.ID).Updates(update).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
			}
		}
		if len(input.RemoveImageRegistries)+len(input.AddImageRegistries) > 0 {
			var defaultImageRegistries int64
			if err := tx.Model(&model1.ClusterProviderImageRegistry{}).Where("`provider_id` = ? AND `default` = ?", cp.ID, true).Count(&defaultImageRegistries).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if defaultImageRegistries > 1 {
				return errInvalidRequest(ctx, "input.addImageRegistries", "multiple image registries are specified as default")
			} else if defaultImageRegistries < 1 {
				return errInvalidRequest(ctx, "input.removeImageRegistries", "default image registry is required")
			}
		}
		if input.LoadBalancer != nil {
			lb := &model1.LoadBalancer{ID: *input.LoadBalancer}
			if err := tx.Take(lb).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, "input.loadBalancer", *input.LoadBalancer)
				}
				panic(&ErrDBConnection{err})
			}
			if lb.OrganizationID.String() != cp.OrganizationID.String() {
				return errInvalidRequest(ctx, "input.loadBalancer", "different organization", lb.ID.String())
			}
			updates["load_balancer_id"] = input.LoadBalancer.String()
		}

		if input.Validate != nil && *input.Validate {
			client, err := k8s.NewTanzuKubernetesGridService(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
				Insecure: spec.InsecureSkipTLSVerify,
				CAData:   spec.CACert,
			})
			if err != nil {
				panic(err)
			}

			if err := client.PingContext(ctx); err != nil {
				return errGatewayError(ctx, err)
			}
		}

		if len(updates) > 0 {
			if err := tx.Model(cp).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(cp).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if input.Dryrun != nil && *input.Dryrun {
			return errDryrun
		}

		if err := appendAuditLog(tx, "update", "ClusterProvider", cp.ID, cp.Name, sess.UserID, sess.UserName, cp.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	result := convertClusterProviderFromSQLModel(cp, spec)
	if input.Dryrun != nil && *input.Dryrun && input.Validate != nil && *input.Validate {
		fields := graphql.CollectFieldsCtx(ctx, []string{"ClusterProvider"})
	LOOP:
		for _, field := range fields {
			if field.Name == "spec" {
				oc := graphql.GetOperationContext(ctx)
				fs := graphql.CollectFields(oc, field.Selections, []string{"TKG2SClusterProviderSpec", "ClusterProviderSpec"})
				for _, f := range fs {
					if f.Name == "namespaces" {
						block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
						if err != nil {
							panic(err)
						}
						ns, err := getTKG2SClusterProviderNamespaces(ctx, cp.ID, spec.Endpoint, spec.InsecureSkipTLSVerify, spec.CACert, spec.AuthUser, block.Decrypt(spec.AuthPassword), []*model.QueryFilter{})
						if err != nil {
							return nil, err
						}
						result.Spec.(*model.TKG2SClusterProviderSpec).Namespaces = ns
						break LOOP
					}
				}
				break
			}
		}
	}
	return result, nil
}

// UpdateTKG2MClusterProvider is the resolver for the UpdateTKG2MClusterProvider field.
func (r *mutationResolver) UpdateTKG2MClusterProvider(ctx context.Context, input model.UpdateTKG2MClusterProvider) (*model.ClusterProvider, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	cp := &model1.ClusterProvider{ID: input.ID}
	updates := make(map[string]interface{})
	if input.Description != nil {
		updates["description"] = *input.Description
	}

	var spec *model1.ClusterProviderTKG2MSpec
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(cp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", cp.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if cp.Class != model1.ClusterClassTKG2M {
			return errInvalidRequest(ctx, "input.id", "must be a TKG2M ClusterProvider", cp.ID.String())
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectClusterProvider, cp.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		spec = &model1.ClusterProviderTKG2MSpec{}
		if err := json.Unmarshal(cp.Spec, spec); err != nil {
			panic(err)
		}
		block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
		if err != nil {
			panic(err)
		}
		if input.Spec != nil {
			if input.Spec.KubeConfig != nil {
				spec.KubeConfig = block.Encrypt([]byte(*input.Spec.KubeConfig))
			}
			if input.Spec.DefaultNamespace != nil && *input.Spec.DefaultNamespace != "" {
				var count int64
				if err := tx.Model(&model1.Cluster{}).Where("`provider_id` = ?", cp.ID).Count(&count).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
				if count > 0 && spec.DefaultNS != *input.Spec.DefaultNamespace {
					return errObjectIsBusy(ctx, "input.spec.defaultNamespace", "Cluster", nil, "", *input.Spec.DefaultNamespace)
				}
				if tkg2mProviderProtectedNamespaces[*input.Spec.DefaultNamespace] {
					return errInvalidRequest(ctx, "input.spec.defaultNamespace", "cannot be a system namespace", *input.Spec.DefaultNamespace)
				}
				spec.DefaultNS = *input.Spec.DefaultNamespace
			}
			if len(input.Spec.AddTrustedCAs)+len(input.Spec.RemoveTrustedCAs) > 0 {
				m := make(map[string][]byte)
				for _, each := range spec.TrustedCAs {
					m[each.Name] = each.Data
				}
				for _, each := range input.Spec.RemoveTrustedCAs {
					delete(m, each.Name)
				}
				for i, each := range input.Spec.AddTrustedCAs {
					if !validateName(each.Name) {
						return errInvalidRequest(ctx, fmt.Sprintf("input.spec.addTrustedCAs.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
					}
					rootcas := x509.NewCertPool()
					if !rootcas.AppendCertsFromPEM([]byte(each.Cert)) {
						return errInvalidRequest(ctx, fmt.Sprintf("input.spec.addTrustedCAs.%d.cert", i), "invalid x509 certificate", each.Name)
					}
					m[each.Name] = []byte(each.Cert)
				}

				spec.TrustedCAs = spec.TrustedCAs[:0]
				for name, data := range m {
					spec.TrustedCAs = append(spec.TrustedCAs, &model1.CACert{
						Name: name,
						Data: data,
					})
				}
			}
			if input.Spec.VsphereNetwork != nil {
				spec.VsphereNetwork = *input.Spec.VsphereNetwork
			}
			if input.Spec.VsphereResourcePool != nil {
				spec.VsphereResourcePool = *input.Spec.VsphereResourcePool
			}
			if input.Spec.VsphereFolder != nil {
				spec.VsphereFolder = *input.Spec.VsphereFolder
			}
			if input.Spec.VsphereDatastore != nil {
				spec.VsphereDatastore = *input.Spec.VsphereDatastore
			}
			if input.Spec.AviServiceEngineGroup != nil {
				spec.AviServiceEngineGroup = *input.Spec.AviServiceEngineGroup
			}
			if input.Spec.AviDataNetwork != nil {
				spec.AviDataNetwork = *input.Spec.AviDataNetwork
			}
			if input.Spec.AviDataNetworkCidr != nil {
				spec.AviDataNetworkCidr = *input.Spec.AviDataNetworkCidr
			}
			b, err := json.Marshal(spec)
			if err != nil {
				panic(err)
			}
			updates["spec"] = string(b)
		}

		for _, each := range input.RemoveImageRegistries {
			if err := tx.Where("`provider_id` = ? AND `image_registry_id` = ?", cp.ID, each.ID).Delete(&model1.ClusterProviderImageRegistry{}).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		for _, each := range input.AddImageRegistries {
			reg := &model1.ClusterProviderImageRegistry{}
			if err := tx.Where("`provider_id` = ? AND `image_registry_id` = ?", cp.ID, each.ID).Take(reg).Error; err != nil {
				if !errors.Is(err, gorm.ErrRecordNotFound) {
					panic(&ErrDBConnection{err})
				}

				reg = &model1.ClusterProviderImageRegistry{
					ID:              uuid.New(),
					ProviderID:      model1.NewNullUUID(cp.ID),
					ImageRegistryID: model1.NewNullUUID(each.ID),
				}
				if each.Default != nil && *each.Default {
					reg.Default = true
				}
				if err := tx.Create(reg).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
				continue
			}

			update := make(map[string]interface{})
			if each.Default != nil {
				update["default"] = *each.Default
			}
			if len(update) > 0 {
				if err := tx.Model(&model1.ClusterProviderImageRegistry{}).Where("`provider_id` = ? AND `image_registry_id` = ?", cp.ID, each.ID).Updates(update).Error; err != nil {
					panic(&ErrDBConnection{err})
				}
			}
		}
		if len(input.RemoveImageRegistries)+len(input.AddImageRegistries) > 0 {
			var defaultImageRegistries int64
			if err := tx.Model(&model1.ClusterProviderImageRegistry{}).Where("`provider_id` = ? AND `default` = ?", cp.ID, true).Count(&defaultImageRegistries).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if defaultImageRegistries > 1 {
				return errInvalidRequest(ctx, "input.addImageRegistries", "multiple image registries are specified as default")
			} else if defaultImageRegistries < 1 {
				return errInvalidRequest(ctx, "input.removeImageRegistries", "default image registry is required")
			}
		}
		if input.LoadBalancer != nil {
			lb := &model1.LoadBalancer{ID: *input.LoadBalancer}
			if err := tx.Take(lb).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, "input.loadBalancer", *input.LoadBalancer)
				}
				panic(&ErrDBConnection{err})
			}
			if lb.OrganizationID.String() != cp.OrganizationID.String() {
				return errInvalidRequest(ctx, "input.loadBalancer", "different organization", lb.ID.String())
			}
			updates["load_balancer_id"] = input.LoadBalancer.String()
		}

		if input.Validate != nil && *input.Validate {
			kubeConfig := block.Decrypt(spec.KubeConfig)

			if err := validateTKG2ManagementCluster(ctx, kubeConfig); err != nil {
				return err
			}
		}

		if len(updates) > 0 {
			if err := tx.Model(cp).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(cp).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if input.Dryrun != nil && *input.Dryrun {
			return errDryrun
		}

		if err := appendAuditLog(tx, "update", "ClusterProvider", cp.ID, cp.Name, sess.UserID, sess.UserName, cp.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	result := convertClusterProviderFromSQLModel(cp, spec)
	if input.Dryrun != nil && *input.Dryrun && input.Validate != nil && *input.Validate {
		fields := graphql.CollectFieldsCtx(ctx, []string{"ClusterProvider"})
	LOOP:
		for _, field := range fields {
			if field.Name == "spec" {
				oc := graphql.GetOperationContext(ctx)
				fs := graphql.CollectFields(oc, field.Selections, []string{"TKG2MClusterProviderSpec", "ClusterProviderSpec"})
				for _, f := range fs {
					if f.Name == "namespaces" {
						block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
						if err != nil {
							panic(err)
						}
						kubeConfig := block.Decrypt(spec.KubeConfig)
						ns, err := getTKG2MClusterProviderNamespaces(ctx, cp.ID, kubeConfig, []*model.QueryFilter{})
						if err != nil {
							return nil, err
						}
						result.Spec.(*model.TKG2MClusterProviderSpec).Namespaces = ns
						break LOOP
					}
				}
				break
			}
		}
	}
	return result, nil
}

// DeleteClusterProvider is the resolver for the DeleteClusterProvider field.
func (r *mutationResolver) DeleteClusterProvider(ctx context.Context, input model.DeleteClusterProvider) (*model.ClusterProvider, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	cp := &model1.ClusterProvider{ID: input.ID}
	var spec interface{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(cp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", cp.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermModify, model1.PolicyObjectApplication, cp.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		switch cp.Class {
		case model1.ClusterClassTKGS:
			v := &model1.ClusterProviderTKGSSpec{}
			if err := json.Unmarshal(cp.Spec, v); err != nil {
				panic(err)
			}
			spec = v
		case model1.ClusterClassTKGM:
			v := &model1.ClusterProviderTKGMSpec{}
			if err := json.Unmarshal(cp.Spec, v); err != nil {
				panic(err)
			}
			spec = v
		case model1.ClusterClassTKG2S:
			v := &model1.ClusterProviderTKG2SSpec{}
			if err := json.Unmarshal(cp.Spec, v); err != nil {
				panic(err)
			}
			spec = v
		case model1.ClusterClassTKG2M:
			v := &model1.ClusterProviderTKG2MSpec{}
			if err := json.Unmarshal(cp.Spec, v); err != nil {
				panic(err)
			}
			spec = v
		}

		var exists int64
		if err := tx.Model(&model1.Cluster{}).Where("`provider_id` = ?", cp.ID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectIsBusy(ctx, "input.id", "Cluster", nil, "", input.ID.String())
		}

		if err := appendAuditLog(tx, "delete", "ClusterProvider", cp.ID, cp.Name, sess.UserID, sess.UserName, cp.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Where("`provider_id` = ?", cp.ID).Delete(&model1.ClusterProviderImageRegistry{}).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Delete(cp).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertClusterProviderFromSQLModel(cp, spec), nil
}

// ClusterProvider is the resolver for the ClusterProvider field.
func (r *queryResolver) ClusterProvider(ctx context.Context, id uuid.UUID) (*model.ClusterProvider, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	var (
		cp   = &model1.ClusterProvider{}
		spec interface{}
	)
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if !IsGuardSilenced(ctx) {
			tx = sess.FilterQueryByOrganization(tx, model1.PermReadOnly, model1.PolicyObjectClusterProvider)
		}
		if err := tx.Where("`id` = ?", id).Take(cp).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "id", id.String())
			}
			panic(&ErrDBConnection{err})
		}
		switch cp.Class {
		case model1.ClusterClassTKGS:
			v := &model1.ClusterProviderTKGSSpec{}
			if err := json.Unmarshal(cp.Spec, v); err != nil {
				panic(err)
			}
			spec = v
		case model1.ClusterClassTKGM:
			v := &model1.ClusterProviderTKGMSpec{}
			if err := json.Unmarshal(cp.Spec, v); err != nil {
				panic(err)
			}
			spec = v
		case model1.ClusterClassTKG2S:
			v := &model1.ClusterProviderTKG2SSpec{}
			if err := json.Unmarshal(cp.Spec, v); err != nil {
				panic(err)
			}
			spec = v
		case model1.ClusterClassTKG2M:
			v := &model1.ClusterProviderTKG2MSpec{}
			if err := json.Unmarshal(cp.Spec, v); err != nil {
				panic(err)
			}
			spec = v
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertClusterProviderFromSQLModel(cp, spec), nil
}

// ClusterProviders is the resolver for the ClusterProviders field.
func (r *queryResolver) ClusterProviders(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.ClusterProviders, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj1 := &model.ClusterProviders{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj1.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj1.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj1.Filters = append(obj1.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.ClusterProvider{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectClusterProvider)
	}
	db = finalizeClusterProviderQueryFilter(db, obj1.Filters...)

	if err := db.Count(&obj1.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj1.Page, obj1.TotalCount)
	return obj1, nil
}

// TanzuKubernetesReleases is the resolver for the tanzuKubernetesReleases field.
func (r *tKG2MClusterProviderSpecResolver) TanzuKubernetesReleases(ctx context.Context, obj *model.TKG2MClusterProviderSpec) ([]*semver.Version, error) {
	cp := &model1.ClusterProvider{ID: obj.ID}
	if err := mapd.Database().Take(cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	spec := &model1.ClusterProviderTKG2MSpec{}
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		panic(err)
	}

	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	client, err := tanzuv2.NewTanzuKubernetesGrid(block.Decrypt(spec.KubeConfig))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	list, err := client.GetAvailableTanzuKubernetesReleasesContext(ctx)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	m := make(map[string]*semver.Version)
	for _, each := range list.Items {
		v, err := abstract.MustParseVersion(each.Name).SetPrerelease("")
		if err != nil {
			panic(err)
		}
		verStr := v.String()
		m[verStr] = abstract.MustParseVersion(verStr)
	}
	var versions []*semver.Version
	for _, each := range m {
		versions = append(versions, each)
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i].LessThan(versions[j]) })
	return versions, nil
}

// Namespaces is the resolver for the namespaces field.
func (r *tKG2MClusterProviderSpecResolver) Namespaces(ctx context.Context, obj *model.TKG2MClusterProviderSpec, filters []*model.QueryFilter) ([]*model.TKG2MClusterProviderSpecNamespace, error) {
	if obj.Namespaces != nil {
		return obj.Namespaces, nil
	}

	cp := &model1.ClusterProvider{ID: obj.ID}
	if err := mapd.Database().Take(cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	spec := &model1.ClusterProviderTKG2MSpec{}
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		panic(err)
	}

	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	return getTKG2MClusterProviderNamespaces(ctx, obj.ID, block.Decrypt(spec.KubeConfig), filters)
}

// VMClasses is the resolver for the vmClasses field.
func (r *tKG2SClusterProviderSpecResolver) VMClasses(ctx context.Context, obj *model.TKG2SClusterProviderSpec) ([]*model.VMClass, error) {
	cp := &model1.ClusterProvider{ID: obj.ID}
	if err := mapd.Database().Take(cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	spec := &model1.ClusterProviderTKG2SSpec{}
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		panic(err)
	}

	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	client, err := tanzuv2.NewVSphereWithTanzu(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
		Insecure: spec.InsecureSkipTLSVerify,
		CAData:   spec.CACert,
	})
	if err != nil {
		panic(err)
	}

	vmClasses, err := client.GetVirtualMachineClasses()
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	var classes []*model.VMClass
	for _, each := range vmClasses.Items {
		d, err := json.Marshal(each.Spec)
		if err != nil {
			panic(err)
		}
		classes = append(classes, &model.VMClass{Name: each.Name, Spec: d})
	}
	return classes, nil
}

// TanzuKubernetesReleases is the resolver for the tanzuKubernetesReleases field.
func (r *tKG2SClusterProviderSpecResolver) TanzuKubernetesReleases(ctx context.Context, obj *model.TKG2SClusterProviderSpec) ([]*semver.Version, error) {
	cp := &model1.ClusterProvider{ID: obj.ID}
	if err := mapd.Database().Take(cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	spec := &model1.ClusterProviderTKG2SSpec{}
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		panic(err)
	}

	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	client, err := tanzuv2.NewVSphereWithTanzu(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
		Insecure: spec.InsecureSkipTLSVerify,
		CAData:   spec.CACert,
	})
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	list, err := client.GetAvailableTanzuKubernetesReleasesContext(ctx)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	m := make(map[string]*semver.Version)
	for _, each := range list.Items {
		v, err := abstract.MustParseVersion(each.Name).SetPrerelease("")
		if err != nil {
			panic(err)
		}
		verStr := v.String()
		m[verStr] = abstract.MustParseVersion(verStr)
	}
	var versions []*semver.Version
	for _, each := range m {
		versions = append(versions, each)
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i].LessThan(versions[j]) })
	return versions, nil
}

// Namespaces is the resolver for the namespaces field.
func (r *tKG2SClusterProviderSpecResolver) Namespaces(ctx context.Context, obj *model.TKG2SClusterProviderSpec, filters []*model.QueryFilter) ([]*model.TKG2SClusterProviderSpecNamespace, error) {
	if obj.Namespaces != nil {
		return obj.Namespaces, nil
	}

	cp := &model1.ClusterProvider{ID: obj.ID}
	if err := mapd.Database().Take(cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	spec := &model1.ClusterProviderTKG2SSpec{}
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		panic(err)
	}

	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	return getTKG2SClusterProviderNamespaces(ctx, obj.ID, spec.Endpoint, spec.InsecureSkipTLSVerify, spec.CACert, spec.AuthUser, block.Decrypt(spec.AuthPassword), filters)
}

// StorageClasses is the resolver for the storageClasses field.
func (r *tKG2SClusterProviderSpecNamespaceResolver) StorageClasses(ctx context.Context, obj *model.TKG2SClusterProviderSpecNamespace) ([]*model.ClusterProviderStorageClass, error) {
	cp := &model1.ClusterProvider{ID: obj.ClusterProvider}
	if err := mapd.Database().Take(cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	spec := &model1.ClusterProviderTKG2SSpec{}
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		panic(err)
	}

	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	client, err := tanzuv2.NewVSphereWithTanzu(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
		Insecure: spec.InsecureSkipTLSVerify,
		CAData:   spec.CACert,
	})
	if err != nil {
		panic(err)
	}

	var storages []*model.ClusterProviderStorageClass
	quotas, err := client.GetResourceQuotasContext(ctx, obj.Name)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	for _, quota := range quotas.Items {
		for key, value := range quota.Spec.Hard {
			s := strings.TrimSuffix(string(key), ".storageclass.storage.k8s.io/requests.storage")
			if s != "" {
				sc := &model.ClusterProviderStorageClass{
					Name:    s,
					TotalGi: value.Value() / util.Gi,
				}
				if quota.Status.Used != nil {
					used := quota.Status.Used[key]
					sc.UsedGi = used.Value() / util.Gi
				}
				storages = append(storages, sc)
			}
		}
	}
	return storages, nil
}

// ClusterReleases is the resolver for the clusterReleases field.
func (r *tKGMClusterProviderSpecResolver) ClusterReleases(ctx context.Context, obj *model.TKGMClusterProviderSpec) ([]*semver.Version, error) {
	return nil, errNoLongerSupport(ctx, "TKGm 1.X is no longer supported. Please consider to migrate your clusters to TKG 2.")
}

// Namespaces is the resolver for the namespaces field.
func (r *tKGMClusterProviderSpecResolver) Namespaces(ctx context.Context, obj *model.TKGMClusterProviderSpec, filters []*model.QueryFilter) ([]*model.TKGMClusterProviderSpecNamespace, error) {
	return nil, errNoLongerSupport(ctx, "TKGm 1.X is no longer supported. Please consider to migrate your clusters to TKG 2.")
}

// VMClasses is the resolver for the vmClasses field.
func (r *tKGSClusterProviderSpecResolver) VMClasses(ctx context.Context, obj *model.TKGSClusterProviderSpec) ([]*model.VMClass, error) {
	cp := &model1.ClusterProvider{ID: obj.ID}
	if err := mapd.Database().Take(cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	spec := &model1.ClusterProviderTKGSSpec{}
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		panic(err)
	}

	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	client, err := k8s.NewTanzuKubernetesGridService(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
		Insecure: spec.InsecureSkipTLSVerify,
		CAData:   spec.CACert,
	})
	if err != nil {
		panic(err)
	}

	vmcs, err := getVirtualMachineClasses(client)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	var classes []*model.VMClass
	for _, each := range vmcs {
		d, err := json.Marshal(each.Spec)
		if err != nil {
			panic(err)
		}
		classes = append(classes, &model.VMClass{Name: each.Name, Spec: d})
	}
	return classes, nil
}

// ClusterReleases is the resolver for the clusterReleases field.
func (r *tKGSClusterProviderSpecResolver) ClusterReleases(ctx context.Context, obj *model.TKGSClusterProviderSpec) ([]*semver.Version, error) {
	cp := &model1.ClusterProvider{ID: obj.ID}
	if err := mapd.Database().Take(cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	spec := &model1.ClusterProviderTKGSSpec{}
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		panic(err)
	}

	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	client, err := k8s.NewTanzuKubernetesGridService(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
		Insecure: spec.InsecureSkipTLSVerify,
		CAData:   spec.CACert,
	})
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	list, err := client.GetTanzuKubernetesReleasesContext(ctx)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	m := make(map[string]*semver.Version)
LOOP:
	for _, each := range list.Items {
		for _, cond := range each.Status.Conditions {
			switch cond.Type {
			case "Ready":
				if cond.Status != corev1.ConditionTrue {
					continue LOOP
				}
			case "Compatible":
				if cond.Status != corev1.ConditionTrue {
					continue LOOP
				}
			}
		}

		v, err := abstract.MustParseVersion(each.Name).SetPrerelease("")
		if err != nil {
			panic(err)
		}
		verStr := v.String()
		m[verStr] = abstract.MustParseVersion(verStr)
	}
	var versions []*semver.Version
	for _, each := range m {
		versions = append(versions, each)
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i].LessThan(versions[j]) })
	return versions, nil
}

// Namespaces is the resolver for the namespaces field.
func (r *tKGSClusterProviderSpecResolver) Namespaces(ctx context.Context, obj *model.TKGSClusterProviderSpec, filters []*model.QueryFilter) ([]*model.TKGSClusterProviderSpecNamespace, error) {
	if obj.Namespaces != nil {
		return obj.Namespaces, nil
	}

	cp := &model1.ClusterProvider{ID: obj.ID}
	if err := mapd.Database().Take(cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	spec := &model1.ClusterProviderTKGSSpec{}
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		panic(err)
	}

	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	return getTKGSClusterProviderNamespaces(ctx, obj.ID, spec.Endpoint, spec.InsecureSkipTLSVerify, spec.CACert, spec.AuthUser, block.Decrypt(spec.AuthPassword), filters)
}

// TrustedCAs is the resolver for the trustedCAs field.
func (r *tKGSClusterProviderSpecResolver) TrustedCAs(ctx context.Context, obj *model.TKGSClusterProviderSpec) ([]*model.ClusterProviderSpecTrustedCa, error) {
	cp := &model1.ClusterProvider{ID: obj.ID}
	if err := mapd.Database().Take(cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	spec := &model1.ClusterProviderTKGSSpec{}
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		panic(err)
	}

	block, err := cryptoutil.NewAESGCMCipher([]byte(cp.ID[:])) // AES-128
	if err != nil {
		panic(err)
	}
	client, err := k8s.NewTanzuKubernetesGridService(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
		Insecure: spec.InsecureSkipTLSVerify,
		CAData:   spec.CACert,
	})
	if err != nil {
		panic(err)
	}
	cfg, err := client.GetTkgServiceConfiguration(defTkgServiceConfiguration)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	var cas []*model.ClusterProviderSpecTrustedCa
	if cfg.Spec.Trust != nil {
		for _, each := range cfg.Spec.Trust.AdditionalTrustedCAs {
			v, err := base64.StdEncoding.DecodeString(each.Data)
			if err != nil {
				panic(err)
			}
			cas = append(cas, &model.ClusterProviderSpecTrustedCa{
				Name: each.Name,
				Cert: string(v),
			})
		}
	}
	return cas, nil
}

// Storages is the resolver for the storages field.
func (r *tKGSClusterProviderSpecNamespaceResolver) Storages(ctx context.Context, obj *model.TKGSClusterProviderSpecNamespace) ([]string, error) {
	cp := &model1.ClusterProvider{ID: obj.ClusterProvider}
	if err := mapd.Database().Take(cp).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	spec := &model1.ClusterProviderTKGSSpec{}
	if err := json.Unmarshal(cp.Spec, spec); err != nil {
		panic(err)
	}

	block, err := cryptoutil.NewAESGCMCipher(cp.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	client, err := k8s.NewTanzuKubernetesGridService(spec.Endpoint, spec.AuthUser, block.Decrypt(spec.AuthPassword), rest.TLSClientConfig{
		Insecure: spec.InsecureSkipTLSVerify,
		CAData:   spec.CACert,
	})
	if err != nil {
		panic(err)
	}

	var storages []string
	quotas, err := client.GetResourceQuotasContext(ctx, obj.Name)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	for _, quota := range quotas.Items {
		for key := range quota.Spec.Hard {
			s := strings.TrimSuffix(string(key), ".storageclass.storage.k8s.io/requests.storage")
			if s != "" {
				storages = append(storages, s)
			}
		}
	}
	return storages, nil
}

// ClusterProvider returns graphql1.ClusterProviderResolver implementation.
func (r *Resolver) ClusterProvider() graphql1.ClusterProviderResolver {
	return &clusterProviderResolver{r}
}

// ClusterProviderImageRegistries returns graphql1.ClusterProviderImageRegistriesResolver implementation.
func (r *Resolver) ClusterProviderImageRegistries() graphql1.ClusterProviderImageRegistriesResolver {
	return &clusterProviderImageRegistriesResolver{r}
}

// ClusterProviderImageRegistry returns graphql1.ClusterProviderImageRegistryResolver implementation.
func (r *Resolver) ClusterProviderImageRegistry() graphql1.ClusterProviderImageRegistryResolver {
	return &clusterProviderImageRegistryResolver{r}
}

// ClusterProviders returns graphql1.ClusterProvidersResolver implementation.
func (r *Resolver) ClusterProviders() graphql1.ClusterProvidersResolver {
	return &clusterProvidersResolver{r}
}

// TKG2MClusterProviderSpec returns graphql1.TKG2MClusterProviderSpecResolver implementation.
func (r *Resolver) TKG2MClusterProviderSpec() graphql1.TKG2MClusterProviderSpecResolver {
	return &tKG2MClusterProviderSpecResolver{r}
}

// TKG2SClusterProviderSpec returns graphql1.TKG2SClusterProviderSpecResolver implementation.
func (r *Resolver) TKG2SClusterProviderSpec() graphql1.TKG2SClusterProviderSpecResolver {
	return &tKG2SClusterProviderSpecResolver{r}
}

// TKG2SClusterProviderSpecNamespace returns graphql1.TKG2SClusterProviderSpecNamespaceResolver implementation.
func (r *Resolver) TKG2SClusterProviderSpecNamespace() graphql1.TKG2SClusterProviderSpecNamespaceResolver {
	return &tKG2SClusterProviderSpecNamespaceResolver{r}
}

// TKGMClusterProviderSpec returns graphql1.TKGMClusterProviderSpecResolver implementation.
func (r *Resolver) TKGMClusterProviderSpec() graphql1.TKGMClusterProviderSpecResolver {
	return &tKGMClusterProviderSpecResolver{r}
}

// TKGSClusterProviderSpec returns graphql1.TKGSClusterProviderSpecResolver implementation.
func (r *Resolver) TKGSClusterProviderSpec() graphql1.TKGSClusterProviderSpecResolver {
	return &tKGSClusterProviderSpecResolver{r}
}

// TKGSClusterProviderSpecNamespace returns graphql1.TKGSClusterProviderSpecNamespaceResolver implementation.
func (r *Resolver) TKGSClusterProviderSpecNamespace() graphql1.TKGSClusterProviderSpecNamespaceResolver {
	return &tKGSClusterProviderSpecNamespaceResolver{r}
}

type clusterProviderResolver struct{ *Resolver }
type clusterProviderImageRegistriesResolver struct{ *Resolver }
type clusterProviderImageRegistryResolver struct{ *Resolver }
type clusterProvidersResolver struct{ *Resolver }
type tKG2MClusterProviderSpecResolver struct{ *Resolver }
type tKG2SClusterProviderSpecResolver struct{ *Resolver }
type tKG2SClusterProviderSpecNamespaceResolver struct{ *Resolver }
type tKGMClusterProviderSpecResolver struct{ *Resolver }
type tKGSClusterProviderSpecResolver struct{ *Resolver }
type tKGSClusterProviderSpecNamespaceResolver struct{ *Resolver }
