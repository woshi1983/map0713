package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/avi"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/cryptoutil"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/k8s"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
)

// NetworkProfile is the resolver for the networkProfile field.
func (r *aviRouteSpecResolver) NetworkProfile(ctx context.Context, obj *model.AviRouteSpec) (string, error) {
	if obj.LoadBalancer == uuid.Nil || obj.NetworkProfileID == "" {
		return "", nil
	}

	lb := &model1.LoadBalancer{ID: obj.LoadBalancer}
	if err := mapd.Database().Take(lb).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
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
	profile, err := client.GetNetworkProfile(obj.NetworkProfileID)
	if err != nil {
		return "", errGatewayError(ctx, err)
	}
	return profile.Name, nil
}

// CreateRoute is the resolver for the CreateRoute field.
func (r *mutationResolver) CreateRoute(ctx context.Context, input model.CreateRoute) (*model.Route, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !validateName(input.Name) {
		return nil, errInvalidRequest(ctx, "input.name", fmt.Sprintf("regexp '%s' is not satisfied", regexpName.String()), input.Name)
	}
	if input.Port < 1 || input.Port > 65535 {
		return nil, errInvalidRequest(ctx, "input.port", "valid port range: 1 - 65535", input.Port)
	}

	route := &model1.Route{
		ID:             uuid.New(),
		Name:           input.Name,
		Port:           input.Port,
		ProvisionPhase: model1.RouteProvisionPrepare,
		ComponentID:    model1.NewNullUUID(input.Component),
		LoadBalancerID: model1.NewNullUUID(input.LoadBalancer),
	}
	if input.Description != nil {
		route.Description = *input.Description
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		component := &model1.Component{ID: route.ComponentID.UUID}
		if err := tx.Take(component).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.component", component.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectRoute, component.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		loadBalancer := &model1.LoadBalancer{ID: route.LoadBalancerID.UUID}
		if err := tx.Take(loadBalancer).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.loadBalancer", loadBalancer.ID.String())
			}
			panic(&ErrDBConnection{err})
		}
		if loadBalancer.Type == model1.LoadBalancerTypeExternal {
			return errInvalidRequest(ctx, "input.loadBalancer", "route can not be managed on external load balancer", loadBalancer.ID.String())
		}
		if component.OrganizationID.String() != loadBalancer.OrganizationID.String() {
			return errInvalidRequest(ctx, "input.loadBalancer", "different organization from `input.federation`", loadBalancer.ID.String())
		}
		route.Type = loadBalancer.Type
		if input.Protocol != nil {
			route.Protocol = model1.RouteProtocol(input.Protocol.String())
		} else {
			route.Protocol = model1.RouteProtocolTCP
		}
		route.ApplicationID = component.ApplicationID
		route.OrganizationID = component.OrganizationID

		var exists int64
		if err := tx.Model(&model1.Route{}).Where("`organization_id` = ? AND `name` = ?", route.OrganizationID, route.Name).Count(&exists).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		if exists > 0 {
			return errObjectDuplicated(ctx, "input.name", route.Name)
		}

		switch loadBalancer.Type {
		case model1.LoadBalancerTypeAVI:
			spec := &model1.RouteAVISpec{
				LBAlgorithm:                "LB_ALGORITHM_LEAST_CONNECTIONS",
				LBAlgorithmHash:            "LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS",
				LBAlgorithmCoreNonaffinity: 2,
			}
			if input.Spec != nil && input.Spec.Avi != nil {
				block, err := cryptoutil.NewAESGCMCipher(loadBalancer.ID[:]) // AES-128
				if err != nil {
					panic(err)
				}

				client, err := avi.NewAVI(loadBalancer.Endpoint, loadBalancer.AuthUser, block.Decrypt(loadBalancer.AuthPassword), loadBalancer.InsecureTLS, loadBalancer.CACert)
				if err != nil {
					panic(err)
				}

				if input.Spec.Avi.NetworkProfile != nil && *input.Spec.Avi.NetworkProfile != "" {
					listNetworkProfileOptions := &avi.ListNetworkProfileOptions{Name: *input.Spec.Avi.NetworkProfile}
					profiles, err := listNetworkProfile(client, listNetworkProfileOptions)
					if err != nil {
						return errGatewayError(ctx, err)
					}
					if len(profiles) < 1 {
						return errObjectNotFound(ctx, "input.spec.avi.networkProfile", listNetworkProfileOptions.Name)
					}
					profile := profiles[0]
					switch route.Protocol {
					case model1.RouteProtocolTCP:
						if profile.Profile != nil && profile.Profile.Type != "PROTOCOL_TYPE_TCP_PROXY" && profile.Profile.Type != "PROTOCOL_TYPE_TCP_FAST_PATH" {
							return errInvalidRequest(ctx, "input.spec.avi.networkProfile", "protocol is not compatible", *input.Spec.Avi.NetworkProfile)
						}
					case model1.RouteProtocolUDP:
						if profile.Profile != nil && profile.Profile.Type != "PROTOCOL_TYPE_UDP_FAST_PATH" && profile.Profile.Type != "PROTOCOL_TYPE_UDP_PROXY" {
							return errInvalidRequest(ctx, "input.spec.avi.networkProfile", "protocol is not compatible", *input.Spec.Avi.NetworkProfile)
						}
					}
					spec.NetworkProfileID = profile.UUID
				}

				if input.Spec.Avi.LbAlgorithm != nil && *input.Spec.Avi.LbAlgorithm != "" {
					if !allowedAviLbAlgorithm[*input.Spec.Avi.LbAlgorithm] {
						return errInvalidRequest(ctx, "input.spec.avi.lbAlgorithm", "value is not allowed", *input.Spec.Avi.LbAlgorithm)
					}
					spec.LBAlgorithm = *input.Spec.Avi.LbAlgorithm
				}

				if spec.LBAlgorithm == "LB_ALGORITHM_CONSISTENT_HASH" &&
					input.Spec.Avi.LbAlgorithmHash != nil && *input.Spec.Avi.LbAlgorithmHash != "" {
					if !allowedAviLbAlgorithmHash[*input.Spec.Avi.LbAlgorithmHash] {
						return errInvalidRequest(ctx, "input.spec.avi.lbAlgorithmHash", "value is not allowed", *input.Spec.Avi.LbAlgorithmHash)
					}
					spec.LBAlgorithmHash = *input.Spec.Avi.LbAlgorithmHash

					if spec.LBAlgorithmHash == "LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER" {
						if input.Spec.Avi.LbAlgorithmConsistentHashHeader == nil || *input.Spec.Avi.LbAlgorithmConsistentHashHeader == "" {
							return errInvalidRequest(ctx, "input.spec.avi.lbAlgorithmConsistentHashHeader", "value cannot be empty")
						}
						spec.LBAlgorithmConsistentHashHeader = *input.Spec.Avi.LbAlgorithmConsistentHashHeader
					}
				}

				if input.Spec.Avi.VsVip != nil && *input.Spec.Avi.VsVip != "" {
					vsVips, err := listVsVip(client, &avi.ListVSVIPOptions{Name: *input.Spec.Avi.VsVip})
					if err != nil {
						return errGatewayError(ctx, err)
					}
					if len(vsVips) == 0 {
						return errObjectNotFound(ctx, "input.spec.avi.vsVip", *input.Spec.Avi.VsVip)
					}
					vsVip := vsVips[0]

					spec.VSVIPID = vsVip.UUID
					spec.ManualAllocatedVSVIP = true
					spec.PersistentVSVIP = true
				}

				if input.Spec.Avi.PersistentVsVip != nil {
					spec.PersistentVSVIP = *input.Spec.Avi.PersistentVsVip
				}
			}

			b, err := json.Marshal(spec)
			if err != nil {
				panic(err)
			}
			route.Spec = b
		}

		if err := appendAuditLog(tx, "create", "Route", route.ID, route.Name, sess.UserID, sess.UserName, route.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if err := tx.Create(route).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		for i, each := range input.Backends {
			instance := &model1.Instance{ID: each.Instance}
			if err := tx.Take(instance).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, fmt.Sprintf("input.backends.%d.instance", i), instance.ID.String())
				}
				panic(&ErrDBConnection{err})
			}
			cluster := &model1.Cluster{ID: instance.ClusterID.UUID}
			if err := tx.Take(cluster).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if !cluster.LoadBalancerID.Valid || cluster.LoadBalancerID.String() != loadBalancer.ID.String() {
				return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.instance", i), "different associated LoadBalancer", instance.ID.String())
			}

			if cluster.Provisioned < model1.ClusterProvisionFinish {
				return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.instance", i), "Cluster is not ready", instance.ID.String())
			}

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

			switch *each.Type {
			case "Service":
				service, err := wc.GetService(instance.Namespace, each.Name)
				if err != nil {
					if k8s.IsErrObjectNotExists(err) {
						return errObjectNotFound(ctx, fmt.Sprintf("input.backends.%d.name", i), each.Name)
					}
					return errGatewayError(ctx, err)
				}
				if len(service.Spec.Ports) == 0 {
					return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.name", i), "no available port found", each.Name)
				}
				if !validateLabels(service.Labels, map[string]string{
					defLabelApplication: instance.ApplicationID.String(),
					defLabelComponent:   instance.ComponentID.String(),
					defLabelInstance:    instance.ID.String(),
				}) {
					return errObjectNotFound(ctx, fmt.Sprintf("input.backends.%d.name", i), each.Name)
				}
				portM := make(map[int32]bool)
				for _, each := range service.Spec.Ports {
					switch route.Protocol {
					case model1.RouteProtocolUnknown, model1.RouteProtocolTCP:
						if each.Protocol == "" || each.Protocol == corev1.ProtocolTCP {
							portM[each.Port] = true
						}
					case model1.RouteProtocolUDP:
						if each.Protocol == corev1.ProtocolUDP {
							portM[each.Port] = true
						}
					}
				}
				if len(service.Status.LoadBalancer.Ingress) == 0 {
					return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.name", i), "no available port ingress", each.Name)
				}
				ipM := make(map[string]bool)
				for _, each := range service.Status.LoadBalancer.Ingress {
					ipM[each.IP] = true
				}
				if each.IP == nil {
					ipStr := service.Status.LoadBalancer.Ingress[0].IP
					each.IP = net.ParseIP(ipStr)
					if each.IP == nil {
						panic(fmt.Errorf("'%s' is not a valid IP address", ipStr))
					}
				}
				if !ipM[each.IP.String()] {
					return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.ip", i), "no such ingress", each.IP.String())
				}
				if each.Port == nil {
					if len(service.Spec.Ports) > 1 {
						return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.port", i), "value cannot be empty")
					}
					p := service.Spec.Ports[0]
					each.Port = util.Int64Ptr(int64(p.Port))
				}
				if !portM[int32(*each.Port)] {
					return errObjectNotFound(ctx, fmt.Sprintf("input.backends.%d.port", i), *each.Port)
				}
			default:
				return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.type", i), "invalid backend type", *each.Type)
			}

			backend := &model1.RouteBackend{
				ID:         uuid.New(),
				Name:       each.Name,
				Namespace:  instance.Namespace,
				Type:       *each.Type,
				IP:         each.IP.String(),
				Port:       *each.Port,
				InstanceID: model1.NewNullUUID(instance.ID),
				RouteID:    model1.NewNullUUID(route.ID),
			}
			if each.Ratio != nil {
				switch {
				case *each.Ratio < 0:
				case *each.Ratio > 20:
					backend.Ratio = 20
				default:
					backend.Ratio = *each.Ratio
				}
			}
			if err := tx.Create(backend).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}

		if err := tx.Take(route).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	mapd.EventEmitterRoute.Next(&mapd.Event{
		Type: mapd.EventTypeAdd,
		OID:  route.ID,
	})
	return convertRouteFromSQLModel(route), nil
}

