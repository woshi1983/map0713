package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/identicon"
	"gorm.io/gorm"
	"helm.sh/helm/v3/pkg/chart/loader"
)

// Application is the resolver for the application field.
func (r *componentResolver) Application(ctx context.Context, obj *model.Component) (*model.Application, error) {
	if obj.Application == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Application(WithGuardSilenced(ctx), obj.Application.ID)
}

// Organization is the resolver for the organization field.
func (r *componentResolver) Organization(ctx context.Context, obj *model.Component) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *componentsResolver) Edge(ctx context.Context, obj *model.Components) (*model.ComponentNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.ComponentNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectComponent)
		}
		db = finalizeComponentQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var coms []model1.Component
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&coms).Error; err != nil {
				return err
			}
			for _, com := range coms {
				nodes.Nodes = append(nodes.Nodes, convertComponentFromSQLModel(&com))
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

// Chart is the resolver for the chart field.
func (r *helmComponentTemplateResolver) Chart(ctx context.Context, obj *model.HelmComponentTemplate) (*model.ChartDetails, error) {
	endpoint := &url.URL{
		Scheme: obj.URL.Scheme,
		Opaque: obj.URL.Opaque,
		User:   obj.URL.User,
		Host:   obj.URL.Host,
	}
	caCert, err := getCACertificateFromMatchedImageRegistries(mapd.Database(), endpoint.String())
	if err != nil {
		panic(&ErrDBConnection{err})
	}
	fpath, err := r.ChartCache.Get(ctx, obj.URL.String(), httputil.TrustRootCACerts(caCert))
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			panic(err)
		}
		return nil, errGatewayError(ctx, err)
	}
	chrt, err := loader.Load(fpath)
	if err != nil {
		panic(err)
	}
	v := convertChartDetailsFromQualifiedChart(chrt)
	v.Download = obj.URL
	return v, nil
}

