package vks

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func TestChooseOSVersionMatchesReleaseAndOS(t *testing.T) {
	images := []unstructured.Unstructured{
		osImage("old", "photon", "5.0", "v1.32.0+vmware.1"),
		osImage("ubuntu", "ubuntu", "24.04", "v1.33.3+vmware.2"),
		osImage("photon", "photon", "5.1", "v1.33.3+vmware.2-vkr.1"),
	}
	version, err := chooseOSVersion(images, "v1.33.3---vmware.2", "photon", "")
	if err != nil || version != "5.1" {
		t.Fatalf("wanted compatible photon 5.1, got %q, %v", version, err)
	}
	if _, err := chooseOSVersion(images, "v1.33.3---vmware.2", "photon", "5.0"); err == nil {
		t.Fatal("explicit incompatible OS version must be rejected")
	}
}

func TestResolveOSVersionQueriesSupervisorOSImages(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "run.tanzu.vmware.com", Version: "v1alpha3", Resource: "osimages"}
	image := osImage("photon", "photon", "5.1", "v1.33.3+vmware.2")
	image.SetAPIVersion("run.tanzu.vmware.com/v1alpha3")
	image.SetKind("OSImage")
	client := &Client{dynamic: dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{gvr: "OSImageList"}, &image)}
	version, err := client.ResolveOSVersion(context.Background(), "v1.33.3---vmware.2", "photon", "")
	if err != nil || version != "5.1" {
		t.Fatalf("OSImage API selection failed: got %q, %v", version, err)
	}
}

func osImage(name, osName, osVersion, kubernetesVersion string) unstructured.Unstructured {
	return unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{"name": name, "labels": map[string]interface{}{
			"os-name": osName, "os-version": osVersion,
		}},
		"spec": map[string]interface{}{"kubernetesVersion": kubernetesVersion},
	}}
}