// UpdateRoute is the resolver for the UpdateRoute field.
func (r *mutationResolver) UpdateRoute(ctx context.Context, input model.UpdateRoute) (*model.Route, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	route := &model1.Route{ID: input.ID}
	updates := make(map[string]interface{})
	if input.Description != nil {
		updates["description"] = *input.Description
	}
	if input.Port != nil {
		if *input.Port < 1 || *input.Port > 65535 {
			return nil, errInvalidRequest(ctx, "input.port", "valid port range: 1 - 65535", *input.Port)
		}
		updates["port"] = *input.Port
	}

	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(route).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", route.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermReadWrite, model1.PolicyObjectRoute, route.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		loadBalancer := &model1.LoadBalancer{ID: route.LoadBalancerID.UUID}
		if err := tx.Take(loadBalancer).Error; err != nil {
			panic(&ErrDBConnection{err})
		}

		var requiresProvision int
		if input.Port != nil && route.Port != *input.Port {
			requiresProvision++
		}
		for _, each := range input.RemoveBackends {
			if err := tx.Where("`route_id` = ?", route.ID).Delete(&model1.RouteBackend{ID: each.ID}).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			requiresProvision++
		}
		for i, each := range input.AddBackends {
			instance := &model1.Instance{ID: each.Instance}
			if err := tx.Take(instance).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errObjectNotFound(ctx, fmt.Sprintf("input.backends.%d.instance", i), instance.ID.String())
				}
				panic(&ErrDBConnection{err})
			}
			cluster := &model1.Cluster{ID: instance.ClusterID.UUID}
			if err := tx.Take(cluster).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			if !cluster.LoadBalancerID.Valid || cluster.LoadBalancerID.String() != loadBalancer.ID.String() {
				return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.instance", i), "different associated LoadBalancer", instance.ID.String())
			}

			if cluster.Provisioned < model1.ClusterProvisionFinish {
				return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.instance", i), "Cluster is not ready", instance.ID.String())
			}

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

			switch *each.Type {
			case "Service":
				service, err := wc.GetService(instance.Namespace, each.Name)
				if err != nil {
					if k8s.IsErrObjectNotExists(err) {
						return errObjectNotFound(ctx, fmt.Sprintf("input.backends.%d.name", i), each.Name)
					}
					return errGatewayError(ctx, err)
				}
				if len(service.Spec.Ports) == 0 {
					return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.name", i), "no available port found", each.Name)
				}
				if !validateLabels(service.Labels, map[string]string{
					defLabelApplication: instance.ApplicationID.String(),
					defLabelComponent:   instance.ComponentID.String(),
					defLabelInstance:    instance.ID.String(),
				}) {
					return errObjectNotFound(ctx, fmt.Sprintf("input.backends.%d.name", i), each.Name)
				}
				portM := make(map[int32]bool)
				for _, each := range service.Spec.Ports {
					switch route.Protocol {
					case model1.RouteProtocolUnknown, model1.RouteProtocolTCP:
						if each.Protocol == "" || each.Protocol == corev1.ProtocolTCP {
							portM[each.Port] = true
						}
					case model1.RouteProtocolUDP:
						if each.Protocol == corev1.ProtocolUDP {
							portM[each.Port] = true
						}
					}
				}
				if len(service.Status.LoadBalancer.Ingress) == 0 {
					return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.name", i), "no available port ingress", each.Name)
				}
				ipM := make(map[string]bool)
				for _, each := range service.Status.LoadBalancer.Ingress {
					ipM[each.IP] = true
				}
				if each.IP == nil {
					ipStr := service.Status.LoadBalancer.Ingress[0].IP
					each.IP = net.ParseIP(ipStr)
					if each.IP == nil {
						panic(fmt.Errorf("'%s' is not a valid IP address", ipStr))
					}
				}
				if !ipM[each.IP.String()] {
					return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.ip", i), "no such ingress", each.IP.String())
				}
				if each.Port == nil {
					if len(service.Spec.Ports) > 1 {
						return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.port", i), "value cannot be empty")
					}
					p := service.Spec.Ports[0]
					each.Port = util.Int64Ptr(int64(p.Port))
				}
				if !portM[int32(*each.Port)] {
					return errObjectNotFound(ctx, fmt.Sprintf("input.backends.%d.port", i), *each.Port)
				}
			default:
				return errInvalidRequest(ctx, fmt.Sprintf("input.backends.%d.type", i), "invalid backend type", *each.Type)
			}

			backend := &model1.RouteBackend{
				ID:         uuid.New(),
				Name:       each.Name,
				Namespace:  instance.Namespace,
				Type:       *each.Type,
				IP:         each.IP.String(),
				Port:       *each.Port,
				InstanceID: model1.NewNullUUID(instance.ID),
				RouteID:    model1.NewNullUUID(route.ID),
			}
			if each.Ratio != nil {
				switch {
				case *each.Ratio < 0:
				case *each.Ratio > 20:
					backend.Ratio = 20
				default:
					backend.Ratio = *each.Ratio
				}
			}
			if err := tx.Create(backend).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			requiresProvision++
		}
		if input.Spec != nil {
			switch route.Type {
			case model1.LoadBalancerTypeAVI:
				if input.Spec.Avi != nil {
					spec := &model1.RouteAVISpec{}
					if err := json.Unmarshal(route.Spec, spec); err != nil {
						panic(err)
					}

					if input.Spec.Avi.LbAlgorithm != nil {
						if *input.Spec.Avi.LbAlgorithm != "" {
							if !allowedAviLbAlgorithm[*input.Spec.Avi.LbAlgorithm] {
								return errInvalidRequest(ctx, "input.spec.avi.lbAlgorithm", "value is not allowed", *input.Spec.Avi.LbAlgorithm)
							}
							spec.LBAlgorithm = *input.Spec.Avi.LbAlgorithm
						} else {
							spec.LBAlgorithm = "LB_ALGORITHM_LEAST_CONNECTIONS"
						}
					}
					if spec.LBAlgorithm == "LB_ALGORITHM_CONSISTENT_HASH" &&
						input.Spec.Avi.LbAlgorithmHash != nil && *input.Spec.Avi.LbAlgorithmHash != "" {
						if *input.Spec.Avi.LbAlgorithmHash != "" {
							if !allowedAviLbAlgorithmHash[*input.Spec.Avi.LbAlgorithmHash] {
								return errInvalidRequest(ctx, "input.spec.avi.lbAlgorithmHash", "value is not allowed", *input.Spec.Avi.LbAlgorithmHash)
							}
							spec.LBAlgorithmHash = *input.Spec.Avi.LbAlgorithmHash
						} else {
							spec.LBAlgorithmHash = "LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS"
						}

						if spec.LBAlgorithmHash == "LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER" {
							if input.Spec.Avi.LbAlgorithmConsistentHashHeader == nil || *input.Spec.Avi.LbAlgorithmConsistentHashHeader == "" {
								return errInvalidRequest(ctx, "input.spec.avi.lbAlgorithmConsistentHashHeader", "value cannot be empty")
							}
							spec.LBAlgorithmConsistentHashHeader = *input.Spec.Avi.LbAlgorithmConsistentHashHeader
						} else {
							spec.LBAlgorithmConsistentHashHeader = ""
						}
					} else if spec.LBAlgorithm != "LB_ALGORITHM_CONSISTENT_HASH" {
						spec.LBAlgorithmHash = "LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS"
						spec.LBAlgorithmConsistentHashHeader = ""
					}

					if input.Spec.Avi.PersistentVsVip != nil {
						spec.PersistentVSVIP = *input.Spec.Avi.PersistentVsVip
					}

					b, err := json.Marshal(spec)
					if err != nil {
						panic(err)
					}
					updates["spec"] = string(b)
				}
			}
		}

		if requiresProvision > 0 {
			updates["provisioned"] = model1.RouteProvisionNone
			updates["provision_phase"] = model1.RouteProvisionPrepare
		}
		// for backward compatibility
		if route.Protocol == model1.RouteProtocolUnknown {
			updates["protocol"] = model1.RouteProtocolTCP
		}

		if err := appendAuditLog(tx, "update", "Route", route.ID, route.Name, sess.UserID, sess.UserName, route.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if len(updates) > 0 {
			if err := tx.Model(route).Updates(updates).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
		}
		if err := tx.Take(route).Error; err != nil {
			panic(&ErrDBConnection{err})
		}
		return nil
	}); err != nil && err != errDryrun {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	mapd.EventEmitterRoute.Next(&mapd.Event{
		Type: mapd.EventTypeModify,
		OID:  route.ID,
	})
	return convertRouteFromSQLModel(route), nil
}