// CreateComponent is the resolver for the CreateComponent field.
func (r *mutationResolver) CreateComponent(ctx context.Context, input model.CreateComponent) (*model.Component, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}

	com := &model1.Component{ID: uuid.New(), Name: input.Name, ApplicationID: model1.NewNullUUID(input.Application)}
	if input.Description != nil {
		com.Description = *input.Description
	}
	if input.Avatar != nil {
		com.Avatar = input.Avatar.String()
	}
	if com.Avatar == "" {
		com.Avatar = identicon.NewDataURL(com.ID[:]).String()
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		app := &model1.Application{ID: com.ApplicationID.UUID}
		if err := tx.Take(app).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.application", app.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectComponent, app.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		com.OrganizationID = app.OrganizationID

		var exists int64
		if err := tx.Model(&model1.Component{}).Where("`organization_id` = ? AND `name` = ?", com.OrganizationID, com.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", com.Name)
		}

		release := &model1.Release{
			ID:             uuid.New(),
			Name:           com.Name,
			Num:            1,
			ComponentID:    model1.NewNullUUID(com.ID),
			ApplicationID:  com.ApplicationID,
			OrganizationID: com.OrganizationID,
		}

		switch {
		case input.Template.Basic != nil:
			com.Type = model1.ComponentTypeBasic
			template := input.Template.Basic
			if len(template.Deployments) < 1 {
				return errInvalidRequest(ctx, "input.template.basic.deployments", "value cannot be empty")
			}

			deploymentsM := make(map[string]*model.AddBasicComponentTemplateDeployment)
			for i, each := range template.Deployments {
				if !validateName(each.Name) {
					return errInvalidRequest(ctx, fmt.Sprintf("input.template.basic.deployments.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
				}

				// this prevents empty label
				labels := make(map[string]string)
				for k, v := range each.Labels {
					labels[k] = v
				}
				labels[defLabelComponent] = com.ID.String()
				each.Labels = labels

				deploymentsM[each.Name] = each
			}
			template.Deployments = template.Deployments[:0]
			for _, each := range deploymentsM {
				template.Deployments = append(template.Deployments, each)
			}

			servicesM := make(map[string]*model.AddBasicComponentTemplateService)
			for i, each := range template.Services {
				if !validateName(each.Name) {
					return errInvalidRequest(ctx, fmt.Sprintf("input.template.basic.services.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
				}

				// this prevents empty label
				labels := make(map[string]string)
				for k, v := range each.Labels {
					labels[k] = v
				}
				labels[defLabelComponent] = com.ID.String()
				each.Labels = labels

				servicesM[each.Name] = each
			}
			template.Services = template.Services[:0]
			for _, each := range servicesM {
				template.Services = append(template.Services, each)
			}

			secretsM := make(map[string]*model.AddBasicComponentTemplateSecret)
			for i, each := range template.Secrets {
				if !validateName(each.Name) {
					return errInvalidRequest(ctx, fmt.Sprintf("input.template.basic.secrets.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
				}
				if each.Generic == nil && each.DockerConfigJSON == nil && each.BasicAuth == nil && each.SSHAuth == nil && each.TLS == nil {
					return errInvalidRequest(ctx, fmt.Sprintf("input.template.basic.secrets.%d.generic", i), "value cannot be empty")
				}

				// this prevents empty label
				labels := make(map[string]string)
				for k, v := range each.Labels {
					labels[k] = v
				}
				labels[defLabelComponent] = com.ID.String()
				each.Labels = labels

				secretsM[each.Name] = each
			}
			template.Secrets = template.Secrets[:0]
			for _, each := range secretsM {
				template.Secrets = append(template.Secrets, each)
			}

			configMapM := make(map[string]*model.AddBasicComponentTemplateConfigMap)
			for i, each := range template.ConfigMaps {
				if !validateName(each.Name) {
					return errInvalidRequest(ctx, fmt.Sprintf("input.template.basic.configMaps.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
				}

				// this prevents empty label
				labels := make(map[string]string)
				for k, v := range each.Labels {
					labels[k] = v
				}
				labels[defLabelComponent] = com.ID.String()
				each.Labels = labels

				configMapM[each.Name] = each
			}
			template.ConfigMaps = template.ConfigMaps[:0]
			for _, each := range configMapM {
				template.ConfigMaps = append(template.ConfigMaps, each)
			}

			persistentVolumeClaimM := make(map[string]*model.AddBasicComponentTemplatePersistentVolumeClaim)
			for i, each := range template.PersistentVolumeClaims {
				if !validateName(each.Name) {
					return errInvalidRequest(ctx, fmt.Sprintf("input.template.basic.persistentVolumeClaims.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
				}

				// this prevents empty label
				labels := make(map[string]string)
				for k, v := range each.Labels {
					labels[k] = v
				}
				labels[defLabelComponent] = com.ID.String()
				each.Labels = labels

				persistentVolumeClaimM[each.Name] = each
			}
			template.PersistentVolumeClaims = template.PersistentVolumeClaims[:0]
			for _, each := range persistentVolumeClaimM {
				template.PersistentVolumeClaims = append(template.PersistentVolumeClaims, each)
			}

			b, err := json.Marshal(template)
			if err != nil {
				panic(err)
			}
			com.Template = b

			release.Manifest, err = renderBasicComponentManifest(template)
			if err != nil {
				return err
			}

		case input.Template.Helm != nil:
			com.Type = model1.ComponentTypeHelm
			values := make(map[string]interface{})
			if err := json.Unmarshal(input.Template.Helm.Values, &values); err != nil {
				return errInvalidRequest(ctx, "input.template.helm.values", "value is not a valid object")
			}

			template := &model1.ComponentHelmTemplate{
				URL:       input.Template.Helm.URL.String(),
				Values:    input.Template.Helm.Values,
				ExtraInfo: input.Template.Helm.ExtraInfo,
			}
			b, err := json.Marshal(template)
			if err != nil {
				panic(err)
			}
			com.Template = b

			endpoint := &url.URL{
				Scheme: input.Template.Helm.URL.Scheme,
				Opaque: input.Template.Helm.URL.Opaque,
				User:   input.Template.Helm.URL.User,
				Host:   input.Template.Helm.URL.Host,
			}
			caCert, err := getCACertificateFromMatchedImageRegistries(tx, endpoint.String())
			if err != nil {
				panic(&ErrDBConnection{err})
			}
			fpath, err := r.ChartCache.Get(ctx, template.URL, httputil.TrustRootCACerts(caCert))
			if err != nil {
				if _, ok := err.(*os.PathError); ok {
					panic(err)
				}
				return errGatewayError(ctx, err)
			}
			release.CRDs, release.Manifest, err = renderChart(com.Name, fpath, values)
			if err != nil {
				panic(err)
			}

		case input.Template.Raw != nil:
			com.Type = model1.ComponentTypeRaw
			manifest := []byte(*input.Template.Raw)
			b, err := validateManifestAndConvertToTemplate(manifest)
			if err != nil {
				return errInvalidRequest(ctx, "input.template.raw", err.Error())
			}

			com.Template = b
			release.Manifest = b
		default:
			return errInvalidRequest(ctx, "input.template", "value cannot be empty")
		}

		if err := appendAuditLog(tx, "create", "Component", com.ID, com.Name, sess.UserID, sess.UserName, com.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		fed := &model1.InstanceFederation{
			ID:             uuid.New(),
			Name:           com.Name,
			Description:    com.Description,
			Status:         model1.FederationStatusProvisioned,
			ComponentID:    model1.NewNullUUID(com.ID),
			ApplicationID:  com.ApplicationID,
			OrganizationID: com.OrganizationID,
		}

		if err := tx.Create(com).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Create(fed).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Create(release).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(com).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertComponentFromSQLModel(com), nil
}

// UpdateComponent is the resolver for the UpdateComponent field.
func (r *mutationResolver) UpdateComponent(ctx context.Context, input model.UpdateComponent) (*model.Component, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	com := &model1.Component{ID: input.ID}
	updates := make(map[string]interface{})
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Avatar != nil {
		updates["avatar"] = input.Avatar.String()
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(com).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", com.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectComponent, com.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if input.Template != nil {
			switch {
			case input.Template.Basic != nil:
				if com.Type != model1.ComponentTypeBasic {
					return errInvalidRequest(ctx, "input.template.basic", "cannot modify type of Component")
				}

				template := &model.CreateBasicComponentTemplate{}
				if err := json.Unmarshal(com.Template, template); err != nil {
					panic(err)
				}

				deploymentsM := make(map[string]*model.AddBasicComponentTemplateDeployment)
				for _, each := range template.Deployments {
					deploymentsM[each.Name] = each
				}
				for _, each := range input.Template.Basic.RemoveDeployments {
					delete(deploymentsM, each.Name)
				}
				for i, each := range input.Template.Basic.AddDeployments {
					if !validateName(each.Name) {
						return errInvalidRequest(ctx, fmt.Sprintf("input.template.basic.addDeployments.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
					}

					// this prevents empty label
					labels := make(map[string]string)
					for k, v := range each.Labels {
						labels[k] = v
					}
					labels[defLabelComponent] = com.ID.String()
					each.Labels = labels

					deploymentsM[each.Name] = each
				}
				template.Deployments = template.Deployments[:0]
				for _, each := range deploymentsM {
					template.Deployments = append(template.Deployments, each)
				}

				servicesM := make(map[string]*model.AddBasicComponentTemplateService)
				for _, each := range template.Services {
					servicesM[each.Name] = each
				}
				for _, each := range input.Template.Basic.RemoveServices {
					delete(servicesM, each.Name)
				}
				for i, each := range input.Template.Basic.AddServices {
					if !validateName(each.Name) {
						return errInvalidRequest(ctx, fmt.Sprintf("input.template.basic.addServices.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
					}

					// this prevents empty label
					labels := make(map[string]string)
					for k, v := range each.Labels {
						labels[k] = v
					}
					labels[defLabelComponent] = com.ID.String()
					each.Labels = labels

					servicesM[each.Name] = each
				}
				template.Services = template.Services[:0]
				for _, each := range servicesM {
					template.Services = append(template.Services, each)
				}

				secretsM := make(map[string]*model.AddBasicComponentTemplateSecret)
				for _, each := range template.Secrets {
					secretsM[each.Name] = each
				}
				for _, each := range input.Template.Basic.RemoveSecrets {
					delete(secretsM, each.Name)
				}
				for i, each := range input.Template.Basic.AddSecrets {
					if !validateName(each.Name) {
						return errInvalidRequest(ctx, fmt.Sprintf("input.template.basic.addSecrets.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
					}
					if each.Generic == nil && each.DockerConfigJSON == nil && each.BasicAuth == nil && each.SSHAuth == nil && each.TLS == nil {
						return errInvalidRequest(ctx, fmt.Sprintf("input.template.basic.addSecrets.%d.generic", i), "value cannot be empty")
					}

					// this prevents empty label
					labels := make(map[string]string)
					for k, v := range each.Labels {
						labels[k] = v
					}
					labels[defLabelComponent] = com.ID.String()
					each.Labels = labels

					secretsM[each.Name] = each
				}
				template.Secrets = template.Secrets[:0]
				for _, each := range secretsM {
					template.Secrets = append(template.Secrets, each)
				}

				configMapM := make(map[string]*model.AddBasicComponentTemplateConfigMap)
				for _, each := range template.ConfigMaps {
					configMapM[each.Name] = each
				}
				for _, each := range input.Template.Basic.RemoveConfigMaps {
					delete(configMapM, each.Name)
				}
				for i, each := range input.Template.Basic.AddConfigMaps {
					if !validateName(each.Name) {
						return errInvalidRequest(ctx, fmt.Sprintf("input.template.basic.addConfigMaps.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
					}

					// this prevents empty label
					labels := make(map[string]string)
					for k, v := range each.Labels {
						labels[k] = v
					}
					labels[defLabelComponent] = com.ID.String()
					each.Labels = labels

					configMapM[each.Name] = each
				}
				template.ConfigMaps = template.ConfigMaps[:0]
				for _, each := range configMapM {
					template.ConfigMaps = append(template.ConfigMaps, each)
				}

				persistentVolumeClaimM := make(map[string]*model.AddBasicComponentTemplatePersistentVolumeClaim)
				for _, each := range template.PersistentVolumeClaims {
					persistentVolumeClaimM[each.Name] = each
				}
				for _, each := range input.Template.Basic.RemovePersistentVolumeClaims {
					delete(persistentVolumeClaimM, each.Name)
				}
				for i, each := range input.Template.Basic.AddPersistentVolumeClaims {
					if !validateName(each.Name) {
						return errInvalidRequest(ctx, fmt.Sprintf("input.template.basic.addPersistentVolumeClaims.%d.name", i), fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), each.Name)
					}

					// this prevents empty label
					labels := make(map[string]string)
					for k, v := range each.Labels {
						labels[k] = v
					}
					labels[defLabelComponent] = com.ID.String()
					each.Labels = labels

					persistentVolumeClaimM[each.Name] = each
				}
				template.PersistentVolumeClaims = template.PersistentVolumeClaims[:0]
				for _, each := range persistentVolumeClaimM {
					template.PersistentVolumeClaims = append(template.PersistentVolumeClaims, each)
				}

				b, err := json.Marshal(template)
				if err != nil {
					panic(err)
				}
				if string(com.Template) != string(b) {
					updates["template"] = string(b)

					var cnt int64
					if err := tx.Model(&model1.Release{}).Where("`component_id` = ?", com.ID).Count(&cnt).Error; err != nil {
						panic(&ErrDBConnection{err})
					}

					release := &model1.Release{
						ID:             uuid.New(),
						Name:           com.Name,
						Num:            cnt + 1,
						ComponentID:    model1.NewNullUUID(com.ID),
						ApplicationID:  com.ApplicationID,
						OrganizationID: com.OrganizationID,
					}

					release.Manifest, err = renderBasicComponentManifest(template)
					if err != nil {
						return err
					}

					if err := tx.Create(release).Error; err != nil {
						panic(&ErrDBConnection{err})
					}
				}
			case input.Template.Helm != nil:
				if com.Type != model1.ComponentTypeHelm {
					return errInvalidRequest(ctx, "input.template.helm", "cannot modify type of Component")
				}

				template := &model1.ComponentHelmTemplate{}
				if err := json.Unmarshal(com.Template, template); err != nil {
					panic(err)
				}

				if input.Template.Helm.URL != nil {
					template.URL = input.Template.Helm.URL.String()
				}
				if input.Template.Helm.Values != nil {
					template.Values = input.Template.Helm.Values
				}
				if input.Template.Helm.ExtraInfo != nil {
					template.ExtraInfo = input.Template.Helm.ExtraInfo
				}
				values := make(map[string]interface{})
				if err := json.Unmarshal(template.Values, &values); err != nil {
					return errInvalidRequest(ctx, "input.template.helm.values", "value is not a valid object")
				}
				b, err := json.Marshal(template)
				if err != nil {
					panic(err)
				}
				if string(com.Template) != string(b) {
					updates["template"] = string(b)

					var cnt int64
					if err := tx.Model(&model1.Release{}).Where("`component_id` = ?", com.ID).Count(&cnt).Error; err != nil {
						panic(&ErrDBConnection{err})
					}

					release := &model1.Release{
						ID:             uuid.New(),
						Name:           com.Name,
						Num:            cnt + 1,
						ComponentID:    model1.NewNullUUID(com.ID),
						ApplicationID:  com.ApplicationID,
						OrganizationID: com.OrganizationID,
					}

					templateURL, err := url.Parse(template.URL)
					if err != nil {
						panic(fmt.Errorf("error parsing template URL: %v", err))
					}
					endpoint := &url.URL{
						Scheme: templateURL.Scheme,
						Opaque: templateURL.Opaque,
						User:   templateURL.User,
						Host:   templateURL.Host,
					}
					caCert, err := getCACertificateFromMatchedImageRegistries(tx, endpoint.String())
					if err != nil {
						panic(&ErrDBConnection{err})
					}
					fpath, err := r.ChartCache.Get(ctx, template.URL, httputil.TrustRootCACerts(caCert))
					if err != nil {
						if _, ok := err.(*os.PathError); ok {
							panic(err)
						}
						return errGatewayError(ctx, err)
					}
					release.CRDs, release.Manifest, err = renderChart(com.Name, fpath, values)
					if err != nil {
						panic(err)
					}
					if err := tx.Create(release).Error; err != nil {
						panic(&ErrDBConnection{err})
					}
				}

			case input.Template.Raw != nil:
				if com.Type != model1.ComponentTypeRaw {
					return errInvalidRequest(ctx, "input.template.raw", "cannot modify type of Component")
				}

				manifest := []byte(*input.Template.Raw)
				b, err := validateManifestAndConvertToTemplate(manifest)
				if err != nil {
					return errInvalidRequest(ctx, "input.template.raw", err.Error())
				}
				oldManifest, err := validateManifestAndConvertToTemplate(com.Template)
				if err != nil {
					panic(err)
				}

				if string(b) != string(oldManifest) {
					updates["template"] = string(b)

					var cnt int64
					if err := tx.Model(&model1.Release{}).Where("`component_id` = ?", com.ID).Count(&cnt).Error; err != nil {
						panic(&ErrDBConnection{err})
					}

					release := &model1.Release{
						ID:             uuid.New(),
						Name:           com.Name,
						Num:            cnt + 1,
						Manifest:       manifest,
						ComponentID:    model1.NewNullUUID(com.ID),
						ApplicationID:  com.ApplicationID,
						OrganizationID: com.OrganizationID,
					}
					if err := tx.Create(release).Error; err != nil {
						panic(&ErrDBConnection{err})
					}
				}
			}
		}

		if err := appendAuditLog(tx, "update", "Component", com.ID, com.Name, sess.UserID, sess.UserName, com.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(updates) > 0 {
			if err := tx.Model(com).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(com).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertComponentFromSQLModel(com), nil
}

// DeleteComponent is the resolver for the DeleteComponent field.
func (r *mutationResolver) DeleteComponent(ctx context.Context, input model.DeleteComponent) (*model.Component, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	com := &model1.Component{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(com).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", com.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermModify, model1.PolicyObjectComponent, com.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.Instance{}).Where("`component_id` = ?", com.ID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectIsBusy(ctx, "input.id", "Instance", nil, "", input.ID.String())
		}

		if err := tx.Where("`component_id` = ?", com.ID).Delete(&model1.InstanceFederation{}).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Where("`component_id` = ?", com.ID).Delete(&model1.ClusterBackupScheduleComponent{}).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := appendAuditLog(tx, "delete", "Component", com.ID, com.Name, sess.UserID, sess.UserName, com.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Delete(com).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertComponentFromSQLModel(com), nil
}

// Component is the resolver for the Component field.
func (r *queryResolver) Component(ctx context.Context, id uuid.UUID) (*model.Component, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectComponent)
	}

	com := &model1.Component{ID: id}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(com).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "id", id.String())
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
	return convertComponentFromSQLModel(com), nil
}

// Components is the resolver for the Components field.
func (r *queryResolver) Components(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.Components, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.Components{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.Component{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectComponent)
	}
	db = finalizeComponentQueryFilter(db, obj.Filters...)

	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// Component returns graphql1.ComponentResolver implementation.
func (r *Resolver) Component() graphql1.ComponentResolver { return &componentResolver{r} }

// Components returns graphql1.ComponentsResolver implementation.
func (r *Resolver) Components() graphql1.ComponentsResolver { return &componentsResolver{r} }

// HelmComponentTemplate returns graphql1.HelmComponentTemplateResolver implementation.
func (r *Resolver) HelmComponentTemplate() graphql1.HelmComponentTemplateResolver {
	return &helmComponentTemplateResolver{r}
}

type componentResolver struct{ *Resolver }
type componentsResolver struct{ *Resolver }
type helmComponentTemplateResolver struct{ *Resolver }
