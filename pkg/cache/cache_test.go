package caching_test

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
	caching "github.com/trento-project/agent/pkg/cache"
	"golang.org/x/sync/errgroup"
)

type CacheTestSuite struct {
	suite.Suite
	returnValue string
	count       int
}

func successfulUpdateFunction(args ...any) (any, error) {
	return args[0], nil
}

func updateFunctionWithError(args ...any) (any, error) {
	return nil, errors.New(args[0].(string))
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CacheTestSuite))
}

func (suite *CacheTestSuite) SetupSuite() {
	suite.returnValue = "value"
}

func (suite *CacheTestSuite) SetupTest() {
	suite.count = 0
}

// nolint:errcheck
func (suite *CacheTestSuite) TestEntries() {
	cache := caching.NewCache()
	cache.GetOrUpdate("entry1", func(_ ...any) (any, error) {
		return "", nil
	})
	cache.GetOrUpdate("entry2", func(_ ...any) (any, error) {
		return "", nil
	})
	entries := cache.Entries()

	suite.ElementsMatch([]string{"entry1", "entry2"}, entries)
}

func (suite *CacheTestSuite) TestGetOrUpdate() {
	cache := caching.NewCache()

	value, err := cache.GetOrUpdate("entry1", successfulUpdateFunction, suite.returnValue)

	suite.Equal(suite.returnValue, value)
	suite.NoError(err)
}

func (suite *CacheTestSuite) TestGetOrUpdateWithError() {
	cache := caching.NewCache()
	someError := "some error"

	_, err := cache.GetOrUpdate("entry", updateFunctionWithError, someError)

	suite.EqualError(err, someError)
}

func (suite *CacheTestSuite) TestGetOrUpdateCacheHit() {
	cache := caching.NewCache()

	updateFunc := func(_ ...any) (any, error) {
		suite.count++
		return suite.returnValue, nil
	}

	// nolint:errcheck
	cache.GetOrUpdate("entry", updateFunc)
	value, err := cache.GetOrUpdate("entry", updateFunc)

	suite.Equal(suite.returnValue, value)
	suite.Equal(1, suite.count)
	suite.NoError(err)
}

func (suite *CacheTestSuite) TestGetOrUpdateWithArgs() {
	cache := caching.NewCache()

	// nolint:forcetypeassert
	updateFunc := func(args ...any) (any, error) {
		arg1 := args[0].(int)
		arg2 := args[1].(string)
		return fmt.Sprintf("%d_%s", arg1, arg2), nil
	}

	value, err := cache.GetOrUpdate("entry", updateFunc, 1, "text")

	suite.Equal("1_text", value)
	suite.NoError(err)
}

// nolint:errcheck
func (suite *CacheTestSuite) TestGetOrUpdateCacheConcurrent() {
	cache := caching.NewCache()
	g := errgroup.Group{}

	key1 := "entry1"
	key2 := "entry2"

	updateFunc := func(args ...any) (any, error) {
		value, _ := args[0].(string)
		time.Sleep(100 * time.Millisecond)
		return value, nil
	}

	g.Go(func() error {
		value, _ := cache.GetOrUpdate(key1, updateFunc, "initialValueEntry1")
		castedValue, _ := value.(string)
		suite.Equal("initialValueEntry1", castedValue)
		return nil
	})
	g.Go(func() error {
		value, _ := cache.GetOrUpdate(key2, updateFunc, "initialValueEntry2")
		castedValue, _ := value.(string)
		suite.Equal("initialValueEntry2", castedValue)
		return nil
	})
	time.Sleep(50 * time.Millisecond)
	// The next 2 calls return the memoized value
	g.Go(func() error {
		value, _ := cache.GetOrUpdate(key1, updateFunc, "newValueEntry1")
		castedValue, _ := value.(string)
		suite.Equal("initialValueEntry1", castedValue)
		return nil
	})

	g.Go(func() error {
		value, _ := cache.GetOrUpdate(key2, updateFunc, "newValueEntry2")
		castedValue, _ := value.(string)
		suite.Equal("initialValueEntry2", castedValue)
		return nil
	})
	g.Wait()
}

func (suite *CacheTestSuite) TestGetMissingEntry() {
	cache := caching.NewCache()
	key := "entry1"

	entry, hit := cache.Get(key)

	suite.False(hit)
	suite.Nil(entry.Content)
	suite.NoError(entry.Err)
}

