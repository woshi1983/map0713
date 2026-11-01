package v2

import (
	"testing"

	capiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

func TestGetClusterStatusWithoutControlPlaneReplicas(t *testing.T) {
	cluster := &capiv1beta1.Cluster{}
	cluster.Status.InfrastructureReady = true
	cluster.Status.ControlPlaneReady = true
	controlPlane := &kubeadmv1beta1.KubeadmControlPlane{}
	controlPlane.Status.ReadyReplicas = 1
	info := &ClusterObjects{
		Cluster:             cluster,
		KubeadmControlPlane: controlPlane,
		MachineDeployments: []capiv1beta1.MachineDeployment{{
			Status: capiv1beta1.MachineDeploymentStatus{ReadyReplicas: 1},
		}},
	}
	if got := getClusterStatus(info); got != TKGClusterPhaseCreating {
		t.Fatalf("status = %s, want %s", got, TKGClusterPhaseCreating)
	}
}
