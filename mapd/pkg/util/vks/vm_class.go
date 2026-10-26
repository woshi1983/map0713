package vks

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var vmClassV1Alpha2 = schema.GroupVersionResource{Group: "vmoperator.vmware.com", Version: "v1alpha2", Resource: "virtualmachineclasses"}
var vmClassV1Alpha1 = schema.GroupVersionResource{Group: "vmoperator.vmware.com", Version: "v1alpha1", Resource: "virtualmachineclasses"}

// Recent Supervisors scope VMClasses to workload namespaces. The older API
// was cluster-scoped, so it remains a fallback for existing installations.
func (c *Client) ListVMClasses(ctx context.Context, namespace string) ([]unstructured.Unstructured, error) {
	if namespace == "" {
		return nil, fmt.Errorf("VKS namespace is required to list VMClasses")
	}
	list, err := c.dynamic.Resource(vmClassV1Alpha2).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if apierrors.IsNotFound(err) || apierrors.IsMethodNotSupported(err) {
		list, err = c.dynamic.Resource(vmClassV1Alpha1).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (c *Client) GetVMClass(ctx context.Context, namespace, name string) (*unstructured.Unstructured, error) {
	if namespace == "" || name == "" {
		return nil, fmt.Errorf("VKS namespace and VMClass name are required")
	}
	items, err := c.ListVMClasses(ctx, namespace)
	if err != nil {
		return nil, err
	}
	for i := range items {
		if items[i].GetName() == name {
			return &items[i], nil
		}
	}
	return nil, fmt.Errorf("VKS VMClass %q is not available in namespace %q", name, namespace)
}

func VMClassResources(item *unstructured.Unstructured) (int64, int64, error) {
	if item == nil {
		return 0, 0, fmt.Errorf("VKS VMClass is empty")
	}
	cpus, found, err := unstructured.NestedInt64(item.Object, "spec", "hardware", "cpus")
	if err != nil || !found || cpus <= 0 {
		return 0, 0, fmt.Errorf("VKS VMClass %s has invalid CPU count", item.GetName())
	}
	memory, found, err := unstructured.NestedString(item.Object, "spec", "hardware", "memory")
	if err != nil || !found {
		return 0, 0, fmt.Errorf("VKS VMClass %s has no memory size", item.GetName())
	}
	quantity, err := resource.ParseQuantity(memory)
	if err != nil || quantity.Sign() <= 0 {
		return 0, 0, fmt.Errorf("VKS VMClass %s has invalid memory size %q", item.GetName(), memory)
	}
	return cpus, quantity.ScaledValue(resource.Mega), nil
}