// DeleteRoute is the resolver for the DeleteRoute field.
func (r *mutationResolver) DeleteRoute(ctx context.Context, input model.DeleteRoute) (*model.Route, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	force := input.Force != nil && *input.Force
	if force && !sess.Admin {
		return nil, errInsufficientPrivileges(ctx)
	}

	route := &model1.Route{ID: input.ID}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(route).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errObjectNotFound(ctx, "input.id", route.ID.String())
			}
			panic(&ErrDBConnection{err})
		}

		if !IsGuardSilenced(ctx) && !sess.Can(model1.PermModify, model1.PolicyObjectRoute, route.OrganizationID.UUID) {
			return errInsufficientPrivileges(ctx)
		}

		if err := appendAuditLog(tx, "delete", "Route", route.ID, route.Name, sess.UserID, sess.UserName, route.OrganizationID.UUID); err != nil {
			panic(&ErrDBConnection{err})
		}

		if force {
			if err := tx.Where("`route_id` = ?", route.ID).Delete(&model.RouteBackend{}).Error; err != nil {
				return err
			}
			if err := tx.Delete(route).Error; err != nil {
				panic(&ErrDBConnection{err})
			}
			return nil
		}

		updates := make(map[string]interface{})
		updates["provision_phase"] = model1.RouteProvisionDeleting
		if err := tx.Model(route).Updates(updates).Error; err != nil {
			return &ErrDBConnection{err}
		}
		return nil
	}); err != nil {
		if !IsGQLError(err) {
			panic(err)
		}
		return nil, err
	}
	if force {
		mapd.EventEmitterRoute.Next(&mapd.Event{
			Type: mapd.EventTypeRemove,
			OID:  route.ID,
		})
	} else {
		mapd.EventEmitterRoute.Next(&mapd.Event{
			Type: mapd.EventTypeModify,
			OID:  route.ID,
		})
	}
	return convertRouteFromSQLModel(route), nil
}

