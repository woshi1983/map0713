// Copyright (c) 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"gitlab.eng.vmware.com/cn-se-solution/mapd/pkg/util/httputil"
)

const DefaultSizeLimit = 1024 * 1024 * 64 // 64MiB

// Store is a file store for caching HTTP contents.
type Store struct {
	mux     sync.RWMutex
	lim     int
	rootdir string
	index   map[string]string      // map download link to sha256 digest
	node    map[string]*sync.Mutex // if file (which is named as its digest) has been cached
}

func (s *Store) getDigest(link string) (string, bool) {
	s.mux.RLock()
	defer s.mux.RUnlock()

	digest, ok := s.index[link]
	if ok {
		mux, exists := s.node[digest]
		if !exists {
			return digest, false
		}

		mux.Lock()
		defer mux.Unlock()

		fpath := filepath.Join(s.rootdir, digest)
		_, err := os.Stat(fpath)
		if err != nil {
			return digest, false
		}
	}
	return digest, ok
}

func (s *Store) getMutex(digest string) (mux *sync.Mutex, ok bool) {
	s.mux.Lock()
	defer s.mux.Unlock()

	mux = s.node[digest]
	if mux == nil {
		mux = &sync.Mutex{}
		s.node[digest] = mux
		return
	}
	return mux, true
}

func (s *Store) Get(ctx context.Context, link string, opts ...httputil.HTTPClientOption) (string, error) {
	if err := s.load(ctx, link, opts...); err != nil {
		return "", err
	}

	digest, _ := s.getDigest(link)
	return filepath.Join(s.rootdir, digest), nil
}

func (s *Store) load(ctx context.Context, link string, opts ...httputil.HTTPClientOption) error {
	if _, ok := s.getDigest(link); ok {
		return nil
	}

	do, err := httputil.NewClient(opts...).New(ctx, link)
	if err != nil {
		return err
	}

	buf := newSizeLimitedBuffer(s.lim)
	h := sha256.New()

	if err := do(func(resp *http.Response) error {
		w := io.MultiWriter(buf, h)
		_, err := io.Copy(w, resp.Body)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	digest := hex.EncodeToString(h.Sum(nil))
	mux, _ := s.getMutex(digest)

	mux.Lock()
	defer mux.Unlock()

	fpath := filepath.Join(s.rootdir, digest)
	fi, err := os.Create(fpath)
	if err != nil {
		return err
	}
	defer fi.Close()

	if _, err := io.Copy(fi, buf); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	s.index[link] = digest
	return nil
}

func NewStore(rootdir string, sizeLimit int) (*Store, error) {
	if err := os.RemoveAll(rootdir); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(rootdir, 0755); err != nil {
		return nil, err
	}
	if sizeLimit < 1024 {
		sizeLimit = DefaultSizeLimit
	}
	return &Store{lim: DefaultSizeLimit, rootdir: rootdir, index: make(map[string]string), node: make(map[string]*sync.Mutex)}, nil
}

// sizeLimitedBuffer downloads content to a in-memory buffer.
type sizeLimitedBuffer struct {
	lim int
	siz int
	buf bytes.Buffer
}

func (b *sizeLimitedBuffer) Write(p []byte) (int, error) {
	n, err := b.buf.Write(p)
	if b.siz+n > b.lim {
		return n, fmt.Errorf("io: buffer overflow")
	}
	b.siz = b.siz + n
	return n, err
}

func (b *sizeLimitedBuffer) Read(p []byte) (int, error) {
	return b.buf.Read(p)
}

func (b *sizeLimitedBuffer) Reset() {
	b.siz = 0
	b.buf.Reset()
}

func newSizeLimitedBuffer(size int) *sizeLimitedBuffer {
	return &sizeLimitedBuffer{lim: size}
}
