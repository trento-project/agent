package gatherers

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/trento-project/agent/internal/core/sapsystem"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	DispWorkGathererName = "disp+work"
)

// nolint:gochecknoglobals
var (
	DispWorkFileSystemError = entities.FactGatheringError{
		Type:    "dispwork-file-system-error",
		Message: "error reading the file system",
	}

	DispWorkCommandError = entities.FactGatheringError{
		Type:    "dispwork-command-error",
		Message: "error running disp+work command",
	}

	DispWorkDecodingError = entities.FactGatheringError{
		Type:    "dispwork-decoding-error",
		Message: "error decoding disp+work output",
	}

	// the names groups values are the values used to compose the resulting fact value map
	entriesPatternCompiled = regexp.MustCompile("(?m)" +
		"^kernel release\\s+(?P<kernel_release>.*)$|" +
		"^compilation mode\\s+(?P<compilation_mode>.*)$|" +
		"^patch number\\s+(?P<patch_number>.*)$")

	groupedNames = entriesPatternCompiled.SubexpNames()[1:]
)

type DispWorkGatherer struct {
	fs       afero.Fs
	executor utils.CommandExecutor
}

type dispWorkData struct {
	CompilationMode string `json:"compilation_mode"`
	KernelRelease   string `json:"kernel_release"`
	PatchNumber     string `json:"patch_number"`
}

func NewDefaultDispWorkGatherer() *DispWorkGatherer {
	return NewDispWorkGatherer(afero.NewOsFs(), utils.Executor{})
}

func NewDispWorkGatherer(fs afero.Fs, executor utils.CommandExecutor) *DispWorkGatherer {
	return &DispWorkGatherer{
		fs:       fs,
		executor: executor,
	}
}

func (g *DispWorkGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting %s facts gathering process", DispWorkGathererName)

	systemPaths, err := sapsystem.FindSystems(g.fs)
	if err != nil {
		return nil, DispWorkFileSystemError.Wrap(err.Error())
	}

	dispWorkMap := make(map[string]dispWorkData)

	for _, systemPath := range systemPaths {
		sid := filepath.Base(systemPath)
		sapUser := fmt.Sprintf("%sadm", strings.ToLower(sid))

		dispWorkOutput, err := g.executor.Exec("su", "-", sapUser, "-c", "\"disp+work\"")
		if err != nil {
			gatheringError := DispWorkCommandError.Wrap(err.Error())
			log.Error(gatheringError)
			dispWorkMap[sid] = dispWorkData{} // fill with empty data
			continue
		}

		result := fillRegexpGroups(string(dispWorkOutput))

		dispWorkMap[sid] = dispWorkData{
			CompilationMode: result["compilation_mode"],
			KernelRelease:   result["kernel_release"],
			PatchNumber:     result["patch_number"],
		}
	}

	factValue, err := dispWorkDataToFactValue(dispWorkMap)
	if err != nil {
		gatheringError := DispWorkDecodingError.Wrap(err.Error())
		log.Error(gatheringError)
		return nil, gatheringError
	}

	for _, factReq := range factsRequests {
		facts = append(facts, entities.NewFactGatheredWithRequest(factReq, factValue))
	}

	log.Infof("Requested %s facts gathered", DispWorkGathererName)
	return facts, nil
}

func fillRegexpGroups(output string) map[string]string {
	result := make(map[string]string)
	for _, match := range entriesPatternCompiled.FindAllStringSubmatch(output, -1) {
		for i, name := range groupedNames {
			if value, found := result[name]; found && value != "" {
				continue
			}
			result[name] = match[i+1]
		}
	}
	return result
}

func dispWorkDataToFactValue(data map[string]dispWorkData) (entities.FactValue, error) {
	marshalled, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}

	var unmarshalled map[string]interface{}
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return nil, err
	}

	return entities.NewFactValue(unmarshalled)
}
