package vks

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

func TestUpdateTopologyScalesOwnedAvailableCluster(t *testing.T) {
	id := uuid.New()
	cluster := fakeCluster("demo", "team-a", id, types.UID("remote-1"))
	cluster.Object["spec"] = map[string]interface{}{"topology": map[string]interface{}{
		"controlPlane": map[string]interface{}{"replicas": int64(1)},
		"workers": map[string]interface{}{"machineDeployments": []interface{}{
			map[string]interface{}{"name": "pool-0", "replicas": int64(2)},
		}},
	}}
	cluster.Object["status"] = map[string]interface{}{"conditions": []interface{}{
		map[string]interface{}{"type": "Available", "status": "True"},
	}}
	client := newFakeClient(cluster)
	spec := model.ClusterVKSSpec{Namespace: "team-a", RemoteUID: "remote-1"}
	newWorkers := int64(4)
	updated, err := client.UpdateTopology(context.Background(), id, "demo", spec, nil, &newWorkers, "")
	if err != nil {
		t.Fatal(err)
	}
	remote, err := client.GetCluster(context.Background(), "team-a", "demo")
	if err != nil {
		t.Fatal(err)
	}
	cp, workers, err := VKSReplicas(remote)
	if err != nil || cp != 1 || workers != 4 || updated.WorkerReplicas != 4 {
		t.Fatalf("unexpected scaled topology: cp=%d workers=%d spec=%+v err=%v", cp, workers, updated, err)
	}
	wrongID := uuid.New()
	if _, err := client.UpdateTopology(context.Background(), wrongID, "demo", spec, nil, &newWorkers, ""); err == nil {
		t.Fatal("foreign cluster update must be rejected")
	}
}

func TestSetVKSReplicasUsesNamedPool(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{"spec": map[string]interface{}{"topology": map[string]interface{}{
		"controlPlane": map[string]interface{}{"replicas": int64(1)},
		"workers": map[string]interface{}{"machineDeployments": []interface{}{
			map[string]interface{}{"name": "other", "replicas": int64(2)},
			map[string]interface{}{"name": "pool-0", "replicas": int64(2)},
		}},
	}}}}
	if err := setVKSReplicas(obj, 3, 5); err != nil {
		t.Fatal(err)
	}
	cp, _, _ := unstructured.NestedInt64(obj.Object, "spec", "topology", "controlPlane", "replicas")
	pools, _, _ := unstructured.NestedSlice(obj.Object, "spec", "topology", "workers", "machineDeployments")
	if cp != 3 || pools[0].(map[string]interface{})["replicas"] != int64(2) || pools[1].(map[string]interface{})["replicas"] != int64(5) {
		t.Fatalf("unexpected scaled topology: %+v", obj.Object)
	}
}

func TestSetVKSReplicasRejectsInvalidCountsAndMissingPool(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{"spec": map[string]interface{}{"topology": map[string]interface{}{
		"workers": map[string]interface{}{"machineDeployments": []interface{}{map[string]interface{}{"name": "other"}}},
	}}}}
	for _, counts := range [][2]int64{{2, 1}, {1, 0}, {1, 101}, {1, 1}} {
		if err := setVKSReplicas(obj.DeepCopy(), counts[0], counts[1]); err == nil {
			t.Fatalf("counts %v should be rejected", counts)
		}
	}
}

func TestVKSUpgradeVersionEligibility(t *testing.T) {
	for _, tt := range []struct {
		current, target string
		valid           bool
	}{
		{"v1.35.2+vmware.1", "v1.35.3+vmware.1", true},
		{"v1.35.2+vmware.1", "v1.36.0+vmware.1", true},
		{"v1.35.2+vmware.1", "v1.37.0+vmware.1", false},
		{"v1.35.2+vmware.1", "v1.35.1+vmware.1", false},
		{"v1.35.2+vmware.1", "v1.35.2+vmware.2", false},
	} {
		if valid := validVKSUpgrade(tt.current, tt.target); valid != tt.valid {
			t.Errorf("upgrade %s -> %s: got %v, want %v", tt.current, tt.target, valid, tt.valid)
		}
	}
}

