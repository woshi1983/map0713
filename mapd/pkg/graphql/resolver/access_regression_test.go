package resolver

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestProviderAndClusterMutationsRejectEmptyID(t *testing.T) {
	ctx := auth.Context(context.Background(), &auth.SessionInfo{Admin: true})
	r := &mutationResolver{}
	tests := []struct {
		name string
		run  func() error
	}{
		{"UpdateTKGS", func() error {
			_, err := r.UpdateTKGSClusterProvider(ctx, model.UpdateTKGSClusterProvider{ID: uuid.Nil})
			return err
		}},
		{"UpdateTKG2S", func() error {
			_, err := r.UpdateTKG2SClusterProvider(ctx, model.UpdateTKG2SClusterProvider{ID: uuid.Nil})
			return err
		}},
		{"UpdateVKS", func() error {
			_, err := r.UpdateVKSClusterProvider(ctx, model.UpdateVKSClusterProvider{ID: uuid.Nil})
			return err
		}},
		{"UpdateTKG2M", func() error {
			_, err := r.UpdateTKG2MClusterProvider(ctx, model.UpdateTKG2MClusterProvider{ID: uuid.Nil})
			return err
		}},
		{"DeleteProvider", func() error {
			_, err := r.DeleteClusterProvider(ctx, model.DeleteClusterProvider{ID: uuid.Nil})
			return err
		}},
		{"UpdateCluster", func() error { _, err := r.UpdateCluster(ctx, model.UpdateCluster{ID: uuid.Nil}); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			if err == nil || !IsGQLError(err) {
				t.Fatalf("expected GraphQL validation error, got %v", err)
			}
		})
	}
}

func TestUsersFilterUsesRoleBindings(t *testing.T) {
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       "test:test@tcp(localhost:3306)/mapd",
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatal(err)
	}
	org := uuid.New()
	sess := &auth.SessionInfo{Perm: map[uuid.UUID]*auth.SessionPerm{
		org: {R: model1.PolicyObjectUser},
	}}
	query := filterVisibleUsers(db.Model(&model1.UserBinding{}), sess).Find(&[]model1.UserBinding{})
	if query.Error != nil {
		t.Fatal(query.Error)
	}
	sql := query.Statement.SQL.String()
	if !strings.Contains(sql, "user_binding_id") || !strings.Contains(sql, "organization_id") || !strings.Contains(sql, " IN ") {
		t.Fatalf("user list is not scoped through role bindings: %s", sql)
	}
	if len(query.Statement.Vars) != 1 || query.Statement.Vars[0] != org {
		t.Fatalf("expected only allowed organization ID, got %v", query.Statement.Vars)
	}
}

func TestCreateOrganizationRequiresWritePermission(t *testing.T) {
	ctx := auth.Context(context.Background(), &auth.SessionInfo{
		UserID: uuid.New(),
		Perm:   map[uuid.UUID]*auth.SessionPerm{},
	})
	r := &mutationResolver{}
	_, err := r.CreateOrganization(ctx, model.CreateOrganization{Name: "forbidden"})
	if err == nil || !IsGQLError(err) || !strings.Contains(err.Error(), "insufficient privileges") {
		t.Fatalf("expected insufficient privileges, got %v", err)
	}
}

func TestUserRejectsEmptyID(t *testing.T) {
	ctx := auth.Context(context.Background(), &auth.SessionInfo{Admin: true})
	r := &queryResolver{}
	_, err := r.User(ctx, uuid.Nil)
	if err == nil || !IsGQLError(err) {
		t.Fatalf("expected GraphQL validation error, got %v", err)
	}
}

func TestObjectQueriesRejectEmptyID(t *testing.T) {
	ctx := auth.Context(context.Background(), &auth.SessionInfo{Admin: true})
	r := &queryResolver{}
	tests := []struct {
		name string
		run  func() error
	}{
		{"Cluster", func() error { _, err := r.Cluster(ctx, uuid.Nil); return err }},
		{"Component", func() error { _, err := r.Component(ctx, uuid.Nil); return err }},
		{"Route", func() error { _, err := r.Route(ctx, uuid.Nil); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			if err == nil || !IsGQLError(err) || !strings.Contains(err.Error(), "invalid request") {
				t.Fatalf("expected invalid request, got %v", err)
			}
		})
	}
}

func TestRemainingMutationsRejectEmptyID(t *testing.T) {
	ctx := auth.Context(context.Background(), &auth.SessionInfo{Admin: true})
	r := &mutationResolver{}
	tests := []struct {
		name string
		run  func() error
	}{
		{"UpdateUser", func() error { _, err := r.UpdateUser(ctx, model.UpdateUser{ID: uuid.Nil}); return err }},
		{"DeleteCluster", func() error { _, err := r.DeleteCluster(ctx, model.DeleteCluster{ID: uuid.Nil}); return err }},
		{"UpdateRoute", func() error { _, err := r.UpdateRoute(ctx, model.UpdateRoute{ID: uuid.Nil}); return err }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.run()
			if err == nil || !IsGQLError(err) || !strings.Contains(err.Error(), "invalid request") {
				t.Fatalf("expected invalid request, got %v", err)
			}
		})
	}
}
