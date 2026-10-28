package conductor

import (
	"testing"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

func TestExtensionLoadBalancerEnabled(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cluster model.Cluster
		want    bool
	}{
		{name: "VKS uses native service load balancer", cluster: model.Cluster{Class: model.ClusterClassVKS}, want: true},
		{name: "TKGM without configured load balancer", cluster: model.Cluster{Class: model.ClusterClassTKGM}, want: false},
		{name: "TKGM with configured load balancer", cluster: model.Cluster{Class: model.ClusterClassTKGM, LoadBalancerID: model.NewNullUUID(uuid.New())}, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := extensionLoadBalancerEnabled(&tc.cluster); got != tc.want {
				t.Fatalf("extensionLoadBalancerEnabled() = %v; want %v", got, tc.want)
			}
		})
	}
}
