package gatherers

import (
	"encoding/json"
	"strings"

	"github.com/d-tux/go-fstab"
	log "github.com/sirupsen/logrus"
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
	Device     string   `json:"device"`
	MountPoint string   `json:"mount_point"`
	FS         string   `json:"fs"`
	Options    []string `json:"options"`
	Backup     uint8    `json:"backup"`
	CheckOrder uint     `json:"check_order"`
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

func (g *FstabGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	log.Infof("Starting %s facts gathering process", FstabGathererName)
	facts := []entities.Fact{}

	mounts, err := fstab.ParseFile(g.fstabFilePath)
	if err != nil {
		return nil, FstabFileError.Wrap(err.Error())
	}

	entries := []FstabEntry{}

	for _, m := range mounts {
		entries = append(entries, FstabEntry{
			MountPoint: m.File,
			Device:     m.SpecValue(),
			FS:         m.VfsType,
			Options:    strings.Split(m.MntOpsString(), ","),
			Backup:     uint8(m.Freq),
			CheckOrder: uint(m.PassNo),
		})
	}

	factValues, err := mapFstabEntriesToFactValue(entries)
	if err != nil {
		return nil, FstabFileDecodingError.Wrap(err.Error())
	}

	for _, requestedFact := range factsRequests {
		facts = append(facts, entities.NewFactGatheredWithRequest(requestedFact, factValues))
	}

	log.Infof("Requested %s facts gathered", FstabGathererName)

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

	return entities.NewFactValue(unmarshalled, entities.WithStringConversion())
}
