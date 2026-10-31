package extension

import (
	"encoding/json"
	"strings"
	"testing"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

func TestUnmarshalGrafanaEnvAcceptsPrometheusEndpointWithPort(t *testing.T) {
	endpoints, _ := json.Marshal([]string{"http://prometheus.example:32699"})
	ext := &model.ClusterExtension{Related: []model.Cluster{{Name: "monitoring", Extensions: []model.ClusterExtension{{Name: model.ClusterExtensionNamePrometheus, State: model.ClusterExtensionStateEnabled, Status: model.ClusterExtensionStatusProvisioned, Endpoints: endpoints}}}}}
	env := map[string]string{}
	if err := unmarshalGrafanaEnv(ext, env); err != nil {
		t.Fatal(err)
	}
	if got := env["CLUSTER0_PROMETHEUS_APISERVER"]; got != "http://prometheus.example:32699" {
		t.Fatalf("Prometheus endpoint = %q", got)
	}
	if got := env["GRAFANA_CONTEXT_PREFIX_MAPPING"]; got != "CLUSTER0" {
		t.Fatalf("mapping = %q", got)
	}
}

func TestUnmarshalGrafanaEnvRequiresPrometheusEndpoint(t *testing.T) {
	ext := &model.ClusterExtension{Related: []model.Cluster{{Name: "monitoring"}}}
	if err := unmarshalGrafanaEnv(ext, map[string]string{}); err == nil || !strings.Contains(err.Error(), "Prometheus") {
		t.Fatalf("expected missing Prometheus endpoint error, got %v", err)
	}
}

func TestUnmarshalGrafanaEnvUsesFirstValidEndpointAsDefault(t *testing.T) {
	endpoints, _ := json.Marshal([]string{"https://prometheus.example:9090"})
	ext := &model.ClusterExtension{Related: []model.Cluster{
		{Name: "a-without-endpoint"},
		{Name: "b-with-endpoint", Extensions: []model.ClusterExtension{{Name: model.ClusterExtensionNamePrometheus, State: model.ClusterExtensionStateEnabled, Status: model.ClusterExtensionStatusProvisioned, Endpoints: endpoints}}},
	}}
	env := map[string]string{}
	if err := unmarshalGrafanaEnv(ext, env); err != nil {
		t.Fatal(err)
	}
	if got := env["CLUSTER0_PROMETHEUS_APISERVER"]; got != "https://prometheus.example:9090" {
		t.Fatalf("default endpoint = %q", got)
	}
}
