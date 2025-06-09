package factscache

import (
	"sync"

	"log/slog"

	"golang.org/x/sync/singleflight"
)

type UpdateCacheFunc func(args ...interface{}) (interface{}, error)

type FactsCache struct {
	entries sync.Map
	group   singleflight.Group
}

type Entry struct {
	content interface{}
	err     error
}

func NewFactsCache() *FactsCache {
	return &FactsCache{
		entries: sync.Map{},
		group:   singleflight.Group{},
	}
}

// GetOrUpdate Runs FactsCache GetOrUpdate with a provided cache
// If the cache is nil, it runs the function, otherwise it returns
// from cache
func GetOrUpdate(
	cache *FactsCache,
	entry string,
	udpateFunc UpdateCacheFunc,
	updateFuncArgs ...interface{},
) (interface{}, error) {
	if cache == nil {
		return udpateFunc(updateFuncArgs...)
	}

	return cache.GetOrUpdate(
		entry,
		udpateFunc,
		updateFuncArgs...,
	)
}

// Entries returns the cached entries list
func (c *FactsCache) Entries() []string {
	keys := []string{}
	c.entries.Range(func(key, _ any) bool {
		// nolint:forcetypeassert
		keys = append(keys, key.(string))
		return true
	})
	return keys
}

// GetOrUpdate returns the cached result providing an entry name
// or runs the updateFunc to generate the entry.
// It locks its usage for each used key, returning the same value of the
// first execution in the additional usages.
// If other function with a different key is asked, it runs in parallel
// without blocking.
func (c *FactsCache) GetOrUpdate(
	entry string,
	udpateFunc UpdateCacheFunc,
	updateFuncArgs ...interface{},
) (interface{}, error) {
	loadedEntry, hit := c.entries.Load(entry)
	if hit {
		// nolint:forcetypeassert
		cacheEntry := loadedEntry.(Entry)
		slog.Debug("Value for entry already cached", "entry", entry)
		return cacheEntry.content, cacheEntry.err
	}

	// singleflight is used to avoid a duplicated function execution at
	// the same moment for a given key (memoization).
	// This way, the code only blocks the execution based on same keys,
	// not blocking other keys execution
	content, err, _ := c.group.Do(entry, func() (interface{}, error) {
		content, err := udpateFunc(updateFuncArgs...)
		newEntry := Entry{
			content: content,
			err:     err,
		}
		c.entries.Store(entry, newEntry)

		return content, err
	})

	if err != nil {
		slog.Debug("New value with error set for entry", "entry", entry)
		return content, err
	}

	slog.Debug("New value for entry set", "entry", entry)
	return content, err
}
