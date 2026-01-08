package docker

import (
	"context"
	"time"

	"github.com/h3rmt/docker-exporter/internal/log"
	"github.com/moby/moby/client"
)

var sizeRefreshInterval = 5 * time.Minute

func SetSizeCacheSeconds(interval time.Duration) {
	sizeRefreshInterval = interval
	log.GetLogger().Debug("Size cache refresh interval set", "interval", interval)
}

// sizeEntry stores container disk size fields
type sizeEntry struct {
	SizeRootFs int64
	SizeRw     int64
}

// getCachedValues ensures a cache exists and refreshes it every 5 minutes.
// Behavior:
// - If cache does not exist: start a single refresh goroutine and block until it completes.
// - If cache exists but is stale (>=5m): start a single background refresh and return current cached values immediately.
// - If a refresh is already in progress: if cache exists, return it immediately; if it doesn't, wait for the refresh.
func (c *Client) getCachedValues(ctx context.Context) map[string]sizeEntry {
	c.sizeMu.Lock()
	cacheExists := c.sizeCache != nil
	stale := !c.sizeLastUpdated.IsZero() && time.Since(c.sizeLastUpdated) >= sizeRefreshInterval

	if cacheExists && !stale {
		// Cache exists and is fresh enough
		cached := copySizeCache(c.sizeCache)
		c.sizeMu.Unlock()
		return cached
	}

	// currently refreshing
	if c.sizeRefreshing {
		// A refresh is already running but we have data in cache
		if cacheExists {
			// Serve existing cache without starting another
			cached := copySizeCache(c.sizeCache)
			c.sizeMu.Unlock()
			return cached
		}

		// Cache missing: wait for the in-flight refresh to complete
		ch := c.sizeRefreshCh
		c.sizeMu.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			// Context canceled; return empty snapshot to avoid blocking caller indefinitely
			return map[string]sizeEntry{}
		}
		c.sizeMu.Lock()
		// After completion, return whatever cache we have (may be empty)
		cached := copySizeCache(c.sizeCache)
		c.sizeMu.Unlock()
		return cached
	}

	// Need to start a refresh
	ch := make(chan struct{})
	c.sizeRefreshCh = ch
	c.sizeRefreshing = true
	// Start background refresh with background context to ensure progress independent of request context
	go c.refreshSizes()
	if !cacheExists {
		// Block until initial cache is ready
		c.sizeMu.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			return map[string]sizeEntry{}
		}
		c.sizeMu.Lock()
		cached := copySizeCache(c.sizeCache)
		c.sizeMu.Unlock()
		return cached
	}
	// If cache exists (but stale), return it immediately
	cached := copySizeCache(c.sizeCache)
	c.sizeMu.Unlock()
	return cached
}

// refreshSizes fetches ContainerList with Size:true and updates the cache.
func (c *Client) refreshSizes() {
	// Perform the expensive call
	ctx := context.Background()
	containers, err := c.client.ContainerList(ctx, client.ContainerListOptions{
		All:  true,
		Size: true,
	})
	// Prepare results
	sizes := make(map[string]sizeEntry)
	if err == nil {
		for _, item := range containers.Items {
			sizes[item.ID] = sizeEntry{SizeRootFs: item.SizeRootFs, SizeRw: item.SizeRw}
		}
	} else {
		log.GetLogger().Error("Failed to refresh container sizes", "error", err)
	}

	// Update cache state
	c.sizeMu.Lock()
	defer c.sizeMu.Unlock()
	if err == nil {
		c.sizeCache = sizes
		c.sizeLastUpdated = time.Now()
	} else if c.sizeCache == nil {
		// Ensure cache becomes non-nil to unblock callers even on error
		c.sizeCache = map[string]sizeEntry{}
	}
	if c.sizeRefreshCh != nil {
		close(c.sizeRefreshCh)
	}
	c.sizeRefreshing = false
}

func copySizeCache(src map[string]sizeEntry) map[string]sizeEntry {
	if src == nil {
		return map[string]sizeEntry{}
	}
	dst := make(map[string]sizeEntry, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
