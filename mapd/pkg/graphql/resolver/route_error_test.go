package resolver

import (
	"testing"

	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

func TestRouteConversionIncludesProvisionFailure(t *testing.T) {
	route := &model1.Route{LastError: "AVI could not allocate an IP address"}
	got := convertRouteFromSQLModel(route)
	if got.ErrorReason == nil || *got.ErrorReason != route.LastError {
		t.Fatalf("expected route error %q, got %v", route.LastError, got.ErrorReason)
	}
}

func TestRouteConversionCopiesProvisionFailure(t *testing.T) {
	route := &model1.Route{LastError: "first route failed"}
	got := convertRouteFromSQLModel(route)
	route.LastError = "second route failed"
	if got.ErrorReason == nil || *got.ErrorReason != "first route failed" {
		t.Fatalf("route error changed with source row: %v", got.ErrorReason)
	}
}