func (suite *CacheTestSuite) TestGetAlreadyCachedEntry() {
	cache := caching.NewCache()
	key := "entry1"
	initialValue := "foo"

	cache.GetOrUpdate(key, successfulUpdateFunction, initialValue)

	entry, hit := cache.Get(key)

	suite.True(hit)
	suite.Equal(initialValue, entry.Content)
	suite.NoError(entry.Err)
}

func (suite *CacheTestSuite) TestGetAlreadyCachedEntryWithError() {
	cache := caching.NewCache()
	key := "entry1"

	someError := "some error"

	cache.GetOrUpdate(key, updateFunctionWithError, someError)

	entry, hit := cache.Get(key)

	suite.True(hit)
	suite.Nil(entry.Content)
	suite.EqualError(entry.Err, someError)
}

func (suite *CacheTestSuite) TestUpdatePreviouslyCachedEntry() {
	cache := caching.NewCache()
	key := "entry1"
	initialValue := "foo"

	cache.GetOrUpdate(key, successfulUpdateFunction, initialValue)

	newValue := "bar"

	entry := cache.Update(key, caching.DefaultUpdateFn(newValue))
	suite.Equal(newValue, entry.Content)
	suite.NoError(entry.Err)
}

func (suite *CacheTestSuite) TestUpdateNonPreviouslyCachedEntry() {
	cache := caching.NewCache()
	key := "entry1"

	value := "bar"

	entry := cache.Update(key, caching.DefaultUpdateFn(value))
	suite.NoError(entry.Err)
	suite.Equal(value, entry.Content)
}

func (suite *CacheTestSuite) TestUpdateCachedEntryWithError() {
	cache := caching.NewCache()
	key := "entry1"

	initialValue := "foo"
	value, err := cache.GetOrUpdate(key, successfulUpdateFunction, initialValue)
	suite.NoError(err)
	suite.Equal(initialValue, value)

	someError := "some error"

	entry := cache.Update(key, updateFunctionWithError, someError)
	suite.Nil(entry.Content)
	suite.EqualError(entry.Err, someError)
}

func (suite *CacheTestSuite) TestConcurrentUpdate() {
	cache := caching.NewCache()
	g := errgroup.Group{}

	key1 := "entry1"
	key2 := "entry2"

	updateFunc := func(args ...any) (any, error) {
		value, _ := args[0].(string)
		time.Sleep(150 * time.Millisecond)
		return value, nil
	}

	g.Go(func() error {
		entry := cache.Update(key1, updateFunc, "fooValue")
		castedValue, _ := entry.Content.(string)
		suite.Equal("fooValue", castedValue)
		return nil
	})
	time.Sleep(50 * time.Millisecond)
	g.Go(func() error {
		// this goroutine will concurrently attempt to update the same key that the first one is updating
		// current call will wait for the first call to finish and will get what has been updated in the first call
		entry := cache.Update(key1, caching.DefaultUpdateFn("fooValue2"))
		castedValue, _ := entry.Content.(string)
		suite.Equal("fooValue", castedValue)
		return nil
	})
	g.Go(func() error {
		// this goroutine will also run concurrently to the first one
		// however since it is updating a different key, it will not be blocked by the first call
		entry := cache.Update(key2, caching.DefaultUpdateFn("barValue"))
		castedValue, _ := entry.Content.(string)
		suite.Equal("barValue", castedValue)
		return nil
	})

	g.Wait()
}

func (suite *CacheTestSuite) TestDeleteNonCachedEntry() {
	cache := caching.NewCache()
	key := "entry1"

	entry, hit := cache.Get(key)
	suite.False(hit)
	suite.Nil(entry.Content)

	cache.Delete(key)

	entry, hit = cache.Get(key)
	suite.False(hit)
	suite.Nil(entry.Content)
}

func (suite *CacheTestSuite) TestDeleteEntry() {
	cache := caching.NewCache()
	key := "entry1"
	initialValue := "foo"

	cache.Update(key, caching.DefaultUpdateFn(initialValue))

	entry, hit := cache.Get(key)
	suite.True(hit)
	suite.Equal(initialValue, entry.Content)

	cache.Delete(key)

	entry, hit = cache.Get(key)
	suite.False(hit)
	suite.Nil(entry.Content)
}

