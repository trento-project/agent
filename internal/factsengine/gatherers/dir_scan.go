package gatherers

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	DirScanGathererName = "dir-scan"
)

// nolint:gochecknoglobals
var (
	DirScanMissingArgumentError = entities.FactGatheringError{
		Type:    "dir-scan-missing-argument",
		Message: "missing required argument",
	}

	DirScanOpenError = entities.FactGatheringError{
		Type:    "dir-scan-open-error",
		Message: "could not open the provided directory",
	}

	DirScanScanningError = entities.FactGatheringError{
		Type:    "dir-scan-scanning-error",
		Message: "error during directory scanning",
	}
)

type DirScanDetails struct {
	Name  string   `json:"-"`
	Owner uint32   `json:"owner"`
	Group uint32   `json:"group"`
	Files []string `json:"files"`
}

type DirScanResult map[string]DirScanDetails

type DirScanGatherer struct {
	fs afero.Fs
}

func NewDirScanGatherer(fs afero.Fs) *DirScanGatherer {
	return &DirScanGatherer{fs: fs}
}

func NewDefaultDirScanGatherer() *DirScanGatherer {
	return NewDirScanGatherer(afero.NewOsFs())
}

func (d *DirScanGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	log.Infof("Starting %s facts gathering process", DirScanGathererName)
	facts := []entities.Fact{}

	for _, requestedFact := range factsRequests {
		if requestedFact.Argument == "" {
			facts = append(facts, entities.NewFactGatheredWithError(requestedFact, &DirScanMissingArgumentError))
			continue
		}
		scanResult, err := d.extractDirScanDetails(requestedFact.Argument)
		if err != nil {
			facts = append(facts, entities.NewFactGatheredWithError(requestedFact, DirScanScanningError.Wrap(err.Error())))
			continue
		}
		factValue, err := mapDirScanResultToFactValue(scanResult)
		if err != nil {
			facts = append(facts, entities.NewFactGatheredWithError(requestedFact, DirScanScanningError.Wrap(err.Error())))
			continue
		}
		facts = append(facts, entities.NewFactGatheredWithRequest(requestedFact, factValue))
	}

	log.Infof("Requested %s facts gathered", DirScanGathererName)

	return facts, nil
}

func (d *DirScanGatherer) extractDirScanDetails(dirscanPath string) (DirScanResult, error) {
	result := DirScanResult{}

	matches, err := afero.Glob(d.fs, dirscanPath)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		fi, err := d.fs.Stat(match)
		if err != nil {
			return nil, err
		}

		resultKey := getDirScanResultKeyFromPath(match, fi)
		resultEntry := getDirScanDetailsFromResultOrDefault(result, resultKey)

		if fi.IsDir() {
			stat, ok := fi.Sys().(*syscall.Stat_t) //nolint
			if !ok {
				return nil, fmt.Errorf("could not extract stat infos for file %s", match)
			}

			resultEntry.Group = stat.Gid
			resultEntry.Owner = stat.Uid
		} else {
			resultEntry.Files = append(resultEntry.Files, match)
		}

		result[resultKey] = resultEntry
	}

	if err != nil {
		return nil, err
	}

	return result, nil
}

func getDirScanResultKeyFromPath(path string, info fs.FileInfo) string {
	if info.IsDir() {
		return path
	}
	return filepath.Dir(path)
}

func getDirScanDetailsFromResultOrDefault(result DirScanResult, key string) DirScanDetails {
	currentEntry, ok := result[key]
	if !ok {
		currentEntry = DirScanDetails{
			Name:  key,
			Files: []string{},
		}
	}

	return currentEntry
}

func mapDirScanResultToFactValue(result DirScanResult) (entities.FactValue, error) {
	marshalled, err := json.Marshal(&result)
	if err != nil {
		return nil, err
	}

	var unmarshalled interface{}
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return nil, err
	}

	return entities.NewFactValue(unmarshalled)
}
