package middleware

import (
	"net/http"
	"strings"

	"github.com/EverythingMe/vertex"
)

import (
	"errors"
	"sync"
	"time"

	"github.com/EverythingMe/groupcache/lru"
	"github.com/dvirsky/go-pylog/logging"
)

type entry struct {
	value  interface{}
	expiry time.Time
}

func newEntry(value interface{}, ttl time.Duration) *entry {
	return &entry{
		value,
		time.Now().Add(ttl),
	}
}

// NewCacheMiddleware creates a new Cache middleware
func NewCacheMiddleware(maxItems int, ttl time.Duration) *CacheMiddleware {
	return &CacheMiddleware{
		cache: lru.New(maxItems),
		mutex: &sync.RWMutex{},
		ttl:   ttl,
	}
}

// Get gets data saved for an URL if present in cache.
func (m *CacheMiddleware) get(key string) (*entry, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	data, ok := m.cache.Get(key)
	if !ok {
		return nil, errors.New("not in cache")
	}

	ret, ok := data.(*entry)
	if !ok {
		return nil, errors.New("Invalid entry")
	}

	// This entry is expired!
	if ret.expiry.Before(time.Now()) {
		return nil, errors.New("Expired")
	}

	return ret, nil
}

// Put puts data of an URL in cache.
func (m *CacheMiddleware) put(key string, ent *entry) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.cache.Add(key, ent)
}

// CacheMiddleware is a middleware that caches responses for requests based on their url, method and params.
//
// The cache uses an LRU cache with a given size, and tries to get/set resonses from and to it.
// The url of the request and an ancoded version of request.Form (GET + POST + path params) are used as the key.
// Headers do not play a part in the cache key.
//
// Note: If the request contains a "Cache-Control: no-cache" header, the middleware will be bypassed
type CacheMiddleware struct {
	cache *lru.Cache
	mutex *sync.RWMutex
	ttl   time.Duration
}

func (m *CacheMiddleware) requestKey(r *vertex.Request) string {

	return r.Method + "/" + r.Request.URL.Path + "::" + r.Form.Encode()

}

func (m *CacheMiddleware) Handle(w http.ResponseWriter, r *vertex.Request, next vertex.HandlerFunc) (interface{}, error) {

	// Do not act on request if they have no-cache header
	if strings.ToLower(r.Header.Get("Cache-Control")) == "no-cache" {
		return next(w, r)
	}

	key := m.requestKey(r)
	logging.Debug("CACHING KEY: %s", key)
	entry, err := m.get(key)
	if err == nil && entry != nil {
		logging.Info("Fetched cache response: %#v", entry.value)
		return entry.value, nil
	}

	v, err := next(w, r)
	if err == nil {
		m.put(key, newEntry(v, m.ttl))
	}

	return v, err

}
