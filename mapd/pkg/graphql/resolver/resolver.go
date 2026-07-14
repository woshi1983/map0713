package resolver

import (
	"fmt"
	"io"
	"sync"

	"github.com/dexidp/dex/api/v2"
	"github.com/google/uuid"
	"gitlab.eng.vmware.com/cn-se-solution/mapd"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil/cache"
)

// Resolver is the root resolver definition of GraphQL.
//
//go:generate go run github.com/99designs/gqlgen generate
type Resolver struct {
	DexClient           api.DexClient
	ChartCache          *cache.Store
	PodExecSubscription *PodExecSubscriptionManager
}

func NewResolver(dex api.DexClient) (*Resolver, error) {
	c, err := cache.NewStore(mapd.GlobalConfig.CacheDir, 0)
	if err != nil {
		return nil, fmt.Errorf("error initializing cache dir: %v", err)
	}
	return &Resolver{
		DexClient:           dex,
		ChartCache:          c,
		PodExecSubscription: NewPodExecSubscriptionManager(),
	}, nil
}

type PodExecSubscriptionManager struct {
	mux sync.RWMutex
	sub map[uuid.UUID]*PodExecSubscriptionSession
}

func (m *PodExecSubscriptionManager) Get(id uuid.UUID) *PodExecSubscriptionSession {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return m.sub[id]
}

func (m *PodExecSubscriptionManager) Delete(id uuid.UUID) {
	m.mux.Lock()
	defer m.mux.Unlock()

	_, ok := m.sub[id]
	if ok {
		delete(m.sub, id)
	}
}

func (m *PodExecSubscriptionManager) New(stdin io.Writer, out io.Reader) *PodExecSubscriptionSession {
	m.mux.Lock()
	defer m.mux.Unlock()

	session := &PodExecSubscriptionSession{
		ID:  uuid.New(),
		Out: out,
		In:  stdin,
	}
	m.sub[session.ID] = session
	return session
}

func NewPodExecSubscriptionManager() *PodExecSubscriptionManager {
	return &PodExecSubscriptionManager{
		sub: make(map[uuid.UUID]*PodExecSubscriptionSession),
	}
}

type PodExecSubscriptionSession struct {
	ID  uuid.UUID
	Out io.Reader
	In  io.Writer
}
