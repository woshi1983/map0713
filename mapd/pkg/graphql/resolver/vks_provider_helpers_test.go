package resolver

import "testing"

func TestProtectedVKSNamespace(t *testing.T) {
	tests := []struct {
		name      string
		protected bool
	}{
		{name: "", protected: true},
		{name: "default", protected: true},
		{name: "svc-cci-service-domain-c9", protected: true},
		{name: "vmware-system-vks-public", protected: true},
		{name: "kube-system", protected: true},
		{name: "wcp-system", protected: true},
		{name: "vlan803", protected: false},
		{name: "vlan805", protected: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := protectedVKSNamespace(tt.name); got != tt.protected {
				t.Fatalf("protectedVKSNamespace(%q) = %v, want %v", tt.name, got, tt.protected)
			}
		})
	}
}
