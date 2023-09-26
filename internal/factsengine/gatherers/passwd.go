package gatherers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	PasswdGathererName = "passwd"
	PasswdFilePath     = "/etc/passwd"
)

// nolint:gochecknoglobals
var (
	PasswdFileError = entities.FactGatheringError{
		Type:    "passwd-file-error",
		Message: "error reading /etc/passwd file",
	}

	PasswdDecodingError = entities.FactGatheringError{
		Type:    "passwd-decoding-error",
		Message: "error decoding file content",
	}
)

// An Entry contains all the fields for a specific user
type Entry struct {
	User        string `json:"user"`
	UID         string `json:"uid"`
	GID         string `json:"gid"`
	Description string `json:"description"`
	Home        string `json:"home"`
	Shell       string `json:"shell"`
}

type PasswdGatherer struct {
	passwdFilePath string
}

func NewDefaultPasswdGatherer() *PasswdGatherer {
	return NewPasswdGatherer(PasswdFilePath)
}

func NewPasswdGatherer(path string) *PasswdGatherer {
	return &PasswdGatherer{
		passwdFilePath: path,
	}
}

func (g *PasswdGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting %s state facts gathering process", PasswdGathererName)

	entries, err := parsePasswdFile(g.passwdFilePath)
	if err != nil {
		return nil, PasswdFileError.Wrap(err.Error())
	}

	factValues, err := convertEntriesToFactValue(entries)
	if err != nil {
		return nil, PasswdDecodingError.Wrap(err.Error())
	}

	for _, requestedFact := range factsRequests {
		fact := entities.NewFactGatheredWithRequest(requestedFact, factValues)

		facts = append(facts, fact)
	}

	log.Infof("Requested %s state facts gathered", PasswdGathererName)
	return facts, nil
}

func parsePasswdFile(filePath string) ([]Entry, error) {
	entries := []Entry{}

	passwdFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	defer func() {
		err := passwdFile.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	fileScanner := bufio.NewScanner(passwdFile)
	fileScanner.Split(bufio.ScanLines)
	var fileLines []string

	for fileScanner.Scan() {
		scannedLine := fileScanner.Text()
		fileLines = append(fileLines, scannedLine)
	}

	for index, line := range fileLines {
		values := strings.Split(line, ":")
		if len(values) != 7 {
			return nil, fmt.Errorf("invalid passwd file: line %d entry does not have 7 values", index+1)
		}
		newEntry := Entry{
			User:        values[0],
			UID:         values[2],
			GID:         values[3],
			Description: values[4],
			Home:        values[5],
			Shell:       values[6],
		}

		entries = append(entries, newEntry)
	}

	return entries, nil
}

func convertEntriesToFactValue(entries []Entry) (entities.FactValue, error) {
	marshalled, err := json.Marshal(&entries)
	if err != nil {
		return nil, err
	}

	var unmarshalled []interface{}
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return nil, err
	}

	return entities.NewFactValue(unmarshalled)
}
