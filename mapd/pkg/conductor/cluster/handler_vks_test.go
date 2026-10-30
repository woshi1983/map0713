package cluster

import (
	"testing"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestVKSHealthStateTracksScaling(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{"generation": int64(2)},
		"spec": map[string]interface{}{"topology": map[string]interface{}{
			"controlPlane": map[string]interface{}{"replicas": int64(1)},
			"workers": map[string]interface{}{"machineDeployments": []interface{}{
				map[string]interface{}{"name": "pool-0", "replicas": int64(5)},
			}},
		}},
		"status": map[string]interface{}{
			"observedGeneration": int64(2),
			"controlPlane":       map[string]interface{}{"desiredReplicas": int64(1), "readyReplicas": int64(1), "availableReplicas": int64(1)},
			"workers":            map[string]interface{}{"desiredReplicas": int64(5), "readyReplicas": int64(4), "availableReplicas": int64(4)},
			"conditions":         []interface{}{map[string]interface{}{"type": "Available", "status": "False"}},
		},
	}}
	state, err := vksHealthState(model.ClusterStateUpdating, obj)
	if state != model.ClusterStateUpdating || err == nil {
		t.Fatalf("scaling must remain updating: %s, %v", state, err)
	}
	_ = unstructured.SetNestedField(obj.Object, int64(5), "status", "workers", "availableReplicas")
	_ = unstructured.SetNestedField(obj.Object, int64(5), "status", "workers", "readyReplicas")
	_ = unstructured.SetNestedSlice(obj.Object, []interface{}{map[string]interface{}{"type": "Available", "status": "True"}}, "status", "conditions")
	state, err = vksHealthState(model.ClusterStateUpdating, obj)
	if state != model.ClusterStateRunning || err != nil {
		t.Fatalf("completed scaling must become running: %s, %v", state, err)
	}
}

func TestVKSHealthStateRecoversPrematureRunningDuringUpgrade(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"status": map[string]interface{}{"conditions": []interface{}{
			map[string]interface{}{"type": "Available", "status": "True"},
			map[string]interface{}{"type": "TopologyReconciled", "status": "False", "reason": "ClusterUpgrading"},
		}},
	}}
	state, err := vksHealthState(model.ClusterStateRunning, obj)
	if state != model.ClusterStateUpdating || err == nil {
		t.Fatalf("active upgrade must restore UPDATING: %s, %v", state, err)
	}
}

func TestVKSHealthStateWaitsForRequestedControlPlaneReplicas(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"spec": map[string]interface{}{"topology": map[string]interface{}{
			"controlPlane": map[string]interface{}{"replicas": int64(3)},
			"workers": map[string]interface{}{"machineDeployments": []interface{}{
				map[string]interface{}{"name": "pool-0", "replicas": int64(1)},
			}},
		}},
		"status": map[string]interface{}{
			"controlPlane": map[string]interface{}{"desiredReplicas": int64(3), "readyReplicas": int64(1), "availableReplicas": int64(1)},
			"workers":      map[string]interface{}{"desiredReplicas": int64(1), "readyReplicas": int64(1), "availableReplicas": int64(1)},
			"conditions":   []interface{}{map[string]interface{}{"type": "Available", "status": "True"}},
		},
	}}
	state, err := vksHealthState(model.ClusterStateRunning, obj)
	if state != model.ClusterStateCreating || err == nil {
		t.Fatalf("1/3 ready control plane must remain initializing: %s, %v", state, err)
	}
	_ = unstructured.SetNestedField(obj.Object, int64(3), "status", "controlPlane", "readyReplicas")
	_ = unstructured.SetNestedField(obj.Object, int64(3), "status", "controlPlane", "availableReplicas")
	state, err = vksHealthState(model.ClusterStateCreating, obj)
	if state != model.ClusterStateRunning || err != nil {
		t.Fatalf("3/3 ready control plane must become running: %s, %v", state, err)
	}
}
