package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/avi"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil"
	"gorm.io/gorm"
)

// SeGroup is the resolver for the seGroup field.
func (r *aviLoadBalancerSpecResolver) SeGroup(ctx context.Context, obj *model.AviLoadBalancerSpec) (*string, error) {
	lb := &model1.LoadBalancer{ID: obj.LoadBalancer}
	if err := mapd.Database().Take(lb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	spec := &model1.LoadBalancerAVISpec{}
	if err := json.Unmarshal(lb.Spec, spec); err != nil {
		panic(err)
	}

	if spec.SEGroupID == "" {
		return nil, nil
	}

	block, err := cryptoutil.NewAESGCMCipher(lb.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	client, err := avi.NewAVI(lb.Endpoint, lb.AuthUser, block.Decrypt(lb.AuthPassword), lb.InsecureTLS, lb.CACert)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	seGroup, err := client.GetSEGroupContext(ctx, spec.SEGroupID)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}
	return &seGroup.Name, nil
}

// Clouds is the resolver for the clouds field.
func (r *aviLoadBalancerStatusResolver) Clouds(ctx context.Context, obj *model.AviLoadBalancerStatus, filters []*model.QueryFilter) ([]*model.AviCloud, error) {
	lb := &model1.LoadBalancer{ID: obj.ID}
	if err := mapd.Database().Take(lb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	block, err := cryptoutil.NewAESGCMCipher(lb.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	client, err := avi.NewAVI(lb.Endpoint, lb.AuthUser, block.Decrypt(lb.AuthPassword), lb.InsecureTLS, lb.CACert)
	if err != nil {
		panic(err)
	}

	options := &avi.ListCloudOptions{}
	for _, each := range filters {
		switch each.Key {
		case "name":
			for _, val := range each.Value {
				options.Name = val
			}
		}
	}

	clouds, err := listCloud(client, options)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	var result []*model.AviCloud
	for _, cloud := range clouds {
		result = append(result, &model.AviCloud{
			LoadBalancer: lb.ID,
			ID:           cloud.UUID,
			Name:         cloud.Name,
			Vtype:        cloud.VType,
		})
	}
	return result, nil
}

// NetworkProfiles is the resolver for the networkProfiles field.
func (r *aviLoadBalancerStatusResolver) NetworkProfiles(ctx context.Context, obj *model.AviLoadBalancerStatus, filters []*model.QueryFilter) ([]*model.AviNetworkProfile, error) {
	lb := &model1.LoadBalancer{ID: obj.ID}
	if err := mapd.Database().Take(lb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	block, err := cryptoutil.NewAESGCMCipher(lb.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	client, err := avi.NewAVI(lb.Endpoint, lb.AuthUser, block.Decrypt(lb.AuthPassword), lb.InsecureTLS, lb.CACert)
	if err != nil {
		panic(err)
	}

	selectedProtocols := make(map[string]bool)

	options := &avi.ListNetworkProfileOptions{}
	for _, each := range filters {
		switch each.Key {
		case "name":
			for _, val := range each.Value {
				options.Name = val
			}
		case "protocol":
			for _, val := range each.Value {
				protocol := strings.ToUpper(val)
				switch protocol {
				case "TCP", "UDP":
					selectedProtocols[protocol] = true
				}
			}
		}
	}

	profiles, err := listNetworkProfile(client, options)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	var protocolFilter func(profile *avi.NetworkProfile) bool
	switch len(selectedProtocols) {
	case 1:
		if selectedProtocols["TCP"] {
			protocolFilter = func(profile *avi.NetworkProfile) bool {
				return profile.Profile != nil && profile.Profile.TCPFastPathProfile != nil || profile.Profile.TCPProxyProfile != nil
			}
		} else {
			protocolFilter = func(profile *avi.NetworkProfile) bool {
				return profile.Profile != nil && profile.Profile.UDPFastPathProfile != nil || profile.Profile.UDPProxyProfile != nil
			}
		}
	default:
		protocolFilter = func(profile *avi.NetworkProfile) bool { return true }
	}

	var result []*model.AviNetworkProfile
	for _, profile := range profiles {
		if protocolFilter(profile) {
			result = append(result, &model.AviNetworkProfile{
				LoadBalancer: lb.ID,
				ID:           profile.UUID,
				Name:         profile.Name,
			})
		}
	}
	return result, nil
}

// Networks is the resolver for the networks field.
func (r *aviLoadBalancerStatusResolver) Networks(ctx context.Context, obj *model.AviLoadBalancerStatus, filters []*model.QueryFilter) ([]*model.AviNetwork, error) {
	lb := &model1.LoadBalancer{ID: obj.ID}
	if err := mapd.Database().Take(lb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	block, err := cryptoutil.NewAESGCMCipher(lb.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	client, err := avi.NewAVI(lb.Endpoint, lb.AuthUser, block.Decrypt(lb.AuthPassword), lb.InsecureTLS, lb.CACert)
	if err != nil {
		panic(err)
	}

	options := &avi.ListNetworkOptions{}
	for _, each := range filters {
		switch each.Key {
		case "name":
			for _, val := range each.Value {
				options.Name = val
			}
		case "cloud":
			for _, val := range each.Value {
				options.CloudName = val
			}
		}
	}

	networks, err := listNetwork(client, options)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	var result []*model.AviNetwork
	for _, network := range networks {
		n := &model.AviNetwork{
			LoadBalancer: lb.ID,
			ID:           network.UUID,
			Name:         network.Name,
		}

		for _, subnet := range network.ConfiguredSubnets {
			_, cidr, _ := net.ParseCIDR(fmt.Sprintf("%s/%d", subnet.Prefix.IPAddr.Addr, subnet.Prefix.Mask))

			sub := &model.AviSubnet{Cidr: cidr}
			for _, ipRange := range subnet.StaticIPRanges {
				switch ipRange.Type {
				case "STATIC_IPS_FOR_VIP", "STATIC_IPS_FOR_VIP_AND_SE":
					r := &model.AviSubnetIPRange{
						Begin: net.ParseIP(ipRange.Range.Begin.Addr),
						End:   net.ParseIP(ipRange.Range.End.Addr),
					}
					sub.StaticIPRanges = append(sub.StaticIPRanges, r)
				}
			}
			n.Subnets = append(n.Subnets, sub)
		}

		result = append(result, n)
	}
	return result, nil
}

// VsVips is the resolver for the vsVips field.
func (r *aviLoadBalancerStatusResolver) VsVips(ctx context.Context, obj *model.AviLoadBalancerStatus, filters []*model.QueryFilter) ([]*model.AviVsVip, error) {
	lb := &model1.LoadBalancer{ID: obj.ID}
	if err := mapd.Database().Take(lb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		panic(&ErrDBConnection{err})
	}

	block, err := cryptoutil.NewAESGCMCipher(lb.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	client, err := avi.NewAVI(lb.Endpoint, lb.AuthUser, block.Decrypt(lb.AuthPassword), lb.InsecureTLS, lb.CACert)
	if err != nil {
		panic(err)
	}

	options := &avi.ListVSVIPOptions{}
	for _, each := range filters {
		switch each.Key {
		case "name":
			for _, val := range each.Value {
				options.Name = val
			}
		case "cloud":
			for _, val := range each.Value {
				options.CloudName = val
			}
		}
	}

	vsVips, err := listVsVip(client, options)
	if err != nil {
		return nil, errGatewayError(ctx, err)
	}

	var result []*model.AviVsVip
	for _, vsVip := range vsVips {
		vv := &model.AviVsVip{
			LoadBalancer: lb.ID,
			ID:           vsVip.UUID,
			Name:         vsVip.Name,
		}

		for _, vip := range vsVip.VIP {
			if !vip.Enabled {
				continue
			}

			v := &model.AviVip{}
			// IPv4 as primary
			if vip.IPAddress != nil {
				v.IP = net.ParseIP(vip.IPAddress.Addr)
			} else if vip.IP6Address != nil {
				v.IP = net.ParseIP(vip.IP6Address.Addr)
			} else {
				continue
			}
			// IPv4 as primary
			if vip.Subnet != nil {
				_, cidr, _ := net.ParseCIDR(fmt.Sprintf("%s/%d", vip.Subnet.IPAddr.Addr, vip.Subnet.Mask))
				v.Subnet = cidr
			} else if vip.Subnet6 != nil {
				_, cidr, _ := net.ParseCIDR(fmt.Sprintf("%s/%d", vip.Subnet6.IPAddr.Addr, vip.Subnet6.Mask))
				v.Subnet = cidr
			} else {
				continue
			}
			vv.Vips = append(vv.Vips, v)
		}

		result = append(result, vv)
	}
	return result, nil
}

// TLSCertChain is the resolver for the tlsCertChain field.
func (r *loadBalancerResolver) TLSCertChain(ctx context.Context, obj *model.LoadBalancer) ([]*model.X509Certificate, error) {
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

// Organization is the resolver for the organization field.
func (r *loadBalancerResolver) Organization(ctx context.Context, obj *model.LoadBalancer) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Edge is the resolver for the edge field.
func (r *loadBalancersResolver) Edge(ctx context.Context, obj *model.LoadBalancers) (*model.LoadBalancerNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.LoadBalancerNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectLoadBalancer)
		}
		db = finalizeLoadBalancerQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var lbs []model1.LoadBalancer
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&lbs).Error; err != nil {
				return err
			}
			for _, lb := range lbs {
				nodes.Nodes = append(nodes.Nodes, convertLoadBalancerFromSQLModel(&lb))
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

// CreateLoadBalancer is the resolver for the CreateLoadBalancer field.
func (r *mutationResolver) CreateLoadBalancer(ctx context.Context, input model.CreateLoadBalancer) (*model.LoadBalancer, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}

	lb := &model1.LoadBalancer{ID: uuid.New(), Name: input.Name, OrganizationID: model1.NewNullUUID(input.Organization)}
	if input.Description != nil {
		lb.Description = *input.Description
	}
	if input.Type == nil || *input.Type == model.LoadBalancerTypeUnknown {
		return nil, errInvalidRequest(ctx, "input.type", "value cannot be empty")
	}
	lb.Type = model1.LoadBalancerType(input.Type.String())
	lb.Endpoint = input.Endpoint.String()
	if input.CaCert != nil {
		rootcas := x509.NewCertPool()
		if !rootcas.AppendCertsFromPEM([]byte(*input.CaCert)) {
			return nil, errInvalidRequest(ctx, "input.caCert", "invalid x509 certificate")
		}
		lb.CACert = []byte(*input.CaCert)
	}
	if input.InsecureSkipTLSVerify != nil {
		lb.InsecureTLS = *input.InsecureSkipTLSVerify
	}
	block, err := cryptoutil.NewAESGCMCipher(lb.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}
	if lb.Type != model1.LoadBalancerTypeExternal {
		if input.AuthUser == nil || *input.AuthUser == "" {
			return nil, errInvalidRequest(ctx, "input.authUser", "value cannot be empty")
		}
		if input.AuthPassword == nil || *input.AuthPassword == "" {
			return nil, errInvalidRequest(ctx, "input.authPassword", "value cannot be empty")
		}
		lb.AuthUser = *input.AuthUser
		lb.AuthPassword = block.Encrypt([]byte(*input.AuthPassword))
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectLoadBalancer, lb.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.Organization{}).Where("`id` = ?", lb.OrganizationID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists < 1 {
			return errObjectNotFound(ctx, "input.organization", lb.OrganizationID.String())
		}

		if err := tx.Model(&model1.LoadBalancer{}).Where("`organization_id` = ? AND `name` = ?", lb.OrganizationID, lb.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", lb.Name)
		}

		if input.Validate != nil && *input.Validate {
			switch lb.Type {
			case model1.LoadBalancerTypeAVI:
				client, err := avi.NewAVI(lb.Endpoint, lb.AuthUser, block.Decrypt(lb.AuthPassword), lb.InsecureTLS, lb.CACert)
				if err != nil {
					panic(err)
				}
				if err := client.PingContext(ctx); err != nil {
					return errGatewayError(ctx, err)
				}
			}
		}

		if input.Dryrun != nil && *input.Dryrun {
			switch lb.Type {
			case model1.LoadBalancerTypeAVI:
				client, err := avi.NewAVI(lb.Endpoint, lb.AuthUser, block.Decrypt(lb.AuthPassword), lb.InsecureTLS, lb.CACert)
				if err != nil {
					panic(err)
				}

				spec := &model1.LoadBalancerAVISpec{}
				if input.Spec != nil && input.Spec.Avi != nil && input.Spec.Avi.PreferredNetwork != nil {
					spec.PreferredNetwork = *input.Spec.Avi.PreferredNetwork
					if spec.PreferredNetwork != "" {
						networks, err := listNetwork(client, &avi.ListNetworkOptions{Name: spec.PreferredNetwork})
						if err != nil {
							return errGatewayError(ctx, err)
						}
						if len(networks) == 0 {
							return errObjectNotFound(ctx, "input.spec.avi.preferedNetwork", spec.PreferredNetwork)
						}
					}
				}

				lb.Spec, err = json.Marshal(spec)
				if err != nil {
					panic(err)
				}
			}
			now := time.Now()
			lb.CreatedAt = now
			lb.UpdatedAt = now
			return errDryrun
		}

		switch lb.Type {
		case model1.LoadBalancerTypeAVI:
			client, err := avi.NewAVI(lb.Endpoint, lb.AuthUser, block.Decrypt(lb.AuthPassword), lb.InsecureTLS, lb.CACert)
			if err != nil {
				panic(err)
			}

			spec, seGroupName := &model1.LoadBalancerAVISpec{}, defAVIServiceEngineGroup
			if input.Spec != nil && input.Spec.Avi != nil {
				if input.Spec.Avi.SeGroup != nil && *input.Spec.Avi.SeGroup != "" {
					seGroupName = *input.Spec.Avi.SeGroup
				}
				if input.Spec.Avi.PreferredNetwork != nil && *input.Spec.Avi.PreferredNetwork != "" {
					spec.PreferredNetwork = *input.Spec.Avi.PreferredNetwork
					if spec.PreferredNetwork != "" {
						networks, err := listNetwork(client, &avi.ListNetworkOptions{Name: spec.PreferredNetwork})
						if err != nil {
							return errGatewayError(ctx, err)
						}
						if len(networks) == 0 {
							return errObjectNotFound(ctx, "input.spec.avi.preferedNetwork", spec.PreferredNetwork)
						}
					}
				}
			}
			var seGroup *avi.ServiceEngineGroup
			seGroups, err := listSEGroup(client, &avi.ListSEGroupOptions{Name: seGroupName})
			if err != nil {
				return errGatewayError(ctx, err)
			}
			if len(seGroups) > 0 {
				seGroup = seGroups[0]
			} else {
				tier, err := getHighestAVILicenseTier(client, false)
				if err != nil {
					return errGatewayError(ctx, err)
				}
				seGroup, err = createServiceEngineGroup(client, seGroupName, tier)
				if err != nil {
					return errGatewayError(ctx, err)
				}
			}
			spec.SEGroupID, spec.LicenseTier = seGroup.UUID, model1.AVILicenseTier(seGroup.LicenseTier)

			lb.Spec, err = json.Marshal(spec)
			if err != nil {
				panic(err)
			}
		}

		if err := appendAuditLog(tx, "create", "LoadBalancer", lb.ID, lb.Name, sess.UserID, sess.UserName, lb.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Create(lb).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if err := tx.Take(lb).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertLoadBalancerFromSQLModel(lb), nil
}

// UpdateLoadBalancer is the resolver for the UpdateLoadBalancer field.
func (r *mutationResolver) UpdateLoadBalancer(ctx context.Context, input model.UpdateLoadBalancer) (*model.LoadBalancer, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	lb := &model1.LoadBalancer{ID: input.ID}
	updates := make(map[string]interface{})
	if input.Name != nil {
		if !validateName(*input.Name) {
			return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), *input.Name)
		}
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Endpoint != nil {
		updates["endpoint"] = input.Endpoint.String()
	}
	if input.CaCert != nil {
		rootcas := x509.NewCertPool()
		if !rootcas.AppendCertsFromPEM([]byte(*input.CaCert)) {
			return nil, errInvalidRequest(ctx, "input.caCert", "invalid x509 certificate")
		}
		updates["ca_cert"] = *input.CaCert
	}
	if input.InsecureSkipTLSVerify != nil {
		updates["insecure_tls"] = *input.InsecureSkipTLSVerify
	}
	block, err := cryptoutil.NewAESGCMCipher(lb.ID[:]) // AES-128
	if err != nil {
		panic(err)
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(lb).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", lb.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectLoadBalancer, lb.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if input.Name != nil {
			var exists int64
			if err := tx.Model(&model1.LoadBalancer{}).Where("`organization_id` = ? AND `id` != ? AND `name` = ?", lb.OrganizationID, lb.ID, *input.Name).Count(&exists).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if exists > 0 {
				return errObjectDuplicated(ctx, "input.name", *input.Name)
			}
		}
		if input.AuthUser != nil && lb.Type != model1.LoadBalancerTypeExternal {
			updates["auth_user"] = *input.AuthUser
		}
		if input.AuthPassword != nil && lb.Type != model1.LoadBalancerTypeExternal {
			updates["auth_password"] = block.Encrypt([]byte(*input.AuthPassword))
		}

		if input.Spec != nil {
			switch lb.Type {
			case model1.LoadBalancerTypeAVI:
				if input.Spec.Avi != nil {
					spec := &model1.LoadBalancerAVISpec{}
					if err := json.Unmarshal(lb.Spec, spec); err != nil {
						panic(err)
					}

					if input.Spec.Avi.PreferredNetwork != nil {
						if *input.Spec.Avi.PreferredNetwork != "" {
							client, err := avi.NewAVI(lb.Endpoint, lb.AuthUser, block.Decrypt(lb.AuthPassword), lb.InsecureTLS, lb.CACert)
							if err != nil {
								panic(err)
							}

							networks, err := listNetwork(client, &avi.ListNetworkOptions{Name: *input.Spec.Avi.PreferredNetwork})
							if err != nil {
								return errGatewayError(ctx, err)
							}
							if len(networks) == 0 {
								return errObjectNotFound(ctx, "input.spec.avi.preferedNetwork", *input.Spec.Avi.PreferredNetwork)
							}
						}
						spec.PreferredNetwork = *input.Spec.Avi.PreferredNetwork
					}

					b, err := json.Marshal(spec)
					if err != nil {
						panic(err)
					}
					updates["spec"] = string(b)
				}
			}
		}

		if len(updates) > 0 {
			if err := tx.Model(lb).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(lb).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		if input.Validate != nil && *input.Validate {
			switch lb.Type {
			case model1.LoadBalancerTypeAVI:
				client, err := avi.NewAVI(lb.Endpoint, lb.AuthUser, block.Decrypt(lb.AuthPassword), lb.InsecureTLS, lb.CACert)
				if err != nil {
					panic(err)
				}
				if err := client.PingContext(ctx); err != nil {
					return errGatewayError(ctx, err)
				}
			}
		}
		if input.Dryrun != nil && *input.Dryrun {
			return errDryrun
		}

		if err := appendAuditLog(tx, "update", "LoadBalancer", lb.ID, lb.Name, sess.UserID, sess.UserName, lb.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertLoadBalancerFromSQLModel(lb), nil
}

// DeleteLoadBalancer is the resolver for the DeleteLoadBalancer field.
func (r *mutationResolver) DeleteLoadBalancer(ctx context.Context, input model.DeleteLoadBalancer) (*model.LoadBalancer, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	lb := &model1.LoadBalancer{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Take(lb).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", lb.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermModify, model1.PolicyObjectLoadBalancer, lb.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		var exists int64
		if err := tx.Model(&model1.ClusterProvider{}).Where("`load_balancer_id` = ?", lb.ID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectIsBusy(ctx, "input.id", "Supervisor", nil, "", input.ID.String())
		}
		if err := tx.Model(&model1.Route{}).Where("`load_balancer_id` = ?", lb.ID).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectIsBusy(ctx, "input.id", "Route", nil, "", input.ID.String())
		}

		if err := appendAuditLog(tx, "delete", "LoadBalancer", lb.ID, lb.Name, sess.UserID, sess.UserName, lb.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Delete(lb).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	return convertLoadBalancerFromSQLModel(lb), nil
}

// LoadBalancer is the resolver for the LoadBalancer field.
func (r *queryResolver) LoadBalancer(ctx context.Context, id uuid.UUID) (*model.LoadBalancer, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectLoadBalancer)
	}

	lb := &model1.LoadBalancer{ID: id}
	if err := db.Take(lb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", id.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertLoadBalancerFromSQLModel(lb), nil
}

// LoadBalancers is the resolver for the LoadBalancers field.
func (r *queryResolver) LoadBalancers(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.LoadBalancers, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.LoadBalancers{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.LoadBalancer{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectLoadBalancer)
	}
	db = finalizeLoadBalancerQueryFilter(db, obj.Filters...)

	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// AviLoadBalancerSpec returns graphql1.AviLoadBalancerSpecResolver implementation.
func (r *Resolver) AviLoadBalancerSpec() graphql1.AviLoadBalancerSpecResolver {
	return &aviLoadBalancerSpecResolver{r}
}

// AviLoadBalancerStatus returns graphql1.AviLoadBalancerStatusResolver implementation.
func (r *Resolver) AviLoadBalancerStatus() graphql1.AviLoadBalancerStatusResolver {
	return &aviLoadBalancerStatusResolver{r}
}

// LoadBalancer returns graphql1.LoadBalancerResolver implementation.
func (r *Resolver) LoadBalancer() graphql1.LoadBalancerResolver { return &loadBalancerResolver{r} }

// LoadBalancers returns graphql1.LoadBalancersResolver implementation.
func (r *Resolver) LoadBalancers() graphql1.LoadBalancersResolver { return &loadBalancersResolver{r} }

type aviLoadBalancerSpecResolver struct{ *Resolver }
type aviLoadBalancerStatusResolver struct{ *Resolver }
type loadBalancerResolver struct{ *Resolver }
type loadBalancersResolver struct{ *Resolver }
