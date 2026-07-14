package kvstore_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/kvstore"
)

func TestMemoryKVStore(t *testing.T) {
	store := kvstore.NewMemoryKVStore()

	store.Set("1", []byte(`1`))
	assert.Equal(t, []byte(`1`), store.Get("1"))

	store.SetTTL("2", []byte(`2`), time.Millisecond)
	assert.Equal(t, []byte(`2`), store.Get("2"))
	time.Sleep(2 * time.Millisecond)
	assert.Nil(t, store.Get("2"))
}
