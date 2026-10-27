package cluster

import (
	"testing"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestVKSHealthStateTracksScaling(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{"generation": int64(2)},
		"status": map[string]interface{}{
			"observedGeneration": int64(2),
			"workers":            map[string]interface{}{"desiredReplicas": int64(5), "availableReplicas": int64(4)},
			"conditions":         []interface{}{map[string]interface{}{"type": "Available", "status": "False"}},
		},
	}}
	state, err := vksHealthState(model.ClusterStateUpdating, obj)
	if state != model.ClusterStateUpdating || err == nil {
		t.Fatalf("scaling must remain updating: %s, %v", state, err)
	}
	_ = unstructured.SetNestedField(obj.Object, int64(5), "status", "workers", "availableReplicas")
	_ = unstructured.SetNestedSlice(obj.Object, []interface{}{map[string]interface{}{"type": "Available", "status": "True"}}, "status", "conditions")
	state, err = vksHealthState(model.ClusterStateUpdating, obj)
	if state != model.ClusterStateRunning || err != nil {
		t.Fatalf("completed scaling must become running: %s, %v", state, err)
	}
}