// Route is the resolver for the Route field.
func (r *queryResolver) Route(ctx context.Context, id uuid.UUID) (*model.Route, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	db := mapd.Database()
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectRoute)
	}

	route := &model1.Route{ID: id}
	if err := db.Take(route).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "id", route.ID.String())
		}
		panic(&ErrDBConnection{err})
	}
	return convertRouteFromSQLModel(route), nil
}

// Routes is the resolver for the Routes field.
func (r *queryResolver) Routes(ctx context.Context, page *model.QueryPage, filters []*model.QueryFilter) (*model.Routes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	obj := &model.Routes{Page: &model.PageInfo{}}
	if page.Limit != nil {
		obj.Page.Limit = *page.Limit
	}
	if page.Number != nil {
		obj.Page.Number = *page.Number
	}
	for _, each := range filters {
		obj.Filters = append(obj.Filters, &model.FilterInfo{Key: each.Key, Value: each.Value})
	}

	db := mapd.Database().Model(&model1.Route{})
	if !IsGuardSilenced(ctx) {
		db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectRoute)
	}
	db = finalizeRouteQueryFilter(db, obj.Filters...)

	if err := db.Count(&obj.TotalCount).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	finalizePageInfo(obj.Page, obj.TotalCount)
	return obj, nil
}

