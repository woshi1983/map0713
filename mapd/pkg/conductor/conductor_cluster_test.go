package conductor

import (
	"errors"
	"testing"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

func TestVKSUpdatingHealthStatus(t *testing.T) {
	if got := clusterHealthStatus(model.ClusterClassVKS, model.ClusterStateUpdating, errors.New("reconciling")); got != model.ClusterStatusInitializing {
		t.Fatalf("VKS updating status = %s, want INITIALIZING", got)
	}
	if got := clusterHealthStatus(model.ClusterClassVKS, model.ClusterStateCreating, errors.New("control plane replicas are not ready")); got != model.ClusterStatusInitializing {
		t.Fatalf("VKS creating status = %s, want INITIALIZING", got)
	}
	if got := clusterHealthStatus(model.ClusterClassVKS, model.ClusterStateRunning, nil); got != model.ClusterStatusHealthy {
		t.Fatalf("VKS ready status = %s, want HEALTHY", got)
	}
	if got := clusterHealthStatus(model.ClusterClassTKGS, model.ClusterStateUpdating, errors.New("reconciling")); got != model.ClusterStatusUnhealthy {
		t.Fatalf("existing TKGS status changed: %s", got)
	}
}
