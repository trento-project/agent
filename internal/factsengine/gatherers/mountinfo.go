package gatherers

import (
	"encoding/json"
	"strings"

	"github.com/hashicorp/go-envparse"
	"github.com/moby/sys/mountinfo"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	MountInfoGathererName = "mount_info"
)

// nolint:gochecknoglobals
var (
	MountInfoParsingError = entities.FactGatheringError{
		Type:    "mount-info-parsing-error",
		Message: "error parsing mount information",
	}

	MountInfoMissingArgumentError = entities.FactGatheringError{
		Type:    "mount-info-missing-argument",
		Message: "missing required argument",
	}

	MountInfoDecodingError = entities.FactGatheringError{
		Type:    "mount-info-decoding-error",
		Message: "error decoding mount information",
	}
)

//go:generate mockery --name=MountParserInterface
type MountParserInterface interface {
	GetMounts(f mountinfo.FilterFunc) ([]*mountinfo.Info, error)
}

type MountParser struct{}

func (x *MountParser) GetMounts(f mountinfo.FilterFunc) ([]*mountinfo.Info, error) {
	return mountinfo.GetMounts(f)
}

type MountInfoGatherer struct {
	mInfo    MountParserInterface
	executor utils.CommandExecutor
}

type MountInfoResult struct {
	BlockUUID  string `json:"block_uuid"`
	FSType     string `json:"fs_type"`
	MountPoint string `json:"mount_point"`
	Options    string `json:"options"`
	Source     string `json:"source"`
}

func NewDefaultMountInfoGatherer() *MountInfoGatherer {
	return NewMountInfoGatherer(&MountParser{}, utils.Executor{})
}

func NewMountInfoGatherer(mInfo MountParserInterface, executor utils.CommandExecutor) *MountInfoGatherer {
	return &MountInfoGatherer{mInfo: mInfo, executor: executor}
}

func (g *MountInfoGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting %s facts gathering process", MountInfoGathererName)
	mounts, err := g.mInfo.GetMounts(nil)
	if err != nil {
		return nil, MountInfoParsingError.Wrap(err.Error())
	}

	for _, requestedFact := range factsRequests {
		if requestedFact.Argument == "" {
			log.Errorf("could not gather facts for %s gatherer, missing argument", MountInfoGathererName)
			facts = append(facts, entities.NewFactGatheredWithError(requestedFact, &MountInfoMissingArgumentError))
			continue
		}

		var foundMountInfoResult = MountInfoResult{}

		for _, mount := range mounts {
			if mount.Mountpoint == requestedFact.Argument {
				foundMountInfoResult = MountInfoResult{
					MountPoint: mount.Mountpoint,
					Source:     mount.Source,
					FSType:     mount.FSType,
					Options:    mount.Options,
				}

				if blkidOuptut, err := g.executor.Exec("blkid", foundMountInfoResult.Source, "-o", "export"); err != nil {
					log.Warnf("blkid command failed for source %s: %s", foundMountInfoResult.Source, err)
				} else if fields, err := envparse.Parse(strings.NewReader(string(blkidOuptut))); err != nil {
					log.Warnf("error parsing the blkid output: %s", err)
				} else {
					foundMountInfoResult.BlockUUID = fields["UUID"]
				}

				continue
			}
		}

		factValue, err := mountInfoResultToFactValue(foundMountInfoResult)
		if err != nil {
			facts = append(facts, entities.NewFactGatheredWithError(requestedFact, MountInfoDecodingError.Wrap(err.Error())))
			continue
		}
		facts = append(facts, entities.NewFactGatheredWithRequest(requestedFact, factValue))
	}

	log.Infof("Requested %s facts gathered", MountInfoGathererName)
	return facts, nil
}

func mountInfoResultToFactValue(result MountInfoResult) (entities.FactValue, error) {
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
