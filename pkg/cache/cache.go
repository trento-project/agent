package caching

import (
	"sync"

	"golang.org/x/sync/singleflight"

	log "github.com/sirupsen/logrus"
)

type UpdateCacheFunc func(args ...any) (any, error)

type Cache struct {
	entries sync.Map
	group   singleflight.Group
}

type Entry struct {
	Content any
	Err     error
}

func NewCache() *Cache {
	return &Cache{
		entries: sync.Map{},
		group:   singleflight.Group{},
	}
}

// GetOrUpdate Runs Cache GetOrUpdate with a provided cache
// If the cache is nil, it runs the function, otherwise it returns
// from cache
func GetOrUpdate(
	cache *Cache,
	key string,
	updateFunc UpdateCacheFunc,
	updateFuncArgs ...any,
) (any, error) {
	if cache == nil {
		return updateFunc(updateFuncArgs...)
	}

	return cache.GetOrUpdate(
		key,
		updateFunc,
		updateFuncArgs...,
	)
}

// Entries returns the cached entries list
func (c *Cache) Entries() []string {
	keys := []string{}
	c.entries.Range(func(key, _ any) bool {
		// nolint:forcetypeassert
		keys = append(keys, key.(string))
		return true
	})
	return keys
}

func (c *Cache) GetOrDefault(key string, defaultValue any) Entry {
	cacheEntry, hit := c.Get(key)
	if hit {
		return cacheEntry
	}

	return Entry{
		Content: defaultValue,
	}
}

func (c *Cache) Get(key string) (Entry, bool) {
	loadedEntry, hit := c.entries.Load(key)
	if hit {
		// nolint:forcetypeassert
		cacheEntry := loadedEntry.(Entry)
		return cacheEntry, true
	}

	return Entry{}, false
}

// GetOrUpdate returns the cached result providing an entry name
// or runs the updateFunc to generate the entry.
// It locks its usage for each used key, returning the same value of the
// first execution in the additional usages.
// If other function with a different key is asked, it runs in parallel
// without blocking.
func (c *Cache) GetOrUpdate(
	key string,
	updateFunc UpdateCacheFunc,
	updateFuncArgs ...any,
) (any, error) {
	cacheEntry, hit := c.Get(key)
	if hit {
		log.Debugf("Value for entry '%s' already cached", key)
		return cacheEntry.Content, cacheEntry.Err
	}

	entry := c.Update(key, updateFunc, updateFuncArgs...)

	return entry.Content, entry.Err
}

func (c *Cache) Update(
	key string,
	updateFunc UpdateCacheFunc,
	updateFuncArgs ...any,
) Entry {
	// singleflight is used to avoid a duplicated function execution at
	// the same moment for a given key (memoization).
	// This way, the code only blocks the execution based on same keys,
	// not blocking other keys execution
	entry, err, _ := c.group.Do(key, func() (any, error) {
		content, err := updateFunc(updateFuncArgs...)
		newEntry := Entry{
			Content: content,
			Err:     err,
		}
		c.entries.Store(key, newEntry)

		return newEntry, err
	})

	typedEntry := entry.(Entry)

	if err != nil {
		log.Debugf("New value with error set for entry '%s'", key)
		return typedEntry
	}

	log.Debugf("New value for entry '%s' set", key)
	return typedEntry
}

func DefaultUpdateFn(value any) UpdateCacheFunc {
	return func(_ ...any) (any, error) {
		return value, nil
	}
}

func (c *Cache) Delete(key string) {
	c.entries.Delete(key)
}
