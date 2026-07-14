package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
)

// Metrics is the resolver for the Metrics field.
func (r *queryResolver) Metrics(ctx context.Context) (*model.Metrics, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !IsGuardSilenced(ctx) && !sess.Admin {
		return nil, errInsufficientPrivileges(ctx)
	}

	return &model.Metrics{
		ActiveUsers:        int64(auth.RuntimeSessionCache.ActiveUsers.Len()),
		LocalUsersDisabled: mapd.GlobalConfig.Auth.DisableLocalUser,
	}, nil
}
