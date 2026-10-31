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