// Backends is the resolver for the backends field.
func (r *routeResolver) Backends(ctx context.Context, obj *model.Route) ([]*model.RouteBackend, error) {
	var backends []model1.RouteBackend
	if err := mapd.Database().Where("`route_id` = ?", obj.ID).Find(&backends).Error; err != nil {
		panic(&ErrDBConnection{err})
	}
	var objs []*model.RouteBackend
	for _, backend := range backends {
		objs = append(objs, convertRouteBackendFromSQLModel(&backend))
	}
	return objs, nil
}

// Component is the resolver for the component field.
func (r *routeResolver) Component(ctx context.Context, obj *model.Route) (*model.Component, error) {
	if obj.Component == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Component(WithGuardSilenced(ctx), obj.Component.ID)
}

// Application is the resolver for the application field.
func (r *routeResolver) Application(ctx context.Context, obj *model.Route) (*model.Application, error) {
	if obj.Application == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Application(WithGuardSilenced(ctx), obj.Application.ID)
}

// LoadBalancer is the resolver for the loadBalancer field.
func (r *routeResolver) LoadBalancer(ctx context.Context, obj *model.Route) (*model.LoadBalancer, error) {
	if obj.LoadBalancer == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).LoadBalancer(WithGuardSilenced(ctx), obj.LoadBalancer.ID)
}

