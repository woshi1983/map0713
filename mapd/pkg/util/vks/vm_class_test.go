package vks

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func TestListVMClassesUsesNamespacedV1Alpha2(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "vmoperator.vmware.com", Version: "v1alpha2", Resource: "virtualmachineclasses"}
	item := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "vmoperator.vmware.com/v1alpha2", "kind": "VirtualMachineClass",
		"metadata": map[string]interface{}{"name": "best-effort-small", "namespace": "vlan803"},
		"spec":     map[string]interface{}{"hardware": map[string]interface{}{"cpus": int64(2), "memory": "4Gi"}},
	}}
	client := &Client{dynamic: dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{gvr: "VirtualMachineClassList"}, item)}
	classes, err := client.ListVMClasses(context.Background(), "vlan803")
	if err != nil || len(classes) != 1 || classes[0].GetName() != "best-effort-small" {
		t.Fatalf("expected namespaced VMClass, got %v, %v", classes, err)
	}
	selected, err := client.GetVMClass(context.Background(), "vlan803", "best-effort-small")
	if err != nil {
		t.Fatal(err)
	}
	cpus, memoryMi, err := VMClassResources(selected)
	if err != nil || cpus != 2 || memoryMi < 4096 {
		t.Fatalf("wrong VMClass resources: CPU=%d memoryMi=%d error=%v", cpus, memoryMi, err)
	}
}
