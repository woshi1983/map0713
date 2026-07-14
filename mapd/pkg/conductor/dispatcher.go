// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conductor

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"go.uber.org/zap"
)

type Dispatcher interface {
	// Init performs initialization tasks.
	Init() error
	// Run receives incoming events and dispatch them to specific handler.
	Run()
}

type dispatcher struct {
	wg           sync.WaitGroup
	mux          sync.Mutex
	ctx          context.Context
	log          *zap.SugaredLogger
	cancel       map[uuid.UUID]context.CancelFunc
	emitter      *mapd.EventEmitter
	subscription *mapd.EventSubscription
	initFn       DispatcherInitFunc
	syncFn       DispatcherSyncerFunc
}

func (d *dispatcher) Init() error {
	d.subscription = d.emitter.Subscribe()
	if d.initFn != nil {
		return d.initFn()
	}
	return nil
}

func (d *dispatcher) Run() {
LOOP:
	for {
		select {
		case ev := <-d.subscription.Chan:
			switch ev.Type {
			case mapd.EventTypeAdd, mapd.EventTypeModify:
				d.wg.Add(1)
				go d.onSet(ev.OID, d.log.Named(ev.OID.String()))
			case mapd.EventTypeRemove:
				d.wg.Add(1)
				go d.onRemove(ev.OID)
			}
		case <-d.ctx.Done():
			break LOOP
		}
	}
	d.subscription.Close()
	d.wg.Wait()
}

func (d *dispatcher) onSet(id uuid.UUID, log *zap.SugaredLogger) {
	d.mux.Lock()
	defer d.mux.Unlock()

	defer d.wg.Done()

	ctx, cancel := context.WithCancel(d.ctx)
	if abort, ok := d.cancel[id]; ok {
		abort()
	}
	d.cancel[id] = cancel
	d.wg.Add(1)
	go d.interate(ctx, id, log)
}

func (d *dispatcher) onRemove(id uuid.UUID) {
	d.mux.Lock()
	defer d.mux.Unlock()

	defer d.wg.Done()

	if abort, ok := d.cancel[id]; ok {
		abort()
	}
	delete(d.cancel, id)
}

func (d *dispatcher) interate(ctx context.Context, id uuid.UUID, log *zap.SugaredLogger) {
	defer d.wg.Done()

	ticker := time.NewTimer(0)
LOOP:
	for {
		select {
		case <-ticker.C:
			next := func() time.Duration {
				defer func() {
					defer log.Sync()

					if r := recover(); r != nil {
						errmsg := fmt.Sprintf("handler has just recovered from a panic: %v. Please contact customer support to report a bug. Stacktrace:\n---\n%s\n", r, debug.Stack())
						log.Errorf(strings.ReplaceAll(errmsg, "\n", "\n> "))
					}
				}()

				return d.syncFn(id, log)
			}()
			if next < 0 {
				next = mapd.GlobalConfig.ConductorSyncerInterval
			}
			ticker.Reset(next)
		case <-ctx.Done():
			break LOOP
		}
	}
	ticker.Stop()
}

func NewDispatcher(ctx context.Context, emitter *mapd.EventEmitter, init DispatcherInitFunc, sink DispatcherSyncerFunc, log *zap.SugaredLogger) *dispatcher {
	return &dispatcher{
		ctx:     ctx,
		log:     log,
		cancel:  make(map[uuid.UUID]context.CancelFunc),
		emitter: emitter,
		initFn:  init,
		syncFn:  sink,
	}
}

func DispatcherFactory(emitter *mapd.EventEmitter, init DispatcherInitFunc, sink DispatcherSyncerFunc) DispatcherRegister {
	return func(ctx context.Context, log *zap.SugaredLogger) Dispatcher {
		return NewDispatcher(ctx, emitter, init, sink, log)
	}
}

type DispatcherRegister func(ctx context.Context, log *zap.SugaredLogger) Dispatcher

type DispatcherRegistry map[string]DispatcherRegister

func (r DispatcherRegistry) Register(name string, register DispatcherRegister) {
	r[name] = register
}

var Registry = make(DispatcherRegistry)

type DispatcherInitFunc func() error

type DispatcherSyncerFunc func(id uuid.UUID, log *zap.SugaredLogger) (next time.Duration)
