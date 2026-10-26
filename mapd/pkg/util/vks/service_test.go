package vks

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func fakeCluster(name, namespace string, owner uuid.UUID, uid types.UID) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "cluster.x-k8s.io/v1beta1", "kind": "Cluster",
		"metadata": map[string]interface{}{"name": name, "namespace": namespace, "uid": string(uid), "labels": map[string]interface{}{OwnerLabel: owner.String()}},
	}}
}

func newFakeClient(objects ...runtime.Object) *Client {
	return &Client{dynamic: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), objects...)}
}

func TestCreateClusterRejectsExistingForeignObject(t *testing.T) {
	ownedByOther := fakeCluster("demo", "team-a", uuid.New(), types.UID("remote-1"))
	client := newFakeClient(ownedByOther)
	spec := model.ClusterVKSSpec{
		Namespace: "team-a", ClusterClass: "builtin", OSName: "photon", StorageClass: "vsan",
		CtlPlaneVmClass: "small", WorkerVmClass: "small", CtlPlaneReplicas: 1, WorkerReplicas: 1,
		ClusterCIDR: "192.168.0.0/16", ServiceCIDR: "10.96.0.0/12", ServiceDomain: "cluster.local",
	}
	if _, err := client.CreateCluster(context.Background(), uuid.New(), "demo", "v1.33.3", spec); err == nil {
		t.Fatal("foreign VKS Cluster must not be updated or adopted")
	}
}

func TestDeleteClusterChecksOwnerAndRemoteUID(t *testing.T) {
	id := uuid.New()
	client := newFakeClient(fakeCluster("demo", "team-a", id, types.UID("remote-1")))
	wrong := model.ClusterVKSSpec{Namespace: "team-a", RemoteUID: "remote-2"}
	if err := client.DeleteCluster(context.Background(), id, "demo", wrong); err == nil {
		t.Fatal("mismatched UID must prevent deletion")
	}
	correct := model.ClusterVKSSpec{Namespace: "team-a", RemoteUID: "remote-1"}
	if err := client.DeleteCluster(context.Background(), id, "demo", correct); err != nil {
		t.Fatalf("matching owner and UID should permit deletion: %v", err)
	}
}

func TestResolveReleasePrefersExactBuildMetadata(t *testing.T) {
	release := func(name, version string) *unstructured.Unstructured {
		return &unstructured.Unstructured{Object: map[string]interface{}{
			"apiVersion": "kubernetes.vmware.com/v1alpha1", "kind": "KubernetesRelease",
			"metadata": map[string]interface{}{"name": name},
			"spec":     map[string]interface{}{"version": version},
			"status": map[string]interface{}{"conditions": []interface{}{
				map[string]interface{}{"type": "Ready", "status": "True"},
				map[string]interface{}{"type": "Compatible", "status": "True"},
			}},
		}}
	}
	client := &Client{dynamic: dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{releaseGVR: "KubernetesReleaseList"},
		release("v1.33.3---vmware.1", "v1.33.3+vmware.1"),
		release("v1.33.3---vmware.2", "v1.33.3+vmware.2"),
	)}
	got, err := client.ResolveRelease(context.Background(), "v1.33.3+vmware.2")
	if err != nil || got != "v1.33.3---vmware.2" {
		t.Fatalf("exact release was not selected: got %q, error %v", got, err)
	}
}

func TestConfigureV1Beta2ClassReference(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{"spec": map[string]interface{}{
		"topology": map[string]interface{}{"class": "builtin", "classNamespace": "public"},
	}}}
	if err := configureClusterAPIVersion(obj, "v1beta2"); err != nil {
		t.Fatal(err)
	}
	name, found, _ := unstructured.NestedString(obj.Object, "spec", "topology", "classRef", "name")
	namespace, _, _ := unstructured.NestedString(obj.Object, "spec", "topology", "classRef", "namespace")
	_, oldField, _ := unstructured.NestedString(obj.Object, "spec", "topology", "class")
	if !found || name != "builtin" || namespace != "public" || oldField {
		t.Fatalf("unexpected v1beta2 topology: %v", obj.Object)
	}
}

func TestAvailableReadsV1Beta1EmbeddedConditions(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]interface{}{"status": map[string]interface{}{
		"v1beta2": map[string]interface{}{"conditions": []interface{}{map[string]interface{}{"type": "Available", "status": "True"}}},
	}}}
	if !Available(obj) {
		t.Fatal("embedded Available condition was ignored")
	}
}
