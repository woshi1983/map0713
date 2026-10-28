package conductor

import (
	"testing"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
)

func TestSelectedExtensionStorageClass(t *testing.T) {
	tests := []struct {
		name    string
		cluster model.Cluster
		want    string
		wantErr bool
	}{
		{name: "VKS selected class", cluster: model.Cluster{Class: model.ClusterClassVKS, Spec: []byte(`{"storageClass":"cluster-optimal-datastore-default-policy-raid1"}`)}, want: "cluster-optimal-datastore-default-policy-raid1"},
		{name: "other provider", cluster: model.Cluster{Class: model.ClusterClassTKGM, Spec: []byte(`{}`)}},
		{name: "invalid VKS spec", cluster: model.Cluster{Class: model.ClusterClassVKS, Spec: []byte(`{`)}, wantErr: true},
		{name: "missing VKS class", cluster: model.Cluster{Class: model.ClusterClassVKS, Spec: []byte(`{}`)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := selectedExtensionStorageClass(&tt.cluster)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("selectedExtensionStorageClass() = %q, %v; want %q, error=%v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}
