// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"encoding/json"
	"net"
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"gorm.io/gorm"
)

func convertRouteFromSQLModel(route *model1.Route) *model.Route {
	if route.Protocol == model1.RouteProtocolUnknown {
		route.Protocol = model1.RouteProtocolTCP
	}
	obj := &model.Route{
		ID:           route.ID,
		Name:         route.Name,
		Description:  route.Description,
		Type:         model.LoadBalancerType(route.Type.String()),
		Protocol:     model.RouteProtocol(route.Protocol.String()),
		IP:           net.ParseIP(route.IP),
		Port:         route.Port,
		CreatedAt:    route.CreatedAt,
		LastModified: route.UpdatedAt,
	}
	switch route.Type {
	case model1.LoadBalancerTypeAVI:
		spec := &model1.RouteAVISpec{}
		if err := json.Unmarshal(route.Spec, spec); err != nil {
			panic(err)
		}

		v := &model.AviRouteSpec{
			NetworkProfileID:                spec.NetworkProfileID,
			LbAlgorithm:                     spec.LBAlgorithm,
			LbAlgorithmHash:                 spec.LBAlgorithmHash,
			LbAlgorithmConsistentHashHeader: spec.LBAlgorithmConsistentHashHeader,
			ManualAllocatedVsVip:            spec.ManualAllocatedVSVIP,
			PersistentVsVip:                 spec.PersistentVSVIP,
		}
		if route.LoadBalancerID.Valid {
			v.LoadBalancer = route.LoadBalancerID.UUID
		}
		obj.Spec = v
	}
	switch {
	case route.ProvisionPhase == model1.RouteProvisionDeleting:
		obj.Status = model.RouteStatusDeleting
	case route.Provisioned != model1.RouteProvisionFinish:
		obj.Status = model.RouteStatusProvisioning
	default:
		obj.Status = model.RouteStatusProvisioned
	}
	if route.ComponentID.Valid {
		obj.Component = &model.Component{ID: route.ComponentID.UUID}
	}
	if route.ApplicationID.Valid {
		obj.Application = &model.Application{ID: route.ApplicationID.UUID}
	}
	if route.LoadBalancerID.Valid {
		obj.LoadBalancer = &model.LoadBalancer{ID: route.LoadBalancerID.UUID}
	}
	if route.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: route.OrganizationID.UUID}
	}
	return obj
}

func convertRouteBackendFromSQLModel(backend *model1.RouteBackend) *model.RouteBackend {
	obj := &model.RouteBackend{
		ID:        backend.ID,
		Name:      backend.Name,
		Namespace: backend.Namespace,
		Type:      backend.Type,
		IP:        net.ParseIP(backend.IP),
		Port:      backend.Port,
		Ratio:     backend.Ratio,
	}
	if backend.InstanceID.Valid {
		obj.Instance = &model.Instance{ID: backend.InstanceID.UUID}
	}
	return obj
}

func finalizeRouteQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
	for _, each := range filters {
		switch each.Key {
		case "name":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`name` LIKE ?")
				args = append(args, "%"+textutil.EscapeSQLString(val)+"%")
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "federation":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`federation_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "component":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`component_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "application":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`application_id` = ?")
				args = append(args, val)
			}
			if len(args) > 0 {
				db = db.Where(strings.Join(exprs, " OR "), args...)
			}
		case "loadBalancer":
			var (
				exprs []string
				args  []interface{}
			)
			for _, val := range each.Value {
				exprs = append(exprs, "`load_balancer_id` = ?")
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
	return db
}

var (
	allowedAviLbAlgorithm = map[string]bool{
		"LB_ALGORITHM_LEAST_CONNECTIONS": true,
		"LB_ALGORITHM_ROUND_ROBIN":       true,
		"LB_ALGORITHM_FASTEST_RESPONSE":  true,
		"LB_ALGORITHM_CONSISTENT_HASH":   true,
		"LB_ALGORITHM_LEAST_LOAD":        true,
		"LB_ALGORITHM_FEWEST_SERVERS":    true,
		"LB_ALGORITHM_RANDOM":            true,
		"LB_ALGORITHM_FEWEST_TASKS":      true,
		"LB_ALGORITHM_CORE_AFFINITY":     true,
	}
	allowedAviLbAlgorithmHash = map[string]bool{
		"LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS":          true,
		"LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS_AND_PORT": true,
		"LB_ALGORITHM_CONSISTENT_HASH_URI":                        true,
		"LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER":              true,
		"LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_STRING":              true,
		"LB_ALGORITHM_CONSISTENT_HASH_CALLID":                     true,
	}
)
