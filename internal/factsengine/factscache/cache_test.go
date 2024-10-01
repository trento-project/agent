package factscache_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/internal/factsengine/factscache"
	"golang.org/x/sync/errgroup"
)

type FactsCacheTestSuite struct {
	suite.Suite
	returnValue string
	count       int
}

func TestFactsCacheTestSuite(t *testing.T) {
	suite.Run(t, new(FactsCacheTestSuite))
}

func (suite *FactsCacheTestSuite) SetupSuite() {
	suite.returnValue = "value"
}

func (suite *FactsCacheTestSuite) SetupTest() {
	suite.count = 0
}

// nolint:errcheck
func (suite *FactsCacheTestSuite) TestEntries() {
	cache := factscache.NewFactsCache()
	cache.GetOrUpdate("entry1", func(_ ...interface{}) (interface{}, error) {
		return "", nil
	})
	cache.GetOrUpdate("entry2", func(_ ...interface{}) (interface{}, error) {
		return "", nil
	})
	entries := cache.Entries()

	suite.ElementsMatch([]string{"entry1", "entry2"}, entries)
}

func (suite *FactsCacheTestSuite) TestGetOrUpdate() {
	cache := factscache.NewFactsCache()

	updateFunc := func(_ ...interface{}) (interface{}, error) {
		return suite.returnValue, nil
	}

	value, err := cache.GetOrUpdate("entry1", updateFunc)

	suite.Equal(suite.returnValue, value)
	suite.NoError(err)
}

func (suite *FactsCacheTestSuite) TestGetOrUpdateWithError() {
	cache := factscache.NewFactsCache()
	someError := "some error"

	updateFunc := func(_ ...interface{}) (interface{}, error) {
		return nil, fmt.Errorf(someError)
	}

	_, err := cache.GetOrUpdate("entry", updateFunc)

	suite.EqualError(err, someError)
}

func (suite *FactsCacheTestSuite) TestGetOrUpdateCacheHit() {
	cache := factscache.NewFactsCache()

	updateFunc := func(_ ...interface{}) (interface{}, error) {
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

func (suite *FactsCacheTestSuite) TestGetOrUpdateWithArgs() {
	cache := factscache.NewFactsCache()

	// nolint:forcetypeassert
	updateFunc := func(args ...interface{}) (interface{}, error) {
		arg1 := args[0].(int)
		arg2 := args[1].(string)
		return fmt.Sprintf("%d_%s", arg1, arg2), nil
	}

	value, err := cache.GetOrUpdate("entry", updateFunc, 1, "text")

	suite.Equal("1_text", value)
	suite.NoError(err)
}

// nolint:errcheck
func (suite *FactsCacheTestSuite) TestGetOrUpdateCacheConcurrent() {
	cache := factscache.NewFactsCache()
	g := errgroup.Group{}

	updateFunc := func(args ...interface{}) (interface{}, error) {
		value, _ := args[0].(string)
		time.Sleep(100 * time.Millisecond)
		return value, nil
	}

	g.Go(func() error {
		value, _ := cache.GetOrUpdate("entry1", updateFunc, "initialValueEntry1")
		castedValue, _ := value.(string)
		suite.Equal("initialValueEntry1", castedValue)
		return nil
	})
	g.Go(func() error {
		value, _ := cache.GetOrUpdate("entry2", updateFunc, "initialValueEntry2")
		castedValue, _ := value.(string)
		suite.Equal("initialValueEntry2", castedValue)
		return nil
	})
	time.Sleep(50 * time.Millisecond)
	// The next 2 calls return the memoized value
	g.Go(func() error {
		value, _ := cache.GetOrUpdate("entry1", updateFunc, "newValueEntry1")
		castedValue, _ := value.(string)
		suite.Equal("initialValueEntry1", castedValue)
		return nil
	})

	g.Go(func() error {
		value, _ := cache.GetOrUpdate("entry2", updateFunc, "newValueEntry2")
		castedValue, _ := value.(string)
		suite.Equal("initialValueEntry2", castedValue)
		return nil
	})
	g.Wait()
}

func (suite *FactsCacheTestSuite) TestPureGetOrUpdate() {
	updateFunc := func(_ ...interface{}) (interface{}, error) {
		suite.count++
		return suite.returnValue, nil
	}

	value, err := factscache.GetOrUpdate(nil, "entry1", updateFunc)

	suite.Equal(suite.returnValue, value)
	suite.Equal(1, suite.count)
	suite.NoError(err)
}

func (suite *FactsCacheTestSuite) TestPureGetOrUpdateCacheHit() {
	cache := factscache.NewFactsCache()

	updateFunc := func(_ ...interface{}) (interface{}, error) {
		suite.count++
		return suite.returnValue, nil
	}

	// nolint:errcheck
	factscache.GetOrUpdate(cache, "entry1", updateFunc)
	value, err := factscache.GetOrUpdate(cache, "entry1", updateFunc)

	suite.Equal(suite.returnValue, value)
	suite.Equal(1, suite.count)
	suite.NoError(err)
}
