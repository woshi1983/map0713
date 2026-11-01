package conductor

import (
	"testing"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/avi"
)

func TestSelectRouteNetworkUsesSelectedCloud(t *testing.T) {
	networks := []*avi.Network{
		{Name: "vlan804", UUID: "wrong", CloudRef: "https://avi/api/cloud/default"},
		{Name: "vlan804", UUID: "right", CloudRef: "https://avi/api/cloud/vc210"},
	}
	if got := selectRouteNetwork(networks, "vlan804", "https://avi/api/cloud/vc210"); got == nil || got.UUID != "right" {
		t.Fatalf("selected network = %+v", got)
	}
	if got := selectRouteNetwork(networks, "vlan804", "https://avi/api/cloud/missing"); got != nil {
		t.Fatalf("unexpected network = %+v", got)
	}
}
