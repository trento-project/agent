package saptune

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/tidwall/gjson"
	"golang.org/x/mod/semver"

	"github.com/trento-project/agent/pkg/utils"
)

const minimalSaptuneVersion = "v3.1.0"

type Saptune interface {
	CheckVersionSupport(ctx context.Context) error
	GetVersion(ctx context.Context) (string, error)
	GetAppliedSolution(ctx context.Context) (string, error)
	Check(ctx context.Context) ([]byte, error)
	GetStatus(ctx context.Context, nonComplianceCheck bool) ([]byte, error)
	ApplySolution(ctx context.Context, solution string) error
	ChangeSolution(ctx context.Context, solution string) error
	RevertSolution(ctx context.Context, solution string) error
	ListSolution(ctx context.Context) ([]byte, error)
	VerifySolution(ctx context.Context) ([]byte, error)
	ListNote(ctx context.Context) ([]byte, error)
	VerifyNote(ctx context.Context) ([]byte, error)
}

type saptuneClient struct {
	executor utils.CommandExecutor
	logger   *slog.Logger
}

func NewSaptuneClient(
	executor utils.CommandExecutor,
	logger *slog.Logger,
) Saptune {
	return &saptuneClient{
		executor: executor,
		logger:   logger,
	}
}

func IsJSONSupported(version string) bool {
	compareOutput := semver.Compare(minimalSaptuneVersion, "v"+strings.TrimSpace(version))

	return compareOutput != 1
}

func (s *saptuneClient) CheckVersionSupport(ctx context.Context) error {
	version, err := s.GetVersion(ctx)
	if err != nil {
		return err
	}

	if supported := IsJSONSupported(version); !supported {
		return fmt.Errorf(
			"saptune version not supported, installed: %s, minimum supported: %s",
			version,
			minimalSaptuneVersion,
		)
	}

	return nil
}

func (s *saptuneClient) GetVersion(ctx context.Context) (string, error) {
	versionOutput, err := s.executor.CombinedOutputContext(
		ctx, "rpm", "-q", "--qf", "%{VERSION}", "saptune")
	if err != nil {
		return "", fmt.Errorf(
			"could not get the installed saptune version: %w",
			err,
		)
	}

	version := string(versionOutput)
	s.logger.Debug("installed saptune version", "version", version)

	return version, nil
}

func (s *saptuneClient) Check(ctx context.Context) ([]byte, error) {
	return s.runSaptuneJSON(ctx, "check")
}

func (s *saptuneClient) GetAppliedSolution(ctx context.Context) (string, error) {
	solutionAppliedOutput, err := s.runSaptuneJSON(ctx, "solution", "applied")
	if err != nil {
		return "", err
	}
	return gjson.GetBytes(solutionAppliedOutput, "result.Solution applied.0.Solution ID").String(), nil
}

func (s *saptuneClient) GetStatus(ctx context.Context, nonComplianceCheck bool) ([]byte, error) {
	args := []string{"status"}
	if nonComplianceCheck {
		args = append(args, "--non-compliance-check")
	}

	return s.runSaptuneJSON(ctx, args...)
}

func (s *saptuneClient) ApplySolution(ctx context.Context, solution string) error {
	_, err := s.runSaptune(ctx, "solution", "apply", solution)
	return err
}

func (s *saptuneClient) ChangeSolution(ctx context.Context, solution string) error {
	_, err := s.runSaptune(ctx, "solution", "change", "--force", solution)
	return err
}

func (s *saptuneClient) RevertSolution(ctx context.Context, solution string) error {
	_, err := s.runSaptune(ctx, "solution", "revert", solution)
	return err
}

func (s *saptuneClient) ListSolution(ctx context.Context) ([]byte, error) {
	return s.runSaptuneJSON(ctx, "solution", "list")
}

func (s *saptuneClient) VerifySolution(ctx context.Context) ([]byte, error) {
	return s.runSaptuneJSON(ctx, "solution", "verify")
}

func (s *saptuneClient) ListNote(ctx context.Context) ([]byte, error) {
	return s.runSaptuneJSON(ctx, "note", "list")
}

func (s *saptuneClient) VerifyNote(ctx context.Context) ([]byte, error) {
	return s.runSaptuneJSON(ctx, "note", "verify")
}

func (s *saptuneClient) runSaptune(ctx context.Context, args ...string) ([]byte, error) {
	slog.Info("Running saptune command", "args", args)
	output, err := s.executor.CombinedOutputContext(ctx, "saptune", args...)
	if err != nil {
		slog.Error("error executing saptune command", "args", args, "error", err)
		return output, fmt.Errorf("error executing saptune command: %w", err)
	}
	slog.Debug("saptune output", "output", string(output))
	slog.Info("Saptune command executed")

	return output, nil
}

func (s *saptuneClient) runSaptuneJSON(ctx context.Context, args ...string) ([]byte, error) {
	prependedArgs := append([]string{"--format", "json"}, args...)
	return s.runSaptune(ctx, prependedArgs...)
}
