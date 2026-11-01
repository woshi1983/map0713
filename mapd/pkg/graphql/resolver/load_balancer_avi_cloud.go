package resolver

import (
	"fmt"
	"strings"

	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/avi"
)

func selectAVICloud(clouds []*avi.Cloud, name string) (*avi.Cloud, error) {
	var match *avi.Cloud
	for _, cloud := range clouds {
		if cloud == nil || (name != "" && cloud.Name != name) {
			continue
		}
		if match != nil {
			return nil, fmt.Errorf("AVI cloud %q is ambiguous; select a specific cloud", name)
		}
		match = cloud
	}
	if match == nil {
		return nil, fmt.Errorf("AVI cloud %q was not found", name)
	}
	return match, nil
}

func selectAVIServiceEngineGroup(groups []*avi.ServiceEngineGroup, name, cloudRef string) (*avi.ServiceEngineGroup, error) {
	for _, group := range groups {
		if group != nil && group.Name == name && sameAVIReference(group.CloudRef, cloudRef) {
			return group, nil
		}
	}
	return nil, fmt.Errorf("AVI service engine group %q was not found in the selected cloud", name)
}

func selectAVINetwork(networks []*avi.Network, name, cloudRef string) (*avi.Network, error) {
	for _, network := range networks {
		if network != nil && network.Name == name && sameAVIReference(network.CloudRef, cloudRef) {
			return network, nil
		}
	}
	return nil, fmt.Errorf("AVI network %q was not found in the selected cloud", name)
}

func sameAVIReference(actual, expected string) bool {
	return actual != "" && expected != "" && strings.TrimRight(actual, "/") == strings.TrimRight(expected, "/")
}

// prepareAVILoadBalancerSpec keeps a Cloud, SE Group and route VIP network in
// the same AVI Cloud. AVI permits identical SE Group names in different Clouds.
func prepareAVILoadBalancerSpec(client *avi.AVI, cloudName, groupName, networkName string, createMissingGroup bool) (*model1.LoadBalancerAVISpec, error) {
	cloudList, err := client.ListCloud(&avi.ListCloudOptions{Name: cloudName})
	if err != nil {
		return nil, err
	}
	cloud, err := selectAVICloud(cloudList.Results, cloudName)
	if err != nil {
		return nil, err
	}
	cloudRef := client.CloudRef(cloud.UUID).String()
	if networkName != "" {
		list, err := client.ListNetwork(&avi.ListNetworkOptions{Name: networkName, CloudName: cloud.Name})
		if err != nil {
			return nil, err
		}
		if _, err := selectAVINetwork(list.Results, networkName, cloudRef); err != nil {
			return nil, err
		}
	}
	list, err := client.ListSEGroup(&avi.ListSEGroupOptions{Name: groupName, CloudName: cloud.Name})
	if err != nil {
		return nil, err
	}
	group, err := selectAVIServiceEngineGroup(list.Results, groupName, cloudRef)
	if err != nil {
		if len(list.Results) != 0 {
			return nil, err
		}
		if !createMissingGroup {
			return &model1.LoadBalancerAVISpec{CloudID: cloud.UUID, CloudName: cloud.Name, PreferredNetwork: networkName}, nil
		}
		tier, err := getHighestAVILicenseTier(client, false)
		if err != nil {
			return nil, err
		}
		group, err = createServiceEngineGroup(client, groupName, tier, cloudRef)
		if err != nil {
			return nil, err
		}
		if !sameAVIReference(group.CloudRef, cloudRef) {
			return nil, fmt.Errorf("AVI created service engine group %q in an unexpected cloud", groupName)
		}
	}
	return &model1.LoadBalancerAVISpec{
		CloudID: cloud.UUID, CloudName: cloud.Name, SEGroupID: group.UUID,
		LicenseTier: model1.AVILicenseTier(group.LicenseTier), PreferredNetwork: networkName,
	}, nil
}
