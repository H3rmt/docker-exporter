package docker

import (
	"context"
	"sync"
	"time"

	"github.com/h3rmt/docker-exporter/internal/log"
)

type Cache[T any] struct {
	name            string
	mu              sync.Mutex
	data            T
	lastUpdated     time.Time
	refreshing      bool
	refreshCh       chan struct{}
	refreshInterval time.Duration
	clone           func(T) T
	load            func(ctx context.Context) (T, error)
}

func NewCache[T any](name string, refreshInterval time.Duration, load func(ctx context.Context) (T, error)) Cache[T] {
	return NewCacheFull(name, refreshInterval, load, func(t T) T { return t })
}

func NewCacheFull[T any](name string, refreshInterval time.Duration, load func(ctx context.Context) (T, error), clone func(T) T) Cache[T] {
	return Cache[T]{name: name, clone: clone, load: load, refreshInterval: refreshInterval}
}

func copyMap[K comparable, T any](src map[K]T) map[K]T {
	if src == nil {
		return map[K]T{}
	}
	dst := make(map[K]T, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (c *Cache[T]) GetValues(ctx context.Context) T {
	c.mu.Lock()
	cacheExists := !c.lastUpdated.IsZero()
	stale := !c.lastUpdated.IsZero() && time.Since(c.lastUpdated) >= c.refreshInterval

	if cacheExists && !stale {
		// Cache exists and is fresh enough
		cached := c.clone(c.data)
		c.mu.Unlock()
		return cached
	}

	// currently refreshing
	if c.refreshing {
		// A refresh is already running, but we have data in the cache
		if cacheExists {
			// Serve the existing cache without starting another
			cached := c.clone(c.data)
			c.mu.Unlock()
			return cached
		}

		// Cache missing: wait for the in-flight refresh to complete
		ch := c.refreshCh
		c.mu.Unlock()
		select {
		case <-ch:
		// return early
		case <-ctx.Done():
			var nilT T
			return nilT
		}
		c.mu.Lock()
		// After completion, return whatever cache we have (maybe empty if loading failed)
		cached := c.clone(c.data)
		c.mu.Unlock()
		return cached
	}

	log.GetLogger().DebugContext(ctx, "Refreshing cache", "name", c.name, "stale", stale, "cacheExists", cacheExists)

	// Need to start a refresh
	ch := make(chan struct{})
	c.refreshCh = ch
	c.refreshing = true
	// Start background refresh with background context to ensure progress independent of the request context
	ctx2 := context.Background()
	go c.loadData(ctx2)
	if !cacheExists {
		// Block until the initial cache is ready
		c.mu.Unlock()
		<-ch
		c.mu.Lock()
		cached := c.clone(c.data)
		c.mu.Unlock()
		return cached
	}
	// If cache exists (but stale), return it immediately
	cached := c.clone(c.data)
	c.mu.Unlock()
	return cached
}

func (c *Cache[T]) loadData(ctx context.Context) {
	// Perform the expensive call
	data, err := c.load(ctx)

	// Update cache state
	c.mu.Lock()
	if err == nil {
		c.data = data
		c.lastUpdated = time.Now()
	}
	if c.refreshCh != nil {
		close(c.refreshCh)
	}
	c.refreshing = false
	c.mu.Unlock()
}