// Organization is the resolver for the organization field.
func (r *routeResolver) Organization(ctx context.Context, obj *model.Route) (*model.Organization, error) {
	if obj.Organization == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Organization(WithGuardSilenced(ctx), obj.Organization.ID)
}

// Instance is the resolver for the instance field.
func (r *routeBackendResolver) Instance(ctx context.Context, obj *model.RouteBackend) (*model.Instance, error) {
	if obj.Instance == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Instance(WithGuardSilenced(ctx), obj.Instance.ID)
}

// Edge is the resolver for the edge field.
func (r *routesResolver) Edge(ctx context.Context, obj *model.Routes) (*model.RouteNodes, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	nodes := &model.RouteNodes{}
	if err := mapd.Database().Transaction(func(tx *gorm.DB) error {
		db := tx.Order("`updated_at` DESC")
		if !IsGuardSilenced(ctx) {
			db = sess.FilterQueryByOrganization(db, model1.PermReadOnly, model1.PolicyObjectRoute)
		}
		db = finalizeRouteQueryFilter(db, obj.Filters...)

		if err := paginate(obj.Page.Count, obj.Page.Offset, func(lim, off int64) error {
			var routes []model1.Route
			if err := db.Limit(int(lim)).Offset(int(off)).Find(&routes).Error; err != nil {
				return err
			}
			for _, route := range routes {
				nodes.Nodes = append(nodes.Nodes, convertRouteFromSQLModel(&route))
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

// AviRouteSpec returns graphql1.AviRouteSpecResolver implementation.
func (r *Resolver) AviRouteSpec() graphql1.AviRouteSpecResolver { return &aviRouteSpecResolver{r} }

// Route returns graphql1.RouteResolver implementation.
func (r *Resolver) Route() graphql1.RouteResolver { return &routeResolver{r} }

// RouteBackend returns graphql1.RouteBackendResolver implementation.
func (r *Resolver) RouteBackend() graphql1.RouteBackendResolver { return &routeBackendResolver{r} }

// Routes returns graphql1.RoutesResolver implementation.
func (r *Resolver) Routes() graphql1.RoutesResolver { return &routesResolver{r} }

type aviRouteSpecResolver struct{ *Resolver }
type routeResolver struct{ *Resolver }
type routeBackendResolver struct{ *Resolver }
type routesResolver struct{ *Resolver }
