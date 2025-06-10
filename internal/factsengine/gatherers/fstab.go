package gatherers

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"strconv"
	"strings"

	"github.com/d-tux/go-fstab"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	FstabGathererName = "fstab"
	FstabFilePath     = "/etc/fstab"
)

// nolint:gochecknoglobals
var (
	FstabFileError = entities.FactGatheringError{
		Type:    "fstab-file-error",
		Message: "error reading /etc/fstab file",
	}

	FstabFileDecodingError = entities.FactGatheringError{
		Type:    "fstab-decoding-error",
		Message: "error decoding fstab file",
	}
)

type FstabEntry struct {
	Device         string   `json:"device"`
	MountPoint     string   `json:"mount_point"`
	FileSystemType string   `json:"file_system_type"`
	Options        []string `json:"options"`
	Backup         uint8    `json:"backup"`
	CheckOrder     uint     `json:"check_order"`
}

type FstabGatherer struct {
	fstabFilePath string
}

func NewFstabGatherer(filePath string) *FstabGatherer {
	return &FstabGatherer{fstabFilePath: filePath}
}

func NewDefaultFstabGatherer() *FstabGatherer {
	return &FstabGatherer{fstabFilePath: FstabFilePath}
}

func (f *FstabGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	slog.Info("Starting facts gathering process", "gatherer", FstabGathererName)
	facts := []entities.Fact{}

	mounts, err := fstab.ParseFile(f.fstabFilePath)
	if err != nil {
		return nil, FstabFileError.Wrap(err.Error())
	}

	entries := []FstabEntry{}

	for i, m := range mounts {
		if m.PassNo < 0 {
			return nil, FstabFileDecodingError.Wrap("invalid check order for mount" + strconv.Itoa(i))
		}
		if m.Freq < 0 || m.Freq > math.MaxUint8 {
			return nil, FstabFileDecodingError.Wrap("invalid backup frequency for mount" + strconv.Itoa(i))
		}
		entries = append(entries, FstabEntry{
			MountPoint:     m.File,
			Device:         m.SpecValue(),
			FileSystemType: m.VfsType,
			Options:        strings.Split(m.MntOpsString(), ","),
			Backup:         uint8(m.Freq),
			CheckOrder:     uint(m.PassNo),
		})
	}

	factValues, err := mapFstabEntriesToFactValue(entries)
	if err != nil {
		return nil, FstabFileDecodingError.Wrap(err.Error())
	}

	for _, requestedFact := range factsRequests {
		facts = append(facts, entities.NewFactGatheredWithRequest(requestedFact, factValues))
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	slog.Info("Requested facts gathered", "gatherer", FstabGathererName)

	return facts, nil
}

func mapFstabEntriesToFactValue(entries []FstabEntry) (entities.FactValue, error) {
	marshalled, err := json.Marshal(&entries)
	if err != nil {
		return nil, err
	}

	var unmarshalled []interface{}
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return nil, err
	}

	return entities.NewFactValue(unmarshalled)
}
