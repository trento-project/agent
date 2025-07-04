package gatherers

import (
	"context"
	"log/slog"

	"github.com/trento-project/agent/internal/factsengine/factscache"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	CibAdminGathererName  = "cibadmin"
	CibAdminGathererCache = "cibadmin"
)

// nolint:gochecknoglobals
var (
	CibAdminCommandError = entities.FactGatheringError{
		Type:    "cibadmin-command-error",
		Message: "error running cibadmin command",
	}

	CibAdminDecodingError = entities.FactGatheringError{
		Type:    "cibadmin-decoding-error",
		Message: "error decoding cibadmin output",
	}
)

type CibAdminGatherer struct {
	executor utils.CommandExecutor
	cache    *factscache.FactsCache
}

func NewDefaultCibAdminGatherer() *CibAdminGatherer {
	return NewCibAdminGatherer(utils.Executor{}, nil)
}

func NewCibAdminGatherer(executor utils.CommandExecutor, cache *factscache.FactsCache) *CibAdminGatherer {
	return &CibAdminGatherer{
		executor: executor,
		cache:    cache,
	}
}

func (g *CibAdminGatherer) SetCache(cache *factscache.FactsCache) {
	g.cache = cache
}

func makeMemoizeCibAdmin(ctx context.Context) func(...interface{}) (interface{}, error) {
	return func(args ...interface{}) (interface{}, error) {
		executor, ok := args[0].(utils.CommandExecutor)
		if !ok {
			return nil, ImplementationError.Wrap("error using memoizeCibAdmin. executor must be 1st argument")
		}
		return executor.ExecContext(ctx, "/usr/sbin/cibadmin", "--query", "--local")
	}

}

func (g *CibAdminGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	slog.Info("Starting facts gathering process", "gatherer", CibAdminGathererName)

	content, err := factscache.GetOrUpdate(g.cache, CibAdminGathererCache, makeMemoizeCibAdmin(ctx), g.executor)

	if err != nil {
		return nil, CibAdminCommandError.Wrap(err.Error())
	}

	cibadmin, ok := content.([]byte)
	if !ok {
		return nil, CibAdminDecodingError.Wrap("error casting the command output")
	}

	elementsToList := map[string]bool{"primitive": true, "clone": true, "master": true, "group": true,
		"nvpair": true, "op": true, "rsc_location": true, "rsc_order": true,
		"rsc_colocation": true, "cluster_property_set": true, "meta_attributes": true}

	factValueMap, err := parseXMLToFactValueMap(cibadmin, elementsToList, entities.WithStringConversion())
	if err != nil {
		return nil, CibAdminDecodingError.Wrap(err.Error())
	}

	facts := []entities.Fact{}

	for _, factReq := range factsRequests {
		var fact entities.Fact

		if value, err := factValueMap.GetValue(factReq.Argument); err == nil {
			fact = entities.NewFactGatheredWithRequest(factReq, value)

		} else {
			slog.Error(err.Error())
			fact = entities.NewFactGatheredWithError(factReq, err)
		}
		facts = append(facts, fact)
	}

	slog.Info("Requested facts gathered", "gatherer", CibAdminGathererName)
	return facts, err
}
