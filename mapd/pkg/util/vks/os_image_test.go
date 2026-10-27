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
		map[schema.GroupVersionResource]string{gvr: "OSImageList", releaseGVR: "KubernetesReleaseList"}, &image)}
	version, err := client.ResolveOSVersion(context.Background(), "v1.33.3---vmware.2", "photon", "")
	if err != nil || version != "5.1" {
		t.Fatalf("OSImage API selection failed: got %q, %v", version, err)
	}
}

func TestListOSImagesReturnsReadySelectableImages(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "run.tanzu.vmware.com", Version: "v1alpha3", Resource: "osimages"}
	release := unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "kubernetes.vmware.com/v1alpha1", "kind": "KubernetesRelease",
		"metadata": map[string]interface{}{"name": "v1.34.2---vmware.2-vkr.2"},
		"spec": map[string]interface{}{"version": "v1.34.2+vmware.2-vkr.2", "osImages": []interface{}{
			map[string]interface{}{"name": "photon-ready"},
		}},
		"status": map[string]interface{}{"conditions": []interface{}{
			map[string]interface{}{"type": "Ready", "status": "True"},
			map[string]interface{}{"type": "Compatible", "status": "True"},
		}},
	}}
	ready := osImage("photon-ready", "VMwarePhoton", "5.0", "v1.34.2+vmware.2")
	ready.SetAPIVersion("run.tanzu.vmware.com/v1alpha3")
	ready.SetKind("OSImage")
	unready := osImage("photon-unready", "photon", "4.0", "v1.34.2+vmware.2")
	unready.SetAPIVersion("run.tanzu.vmware.com/v1alpha3")
	unready.SetKind("OSImage")
	unready.Object["status"] = map[string]interface{}{"conditions": []interface{}{
		map[string]interface{}{"type": "Ready", "status": "False"},
	}}
	client := &Client{dynamic: dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{gvr: "OSImageList", releaseGVR: "KubernetesReleaseList"}, &ready, &unready, &release)}
	images, err := client.ListOSImages(context.Background(), "vlan803")
	if err != nil {
		t.Fatal(err)
	}
	if len(images) != 1 || images[0].Name != "photon-ready" || images[0].OSName != "photon" ||
		images[0].OSVersion != "5.0" || images[0].KubernetesVersion != "v1.34.2+vmware.2" ||
		len(images[0].Releases) != 1 || images[0].Releases[0] != "v1.34.2+vmware.2-vkr.2" {
		t.Fatalf("unexpected selectable OS images: %+v", images)
	}
}

func TestResolveOSVersionHonorsReleaseImageReferences(t *testing.T) {
	gvr := schema.GroupVersionResource{Group: "run.tanzu.vmware.com", Version: "v1alpha3", Resource: "osimages"}
	newImage := osImage("photon-new", "photon", "6.0", "v1.34.2+vmware.2")
	oldImage := osImage("photon-old", "photon", "5.0", "v1.34.2+vmware.2")
	for _, image := range []*unstructured.Unstructured{&newImage, &oldImage} {
		image.SetAPIVersion("run.tanzu.vmware.com/v1alpha3")
		image.SetKind("OSImage")
	}
	release := unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "kubernetes.vmware.com/v1alpha1", "kind": "KubernetesRelease",
		"metadata": map[string]interface{}{"name": "v1.34.2---vmware.2-vkr.2"},
		"spec":     map[string]interface{}{"osImages": []interface{}{map[string]interface{}{"name": "photon-old"}}},
	}}
	client := &Client{dynamic: dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{gvr: "OSImageList", releaseGVR: "KubernetesReleaseList"},
		&newImage, &oldImage, &release)}
	version, err := client.ResolveOSVersion(context.Background(), release.GetName(), "photon", "")
	if err != nil || version != "5.0" {
		t.Fatalf("wanted release-referenced photon 5.0, got %q, %v", version, err)
	}
}

func TestSelectableVMImageUsesOSVersionRatherThanProductVersion(t *testing.T) {
	image := unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{"name": "photon-vmi", "labels": map[string]interface{}{
			"vmi.vmoperator.vmware.com/kubernetes-version": "v1.34.2+vmware.2",
			"vmi.vmoperator.vmware.com/os-type":            "vmwarePhoton64Guest",
			"vmi.vmoperator.vmware.com/os-version":         "5.0",
		}},
		"spec": map[string]interface{}{"productInfo": map[string]interface{}{"version": "v1.34.2+vmware.2"}},
	}}
	info, ok := selectableVMImage(&image)
	if !ok || info.OSName != "photon" || info.OSVersion != "5.0" {
		t.Fatalf("unexpected VMI OS information: %+v, %v", info, ok)
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
