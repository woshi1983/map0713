// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package resolver

import (
	"encoding/json"
	"sort"
	"strings"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/avi"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/textutil"
	"gorm.io/gorm"
)

func convertLoadBalancerFromSQLModel(lb *model1.LoadBalancer) *model.LoadBalancer {
	obj := &model.LoadBalancer{
		ID:                    lb.ID,
		Name:                  lb.Name,
		Description:           lb.Description,
		Type:                  model.LoadBalancerType(lb.Type.String()),
		Endpoint:              abstract.MustParseURL(lb.Endpoint),
		CaCert:                string(lb.CACert),
		InsecureSkipTLSVerify: lb.InsecureTLS,
		CreatedAt:             lb.CreatedAt,
		LastModified:          lb.UpdatedAt,
	}
	switch lb.Type {
	case model1.LoadBalancerTypeAVI:
		obj.AuthUser = util.StringPtr(lb.AuthUser)

		spec := &model1.LoadBalancerAVISpec{}
		if err := json.Unmarshal(lb.Spec, spec); err != nil {
			panic(err)
		}

		v := &model.AviLoadBalancerSpec{LoadBalancer: lb.ID}
		if spec.PreferredNetwork != "" {
			v.PreferredNetwork = &spec.PreferredNetwork
		}
		if spec.LicenseTier != model1.AVILicenseTierUnknown {
			v.LicenseTier = util.StringPtr(string(spec.LicenseTier))
		}
		obj.Spec = v

		obj.Status = &model.AviLoadBalancerStatus{ID: lb.ID}
	}
	if lb.OrganizationID.Valid {
		obj.Organization = &model.Organization{ID: lb.OrganizationID.UUID}
	}
	return obj
}

func finalizeLoadBalancerQueryFilter(db *gorm.DB, filters ...*model.FilterInfo) *gorm.DB {
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

func createServiceEngineGroup(client *avi.AVI, name string, licenseTier model1.AVILicenseTier) (*avi.ServiceEngineGroup, error) {
	obj := &avi.ServiceEngineGroup{
		Name:             name,
		MinScaleoutPerVS: 2,
		MaxScaleoutPerVS: 4,
		MinSE:            2,
		MaxSE:            1000,
		MaxVSPerSE:       1000,
		VCPUsPerSE:       2,
		MemoryPerSE:      2048,
		LicenseTier:      string(licenseTier),
	}
	if licenseTier == model1.AVILicenseTierEnterprise {
		obj.HAMode = "HA_MODE_SHARED_PAIR"
	} else {
		obj.HAMode = "HA_MODE_LEGACY_ACTIVE_STANDBY"
	}
	return client.CreateSEGroup(obj)
}

func listSEGroup(client *avi.AVI, options *avi.ListSEGroupOptions) ([]*avi.ServiceEngineGroup, error) {
	list, err := client.ListSEGroup(options)
	if err != nil {
		return nil, err
	}
	return list.Results, nil
}

func listCloud(client *avi.AVI, options *avi.ListCloudOptions) ([]*avi.Cloud, error) {
	list, err := client.ListCloud(options)
	if err != nil {
		return nil, err
	}
	return list.Results, nil
}

func listNetworkProfile(client *avi.AVI, options *avi.ListNetworkProfileOptions) ([]*avi.NetworkProfile, error) {
	list, err := client.ListNetworkProfile(options)
	if err != nil {
		return nil, err
	}
	return list.Results, nil
}

func listNetwork(client *avi.AVI, options *avi.ListNetworkOptions) ([]*avi.Network, error) {
	list, err := client.ListNetwork(options)
	if err != nil {
		return nil, err
	}
	return list.Results, nil
}

func listVsVip(client *avi.AVI, options *avi.ListVSVIPOptions) ([]*avi.VSVIP, error) {
	list, err := client.ListVSVIP(options)
	if err != nil {
		return nil, err
	}
	return list.Results, nil
}

func getHighestAVILicenseTier(client *avi.AVI, all bool) (model1.AVILicenseTier, error) {
	details, err := client.GetLicenseLedgerDetails()
	if err != nil {
		return model1.AVILicenseTierUnknown, err
	}
	return GetHighestAVILicense(details.TierUsages, all), nil
}

var (
	aviLicenseTierMapper = map[model1.AVILicenseTier]int{
		model1.AVILicenseTierEssentials: 1,
		model1.AVILicenseTierBasic:      2,
		model1.AVILicenseTierEnterprise: 3,
	}
)

func GetHighestAVILicense(usages []*avi.LicenseTierUsage, all bool) model1.AVILicenseTier {
	if len(usages) == 0 {
		return model1.AVILicenseTierUnknown
	}
	var availables []model1.AVILicenseTier
	for _, usage := range usages {
		if all || usage.Usage.Remaining > 0 {
			availables = append(availables, model1.AVILicenseTier(usage.Tier))
		}
	}
	sort.Slice(availables, func(i, j int) bool { return aviLicenseTierMapper[availables[i]] > aviLicenseTierMapper[availables[j]] })
	return availables[0]
}
