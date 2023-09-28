package gatherers

import (
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

	log.Infof("Requested %s facts gathered", DirScanGathererName)

	return facts, nil
}

func (d *DirScanGatherer) extractDirScanDetails(dirscanPath string) (DirScanResult, error) {
	result := DirScanResult{}
	err := afero.Walk(d.fs, dirscanPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("could not extract stat infos for file %s", path)
		}

		owner := stat.Uid
		group := stat.Gid

		resultKey := path
		if info.Mode().IsRegular() {
			resultKey = filepath.Dir(path)
		}

		// todo: distinguish between a dir and not a dir, and if is a directory
		// create the entry in the map with the directory path as key
	})

	if err != nil {
		return nil, DirScanScanningError.Wrap(err.Error())
	}

	return result, nil
}
