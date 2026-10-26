package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	model1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/model"
	tanzuv2 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/tanzu/v2"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/vks"
	"gorm.io/gorm"
)

func protectedVKSNamespace(name string) bool {
	return name == "" || tkg2sProviderProtectedNamespaces[name] ||
		strings.HasPrefix(name, "svc-") || strings.HasPrefix(name, "vmware-") ||
		strings.HasPrefix(name, "kube-") || strings.HasPrefix(name, "wcp-")
}

func vksProviderSpec(ctx context.Context, id uuid.UUID) (*model1.ClusterProviderVKSSpec, error) {
	provider := &model1.ClusterProvider{ID: id}
	if err := mapd.Database().Take(provider).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errObjectNotFound(ctx, "provider", id.String())
		}
		return nil, &ErrDBConnection{err}
	}
	if provider.Class != model1.ClusterClassVKS {
		return nil, errInvalidRequest(ctx, "provider", "must be a VKS provider")
	}
	spec := &model1.ClusterProviderVKSSpec{}
	if err := json.Unmarshal(provider.Spec, spec); err != nil {
		return nil, err
	}
	return spec, nil
}

func vksProviderClient(ctx context.Context, id uuid.UUID) (*vks.Client, *tanzuv2.VSphereWithTanzu, error) {
	spec, err := vksProviderSpec(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	client, supervisor, err := vks.NewClientFromProvider(id, spec)
	if err != nil {
		return nil, nil, errGatewayError(ctx, err)
	}
	return client, supervisor, nil
}
