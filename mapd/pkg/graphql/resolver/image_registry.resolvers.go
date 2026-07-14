package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/harbor"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil"
	"gorm.io/gorm"
)

// Versions is the resolver for the versions field.
func (r *chartRepoResolver) Versions(ctx context.Context, obj *model.ChartRepo) ([]*model.ChartVersion, error) {
	if obj.ImageRegistry == nil {
		return nil, nil
	}

	registry := &model1.ImageRegistry{ID: obj.ImageRegistry.ID}
	if err := mapd.Database().Take(registry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}
	block, err := cryptoutil.NewAESGCMCipher(registry.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	var versions []*model.ChartVersion
	switch registry.Type {
	case model1.ImageRegistryTypeHarbor:
		var (
			authUser     string
			authPassword []byte
		)
		if registry.AuthUser != "" {
			authUser, authPassword = registry.AuthUser, block.Decrypt(registry.AuthPassword)
		}
		client, err := harbor.NewV2Client(registry.Endpoint, authUser, authPassword, false, registry.CACert)
		if err != nil {
			panic(err)
		}
		items, err := client.GetChartsContext(ctx, obj.Project, obj.Name)
		if err != nil {
			return nil, errGatewayError(ctx, err)
		}
		for _, chart := range items {
			obj := &model.ChartVersion{
				Repo:          obj.Name,
				Project:       obj.Project,
				ImageRegistry: obj.ImageRegistry.ID,
				Version:       chart.Version,
			}
			versions = append(versions, obj)
		}
	}
	return versions, nil
}

// Details is the resolver for the details field.
func (r *chartVersionResolver) Details(ctx context.Context, obj *model.ChartVersion) (*model.ChartDetails, error) {
	return r.Resolver.Query().(*queryResolver).ChartDetails(WithGuardSilenced(ctx), obj.ImageRegistry, obj.Project, obj.Repo, obj.Version)
}

// Edge is the resolver for the edge field.
func (r *imageRegistriesResolver) Edge(ctx context.Context, obj *model.ImageRegistries) (*model.ImageRegistryNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.ImageRegistryNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectImageRegistry)
		}
		db = finalizeImageRegistryQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var registries []model1.ImageRegistry
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&registries).Error; err != nil {
				return err
			}
			for _, registry := range registries {
				nodes.Nodes = append(nodes.Nodes, convertImageRegistryFromSQLModel(&registry))
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

// TLSCertChain is the resolver for the tlsCertChain field.
func (r *imageRegistryResolver) TLSCertChain(ctx context.Context, obj *model.ImageRegistry) ([]*model.X509Certificate, error) {
	zctx, zcancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
	defer zcancel()

	chain, err := httputil.EZGetTLSCertificateChainContext(zctx, obj.Endpoint.String(), false, []byte(obj.CaCert+"\n"))
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	var certs []*model.X509Certificate
	for _, each := range chain {
		certs = append(certs, convertX509Certificate(each))
	}
	return certs, nil
}

// Projects is the resolver for the projects field.
func (r *imageRegistryResolver) Projects(ctx context.Context, obj *model.ImageRegistry) ([]string, error) {
	registry := &model1.ImageRegistry{ID: obj.ID}
	if err := mapd.Database().Take(registry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}
	block, err := cryptoutil.NewAESGCMCipher(registry.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	var projects []string
	switch registry.Type {
	case model1.ImageRegistryTypeHarbor:
		var (
			authUser     string
			authPassword []byte
		)
		if registry.AuthUser != "" {
			authUser, authPassword = registry.AuthUser, block.Decrypt(registry.AuthPassword)
		}
		client, err := harbor.NewV2Client(registry.Endpoint, authUser, authPassword, false, registry.CACert)
		if err != nil {
			panic(err)
		}
		items, err := client.GetProjectsContext(ctx)
		if err != nil {
			return nil, errGatewayError(ctx, err)
		}
		for _, project := range items {
			projects = append(projects, project.Name)
		}
	}
	return projects, nil
}

// Organization is the resolver for the organization field.
func (r *imageRegistryResolver) Organization(ctx context.Context, obj *model.ImageRegistry) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Tags is the resolver for the tags field.
func (r *imageRepoResolver) Tags(ctx context.Context, obj *model.ImageRepo) ([]string, error) {
	if obj.ImageRegistry == nil {
		return nil, nil
	}

	registry := &model1.ImageRegistry{ID: obj.ImageRegistry.ID}
	if err := mapd.Database().Take(registry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}
	block, err := cryptoutil.NewAESGCMCipher(registry.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	var tags []string
	switch registry.Type {
	case model1.ImageRegistryTypeHarbor:
		var (
			authUser     string
			authPassword []byte
		)
		if registry.AuthUser != "" {
			authUser, authPassword = registry.AuthUser, block.Decrypt(registry.AuthPassword)
		}
		client, err := harbor.NewV2Client(registry.Endpoint, authUser, authPassword, false, registry.CACert)
		if err != nil {
			panic(err)
		}
		items, err := client.GetArtifactsContext(ctx, obj.Project, obj.Name)
		if err != nil {
			return nil, errGatewayError(ctx, err)
		}
		for _, image := range items {
			for _, tag := range image.Tags {
				tags = append(tags, tag.Name)
			}
		}
	}
	return tags, nil
}

// CreateImageRegistry is the resolver for the CreateImageRegistry field.
func (r *mutationResolver) CreateImageRegistry(ctx context.Context, input model.CreateImageRegistry) (*model.ImageRegistry, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}

	registry := &model1.ImageRegistry{ID: uuid.New(), Name: input.Name, OrganizationID: model1.NewNullUUID(input.Organization)}
	if input.Description != nil {
		registry.Description = *input.Description
	}
	registry.Endpoint = input.Endpoint.String()
	block, err := cryptoutil.NewAESGCMCipher([]byte(registry.ID[:])) // AES-128
	if err != nil {
		panic(err)
	}
	if input.AuthUser != nil && *input.AuthUser != "" {
		registry.AuthUser = *input.AuthUser

		if input.AuthPassword == nil || *input.AuthPassword == "" {
			return nil, errInvalidRequest(ctx, "input.authPassword", "value cannot be empty")
		}
		registry.AuthPassword = block.Encrypt([]byte(*input.AuthPassword))
	}
	if input.CaCert != nil && *input.CaCert != "" {
		rootcas := x509.NewCertPool()
		if !rootcas.AppendCertsFromPEM([]byte(*input.CaCert)) {
			return nil, errInvalidRequest(ctx, "input.caCert", "invalid x509 certificate")
		}
		registry.CACert = []byte(*input.CaCert)
	}
	registry.Type.FromString(input.Type.String())

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectImageRegistry, registry.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.Organization{}).Where("`id` = ?", registry.OrganizationID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists < 1 {
			return errObjectNotFound(ctx, "input.organization", registry.OrganizationID.String())
		}

		if err := tx.Model(&model1.ImageRegistry{}).Where("`organization_id` = ? AND `name` = ?", registry.OrganizationID, registry.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", registry.Name)
		}

		if input.Validate != nil && *input.Validate {
			var (
				authUser     string
				authPassword []byte
			)
			if registry.AuthUser != "" {
				authUser, authPassword = registry.AuthUser, block.Decrypt(registry.AuthPassword)
			}
			client, err := harbor.NewV2Client(registry.Endpoint, authUser, authPassword, false, registry.CACert)
			if err != nil {
				panic(err)
			}

			if _, err := client.GetCurrentUserContext(ctx); err != nil {
				return errGatewayError(ctx, err)
			}
		}

		if err := tx.Create(registry).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(registry).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if input.Dryrun != nil && *input.Dryrun {
			return errDryrun
		}

		if err := appendAuditLog(tx, "create", "ImageRegistry", registry.ID, registry.Name, sess.UserID, sess.UserName, registry.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertImageRegistryFromSQLModel(registry), nil
}

// UpdateImageRegistry is the resolver for the UpdateImageRegistry field.
func (r *mutationResolver) UpdateImageRegistry(ctx context.Context, input model.UpdateImageRegistry) (*model.ImageRegistry, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	registry := &model1.ImageRegistry{ID: input.ID}
	updates := make(map[string]interface{})
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Endpoint != nil {
		updates["endpoint"] = input.Endpoint.String()
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(registry).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", registry.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectImageRegistry, registry.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if input.Name != nil {
			if !validateName(*input.Name) {
				return errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), *input.Name)
			}
			updates["name"] = *input.Name

			var exists int64
			if err := tx.Model(&model1.ImageRegistry{}).Where("`organization_id` = ? AND `id` != ? AND `name` = ?", registry.OrganizationID, registry.ID, *input.Name).Count(&exists).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if exists > 0 {
				return errObjectDuplicated(ctx, "input.name", *input.Name)
			}
		}

		block, err := cryptoutil.NewAESGCMCipher([]byte(registry.ID[:])) // AES-128
		if err != nil {
			panic(err)
		}

		if input.AuthUser != nil {
			updates["auth_user"] = *input.AuthUser
			if *input.AuthUser == "" {
				updates["auth_password"] = block.Encrypt([]byte(""))
			}
		}
		if input.AuthPassword != nil {
			_, ok := updates["auth_password"]
			if !ok && registry.AuthUser != "" {
				updates["auth_password"] = block.Encrypt([]byte(*input.AuthPassword))
			}
		}
		if input.CaCert != nil {
			if *input.CaCert != "" {
				rootcas := x509.NewCertPool()
				if !rootcas.AppendCertsFromPEM([]byte(*input.CaCert)) {
					return errInvalidRequest(ctx, "input.caCert", "invalid x509 certificate")
				}
			}
			updates["ca_cert"] = *input.CaCert
		}

		if err := appendAuditLog(tx, "update", "ImageRegistry", registry.ID, registry.Name, sess.UserID, sess.UserName, registry.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(updates) > 0 {
			if err := tx.Model(registry).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(registry).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if input.Validate != nil && *input.Validate {
			var (
				authUser     string
				authPassword []byte
			)
			if registry.AuthUser != "" {
				authUser, authPassword = registry.AuthUser, block.Decrypt(registry.AuthPassword)
			}
			client, err := harbor.NewV2Client(registry.Endpoint, authUser, authPassword, false, registry.CACert)
			if err != nil {
				panic(err)
			}

			if _, err := client.GetCurrentUserContext(ctx); err != nil {
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
	return convertImageRegistryFromSQLModel(registry), nil
}

// DeleteImageRegistry is the resolver for the DeleteImageRegistry field.
func (r *mutationResolver) DeleteImageRegistry(ctx context.Context, input model.DeleteImageRegistry) (*model.ImageRegistry, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	registry := &model1.ImageRegistry{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(registry).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", input.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermModify, model1.PolicyObjectImageRegistry, registry.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.ClusterProvider{}).Where("`id` IN (SELECT `provider_id` FROM `cluster_provider_image_registry` WHERE `image_registry_id` = ?)", registry.ID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectIsBusy(ctx, "input.id", "ClusterProvider", nil, "", input.ID.String())
		}
		if err := tx.Model(&model1.Cluster{}).Where("`id` IN (SELECT `cluster_id` FROM `cluster_image_registry` WHERE `image_registry_id` = ?)", registry.ID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectIsBusy(ctx, "input.id", "Cluster", nil, "", input.ID.String())
		}

		if err := appendAuditLog(tx, "delete", "ImageRegistry", registry.ID, registry.Name, sess.UserID, sess.UserName, registry.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Delete(registry).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertImageRegistryFromSQLModel(registry), nil
}

// ImageRegistry is the resolver for the ImageRegistry field.
func (r *queryResolver) ImageRegistry(ctx context.Context, id uuid.UUID) (*model.ImageRegistry, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectImageRegistry)
	}

	registry := &model1.ImageRegistry{ID: id}
	if err := db.Take(registry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", id.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertImageRegistryFromSQLModel(registry), nil
}

// ImageRegistries is the resolver for the ImageRegistries field.
func (r *queryResolver) ImageRegistries(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.ImageRegistries, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.ImageRegistries{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.ImageRegistry{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectImageRegistry)
	}
	db = finalizeImageRegistryQueryFilter(db, obj.Filters...)
	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// ImageRepos is the resolver for the ImageRepos field.
func (r *queryResolver) ImageRepos(ctx context.Context, imageRegistry uuid.UUID, project string, filters []*model.QueryFilter) ([]*model.ImageRepo, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectImageRegistry)
	}

	registry := &model1.ImageRegistry{ID: imageRegistry}
	if err := db.Take(registry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "imageRegistry", imageRegistry.String())
		}
		panic(&ErrDBConnection{err})
	}
	block, err := cryptoutil.NewAESGCMCipher(registry.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	var name string
	for _, each := range filters {
		switch each.Key {
		case "name":
			for _, val := range each.Value {
				name = val
			}
		}
	}

	ctx, cancel := context.WithTimeout(ctx, mapd.GlobalConfig.HTTPClientTimeout)
	defer cancel()

	var repos []*model.ImageRepo
	switch registry.Type {
	case model1.ImageRegistryTypeHarbor:
		var (
			authUser     string
			authPassword []byte
		)
		if registry.AuthUser != "" {
			authUser, authPassword = registry.AuthUser, block.Decrypt(registry.AuthPassword)
		}
		client, err := harbor.NewV2Client(registry.Endpoint, authUser, authPassword, false, registry.CACert)
		if err != nil {
			panic(err)
		}

		items, err := client.GetRepositoriesContext(ctx, project)
		if err != nil {
			return nil, errGatewayError(ctx, err)
		}
		for _, repo := range items {
			ss := strings.Split(repo.Name, "/")
			obj := &model.ImageRepo{
				Name:          strings.Join(ss[1:], "/"),
				Project:       project,
				ImageRegistry: &model.ImageRegistry{ID: imageRegistry},
			}
			if name != "" && !strings.Contains(obj.Name, name) {
				continue
			}
			repos = append(repos, obj)
		}
	}
	return repos, nil
}

// ChartRepos is the resolver for the ChartRepos field.
func (r *queryResolver) ChartRepos(ctx context.Context, imageRegistry uuid.UUID, project string, filters []*model.QueryFilter) ([]*model.ChartRepo, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectImageRegistry)
	}

	registry := &model1.ImageRegistry{ID: imageRegistry}
	if err := db.Take(registry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "imageRegistry", imageRegistry.String())
		}
		panic(&ErrDBConnection{err})
	}
	block, err := cryptoutil.NewAESGCMCipher(registry.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	var name string
	for _, each := range filters {
		switch each.Key {
		case "name":
			for _, val := range each.Value {
				name = val
			}
		}
	}

	var repos []*model.ChartRepo
	switch registry.Type {
	case model1.ImageRegistryTypeHarbor:
		var (
			authUser     string
			authPassword []byte
		)
		if registry.AuthUser != "" {
			authUser, authPassword = registry.AuthUser, block.Decrypt(registry.AuthPassword)
		}
		client, err := harbor.NewV2Client(registry.Endpoint, authUser, authPassword, false, registry.CACert)
		if err != nil {
			panic(err)
		}
		items, err := client.GetChartReposContext(ctx, project)
		if err != nil {
			return nil, errGatewayError(ctx, err)
		}
		for _, repo := range items {
			obj := &model.ChartRepo{
				Name:          repo.Name,
				LatestVersion: repo.LatestVersion,
				TotalVersions: repo.TotalVersions,
				Deprecated:    repo.Deprecated,
				Project:       project,
				ImageRegistry: &model.ImageRegistry{ID: imageRegistry},
			}
			if repo.Icon != "" {
				v, err := url.Parse(repo.Icon)
				if err == nil {
					obj.Avatar = v
				}
			}
			if repo.Home != "" {
				v, err := url.Parse(repo.Home)
				if err == nil {
					obj.Homepage = v
				}
			}
			if name != "" && !strings.Contains(obj.Name, name) {
				continue
			}
			repos = append(repos, obj)
		}
	}
	return repos, nil
}

