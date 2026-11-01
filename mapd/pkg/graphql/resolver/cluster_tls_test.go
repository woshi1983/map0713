package resolver

import (
	"context"
	"net/url"
	"testing"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
)

func TestUnhealthyClusterDoesNotProbeTLSCertificate(t *testing.T) {
	endpoint, err := url.Parse("https://127.0.0.1:1")
	if err != nil {
		t.Fatal(err)
	}
	cluster := &model.Cluster{Endpoint: endpoint, Status: model.ClusterStatusUnhealthy}
	certs, err := (&clusterResolver{}).TLSCertChain(context.Background(), cluster)
	if err != nil || certs != nil {
		t.Fatalf("unhealthy cluster TLS preview: certs=%v err=%v", certs, err)
	}
}
