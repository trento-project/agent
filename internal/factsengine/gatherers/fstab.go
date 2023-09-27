package gatherers

import (
	"encoding/json"
	"os"

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
	DeviceID   string   `json:"device_id"`
	MountPoint string   `json:"mount_point"`
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
	facts := []entities.Fact{}

	fstabFile, err := os.Open(g.fstabFilePath)
	if err != nil {
		return nil, FstabFileError.Wrap(err.Error())
	}

	defer func() {
		err := fstabFile.Close()
		if err != nil {
			log.Errorf("could not close fstab file %s, error: %s", g.fstabFilePath, err)
		}
	}()

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
