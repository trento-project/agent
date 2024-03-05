package factscache

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

type FactsCache struct {
	entries map[string]Entry
	lock    sync.Mutex
}

type Entry struct {
	content interface{}
	err     error
}

func NewFactsCache() *FactsCache {
	return &FactsCache{
		entries: make(map[string]Entry),
		lock:    sync.Mutex{},
	}
}

// GetOrUpdate Runs FactsCache GetOrUpdate with a provided cache
// If the cache is nil, it runs the function, otherwise it returns
// from cache
func GetOrUpdate(
	cache *FactsCache,
	entry string,
	udpateFunc func(args ...interface{}) (interface{}, error),
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
	c.lock.Lock()
	defer c.lock.Unlock()

	keys := []string{}
	for key := range c.entries {
		keys = append(keys, key)
	}
	return keys
}

// GetOrUpdate returns the cached result providing an entry name
// or runs the updateFunc to generate the entry.
// It locks its usage, so only one user at a time uses it
func (c *FactsCache) GetOrUpdate(
	entry string,
	udpateFunc func(args ...interface{}) (interface{}, error),
	updateFuncArgs ...interface{},
) (interface{}, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	cacheEntry, hit := c.entries[entry]
	if hit {
		log.Debugf("Value for entry %s already cached", entry)
		return cacheEntry.content, cacheEntry.err
	}

	content, err := udpateFunc(updateFuncArgs...)
	c.entries[entry] = Entry{
		content: content,
		err:     err,
	}

	if err != nil {
		log.Debugf("New value with error set for entry %s", entry)
		return content, err
	}

	log.Debugf("New value for entry %s set", entry)
	return content, err
}
