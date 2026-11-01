package resolver

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
)

func TestRemainingObjectQueriesRejectEmptyID(t *testing.T) {
	ctx := auth.Context(context.Background(), &auth.SessionInfo{Admin: true})
	r := &queryResolver{}
	tests := []struct {
		name string
		run  func() error
	}{
		{"ClusterExtension", func() error { _, err := r.ClusterExtension(ctx, uuid.Nil); return err }},
		{"ImageRegistry", func() error { _, err := r.ImageRegistry(ctx, uuid.Nil); return err }},
		{"LoadBalancer", func() error { _, err := r.LoadBalancer(ctx, uuid.Nil); return err }},
		{"Release", func() error { _, err := r.Release(ctx, uuid.Nil); return err }},
		{"RoleBinding", func() error { _, err := r.RoleBinding(ctx, uuid.Nil); return err }},
		{"StandardPackageRepository", func() error { _, err := r.StandardPackageRepository(ctx, uuid.Nil); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			if err == nil || !IsGQLError(err) || !strings.Contains(err.Error(), "invalid request") {
				t.Fatalf("expected invalid request for empty ID, got %v", err)
			}
		})
	}
}
