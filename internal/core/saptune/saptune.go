package saptune

import (
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
	Version  string `json:"version"`
	executor utils.CommandExecutor
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

func isSaptuneVersionSupported(version string) bool {
	compareOutput := semver.Compare(MinimalSaptuneVersion, "v"+version)

	return compareOutput != 1
}

func NewSaptune(commandExecutor utils.CommandExecutor) (Saptune, error) {
	saptuneVersion, err := getSaptuneVersion(commandExecutor)
	if err != nil {
		return Saptune{Version: ""}, err
	}

	return Saptune{Version: saptuneVersion, executor: commandExecutor}, nil
}

func (s *Saptune) RunCommand(args ...string) ([]byte, error) {
	if !isSaptuneVersionSupported(s.Version) {
		return nil, ErrUnsupportedSaptuneVer
	}

	log.Info("Requesting Saptune status...")

	log.Infof("Saptune status discovered")
	log.Infof("Running saptune command: saptune %v", args)
	output, err := s.executor.Exec("saptune", args...)
	log.Debugf("saptune output: %s", string(output))
	if err != nil {
		return output, errors.Wrap(err, "non-zero return code while calling saptune")
	}

	log.Debugf("saptune output: %s", string(output))

	log.Infof("Saptune command executed")

	return output, nil
}
