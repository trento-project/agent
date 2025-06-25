package gatherers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"syscall"

	"log/slog"

	"github.com/spf13/afero"

	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	DirScanGathererName = "dir_scan"
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
	Name  string `json:"name"`
	Owner string `json:"owner"`
	Group string `json:"group"`
}

type DirScanResult []DirScanDetails

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

func (d *DirScanGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	slog.Info("Starting facts gathering process", "gatherer", DirScanGathererName)

	results := make(chan []entities.Fact)

	go func() {
		facts := []entities.Fact{}
		for _, requestedFact := range factsRequests {
			fact := d.gatherSingle(requestedFact)
			facts = append(facts, fact)
		}
		results <- facts
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case facts := <-results:
		slog.Info("Requested facts gathered", "gatherer", DirScanGathererName)
		return facts, nil
	}
}

func (d *DirScanGatherer) gatherSingle(requestedFact entities.FactRequest) entities.Fact {
	if requestedFact.Argument == "" {
		slog.Error("could not gather facts for gatherer, missing argument", "gatherer", DirScanGathererName)
		return entities.NewFactGatheredWithError(requestedFact, &DirScanMissingArgumentError)

	}
	scanResult, err := d.extractDirScanDetails(requestedFact.Argument)
	if err != nil {
		return entities.NewFactGatheredWithError(requestedFact, DirScanScanningError.Wrap(err.Error()))
	}
	factValue, err := mapDirScanResultToFactValue(scanResult)
	if err != nil {
		return entities.NewFactGatheredWithError(requestedFact, DirScanScanningError.Wrap(err.Error()))

	}
	return entities.NewFactGatheredWithRequest(requestedFact, factValue)
}

func (d *DirScanGatherer) extractDirScanDetails(dirscanPath string) (DirScanResult, error) {
	result := DirScanResult{}

	matches, err := afero.Glob(d.fs, dirscanPath)
	if err != nil {
		return nil, err
	}

	for _, match := range matches {
		scanDetails, err := d.getDirScanDetailsForPath(match)
		if err != nil {
			return nil, err
		}

		result = append(result, *scanDetails)
	}
	return result, nil
}

func (d *DirScanGatherer) getDirScanDetailsForPath(path string) (*DirScanDetails, error) {
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
		return nil, fmt.Errorf("could not retrieve group for gid %s", gid)
	}
	user, err := d.userSearcher.GetUsernameByID(uid)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve username for uid %s", uid)
	}

	return &DirScanDetails{
		Group: group,
		Owner: user,
		Name:  path,
	}, nil
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
