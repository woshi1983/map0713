package resolver

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"time"

	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/auth"
	graphql1 "gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/abstract"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/graphql/model"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/logging"
)

// Ping is the resolver for the Ping field.
func (r *mutationResolver) Ping(ctx context.Context) (*model.Pong, error) {
	return &model.Pong{Pong: time.Now()}, nil
}

// Version is the resolver for the Version field.
func (r *queryResolver) Version(ctx context.Context) (*model.SoftwareVersion, error) {
	return &model.SoftwareVersion{Version: abstract.MustParseVersion(mapd.Version), Revision: mapd.Revision}, nil
}

// Ping is the resolver for the Ping field.
func (r *queryResolver) Ping(ctx context.Context) (*model.Pong, error) {
	return &model.Pong{Pong: time.Now()}, nil
}

// Ping is the resolver for the Ping field.
func (r *subscriptionResolver) Ping(ctx context.Context) (<-chan *model.Pong, error) {
	msgChan := make(chan *model.Pong, 1)
	timer := time.NewTimer(0)

	go func() {
		defer timer.Stop()

	LOOP:
		for {
			select {
			case <-timer.C:
				msgChan <- &model.Pong{Pong: time.Now()}
				timer.Reset(time.Second)
			case <-ctx.Done():
				close(msgChan)
				break LOOP
			}
		}
	}()
	return msgChan, nil
}

// ConsoleLog is the resolver for the ConsoleLog field.
func (r *subscriptionResolver) ConsoleLog(ctx context.Context) (<-chan string, error) {
	sess, ok := auth.FromContext(ctx)
	if !ok {
		panic(ErrIdentityNotFound)
	}

	if !IsGuardSilenced(ctx) && !sess.Admin {
		return nil, errInsufficientPrivileges(ctx)
	}

	msgChan := make(chan string, 10)
	ch := logging.LoggingDispatcher.Subscribe(ctx)

	go func() {
	LOOP:
		for {
			select {
			case p := <-ch:
				msgChan <- string(p)
			case <-ctx.Done():
				close(msgChan)
				break LOOP
			}
		}
	}()
	return msgChan, nil
}

// Mutation returns graphql1.MutationResolver implementation.
func (r *Resolver) Mutation() graphql1.MutationResolver { return &mutationResolver{r} }

// Query returns graphql1.QueryResolver implementation.
func (r *Resolver) Query() graphql1.QueryResolver { return &queryResolver{r} }

// Subscription returns graphql1.SubscriptionResolver implementation.
func (r *Resolver) Subscription() graphql1.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
