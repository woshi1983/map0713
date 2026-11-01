package resolver

import (
	"testing"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/avi"
)

func TestSelectAVICloudByName(t *testing.T) {
	clouds := []*avi.Cloud{
		{Name: "Default-Cloud", UUID: "default"},
		{Name: "vc210", UUID: "vc210"},
	}
	cloud, err := selectAVICloud(clouds, "vc210")
	if err != nil || cloud.UUID != "vc210" {
		t.Fatalf("cloud = %+v, err = %v", cloud, err)
	}
	if _, err := selectAVICloud(clouds, "missing"); err == nil {
		t.Fatal("missing cloud should fail")
	}
	if _, err := selectAVICloud(clouds, ""); err == nil {
		t.Fatal("ambiguous cloud should fail")
	}
}

func TestSelectAVIServiceEngineGroupUsesCloud(t *testing.T) {
	groups := []*avi.ServiceEngineGroup{
		{Name: "Default-Group", UUID: "wrong", CloudRef: "https://avi/api/cloud/default"},
		{Name: "Default-Group", UUID: "correct", CloudRef: "https://avi/api/cloud/vc210"},
	}
	group, err := selectAVIServiceEngineGroup(groups, "Default-Group", "https://avi/api/cloud/vc210")
	if err != nil || group.UUID != "correct" {
		t.Fatalf("group = %+v, err = %v", group, err)
	}
	if _, err := selectAVIServiceEngineGroup(groups, "Default-Group", "https://avi/api/cloud/missing"); err == nil {
		t.Fatal("group from another cloud should fail")
	}
}

func TestSelectAVINetworkRejectsDifferentCloud(t *testing.T) {
	networks := []*avi.Network{{Name: "vlan804", CloudRef: "https://avi/api/cloud/vc210"}}
	if _, err := selectAVINetwork(networks, "vlan804", "https://avi/api/cloud/default"); err == nil {
		t.Fatal("network from another cloud should fail")
	}
	if network, err := selectAVINetwork(networks, "vlan804", "https://avi/api/cloud/vc210"); err != nil || network.Name != "vlan804" {
		t.Fatalf("network = %+v, err = %v", network, err)
	}
}
