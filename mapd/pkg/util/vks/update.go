package vks

import (
	"context"
	"fmt"
	"strings"

	semver "github.com/Masterminds/semver/v3"
	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// UpdateTopology updates the owned Supervisor Cluster using its live resourceVersion.
// A nil count leaves that dimension unchanged; targetVersion is the MAP semver.
func (c *Client) UpdateTopology(ctx context.Context, id uuid.UUID, name string, spec model.ClusterVKSSpec, controlPlane, workers *int64, targetVersion string) (model.ClusterVKSSpec, error) {
	obj, err := c.GetCluster(ctx, spec.Namespace, name)
	if err != nil {
		return spec, err
	}
	if obj.GetLabels()[OwnerLabel] != id.String() || spec.RemoteUID == "" || string(obj.GetUID()) != spec.RemoteUID {
		return spec, fmt.Errorf("VKS cluster %s/%s ownership or UID mismatch", spec.Namespace, name)
	}
	if !Available(obj) || strings.EqualFold(obj.GetAnnotations()["cluster.x-k8s.io/paused"], "true") || vksUpgradeInProgress(obj) {
		return spec, fmt.Errorf("VKS cluster %s/%s is not available, is paused, or is upgrading", spec.Namespace, name)
	}
	currentCP, currentWorkers, err := VKSReplicas(obj)
	if err != nil {
		return spec, err
	}
	newCP, newWorkers := currentCP, currentWorkers
	if controlPlane != nil {
		newCP = *controlPlane
	}
	if workers != nil {
		newWorkers = *workers
	}
	if newCP != 1 && newCP != 3 {
		return spec, fmt.Errorf("VKS control plane replicas must be 1 or 3")
	}
	if newWorkers < 1 || newWorkers > 100 {
		return spec, fmt.Errorf("VKS worker replicas must be between 1 and 100")
	}
	if controlPlane != nil || workers != nil {
		if err := setVKSReplicas(obj, newCP, newWorkers); err != nil {
			return spec, err
		}
	}
	if targetVersion != "" {
		liveVersion, _, _ := unstructured.NestedString(obj.Object, "spec", "topology", "version")
		if !validVKSUpgrade(liveVersion, targetVersion) {
			return spec, fmt.Errorf("VKS upgrade must target a higher patch in the current minor or the next minor")
		}
		release, err := c.ResolveRelease(ctx, targetVersion)
		if err != nil {
			return spec, err
		}
		osVersion, err := c.ResolveOSVersion(ctx, release, spec.OSName, "")
		if err != nil {
			return spec, err
		}
		if osVersion == "" {
			return spec, fmt.Errorf("no ready VKS OSImage for %s and release %s", spec.OSName, release)
		}
		if err := setVKSOSSelector(obj, spec.OSName, osVersion); err != nil {
			return spec, err
		}
		if err := unstructured.SetNestedField(obj.Object, strings.ReplaceAll(release, "---", "+"), "spec", "topology", "version"); err != nil {
			return spec, err
		}
		spec.OSVersion = osVersion
	}
	if _, err := c.dynamic.Resource(c.capiResource("clusters")).Namespace(spec.Namespace).Update(ctx, obj, metav1.UpdateOptions{}); err != nil {
		return spec, err
	}
	spec.CtlPlaneReplicas, spec.WorkerReplicas = newCP, newWorkers
	return spec, nil
}

func vksUpgradeInProgress(obj *unstructured.Unstructured) bool {
	for _, path := range [][]string{{"status", "conditions"}, {"status", "v1beta2", "conditions"}} {
		conditions, _, _ := unstructured.NestedSlice(obj.Object, path...)
		for _, raw := range conditions {
			condition, ok := raw.(map[string]interface{})
			if !ok || condition["status"] == "True" {
				continue
			}
			if condition["type"] == "UpToDate" || strings.Contains(strings.ToLower(fmt.Sprint(condition["message"])), "upgrad") {
				return true
			}
		}
	}
	return false
}

// ReconciliationPending reports whether the Supervisor has observed the latest
// Cluster spec and brought all requested replicas up to date and available.
func ReconciliationPending(obj *unstructured.Unstructured) bool {
	observed, found, _ := unstructured.NestedInt64(obj.Object, "status", "observedGeneration")
	if found && observed < obj.GetGeneration() {
		return true
	}
	for _, role := range []string{"controlPlane", "workers"} {
		desired, found, _ := unstructured.NestedInt64(obj.Object, "status", role, "desiredReplicas")
		if !found || desired == 0 {
			continue
		}
		for _, field := range []string{"replicas", "readyReplicas", "availableReplicas", "upToDateReplicas"} {
			count, reported, _ := unstructured.NestedInt64(obj.Object, "status", role, field)
			if reported && count != desired {
				return true
			}
		}
	}
	for _, path := range [][]string{{"status", "conditions"}, {"status", "v1beta2", "conditions"}} {
		conditions, _, _ := unstructured.NestedSlice(obj.Object, path...)
		for _, raw := range conditions {
			condition, ok := raw.(map[string]interface{})
			if !ok {
				continue
			}
			switch condition["type"] {
			case "ScalingUp", "ScalingDown", "RollingOut":
				if condition["status"] == "True" {
					return true
				}
			}
		}
	}
	return false
}

func VKSReplicas(obj *unstructured.Unstructured) (int64, int64, error) {
	cp, found, err := unstructured.NestedInt64(obj.Object, "spec", "topology", "controlPlane", "replicas")
	if err != nil || !found {
		return 0, 0, fmt.Errorf("VKS control plane replica count is missing")
	}
	pools, found, err := unstructured.NestedSlice(obj.Object, "spec", "topology", "workers", "machineDeployments")
	if err != nil || !found {
		return 0, 0, fmt.Errorf("VKS worker pools are missing")
	}
	for _, raw := range pools {
		pool, ok := raw.(map[string]interface{})
		if !ok || pool["name"] != "pool-0" {
			continue
		}
		count, ok := pool["replicas"].(int64)
		if ok {
			return cp, count, nil
		}
	}
	return 0, 0, fmt.Errorf("VKS worker pool pool-0 is missing or has no replica count")
}

func setVKSReplicas(obj *unstructured.Unstructured, controlPlane, workers int64) error {
	if controlPlane != 1 && controlPlane != 3 || workers < 1 || workers > 100 {
		return fmt.Errorf("VKS control plane replicas must be 1 or 3 and worker replicas must be between 1 and 100")
	}
	pools, found, err := unstructured.NestedSlice(obj.Object, "spec", "topology", "workers", "machineDeployments")
	if err != nil || !found {
		return fmt.Errorf("VKS worker pools are missing")
	}
	matched := false
	for i, raw := range pools {
		pool, ok := raw.(map[string]interface{})
		if ok && pool["name"] == "pool-0" {
			updated := make(map[string]interface{}, len(pool))
			for key, value := range pool {
				updated[key] = value
			}
			updated["replicas"] = workers
			pools[i] = updated
			matched = true
		}
	}
	if !matched {
		return fmt.Errorf("VKS worker pool pool-0 is missing")
	}
	if err := unstructured.SetNestedField(obj.Object, controlPlane, "spec", "topology", "controlPlane", "replicas"); err != nil {
		return err
	}
	return unstructured.SetNestedSlice(obj.Object, pools, "spec", "topology", "workers", "machineDeployments")
}

func validVKSUpgrade(current, target string) bool {
	old, err := semver.NewVersion(strings.ReplaceAll(current, "---", "+"))
	if err != nil {
		return false
	}
	next, err := semver.NewVersion(strings.ReplaceAll(target, "---", "+"))
	if err != nil || old.Major() != next.Major() {
		return false
	}
	return next.Minor() == old.Minor()+1 || next.Minor() == old.Minor() && next.Patch() > old.Patch()
}

func setVKSOSSelector(obj *unstructured.Unstructured, osName, osVersion string) error {
	selector := "os-name=" + osName + ",os-version=" + osVersion
	paths := [][]string{{"spec", "topology", "controlPlane", "metadata", "annotations"}}
	pools, found, err := unstructured.NestedSlice(obj.Object, "spec", "topology", "workers", "machineDeployments")
	if err != nil || !found {
		return fmt.Errorf("VKS worker pools are missing")
	}
	for i, raw := range pools {
		pool, ok := raw.(map[string]interface{})
		if !ok {
			return fmt.Errorf("VKS worker pool %d is invalid", i)
		}
		metadata, _ := pool["metadata"].(map[string]interface{})
		annotations, _ := metadata["annotations"].(map[string]interface{})
		newAnnotations := make(map[string]interface{}, len(annotations)+1)
		for key, value := range annotations {
			newAnnotations[key] = value
		}
		newAnnotations["run.tanzu.vmware.com/resolve-os-image"] = selector
		newMetadata := make(map[string]interface{}, len(metadata)+1)
		for key, value := range metadata {
			newMetadata[key] = value
		}
		newMetadata["annotations"] = newAnnotations
		newPool := make(map[string]interface{}, len(pool))
		for key, value := range pool {
			newPool[key] = value
		}
		newPool["metadata"] = newMetadata
		pools[i] = newPool
	}
	annotations, _, _ := unstructured.NestedStringMap(obj.Object, paths[0]...)
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations["run.tanzu.vmware.com/resolve-os-image"] = selector
	if err := unstructured.SetNestedStringMap(obj.Object, annotations, paths[0]...); err != nil {
		return err
	}
	return unstructured.SetNestedSlice(obj.Object, pools, "spec", "topology", "workers", "machineDeployments")
}
