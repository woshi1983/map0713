// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package mapd

import (
	"sync"

	"github.com/google/uuid"
)

// EventType indicates the type of Event.
type EventType int8

const (
	EventTypeUnknown EventType = iota
	EventTypeAdd
	EventTypeModify
	EventTypeRemove
)

// Event is used for notifying conductor to perform specific tasks.
type Event struct {
	Type EventType
	OID  uuid.UUID
}

// EventEmitter is used for triggering an event. It forwards received
// events to all subscriptions.
type EventEmitter struct {
	mux sync.RWMutex
	chn chan *Event
	sub map[uuid.UUID]chan *Event
	clz chan struct{}
}

func (ee *EventEmitter) broadcast(ev *Event) {
	if ev == nil {
		return
	}

	ee.mux.RLock()
	defer ee.mux.RUnlock()

	for id := range ee.sub {
		ee.sub[id] <- ev
	}
}

func (ee *EventEmitter) forward() {
LOOP:
	for {
		select {
		case ev := <-ee.chn:
			ee.broadcast(ev)
		case <-ee.clz:
			break LOOP
		}
	}
	close(ee.chn)
}

// Next triggers a new event.
func (ee *EventEmitter) Next(ev *Event) {
	if ev == nil {
		return
	}

	ee.chn <- ev
}

// Subscribe creates a new EventSubscription. If the subscription is no longer
// needed, please call *EventSubscription.Close() for garbage collection.
func (ee *EventEmitter) Subscribe() *EventSubscription {
	id := uuid.New()
	ch := make(chan *Event, 512)

	ee.mux.Lock()
	defer ee.mux.Unlock()

	ee.sub[id] = ch
	return &EventSubscription{
		ID:   id,
		Chan: ch,
		Close: func() {
			ee.mux.Lock()
			defer ee.mux.Unlock()

			close(ch)
			delete(ee.sub, id)
		},
	}
}

// Close stops forwarding income events.
func (ee *EventEmitter) Close() {
	close(ee.clz)
}

// NewEventEmitter creates a new EventEmitter.
func NewEventEmitter() *EventEmitter {
	ee := &EventEmitter{
		chn: make(chan *Event, 512),
		sub: make(map[uuid.UUID]chan *Event),
		clz: make(chan struct{}),
	}
	go ee.forward()
	return ee
}

// EventSubscription is a subscription of EventEmitter.
type EventSubscription struct {
	ID    uuid.UUID
	Chan  <-chan *Event
	Close func()
}

var (
	EventEmitterClusterBackupSchedule = NewEventEmitter()
	EventEmitterClusterExtension      = NewEventEmitter()
	EventEmitterClusterRestore        = NewEventEmitter()
	EventEmitterCluster               = NewEventEmitter()
	EventEmitterInstanceFederation    = NewEventEmitter()
	EventEmitterInstance              = NewEventEmitter()
	EventEmitterRoute                 = NewEventEmitter()
)
