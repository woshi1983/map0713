package resolver

import (
	"strings"
	"testing"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

func TestPrometheusCanBeQueuedForGrafana(t *testing.T) {
	for _, status := range []model.ClusterExtensionStatus{
		model.ClusterExtensionStatusPending,
		model.ClusterExtensionStatusProvisioning,
	} {
		prometheus := &model.ClusterExtension{State: model.ClusterExtensionStateEnabling, Status: status}
		if reason := prometheusCannotServeGrafana(prometheus); reason != "" {
			t.Fatalf("status %s unexpectedly rejected: %s", status, reason)
		}
	}
}

func TestFailedPrometheusCannotBeQueuedForGrafana(t *testing.T) {
	prometheus := &model.ClusterExtension{State: model.ClusterExtensionStateEnabling, Status: model.ClusterExtensionStatusFailed}
	if reason := prometheusCannotServeGrafana(prometheus); !strings.Contains(reason, "failed") {
		t.Fatalf("failed dependency reason = %q", reason)
	}
}

func TestProvisionedPrometheusNeedsEndpointForGrafana(t *testing.T) {
	prometheus := &model.ClusterExtension{State: model.ClusterExtensionStateEnabled, Status: model.ClusterExtensionStatusProvisioned}
	if reason := prometheusCannotServeGrafana(prometheus); !strings.Contains(reason, "endpoint") {
		t.Fatalf("missing endpoint reason = %q", reason)
	}
	prometheus.Endpoints = []byte(`["http://prometheus.example:9090"]`)
	if reason := prometheusCannotServeGrafana(prometheus); reason != "" {
		t.Fatalf("ready dependency rejected: %s", reason)
	}
}

func TestClusterExtensionExposesStoredErrorReason(t *testing.T) {
	ext := &model.ClusterExtension{LastError: "waiting for Prometheus to finish installing"}
	got := convertClusterExtensionFromSQLModel(ext)
	if got.ErrorReason == nil || *got.ErrorReason != ext.LastError {
		t.Fatalf("error reason = %v, want %q", got.ErrorReason, ext.LastError)
	}
	ext.LastError = ""
	if got = convertClusterExtensionFromSQLModel(ext); got.ErrorReason != nil {
		t.Fatalf("empty stored error should be omitted, got %v", *got.ErrorReason)
	}
}

func TestClusterExtensionConversionCopiesErrorReason(t *testing.T) {
	ext := &model.ClusterExtension{LastError: "Grafana failed"}
	got := convertClusterExtensionFromSQLModel(ext)
	ext.LastError = "Fluent Bit failed"
	if got.ErrorReason == nil || *got.ErrorReason != "Grafana failed" {
		t.Fatalf("extension error changed with source row: %v", got.ErrorReason)
	}
}
