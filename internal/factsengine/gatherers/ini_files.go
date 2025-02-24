package gatherers

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/trento-project/agent/internal/core/sapsystem"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"gopkg.in/ini.v1"
)

const (
	IniFilesGathererName = "ini_files"
)

// nolint:gochecknoglobals
var (
	IniFilesError = entities.FactGatheringError{
		Type:    "ini-files-error",
		Message: "ini file error",
	}

	IniFilesNotFoundError = entities.FactGatheringError{
		Type:    "ini-files-not-found-error",
		Message: "cannot find ini file",
	}

	IniFilesParseError = entities.FactGatheringError{
		Type:    "ini-files-parse-error",
		Message: "cannot parse ini file",
	}

	IniFilesFormatError = entities.FactGatheringError{
		Type:    "ini-files-format-error",
		Message: "cannot format ini file content to fact value",
	}
)

type IniFilesGatherer struct {
	fs afero.Fs
}

func NewIniFilesGatherer(fs afero.Fs) *IniFilesGatherer {
	return &IniFilesGatherer{fs: fs}
}

func NewDefaultIniFilesGatherer() *IniFilesGatherer {
	return &IniFilesGatherer{fs: afero.NewOsFs()}
}

func (g *IniFilesGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	log.Infof("Starting %s facts gathering process", IniFilesGathererName)
	facts := []entities.Fact{}

	for _, factReq := range factsRequests {

		switch factReq.Argument {
		case "global.ini":
			fact, err := g.gatherGlobalIni(ctx, factReq)
			if err != nil {
				return nil, errors.Wrap(err, "error gathering global.ini")
			}
			facts = append(facts, fact)
		default:
			return nil, fmt.Errorf("unsupported ini file for request %s, file: %s", factReq.Name, factReq.Argument)
		}

	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return facts, nil
}

func (g *IniFilesGatherer) gatherGlobalIni(_ context.Context, factRequest entities.FactRequest) (entities.Fact, error) {
	sids, err := findSIDs(g.fs)
	if err != nil {
		return entities.Fact{}, err
	}
	if len(sids) == 0 {
		return entities.Fact{}, errors.New("no SAP system found")
	}

	values := entities.FactValueList{}

	for _, sid := range sids {
		path := globalIniPath(sid)

		content, err := afero.ReadFile(g.fs, path)
		if err != nil {
			return entities.NewFactGatheredWithError(factRequest, IniFilesNotFoundError.Wrap(err.Error())), nil
		}

		parsed, err := parseIni(content)
		if err != nil {
			return entities.NewFactGatheredWithError(factRequest, IniFilesParseError.Wrap(err.Error())), nil
		}

		value, err := entities.NewFactValue(map[string]interface{}{
			"sid":     sid,
			"content": parsed,
		},
			entities.WithStringConversion())
		if err != nil {
			return entities.NewFactGatheredWithError(factRequest, IniFilesFormatError.Wrap(err.Error())), nil
		}

		values.AppendValue(value)

	}

	return entities.NewFactGatheredWithRequest(factRequest, &entities.FactValueList{Value: values.Value}), nil

}

func globalIniPath(sid string) string {
	return fmt.Sprintf("/usr/sap/%s/SYS/global/hdb/custom/config/global.ini", sid)
}

func parseIni(content []byte) (map[string]interface{}, error) {

	cfg, err := ini.Load(content)
	if err != nil {
		return nil, errors.Wrap(err, "error loading ini file")
	}

	result := make(map[string]interface{})

	for _, section := range cfg.Sections() {
		sectionMap := make(map[string]interface{})
		for _, key := range section.Keys() {
			sectionMap[key.Name()] = key.String()
		}
		if section.Name() == ini.DefaultSection {
			for k, v := range sectionMap {
				result[k] = v
			}
		} else {
			result[section.Name()] = sectionMap
		}
	}

	return result, nil
}

func findSIDs(fs afero.Fs) ([]string, error) {
	sids := []string{}

	systemPaths, err := sapsystem.FindSystems(fs)
	if err != nil {
		return nil, SapProfilesFileSystemError.Wrap(err.Error())
	}

	for _, systemPath := range systemPaths {
		sids = append(sids, filepath.Base(systemPath))
	}

	return sids, nil
}
