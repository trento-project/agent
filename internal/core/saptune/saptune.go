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
	Version         string
	IsJSONSupported bool
	executor        utils.CommandExecutor
}

func getSaptuneVersion(commandExecutor utils.CommandExecutor) (string, error) {
	log.Info("Requesting Saptune version...")
	versionOutput, err := commandExecutor.Exec("rpm", "-q", "--qf", "%{VERSION}", "saptune")
	if err != nil {
		return "", errors.Wrap(err, ErrSaptuneVersionUnknown.Error())
	}

	log.Infof("saptune version output: %s", string(versionOutput))

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

	return Saptune{Version: saptuneVersion, executor: commandExecutor, IsJSONSupported: isSaptuneVersionSupported(saptuneVersion)}, nil
}

func (s *Saptune) RunCommand(args ...string) ([]byte, error) {
	if !s.IsJSONSupported {
		return nil, ErrUnsupportedSaptuneVer
	}

	log.Infof("Running saptune command: saptune %v", args)
	output, err := s.executor.Exec("saptune", args...)
	if err != nil {
		log.Debugf(err.Error())
	}
	log.Debugf("saptune output: %s", string(output))
	log.Infof("Saptune command executed")

	return output, nil
}
