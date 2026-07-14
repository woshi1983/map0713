// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package mapd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/liberr"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

// Module is a global endpoint which determines how the program works.
// Its running state is fully controlled by system signal.
type Module struct {
	log       *zap.SugaredLogger
	preRunFn  func(context.Context) error
	runFn     func(context.Context) error
	postRunFn func(context.Context) error
}

// Run start workers of the module, and returns any unexpected error.
func (m *Module) Run() error {
	defer m.log.Sync()
	sigChan := make(chan os.Signal, 2)
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var (
		cancels []context.CancelFunc
		mux     sync.Mutex
	)
	ctx, cancel := context.WithCancel(context.Background())
	cancels = append(cancels, cancel)

	if err := m.preRunFn(ctx); err != nil {
		cancel()
		return err
	}
	m.log.Info("module is successfully loaded.")

	go func() {
		defer m.log.Sync()

		var i int
		for sig := range sigChan {
			if sig != nil {
				m.log.Infof("%s signal was received, notifying components to shutdown...", sig)
				mux.Lock()
				if i < len(cancels) {
					cancels[i]()
				}
				mux.Unlock()
			}
			i++
		}
	}()

	var err error
	if e := m.runFn(ctx); e != nil {
		err = multierr.Append(err, fmt.Errorf("component was terminated with an error: %v", e))
	}

	ctx, cancel = context.WithTimeout(context.Background(), GlobalConfig.DaemonCleanupTimeout)
	mux.Lock()
	cancels = append(cancels, cancel)
	mux.Unlock()

	if e := m.postRunFn(ctx); e != nil {
		err = multierr.Append(err, fmt.Errorf("failed to cleanup component resources: %v", e))
	}
	m.log.Info("module is terminated.")
	return err
}

// NewModule initializes a new Module with given submodule injected in
// given order.
func NewModule(logger *zap.Logger, submodule ...SubModule) *Module {
	// PreRun will be called in preset order
	preRunFn := func(ctx context.Context) error { return nil }
	for _, each := range submodule {
		if sm, ok := each.(PreRunnableSubModule); ok {
			oldFn := preRunFn
			preRunFn = func(ctx context.Context) error {
				err := oldFn(ctx)
				select {
				case <-ctx.Done():
				default:
					err = multierr.Append(err, sm.PreRun())
				}
				return err
			}
		}
	}
	// PostRun will be called in reversed preset order
	postRunFn := func(ctx context.Context) error { return nil }
	for i := 0; i < len(submodule); i++ {
		each := submodule[len(submodule)-i-1]
		if sm, ok := each.(PostRunnableSubModule); ok {
			oldFn := postRunFn
			postRunFn = func(ctx context.Context) error {
				err := oldFn(ctx)
				select {
				case <-ctx.Done():
				default:
					err = multierr.Append(err, sm.PostRun())
				}
				return err
			}
		}
	}
	l := logger.Sugar()
	return &Module{
		log:      logger.Sugar(),
		preRunFn: preRunFn,
		runFn: func(ctx context.Context) error {
			defer l.Sync()

			l.Infof("starting %d component(s)...", len(submodule))

			var (
				wg       sync.WaitGroup
				errChans []<-chan error
			)
			wg.Add(len(submodule))
			// Run will be called concurrently
			for _, each := range submodule {
				run := each.Run
				errChans = append(errChans, liberr.Take(1, func() <-chan error {
					errChan := make(chan error, 1)
					go func() {
						defer wg.Done()
						errChan <- <-run()
						close(errChan)
					}()
					return errChan
				}()))
			}
			errChan := liberr.Take(len(submodule), liberr.MergeAll(errChans...))

			var (
				err  error
				done int
			)

		LOOP:
			for {
				select {
				case e := <-errChan:
					// a critical error occurred
					err = multierr.Append(err, e)
					done++
				case <-ctx.Done():
				}
				break LOOP
			}

			l.Info("module is terminating due to component shutdown.")

			for i := 0; i < len(submodule); i++ {
				each := submodule[len(submodule)-i-1]
				ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.DaemonShutdownTimeout)
				each.Close(ctx)
				cancel()
			}

			wg.Wait()

			for done < len(submodule) {
				err = multierr.Append(err, <-errChan)
				done++
			}
			return err
		},
		postRunFn: postRunFn,
	}
}

// SubModule is a part of a Module, provides basic control interface.
type SubModule interface {
	// Run returns immediately but with an error channel.
	Run() <-chan error
	// Close notifies the submodule to enter cleanup process, wait until all
	// jobs are finished or terminated.
	Close(ctx context.Context) error
}

// PreRunnableSubModule is like SubModule but with PreRun method support.
type PreRunnableSubModule interface {
	SubModule

	// PreRun prepares for running (if available), returns any encountered error.
	PreRun() error
}

// PostRunnableSubModule is like SubModule but with PostRun method support.
type PostRunnableSubModule interface {
	SubModule

	// PostRun cleanup allocated resources (if available) and returns any
	// encountered error.
	PostRun() error
}
