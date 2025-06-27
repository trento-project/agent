package gatherers

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"

	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/factsengine/plugininterface"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	FSUsageGathererName = "fs_usage"
)

// nolint:gochecknoglobals
var (
	FSUsageInvalidFormatError = entities.FactGatheringError{
		Type:    "fs-usage-invalid-format-error",
		Message: "the df output is invalid",
	}

	FSUsageExecutionError = entities.FactGatheringError{
		Type:    "fs-usage-execution-error",
		Message: "the execution of df failed",
	}

	FSUsageConversionError = entities.FactGatheringError{
		Type:    "fs-usage-conversion-error",
		Message: "failed to convert usage information to a fact",
	}
)

type FSUsageEntry struct {
	// Filesystem specifies the type of the filesystem.  This can either be the
	// device backing the filesystem or a virtual filesystem like tmpfs.  The
	// content of this field is dependent on the implementation of df used by the
	// system.  In SUSE this will most likely be GNU df
	Filesystem string
	// Blocks specifies the amount of 1024 byte blocks, the filesystem consists
	// out of
	Blocks int
	// Used specifies how many 1024 byte blocks are used
	Used int
	// Available specifies how many 1024 byte blocks are free.  The number may be
	// negative
	Available int
	// Capacity specifies the capacity in percent used by the file system.  The
	// number may be bigger than 100 when Available is negative
	Capacity int
	// Mountpoint specifies the path the filesystem is mounted at
	Mountpoint string
}

var _ plugininterface.Gatherer = (*FSUsageGatherer)(nil)

type FSUsageGatherer struct {
	executor utils.CommandExecutor
}

func NewDefaultFSUsageGatherer() *FSUsageGatherer {
	return NewFSUsageGatherer(utils.Executor{})
}

func NewFSUsageGatherer(executor utils.CommandExecutor) *FSUsageGatherer {
	return &FSUsageGatherer{
		executor: executor,
	}
}

func (f *FSUsageGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	slog.Info("Starting facts gathering process", "gatherer", FSUsageGathererName)
	facts := []entities.Fact{}

	for _, factReq := range factsRequests {
		var data []FSUsageEntry
		var err *entities.FactGatheringError

		if factReq.Argument == "" {
			data, err = f.gatherAll(ctx)
		} else {
			data, err = f.gatherSingle(ctx, factReq.Argument)
		}
		if err != nil {
			facts = append(facts, entities.NewFactGatheredWithError(factReq, err))
			continue
		}

		factValue, conversionErr := fsUsageEntriesToFactValue(data)
		if conversionErr != nil {
			facts = append(facts, entities.NewFactGatheredWithError(factReq, FSUsageConversionError.Wrap(err.Error())))
			continue
		}

		facts = append(facts, entities.NewFactGatheredWithRequest(factReq, factValue))
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	slog.Info("Requested facts gathered", "gatherer", FSUsageGathererName)

	return facts, nil
}

// parseFSUsageOutput parses the output of the df -P command
func (f *FSUsageGatherer) parseFSUsageOutput(b []byte) ([]FSUsageEntry, error) {
	entries := []FSUsageEntry{}

	lines := strings.Split(string(b), "\n")
	dfRegex := regexp.MustCompile(`^(.*?) +(\d+) +(\d+) +(-?\d+) +(\d+)% +(.*)$`)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		submatches := dfRegex.FindStringSubmatch(line)
		isMatch := dfRegex.MatchString(line)

		if (!isMatch && i == 0) || line == "" {
			// Skip parsing the heading or trailing newlines
			continue
		}

		if !isMatch || len(submatches) != 7 {
			return nil, fmt.Errorf("parse df: unexpected format")
		}

		filesystem := strings.TrimSpace(submatches[1])
		// These errors can not happen, because the above matched
		blocks, err := strconv.Atoi(submatches[2])
		if err != nil {
			return nil, fmt.Errorf("parse df: blocks: invalid format: %w", err)
		}
		used, err := strconv.Atoi(submatches[3])
		if err != nil {
			return nil, fmt.Errorf("parse df: used: invalid format: %w", err)
		}
		available, err := strconv.Atoi(submatches[4])
		if err != nil {
			return nil, fmt.Errorf("parse df: output: available: invalid format: %w", err)
		}
		capacity, err := strconv.Atoi(submatches[5])
		if err != nil {
			return nil, fmt.Errorf("parse df: capacity: invalid format: %w", err)
		}
		mountpoint := strings.TrimSpace(submatches[6])

		entries = append(entries, FSUsageEntry{
			Filesystem: filesystem,
			Blocks:     blocks,
			Used:       used,
			Available:  available,
			Capacity:   capacity,
			Mountpoint: mountpoint,
		})
	}

	return entries, nil
}

func (f *FSUsageGatherer) gatherAll(ctx context.Context) ([]FSUsageEntry, *entities.FactGatheringError) {
	// Output in 1024 Blocks
	content, err := f.executor.ExecContext(ctx, "/usr/bin/df", "-k", "-P", "--")
	if err != nil {
		return nil, FSUsageExecutionError.Wrap(err.Error())
	}

	entries, err := f.parseFSUsageOutput(content)
	if err != nil {
		return nil, FSUsageInvalidFormatError.Wrap(err.Error())
	}

	return entries, nil
}

// nolint:lll
func (f *FSUsageGatherer) gatherSingle(ctx context.Context, file string) ([]FSUsageEntry, *entities.FactGatheringError) {
	// Output in 1024 Blocks
	content, err := f.executor.ExecContext(ctx, "/usr/bin/df", "-k", "-P", "--", file)
	if err != nil {
		return nil, FSUsageExecutionError.Wrap(err.Error())
	}

	entries, err := f.parseFSUsageOutput(content)
	if err != nil {
		return nil, FSUsageInvalidFormatError.Wrap(err.Error())
	}

	return entries, nil
}

func (f *FSUsageEntry) toFactValue() (entities.FactValue, error) {
	erased := map[string]any{
		"filesystem": f.Filesystem,
		"blocks":     f.Blocks,
		"used":       f.Used,
		"available":  f.Available,
		"capacity":   f.Capacity,
		"mountpoint": f.Mountpoint,
	}

	return entities.NewFactValue(erased)
}

func fsUsageEntriesToFactValue(entries []FSUsageEntry) (entities.FactValue, error) {
	factValues := make([]entities.FactValue, 0)

	for _, entry := range entries {
		v, err := entry.toFactValue()
		if err != nil {
			return nil, err
		}

		factValues = append(factValues, v)
	}

	return &entities.FactValueList{Value: factValues}, nil
}
