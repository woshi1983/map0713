package resolver

import (
	"testing"

	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

func TestClusterConversionCopiesErrorReason(t *testing.T) {
	cluster := &model1.Cluster{Version: "1.24.0", LastError: "first cluster failed"}
	got := convertClusterFromSQLModel(cluster)
	cluster.LastError = "second cluster failed"
	if got.ErrorReason == nil || *got.ErrorReason != "first cluster failed" {
		t.Fatalf("cluster error changed with source row: %v", got.ErrorReason)
	}
}
