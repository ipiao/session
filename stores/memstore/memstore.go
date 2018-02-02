// Package memstore is a in-memory session store for the SCS session package.
//
// Warning: Because memstore uses in-memory storage only, all session data will
// be lost when your Go program is stopped or restarted. On the upside though,
// it is blazingly fast.
//
// In production, memstore should only be used where this volatility is an acceptable
// trade-off for the high performance, and where lost session data will have a
// negligible impact on users.
//
// The memstore package provides a background 'cleanup' goroutine to delete
// expired session data. This stops the underlying cache from holding on to invalid
// sessions forever and taking up unnecessary memory.
package memstore

import (
	"errors"
	"os"
	"time"

	"github.com/patrickmn/go-cache"
)

var errTypeAssertionFailed = errors.New("type assertion failed: could not convert interface{} to []byte")

// MemStore represents the currently configured session session store. It is essentially
// a wrapper around a go-cache instance (see https://github.com/patrickmn/go-cache).
type MemStore struct {
	cache    *cache.Cache
	dumpfile string
}

// New returns a new MemStore instance.
//
// The cleanupInterval parameter controls how frequently expired session data
// is removed by the background 'cleanup' goroutine. Setting it to 0 prevents
// the cleanup goroutine from running (i.e. expired sessions will not be removed).
func New(cleanupInterval time.Duration) *MemStore {
	return &MemStore{
		cache: cache.New(cache.DefaultExpiration, cleanupInterval),
	}
}

// SetDumpFile 设置落地文件
func (m *MemStore) SetDumpFile(f string) {
	m.dumpfile = f
}

// Find returns the data for a given session token from the MemStore instance. If the session
// token is not found or is expired, the returned exists flag will be set to false.
func (m *MemStore) Find(token string) (b []byte, exists bool, err error) {
	v, exists := m.cache.Get(token)
	if exists == false {
		return nil, exists, nil
	}

	b, ok := v.([]byte)
	if ok == false {
		return nil, exists, errTypeAssertionFailed
	}

	return b, exists, nil
}

// Save adds a session token and data to the MemStore instance with the given expiry time.
// If the session token already exists then the data and expiry time are updated.
func (m *MemStore) Save(token string, b []byte, expiry time.Time) error {
	m.cache.Set(token, b, expiry.Sub(time.Now()))
	return nil
}

// Delete removes a session token and corresponding data from the MemStore instance.
func (m *MemStore) Delete(token string) error {
	m.cache.Delete(token)
	return nil
}

// FindAll 查找所有
func (m *MemStore) FindAll() (bs [][]byte, err error) {
	items := m.cache.Items()
	for _, v := range items {
		b, ok := v.Object.([]byte)
		if ok == false {
			err = errTypeAssertionFailed
			continue
		}
		bs = append(bs, b)
	}
	return
}

// Loads 加载
func (m *MemStore) Loads() (bs [][]byte, err error) {
	e := m.cache.LoadFile(m.dumpfile)
	if e != nil {
		if _, ok := e.(*os.PathError); ok {
			return bs, nil
		}
		return
	}
	bs, err = m.FindAll()
	return
}

// Dumps 数据存储
func (m *MemStore) Dumps() (err error) {
	if m.dumpfile == "" {
		return nil
	}
	return m.cache.SaveFile(m.dumpfile)
}
