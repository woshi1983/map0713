package resolver

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	graphqlmodel "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	dbmodel "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/vks"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

func TestRefreshVKSClusterReplicaCountsPreservesStoredValuesOnSupervisorFailure(t *testing.T) {
	id := uuid.New()
	cluster := &dbmodel.Cluster{ID: id, Name: "demo"}
	result := &graphqlmodel.Cluster{Spec: &graphqlmodel.VKSClusterSpec{Namespace: "team", RemoteUID: "remote-1", CtlPlaneReplicas: 3, WorkerReplicas: 5}}
	if err := refreshVKSClusterReplicaCounts(context.Background(), cluster, result, func(_ context.Context, namespace, name string) (*unstructured.Unstructured, error) {
		if namespace != "team" || name != "demo" {
			t.Fatalf("unexpected Supervisor lookup %s/%s", namespace, name)
		}
		return nil, errors.New("Unauthorized")
	}); err == nil {
		t.Fatal("Supervisor error must be reported for logging")
	}
	spec := result.Spec.(*graphqlmodel.VKSClusterSpec)
	if spec.CtlPlaneReplicas != 3 || spec.WorkerReplicas != 5 {
		t.Fatalf("stored counts changed after transient error: %+v", spec)
	}
	remote := &unstructured.Unstructured{Object: map[string]interface{}{
		"spec": map[string]interface{}{"topology": map[string]interface{}{
			"controlPlane": map[string]interface{}{"replicas": int64(1)},
			"workers": map[string]interface{}{"machineDeployments": []interface{}{
				map[string]interface{}{"name": "pool-0", "replicas": int64(4)},
			}},
		}},
	}}
	remote.SetUID(types.UID("remote-1"))
	remote.SetLabels(map[string]string{vks.OwnerLabel: id.String()})
	foreign := remote.DeepCopy()
	foreign.SetLabels(map[string]string{vks.OwnerLabel: uuid.NewString()})
	if err := refreshVKSClusterReplicaCounts(context.Background(), cluster, result, func(context.Context, string, string) (*unstructured.Unstructured, error) {
		return foreign, nil
	}); !errors.Is(err, errVKSClusterOwnershipMismatch) {
		t.Fatalf("foreign Supervisor object must raise integrity error, got %v", err)
	}
	if err := refreshVKSClusterReplicaCounts(context.Background(), cluster, result, func(context.Context, string, string) (*unstructured.Unstructured, error) {
		return remote, nil
	}); err != nil {
		t.Fatalf("valid Supervisor refresh failed: %v", err)
	}
	if spec.CtlPlaneReplicas != 1 || spec.WorkerReplicas != 4 {
		t.Fatalf("replica counts were not refreshed: %+v", spec)
	}
}
