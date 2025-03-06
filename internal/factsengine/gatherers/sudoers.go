package gatherers

import (
	"bufio"
	"context"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/core/sapsystem"
	"github.com/trento-project/agent/pkg/factsengine/entities"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	SudoersGathererName = "sudoers"
)

// nolint:gochecknoglobals
var (
	SudoersReadError = entities.FactGatheringError{
		Type:    "sudoers-read-error",
		Message: "error reading sudoers",
	}

	SudoersParseError = entities.FactGatheringError{
		Type:    "sudoers parse error",
		Message: "error parsing sudoers",
	}

	SudoersUserError = entities.FactGatheringError{
		Type:    "sudoers-user-error",
		Message: "error finding sudoers users",
	}
)

type SudoersGatherer struct {
	executor utils.CommandExecutor
	fs       afero.Fs
}

type privilegeEntry struct {
	runAsUser  string
	runAsGroup string
	noPassword bool
	commands   []string
}

type parsedSudoers struct {
	CommandsAsRoot []privilegeEntry
	User           string
}

func (p parsedSudoers) AsInterface() interface{} {
	return map[string]interface{}{
		"user":       p.User,
		"privileges": p.CommandsAsRoot,
	}
}

func NewDefaultSudoersGatherer() *SudoersGatherer {
	return NewSudoersGatherer(utils.Executor{}, afero.NewOsFs())
}

func NewSudoersGatherer(executor utils.CommandExecutor, fs afero.Fs) *SudoersGatherer {
	return &SudoersGatherer{
		executor: executor,
		fs:       fs,
	}
}

