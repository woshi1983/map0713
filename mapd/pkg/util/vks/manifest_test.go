package vks

import (
	"testing"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestBuildClusterUsesSingleWorkerPoolAndOwner(t *testing.T) {
	id := uuid.New()
	spec := model.ClusterVKSSpec{Namespace: "app", ClusterClass: "builtin-generic-v3.4.0", ClassNamespace: "vmware-system-vks-public", OSName: "ubuntu", StorageClass: "vsan", CtlPlaneVmClass: "best-effort-medium", WorkerVmClass: "best-effort-medium", CtlPlaneReplicas: 3, WorkerReplicas: 2, ClusterCIDR: "192.168.0.0/16", ServiceCIDR: "10.96.0.0/12", ServiceDomain: "cluster.local"}
	obj, err := BuildCluster(id, "test-vks", "v1.33.3---vmware.1", spec)
	if err != nil {
		t.Fatal(err)
	}
	if obj.GetLabels()[OwnerLabel] != id.String() {
		t.Fatalf("missing MAP ownership label: %v", obj.GetLabels())
	}
	version, _, _ := unstructured.NestedString(obj.Object, "spec", "topology", "version")
	if version != "v1.33.3+vmware.1" {
		t.Fatalf("wrong topology version: %q", version)
	}
	classNS, _, _ := unstructured.NestedString(obj.Object, "spec", "topology", "classNamespace")
	if classNS != spec.ClassNamespace {
		t.Fatalf("wrong class namespace: %q", classNS)
	}
	pools, _, _ := unstructured.NestedSlice(obj.Object, "spec", "topology", "workers", "machineDeployments")
	if len(pools) != 1 {
		t.Fatalf("want one worker pool, got %d", len(pools))
	}
}

func TestBuildClusterRejectsUnsupportedCNI(t *testing.T) {
	_, err := BuildCluster(uuid.New(), "test-vks", "v1.33.3", model.ClusterVKSSpec{CNI: "calico"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestBuildClusterPassesTrustedCAsToNodeConfiguration(t *testing.T) {
	spec := model.ClusterVKSSpec{
		Namespace: "team-a", ClusterClass: "builtin-generic-v3.4.0", OSName: "photon",
		StorageClass: "vsan", CtlPlaneVmClass: "small", WorkerVmClass: "medium",
		CtlPlaneReplicas: 1, WorkerReplicas: 1, ClusterCIDR: "192.168.0.0/16",
		ServiceCIDR: "10.96.0.0/12", ServiceDomain: "cluster.local",
		TrustedCAs: [][]byte{[]byte("-----BEGIN CERTIFICATE-----\nexample\n-----END CERTIFICATE-----\n")},
	}
	obj, err := BuildCluster(uuid.New(), "demo", "v1.33.3", spec)
	if err != nil {
		t.Fatal(err)
	}
	cas, found, err := unstructured.NestedSlice(obj.Object, "spec", "topology", "variables")
	if err != nil || !found || len(cas) != 3 {
		t.Fatalf("expected osConfiguration trust variable, got %v: %v", cas, err)
	}
	trust := cas[2].(map[string]interface{})
	if trust["name"] != "osConfiguration" {
		t.Fatalf("unexpected trust variable: %v", trust)
	}
	configuration := trust["value"].(map[string]interface{})
	caEntries := configuration["trust"].(map[string]interface{})["additionalTrustedCAs"].([]interface{})
	if len(caEntries) != 1 || caEntries[0].(map[string]interface{})["caCert"].(map[string]interface{})["content"] != string(spec.TrustedCAs[0]) {
		t.Fatalf("trusted CA content was not propagated: %v", caEntries)
	}
}

func TestBuildClusterRequiresTopologyAndPositiveReplicas(t *testing.T) {
	base := model.ClusterVKSSpec{
		Namespace: "team-a", ClusterClass: "builtin-generic-v3.4.0", OSName: "photon",
		StorageClass: "vsan", CtlPlaneVmClass: "small", WorkerVmClass: "medium",
		CtlPlaneReplicas: 1, WorkerReplicas: 2, ClusterCIDR: "192.168.0.0/16",
		ServiceCIDR: "10.96.0.0/12", ServiceDomain: "cluster.local",
	}
	for _, tc := range []struct {
		name string
		edit func(*model.ClusterVKSSpec)
	}{
		{"no namespace", func(s *model.ClusterVKSSpec) { s.Namespace = "" }},
		{"no storage", func(s *model.ClusterVKSSpec) { s.StorageClass = "" }},
		{"no worker", func(s *model.ClusterVKSSpec) { s.WorkerReplicas = 0 }},
		{"no network", func(s *model.ClusterVKSSpec) { s.ClusterCIDR = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := base
			tc.edit(&s)
			if _, err := BuildCluster(uuid.New(), "demo", "v1.33.3", s); err == nil {
				t.Fatal("expected invalid VKS topology to be rejected")
			}
		})
	}
}