// ChartDetails is the resolver for the ChartDetails field.
func (r *queryResolver) ChartDetails(ctx context.Context, imageRegistry uuid.UUID, project string, name string, version string) (*model.ChartDetails, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectImageRegistry)
	}

	registry := &model1.ImageRegistry{ID: imageRegistry}
	if err := db.Take(registry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "imageRegistry", imageRegistry.String())
		}
		panic(&ErrDBConnection{err})
	}
	block, err := cryptoutil.NewAESGCMCipher(registry.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	switch registry.Type {
	case model1.ImageRegistryTypeHarbor:
		var (
			authUser     string
			authPassword []byte
		)
		if registry.AuthUser != "" {
			authUser, authPassword = registry.AuthUser, block.Decrypt(registry.AuthPassword)
		}
		client, err := harbor.NewV2Client(registry.Endpoint, authUser, authPassword, false, registry.CACert)
		if err != nil {
			panic(err)
		}
		chrt, err := client.GetChartContext(ctx, project, name, version)
		if err != nil {
			return nil, errGatewayError(ctx, err)
		}
		obj := &model.ChartDetails{
			Name:        chrt.Metadata.Name,
			Description: chrt.Metadata.Description,
			Version:     version,
			Keywords:    chrt.Metadata.Keywords,
			Files:       chrt.Files,
		}
		if chrt.Metadata.Icon != "" {
			v, err := url.Parse(chrt.Metadata.Icon)
			if err == nil {
				obj.Avatar = v
			}
		}
		for _, each := range chrt.Metadata.Sources {
			v, err := url.Parse(each)
			if err != nil {
				continue
			}
			obj.Sources = append(obj.Sources, v)
		}
		if chrt.Metadata.Home != "" {
			v, err := url.Parse(chrt.Metadata.Home)
			if err == nil {
				obj.Homepage = v
			}
		}
		if len(chrt.Metadata.URLs) > 0 {
			download := client.HelmChartRepo(project)
			path, err := url.Parse(chrt.Metadata.URLs[0])
			if err == nil {
				download.Path = filepath.Join(download.Path, path.Path)
				obj.Download = download
			}
		}
		for _, each := range chrt.Dependencies {
			dep := &model.ChartDetailsDependency{
				Name: each.Name,
			}
			if each.Repository != "" {
				dep.Repository = abstract.MustParseURL(each.Repository)
			}
			if each.Version != "" {
				dep.Version = util.StringPtr(each.Version)
			}
			obj.Dependencies = append(obj.Dependencies, dep)
		}
		b, err := json.Marshal(chrt.Values)
		if err != nil {
			panic(err)
		}
		obj.Values = b
		return obj, nil
	}
	return nil, nil
}

// ChartRepo returns graphql1.ChartRepoResolver implementation.
func (r *Resolver) ChartRepo() graphql1.ChartRepoResolver { return &chartRepoResolver{r} }

// ChartVersion returns graphql1.ChartVersionResolver implementation.
func (r *Resolver) ChartVersion() graphql1.ChartVersionResolver { return &chartVersionResolver{r} }

// ImageRegistries returns graphql1.ImageRegistriesResolver implementation.
func (r *Resolver) ImageRegistries() graphql1.ImageRegistriesResolver {
	return &imageRegistriesResolver{r}
}

// ImageRegistry returns graphql1.ImageRegistryResolver implementation.
func (r *Resolver) ImageRegistry() graphql1.ImageRegistryResolver { return &imageRegistryResolver{r} }

// ImageRepo returns graphql1.ImageRepoResolver implementation.
func (r *Resolver) ImageRepo() graphql1.ImageRepoResolver { return &imageRepoResolver{r} }

type chartRepoResolver struct{ *Resolver }
type chartVersionResolver struct{ *Resolver }
type imageRegistriesResolver struct{ *Resolver }
type imageRegistryResolver struct{ *Resolver }
type imageRepoResolver struct{ *Resolver }
