package saptune

import (
	"encoding/json"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/utils"
	"golang.org/x/mod/semver"
)

var (
	ErrSaptuneVersionUnknown = errors.New("could not determine saptune version")
	ErrUnsupportedSaptuneVer = errors.New("saptune version is not supported")
)

const (
	MinimalSaptuneVersion = "v3.1.0"
)

type Saptune struct {
	Version		*string				`json:"version"`
	Commands	[]SaptuneOutput		`json:"commands"`
}
type SaptuneOutput struct {
	Schema      string    `json:"$schema"`
	PublishTime string    `json:"publish time"`
	Argv        string    `json:"argv"`
	Pid         int       `json:"pid"`
	Command     string    `json:"command"`
	ExitCode    int       `json:"exit code"`
	Result      Result    `json:"result"`
	Messages    []Message `json:"messages"`
}

type Result struct {
	Services                 Services `json:"services"`
	SystemdSystemState       string   `json:"systemd system state"`
	TuningState              string   `json:"tuning state"`
	Virtualization           string   `json:"virtualization"`
	ConfiguredVersion        string   `json:"configured version"`
	PackageVersion           string   `json:"package version"`
	SolutionEnabled          []string `json:"Solution enabled"`
	NotesEnabledBySolution   []string `json:"Notes enabled by Solution"`
	SolutionApplied          []string `json:"Solution applied"`
	NotesAppliedBySolution   []string `json:"Notes applied by Solution"`
	NotesEnabledAdditionally []string `json:"Notes enabled additionally"`
	NotesEnabled             []string `json:"Notes enabled"`
	NotesApplied             []string `json:"Notes applied"`
	Staging                  Staging  `json:"staging"`
	RememberMessage          string   `json:"remember message"`
}

type Services struct {
	Saptune []string `json:"saptune"`
	Sapconf []string `json:"sapconf"`
	Tuned   []string `json:"tuned"`
}

type Staging struct {
	StagingEnabled  bool     `json:"staging enabled"`
	NotesStaged     []string `json:"Notes staged"`
	SolutionsStaged []string `json:"Solutions staged"`
}

type Message struct {
	Priority string `json:"priority"`
	Message  string `json:"message"`
}

func getSaptuneVersion(commandExecutor utils.CommandExecutor) (string, error) {
	log.Info("Requesting Saptune version...")
	versionOutput, err := commandExecutor.Exec("rpm", "-q", "--qf", "%{VERSION}", "saptune")
	if err != nil {
		return "", ErrSaptuneVersionUnknown
	}
	
	log.Debugf("saptune version output: %s", string(versionOutput))

	return string(versionOutput), nil
}

func isSaptuneVersionSupported(commandExecutor utils.CommandExecutor, version string) bool {
	
	compareOutput := semver.Compare(MinimalSaptuneVersion, "v" + version)
	
	return compareOutput != 1
}

func NewSaptune(commandExecutor utils.CommandExecutor) (Saptune, error) {
	saptuneVersion, err := getSaptuneVersion(commandExecutor)
	if err != nil {
		return Saptune{Version: nil, Commands: []SaptuneOutput{}}, err
	}

	if !isSaptuneVersionSupported(commandExecutor, saptuneVersion) {
		return Saptune{Version: &saptuneVersion, Commands: []SaptuneOutput{}}, ErrUnsupportedSaptuneVer
	}

	log.Info("Requesting Saptune status...")
	output, err := commandExecutor.Exec("saptune", "--format", "json", "status")
	if err != nil {
		return Saptune{Version: &saptuneVersion, Commands: []SaptuneOutput{}}, errors.Wrap(err, "unexpected error while calling saptune")
	}

	log.Debugf("saptune output: %s", string(output))

	var saptuneOutput SaptuneOutput
	err = json.Unmarshal(output, &saptuneOutput)
	if err != nil {
		return Saptune{Version: &saptuneVersion, Commands: []SaptuneOutput{}}, errors.Wrap(err, "unexpected error while parsing saptune output")
	}
	log.Infof("Saptune status discovered")

	return Saptune{Version: &saptuneVersion, Commands: []SaptuneOutput{
		saptuneOutput,
	}}, nil
}
