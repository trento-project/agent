package gatherers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
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
	Name  string   `json:"name,omitempty"`
	Owner string   `json:"owner"`
	Group string   `json:"group"`
	Files []string `json:"files"`
}

type DirScanStatInfo struct {
	Group string
	Owner string
}

type DirScanResult map[string]DirScanDetails

//go:generate mockery --name=UserSearcher
type UserSearcher interface {
	GetUsernameByID(userID string) (string, error)
}

//go:generate mockery --name=GroupSearcher
type GroupSearcher interface {
	GetGroupByID(groupID string) (string, error)
}

type DirScanGatherer struct {
	fs            afero.Fs
	userSearcher  UserSearcher
	groupSearcher GroupSearcher
}

func NewDirScanGatherer(fs afero.Fs, userSearcher UserSearcher, groupSearcher GroupSearcher) *DirScanGatherer {
	return &DirScanGatherer{fs: fs, userSearcher: userSearcher, groupSearcher: groupSearcher}
}

func NewDefaultDirScanGatherer() *DirScanGatherer {
	cf := CredentialsFetcher{}
	return NewDirScanGatherer(afero.NewOsFs(), &cf, &cf)
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

		resultKey := getDirScanResultKeyFromPath(match)

		if resultEntry, found := result[resultKey]; found {
			resultEntry.Files = append(resultEntry.Files, match)
			result[resultKey] = resultEntry
			continue
		}

		statInfo, err := d.getStatInfoForPath(match)
		if err != nil {
			return nil, err
		}

		newEntry := DirScanDetails{
			Owner: statInfo.Owner,
			Group: statInfo.Group,
			Files: []string{match},
		}

		result[resultKey] = newEntry
	}
	return result, nil
}

func (d *DirScanGatherer) getStatInfoForPath(path string) (*DirScanStatInfo, error) {
	fi, err := d.fs.Stat(path)
	if err != nil {
		return nil, err
	}

	stat, ok := fi.Sys().(*syscall.Stat_t) //nolint
	if !ok {
		return nil, fmt.Errorf("could not extract stat infos for file %s", path)
	}
	uid := strconv.Itoa(int(stat.Uid))
	gid := strconv.Itoa(int(stat.Gid))

	group, err := d.groupSearcher.GetGroupByID(gid)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve group for gigroupd %s", gid)
	}
	user, err := d.userSearcher.GetUsernameByID(uid)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve group for uid %s", uid)
	}

	return &DirScanStatInfo{
		Group: group,
		Owner: user,
	}, nil
}

func getDirScanResultKeyFromPath(path string) string {
	return filepath.Dir(path)
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
