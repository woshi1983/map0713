package vks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// setDefaultStorageClass selects the VKS guest-cluster default. An existing
// availableStorageClasses restriction is retained and extended only when it
// does not already include the selected class.
func setDefaultStorageClass(obj *unstructured.Unstructured, name string) (*unstructured.Unstructured, bool, error) {
	updated := obj.DeepCopy()
	variables, found, err := unstructured.NestedSlice(updated.Object, "spec", "topology", "variables")
	if err != nil || !found {
		return nil, false, fmt.Errorf("VKS topology variables are missing or invalid: %v", err)
	}
	for i, raw := range variables {
		variable, ok := raw.(map[string]interface{})
		if !ok {
			return nil, false, fmt.Errorf("VKS topology variable %d is invalid", i)
		}
		if variable["name"] != "vsphereOptions" {
			continue
		}
		options, ok := variable["value"].(map[string]interface{})
		if !ok {
			return nil, false, fmt.Errorf("VKS vsphereOptions is invalid")
		}
		persistent, exists := options["persistentVolumes"]
		if !exists {
			persistent = map[string]interface{}{}
		}
		volumes, ok := persistent.(map[string]interface{})
		if !ok {
			return nil, false, fmt.Errorf("VKS persistentVolumes is invalid")
		}
		changed := volumes["defaultStorageClass"] != name
		if available, restricted := volumes["availableStorageClasses"]; restricted {
			classes, ok := available.([]interface{})
			if !ok {
				return nil, false, fmt.Errorf("VKS availableStorageClasses is invalid")
			}
			present := false
			for _, existing := range classes {
				if existing == name {
					present = true
					break
				}
			}
			if !present {
				volumes["availableStorageClasses"] = append(classes, name)
				changed = true
			}
		}
		if !changed {
			return obj, false, nil
		}
		volumes["defaultStorageClass"] = name
		options["persistentVolumes"] = volumes
		variable["value"] = options
		variables[i] = variable
		if err := unstructured.SetNestedSlice(updated.Object, variables, "spec", "topology", "variables"); err != nil {
			return nil, false, err
		}
		return updated, true, nil
	}
	variables = append(variables, map[string]interface{}{"name": "vsphereOptions", "value": map[string]interface{}{
		"persistentVolumes": map[string]interface{}{"defaultStorageClass": name},
	}})
	if err := unstructured.SetNestedSlice(updated.Object, variables, "spec", "topology", "variables"); err != nil {
		return nil, false, err
	}
	return updated, true, nil
}

// EnsureDefaultStorageClass repairs VKS clusters created before MAP passed the
// selected StorageClass to the guest cluster. Ownership and UID are checked
// before changing the Supervisor Cluster.
func (c *Client) EnsureDefaultStorageClass(ctx context.Context, id uuid.UUID, name string, spec model.ClusterVKSSpec) (bool, error) {
	obj, err := c.GetCluster(ctx, spec.Namespace, name)
	if err != nil {
		return false, err
	}
	if obj.GetLabels()[OwnerLabel] != id.String() || spec.RemoteUID == "" || string(obj.GetUID()) != spec.RemoteUID {
		return false, fmt.Errorf("VKS cluster %s/%s ownership or UID mismatch", spec.Namespace, name)
	}
	updated, changed, err := setDefaultStorageClass(obj, spec.StorageClass)
	if err != nil || !changed {
		return false, err
	}
	if _, err := c.dynamic.Resource(c.capiResource("clusters")).Namespace(spec.Namespace).Update(ctx, updated, metav1.UpdateOptions{}); err != nil {
		return false, err
	}
	return true, nil
}
