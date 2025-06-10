package saptune

import (
	"log/slog"

	"github.com/pkg/errors"
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
	slog.Info("Requesting Saptune version...")
	versionOutput, err := commandExecutor.Exec("rpm", "-q", "--qf", "%{VERSION}", "saptune")
	if err != nil {
		return "", errors.Wrap(err, ErrSaptuneVersionUnknown.Error())
	}

	slog.Info("saptune version output", "output", string(versionOutput))

	return string(versionOutput), nil
}

func isSaptuneVersionSupported(version string) bool {
	compareOutput := semver.Compare(MinimalSaptuneVersion, "v"+version)

	return compareOutput != 1
}

func NewSaptune(commandExecutor utils.CommandExecutor) (Saptune, error) {
	saptuneVersion, err := getSaptuneVersion(commandExecutor)
	if err != nil {
		return Saptune{}, err
	}

	saptune := Saptune{
		Version:         saptuneVersion,
		executor:        commandExecutor,
		IsJSONSupported: isSaptuneVersionSupported(saptuneVersion),
	}

	return saptune, nil
}

func (s *Saptune) RunCommand(args ...string) ([]byte, error) {
	slog.Info("Running saptune command", "args", args)
	output, err := s.executor.Exec("saptune", args...)
	if err != nil {
		slog.Debug("error executing saptune command", "error", err)
	}
	slog.Debug("saptune output", "output", string(output))
	slog.Info("Saptune command executed")

	return output, nil
}

func (s *Saptune) RunCommandJSON(args ...string) ([]byte, error) {
	if !s.IsJSONSupported {
		return nil, ErrUnsupportedSaptuneVer
	}

	prependedArgs := append([]string{"--format", "json"}, args...)
	return s.RunCommand(prependedArgs...)
}
