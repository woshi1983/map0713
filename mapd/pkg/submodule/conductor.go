// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package submodule

import (
	"context"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/conductor"
	"go.uber.org/zap"
)

type Conductor struct {
	svc *conductor.Conductor
	log *zap.SugaredLogger
}

func (m *Conductor) Run() <-chan error {
	defer m.log.Sync()

	errCh := make(chan error, 1)
	go func() {
		defer m.log.Sync()
		err := m.svc.Run()
		errCh <- err
		m.log.Info("conductor is stopped.")
	}()
	m.log.Info("conductor started to run.")

	clozCh := make(chan error, 1)
	go func() {
		err := <-errCh
		if err != nil {
			m.log.Errorf("shutting down immediately due to a critical error: %v.", err)
			m.svc.Stop()
		}
		clozCh <- err
	}()
	return clozCh
}

func (m *Conductor) Close(ctx context.Context) error {
	return m.svc.Stop()
}

func NewConductor(logger *zap.Logger) *Conductor {
	return &Conductor{svc: conductor.NewConductor(logger), log: logger.Named("conductor").Sugar()}
}
