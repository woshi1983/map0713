package extension

import (
	"encoding/json"
	"strings"
	"testing"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

func TestGrafanaWaitsForPrometheus(t *testing.T) {
	ext := &model.ClusterExtension{Related: []model.Cluster{{Name: "monitoring", Extensions: []model.ClusterExtension{{
		Name: model.ClusterExtensionNamePrometheus, State: model.ClusterExtensionStateEnabling,
		Status: model.ClusterExtensionStatusProvisioning,
	}}}}}
	err := unmarshalGrafanaEnv(ext, map[string]string{})
	if err == nil || IsCriticalError(err) || !strings.Contains(err.Error(), "waiting for Prometheus") {
		t.Fatalf("expected retryable Prometheus dependency error, got %v", err)
	}
}

func TestGrafanaFailsWhenPrometheusFails(t *testing.T) {
	ext := &model.ClusterExtension{Related: []model.Cluster{{Name: "monitoring", Extensions: []model.ClusterExtension{{
		Name: model.ClusterExtensionNamePrometheus, State: model.ClusterExtensionStateEnabling,
		Status: model.ClusterExtensionStatusFailed,
	}}}}}
	err := unmarshalGrafanaEnv(ext, map[string]string{})
	if err == nil || !IsCriticalError(err) || !strings.Contains(err.Error(), "Prometheus") {
		t.Fatalf("expected critical Prometheus dependency error, got %v", err)
	}
}

func TestFluentBitWaitsForElasticsearchWithoutParsingNull(t *testing.T) {
	ext := &model.ClusterExtension{Related: []model.Cluster{{Name: "logging", Extensions: []model.ClusterExtension{{
		Name: model.ClusterExtensionNameElasticSearch, State: model.ClusterExtensionStateEnabling,
		Status: model.ClusterExtensionStatusProvisioning,
	}}}}}
	err := unmarshalFluentBitEnv(ext, map[string]string{})
	if err == nil || IsCriticalError(err) || !strings.Contains(err.Error(), "Elasticsearch") || strings.Contains(err.Error(), "JSON") {
		t.Fatalf("expected retryable Elasticsearch dependency error, got %v", err)
	}
}

func TestFluentBitFailsWhenElasticsearchFails(t *testing.T) {
	ext := &model.ClusterExtension{Related: []model.Cluster{{Name: "logging", Extensions: []model.ClusterExtension{{
		Name: model.ClusterExtensionNameElasticSearch, Status: model.ClusterExtensionStatusFailed,
	}}}}}
	err := unmarshalFluentBitEnv(ext, map[string]string{})
	if err == nil || !IsCriticalError(err) || !strings.Contains(err.Error(), "Elasticsearch") {
		t.Fatalf("expected critical Elasticsearch dependency error, got %v", err)
	}
}

func TestFluentBitFailsWhenElasticsearchIsDisabled(t *testing.T) {
	ext := &model.ClusterExtension{Related: []model.Cluster{{Name: "logging", Extensions: []model.ClusterExtension{{
		Name: model.ClusterExtensionNameElasticSearch, State: model.ClusterExtensionStateUnknown,
	}}}}}
	err := unmarshalFluentBitEnv(ext, map[string]string{})
	if err == nil || !IsCriticalError(err) || !strings.Contains(err.Error(), "not enabled") {
		t.Fatalf("expected disabled Elasticsearch dependency error, got %v", err)
	}
}

func TestFluentBitUsesProvisionedElasticsearch(t *testing.T) {
	endpoints, _ := json.Marshal([]string{"http://elasticsearch.example:9200"})
	ext := &model.ClusterExtension{Related: []model.Cluster{{Name: "logging", Extensions: []model.ClusterExtension{{
		Name: model.ClusterExtensionNameElasticSearch, State: model.ClusterExtensionStateEnabled,
		Status: model.ClusterExtensionStatusProvisioned, Endpoints: endpoints,
	}}}}}
	env := map[string]string{}
	if err := unmarshalFluentBitEnv(ext, env); err != nil {
		t.Fatal(err)
	}
	if got := env["CLUSTER0_ELASTICSEARCH_API_HOST"]; got != "elasticsearch.example" {
		t.Fatalf("Elasticsearch host = %q", got)
	}
}

func TestProvisionedDependencyWithEmptyEndpointListFailsClearly(t *testing.T) {
	dependency := &model.ClusterExtension{
		Name:      model.ClusterExtensionNamePrometheus,
		State:     model.ClusterExtensionStateEnabled,
		Status:    model.ClusterExtensionStatusProvisioned,
		Endpoints: []byte("[]"),
	}
	err := requireRelatedExtensionReady("monitoring", "Prometheus", dependency)
	if err == nil || !IsCriticalError(err) || !strings.Contains(err.Error(), "has no endpoint") {
		t.Fatalf("expected clear missing endpoint error, got %v", err)
	}
}

func TestFluentBitWithoutRelatedElasticsearchKeepsEmptyMapping(t *testing.T) {
	env := map[string]string{}
	if err := unmarshalFluentBitEnv(&model.ClusterExtension{}, env); err != nil {
		t.Fatal(err)
	}
	if got := env["FLUENTBIT_CONTEXT_PREFIX_MAPPING"]; got != "" {
		t.Fatalf("empty mapping = %q", got)
	}
}