func (g *SudoersGatherer) Gather(ctx context.Context, factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting %s facts gathering process", SudoersGathererName)

	for _, factReq := range factsRequests {
		var fact entities.Fact
		var data []parsedSudoers

		if factReq.Argument == "" {
			result, err := g.gatherAll(ctx)
			if err != nil {
				return nil, err
			}
			data = result
		} else {
			result, err := g.gatherSingle(ctx, factReq.Argument)
			if err != nil {
				return nil, err
			}
			data = []parsedSudoers{result}
		}

		value, err := toFactValue(data)
		if err != nil {
			return nil, err
		}
		fact = entities.NewFactGatheredWithRequest(factReq, value)
		facts = append(facts, fact)
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	log.Infof("Requested %s facts gathered", SudoersGathererName)
	return facts, nil
}

func (g *SudoersGatherer) gatherAll(ctx context.Context) ([]parsedSudoers, *entities.FactGatheringError) {
	usernames, err := findUsernames(g.fs)
	if err != nil {
		return nil, SudoersUserError.Wrap(err.Error())
	}
	allUsers := []parsedSudoers{}
	for _, username := range usernames {
		single, err := g.gatherSingle(ctx, username)
		if err != nil {
			return nil, err
		}
		allUsers = append(allUsers, single)
	}
	return allUsers, nil
}

func (g *SudoersGatherer) gatherSingle(
	ctx context.Context,
	username string,
) (parsedSudoers, *entities.FactGatheringError) {
	output, err := g.readUserPrivileges(ctx, username)
	if err != nil {
		return parsedSudoers{}, SudoersReadError.Wrap(err.Error())
	}

	privileges, err := g.parseUserPrivileges(string(output))
	if err != nil {
		return parsedSudoers{}, SudoersParseError.Wrap(err.Error())
	}

	return parsedSudoers{
		CommandsAsRoot: privileges,
		User:           username,
	}, nil

}

func (g *SudoersGatherer) readUserPrivileges(ctx context.Context, username string) ([]byte, error) {
	err := validateUsername(username)
	if err != nil {
		return nil, errors.Wrap(err, "invalid username "+username)
	}
	output, err := g.executor.ExecContext(ctx, "/usr/bin/sudo", "-l", "-U", username)
	if err != nil {
		return nil, errors.Wrap(err, "error running sudo command")
	}
	return output, nil
}

func (g *SudoersGatherer) parseUserPrivileges(output string) ([]privilegeEntry, error) {
	var privileges []privilegeEntry
	scanner := bufio.NewScanner(strings.NewReader(output))
	var inPrivilegesSection bool

	// This regex detects the header that indicates the start of the command privileges.
	privilegesStartRegex := regexp.MustCompile(`^User\s+\S+\s+may run the following commands on`)
	// This regex parses each privilege entry line.
	// Group 1 captures the run-as specifier inside parentheses.
	// Group 2 optionally captures "NOPASSWD:" or "PASSWD:".
	// Group 3 captures the commands and any arguments.
	entryRegex := regexp.MustCompile(`^\s*\(([^)]+)\)\s*(?:(NOPASSWD:|PASSWD:)\s*)?(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for the header line to start processing privilege entries.
		if privilegesStartRegex.MatchString(line) {
			inPrivilegesSection = true
			continue // Skip the header line itself.
		}

		if inPrivilegesSection {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}

			// Attempt to extract all expected parts from the line.
			matches := entryRegex.FindStringSubmatch(line)
			if len(matches) == 0 {
				// If the line does not match, it might be a continuation of the previous entry.
				// Advanced parsing would be required to handle multi-line entries.
				continue
			}

			// Matches:
			// matches[1]: run-as specifier (e.g. "ALL : ALL" or "root")
			// matches[2]: optional password flag (e.g. "NOPASSWD:")
			// matches[3]: the commands string
			runas := matches[1]
			flag := matches[2]
			commandsStr := matches[3]

			var runAsUser, runAsGroup string
			parts := strings.Split(runas, ":")
			runAsUser = strings.TrimSpace(parts[0])
			if len(parts) > 1 {
				runAsGroup = strings.TrimSpace(parts[1])
			}

			noPassword := false
			if flag == "NOPASSWD:" {
				noPassword = true
			}

			// Commands might be a comma-separated list. Split and trim each command.
			rawCommands := strings.Split(commandsStr, ",")
			var commands []string
			for _, cmd := range rawCommands {
				trimmedCmd := strings.TrimSpace(cmd)
				if trimmedCmd != "" {
					commands = append(commands, trimmedCmd)
				}
			}

			entry := privilegeEntry{
				runAsUser:  runAsUser,
				runAsGroup: runAsGroup,
				noPassword: noPassword,
				commands:   commands,
			}
			privileges = append(privileges, entry)
		}
	}

	if !inPrivilegesSection {
		return nil, errors.New("failed to parse all privilege entries")
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return privileges, nil
}

func validateUsername(username string) error {
	// the validation fails if the username contains special characters
	if strings.ContainsAny(username, " !\"#$%&'()*+,./:;<=>?@[\\]^`{|}~") {
		return errors.New("username contains special characters")
	}
	return nil
}

func sidToUsername(sid string) string {
	return strings.ToLower(sid) + "adm"
}

func findUsernames(fs afero.Fs) ([]string, error) {
	usernames := []string{}

	systemPaths, err := sapsystem.FindSystems(fs)
	if err != nil {
		return nil, err
	}

	for _, systemPath := range systemPaths {
		username := sidToUsername(filepath.Base(systemPath))
		usernames = append(usernames, username)
	}

	return usernames, nil
}

func toFactValue(allUsers []parsedSudoers) (entities.FactValue, error) {
	values := make([]interface{}, 0, len(allUsers))
	for _, data := range allUsers {
		for _, commandEntry := range data.CommandsAsRoot {
			for _, command := range commandEntry.commands {
				value := make(map[string]interface{})
				value["command"] = command
				value["no_password"] = commandEntry.noPassword
				value["run_as_user"] = commandEntry.runAsUser
				value["run_as_group"] = commandEntry.runAsGroup
				value["user"] = data.User
				values = append(values, value)
			}
		}
	}

	fact, err := entities.NewFactValue(values)
	if err != nil {
		return nil, errors.Wrap(err, "failed to format fact value")
	}
	return fact, nil
}
