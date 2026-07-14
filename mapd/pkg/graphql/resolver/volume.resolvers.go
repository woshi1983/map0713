package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
)

// Ref is the resolver for the ref field.
func (r *configMapVolumeResolver) Ref(ctx context.Context, obj *model.ConfigMapVolume) (*model.ConfigMap, error) {
	if obj.Ref == nil || obj.Ref.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).ConfigMap(WithGuardSilenced(ctx), obj.Ref.Cluster.ID, obj.Ref.Namespace, obj.Ref.Name)
}

// Ref is the resolver for the ref field.
func (r *persistentVolumeClaimVolumeResolver) Ref(ctx context.Context, obj *model.PersistentVolumeClaimVolume) (*model.PersistentVolumeClaim, error) {
	if obj.Ref == nil || obj.Ref.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).PersistentVolumeClaim(WithGuardSilenced(ctx), obj.Ref.Cluster.ID, obj.Ref.Namespace, obj.Ref.Name)
}

// Ref is the resolver for the ref field.
func (r *secretVolumeResolver) Ref(ctx context.Context, obj *model.SecretVolume) (*model.Secret, error) {
	if obj.Ref == nil || obj.Ref.Cluster == nil {
		return nil, nil
	}
	return r.Resolver.Query().(*queryResolver).Secret(WithGuardSilenced(ctx), obj.Ref.Cluster.ID, obj.Ref.Namespace, obj.Ref.Name)
}

// ConfigMapVolume returns graphql1.ConfigMapVolumeResolver implementation.
func (r *Resolver) ConfigMapVolume() graphql1.ConfigMapVolumeResolver {
	return &configMapVolumeResolver{r}
}

// PersistentVolumeClaimVolume returns graphql1.PersistentVolumeClaimVolumeResolver implementation.
func (r *Resolver) PersistentVolumeClaimVolume() graphql1.PersistentVolumeClaimVolumeResolver {
	return &persistentVolumeClaimVolumeResolver{r}
}

// SecretVolume returns graphql1.SecretVolumeResolver implementation.
func (r *Resolver) SecretVolume() graphql1.SecretVolumeResolver { return &secretVolumeResolver{r} }

type configMapVolumeResolver struct{ *Resolver }
type persistentVolumeClaimVolumeResolver struct{ *Resolver }
type secretVolumeResolver struct{ *Resolver }