func TestSetVKSOSSelectorPreservesPoolMetadata(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{"spec": map[string]interface{}{"topology": map[string]interface{}{
		"controlPlane": map[string]interface{}{"metadata": map[string]interface{}{"annotations": map[string]interface{}{"custom": "keep"}}},
		"workers": map[string]interface{}{"machineDeployments": []interface{}{map[string]interface{}{
			"name": "pool-0", "metadata": map[string]interface{}{"annotations": map[string]interface{}{"custom": "keep"}},
		}}},
	}}}}
	if err := setVKSOSSelector(obj, "ubuntu", "24.04"); err != nil {
		t.Fatal(err)
	}
	cpAnnotations, _, _ := unstructured.NestedStringMap(obj.Object, "spec", "topology", "controlPlane", "metadata", "annotations")
	pools, _, _ := unstructured.NestedSlice(obj.Object, "spec", "topology", "workers", "machineDeployments")
	workerAnnotations := pools[0].(map[string]interface{})["metadata"].(map[string]interface{})["annotations"].(map[string]interface{})
	if cpAnnotations["custom"] != "keep" || cpAnnotations["run.tanzu.vmware.com/resolve-os-image"] != "os-name=ubuntu,os-version=24.04" || workerAnnotations["custom"] != "keep" || workerAnnotations["run.tanzu.vmware.com/resolve-os-image"] != "os-name=ubuntu,os-version=24.04" {
		t.Fatalf("unexpected OS selectors: CP=%v worker=%v", cpAnnotations, workerAnnotations)
	}
}

func TestVKSUpgradeInProgressIgnoresResolvedCondition(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{"status": map[string]interface{}{"conditions": []interface{}{
		map[string]interface{}{"type": "Ready", "status": "True", "message": "Upgrading to version v1.36.0"},
	}}}}
	if UpgradeInProgress(obj) {
		t.Fatal("resolved upgrade message must not block updates")
	}
	obj.Object["status"].(map[string]interface{})["conditions"] = []interface{}{
		map[string]interface{}{"type": "UpToDate", "status": "False"},
	}
	if !UpgradeInProgress(obj) {
		t.Fatal("UpToDate=False must block new updates")
	}
}

func TestVKSReconciliationPending(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{"generation": int64(5)},
		"status": map[string]interface{}{
			"observedGeneration": int64(4),
			"controlPlane":       map[string]interface{}{"desiredReplicas": int64(3), "availableReplicas": int64(3), "upToDateReplicas": int64(3)},
			"workers":            map[string]interface{}{"desiredReplicas": int64(5), "availableReplicas": int64(4), "upToDateReplicas": int64(5)},
		},
	}}
	if !ReconciliationPending(obj) {
		t.Fatal("unobserved generation and unavailable worker must remain updating")
	}
	_ = unstructured.SetNestedField(obj.Object, int64(5), "status", "observedGeneration")
	_ = unstructured.SetNestedField(obj.Object, int64(5), "status", "workers", "availableReplicas")
	_ = unstructured.SetNestedField(obj.Object, int64(6), "status", "workers", "replicas")
	if !ReconciliationPending(obj) {
		t.Fatal("extra worker being removed must remain updating")
	}
	_ = unstructured.SetNestedField(obj.Object, int64(5), "status", "workers", "replicas")
	_ = unstructured.SetNestedSlice(obj.Object, []interface{}{map[string]interface{}{"type": "ScalingDown", "status": "True"}}, "status", "conditions")
	if !ReconciliationPending(obj) {
		t.Fatal("ScalingDown condition must remain updating")
	}
	_ = unstructured.SetNestedSlice(obj.Object, []interface{}{map[string]interface{}{"type": "ScalingDown", "status": "False"}}, "status", "conditions")
	if ReconciliationPending(obj) {
		t.Fatal("fully available cluster must complete updating")
	}
}

func TestVKSUpgradeConditionKeepsReconciliationPending(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{"generation": int64(5)},
		"status": map[string]interface{}{
			"observedGeneration": int64(5),
			"controlPlane":       map[string]interface{}{"desiredReplicas": int64(3), "replicas": int64(3), "readyReplicas": int64(3), "availableReplicas": int64(3), "upToDateReplicas": int64(3)},
			"workers":            map[string]interface{}{"desiredReplicas": int64(5), "replicas": int64(5), "readyReplicas": int64(5), "availableReplicas": int64(5), "upToDateReplicas": int64(5)},
			"conditions": []interface{}{
				map[string]interface{}{"type": "Available", "status": "True"},
				map[string]interface{}{"type": "TopologyReconciled", "status": "False", "reason": "ClusterUpgrading", "message": "Cluster is upgrading"},
			},
		},
	}}
	if !ReconciliationPending(obj) {
		t.Fatal("upgrade must remain pending even when replica counts temporarily match")
	}
	if !UpgradeInProgress(obj) {
		t.Fatal("active Supervisor upgrade must be recognized after MAP state becomes RUNNING")
	}
	_ = unstructured.SetNestedSlice(obj.Object, []interface{}{
		map[string]interface{}{"type": "Available", "status": "True"},
		map[string]interface{}{"type": "TopologyReconciled", "status": "True"},
	}, "status", "conditions")
	if ReconciliationPending(obj) || UpgradeInProgress(obj) {
		t.Fatal("completed topology reconciliation must clear upgrade")
	}
}