func (suite *CacheTestSuite) TestConcurrentlyNotConflictingDelete() {
	cache := caching.NewCache()
	g := errgroup.Group{}

	key1 := "entry1"
	key2 := "entry2"

	cache.Update(key1, caching.DefaultUpdateFn("foo"))
	entry, hit := cache.Get(key1)
	suite.True(hit)
	suite.Equal("foo", entry.Content)

	cache.Update(key2, caching.DefaultUpdateFn("bar"))
	entry, hit = cache.Get(key2)
	suite.True(hit)
	suite.Equal("bar", entry.Content)

	deleteKey1 := func() error {
		cache.Delete(key1)
		entry, hit := cache.Get(key1)
		suite.False(hit)
		suite.Nil(entry.Content)
		return nil
	}
	deleteKey2 := func() error {
		cache.Delete(key2)
		entry, hit := cache.Get(key2)
		suite.False(hit)
		suite.Nil(entry.Content)
		return nil
	}

	options := []func() error{deleteKey1, deleteKey2}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for range 20 {
		g.Go(options[rng.Intn(len(options))])
	}

	g.Wait()
}

func (suite *CacheTestSuite) TestConcurrentDelete() {
	cache := caching.NewCache()

	// Helper function to simulate cache updates (to fill the cache first)
	updateFunc := caching.DefaultUpdateFn("TestContent")

	// Simulate adding some entries to the cache
	keys := []string{"key1", "key2", "key3", "key4"}
	for _, key := range keys {
		_, err := caching.GetOrUpdate(cache, key, updateFunc)
		if err != nil {
			suite.T().Fatalf("Failed to update cache for %s: %v", key, err)
		}
	}

	// Check initial cache state
	if len(cache.Entries()) != len(keys) {
		suite.T().Errorf("Expected cache entries: %v, but got: %v", keys, cache.Entries())
	}

	// Create a WaitGroup to synchronize concurrent Delete operations
	var wg sync.WaitGroup
	var mu sync.Mutex // Mutex to check for correctness in test

	// Run concurrent delete operations
	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			cache.Delete(k)

			// Ensure that the entry is deleted
			mu.Lock()
			_, exists := cache.Get(k)
			if exists {
				suite.T().Errorf("Cache entry %s was not deleted", k)
			}
			mu.Unlock()
		}(key)
	}

	// Wait for all delete operations to finish
	wg.Wait()

	// Check that the cache is empty after deletion
	if len(cache.Entries()) != 0 {
		suite.T().Errorf("Expected cache to be empty after deletions, but got: %v", cache.Entries())
	}
}

func (suite *CacheTestSuite) TestConcurrentDeleteAndUpdated() {
	cache := caching.NewCache()

	// Helper function to simulate cache updates (to fill the cache first)
	updateFunc := caching.DefaultUpdateFn("TestContent")

	// Simulate adding some entries to the cache
	keys := []string{"key1", "key2", "key3", "key4"}
	for _, key := range keys {
		_, err := caching.GetOrUpdate(cache, key, updateFunc)
		if err != nil {
			suite.T().Fatalf("Failed to update cache for %s: %v", key, err)
		}
	}

	// Concurrently delete and add entries
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := range 10 {
		wg.Add(2)

		// Delete an entry concurrently
		go func(k string) {
			defer wg.Done()
			cache.Delete(k)

			// Ensure that the entry is deleted
			mu.Lock()
			_, exists := cache.Get(k)
			if exists {
				suite.T().Errorf("Cache entry %s was not deleted", k)
			}
			mu.Unlock()
		}(keys[i%len(keys)])

		// Add a new entry concurrently
		go func(k string) {
			defer wg.Done()
			_, err := caching.GetOrUpdate(cache, k, updateFunc)
			if err != nil {
				suite.T().Errorf("Failed to update cache for %s: %v", k, err)
			}
		}(keys[i%len(keys)])
	}

	// Wait for all operations to finish
	wg.Wait()

	// Check the cache entries after operations
	remainingEntries := cache.Entries()
	if len(remainingEntries) == 0 {
		suite.T().Errorf("Expected some entries to remain in cache, but found none.")
	}
}

func (suite *CacheTestSuite) TestPureGetOrUpdate() {
	updateFunc := func(_ ...any) (any, error) {
		suite.count++
		return suite.returnValue, nil
	}

	value, err := caching.GetOrUpdate(nil, "entry1", updateFunc)

	suite.Equal(suite.returnValue, value)
	suite.Equal(1, suite.count)
	suite.NoError(err)
}

func (suite *CacheTestSuite) TestPureGetOrUpdateCacheHit() {
	cache := caching.NewCache()

	updateFunc := func(_ ...any) (any, error) {
		suite.count++
		return suite.returnValue, nil
	}

	// nolint:errcheck
	caching.GetOrUpdate(cache, "entry1", updateFunc)
	value, err := caching.GetOrUpdate(cache, "entry1", updateFunc)

	suite.Equal(suite.returnValue, value)
	suite.Equal(1, suite.count)
	suite.NoError(err)
}
