// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package conductor

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type Conductor struct {
	mux sync.Mutex
	ctx context.Context
	clz context.CancelFunc
	log *zap.SugaredLogger
}

func (c *Conductor) Run() error {
	var (
		wg          sync.WaitGroup
		dispatchers []Dispatcher
	)
	for name, register := range Registry {
		log := c.log.Named(name)
		dispatcher := register(c.ctx, log)
		if err := dispatcher.Init(); err != nil {
			return err
		}
		dispatchers = append(dispatchers, dispatcher)
	}

	wg.Add(len(dispatchers))
	for _, each := range dispatchers {
		run := each.Run
		go func() {
			defer wg.Done()

			run()
		}()
	}
	wg.Wait()
	return nil
}

func (c *Conductor) Stop() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.clz()
	return nil
}

func NewConductor(logger *zap.Logger) *Conductor {
	ctx, cancel := context.WithCancel(context.Background())
	return &Conductor{
		ctx: ctx,
		clz: cancel,
		log: logger.Named("conductor").Sugar(),
	}
}
