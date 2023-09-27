package gatherers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

const (
	GroupsGathererName = "groups"
	GroupsFilePath     = "/etc/group"
)

// nolint:gochecknoglobals
var (
	GroupsFileError = entities.FactGatheringError{
		Type:    "groups-file-error",
		Message: "error reading /etc/group file",
	}

	GroupsFileDecodingError = entities.FactGatheringError{
		Type:    "groups-decoding-error",
		Message: "error deconding groups file",
	}
)

type GroupsEntry struct {
	Name  string   `json:"name"`
	ID    uint64   `json:"id"`
	Users []string `json:"users"`
}

type GroupsGatherer struct {
	groupsFilePath string
}

func NewDefaultGroupsGatherer() *GroupsGatherer {
	return NewGroupsGatherer(GroupsFilePath)
}

func NewGroupsGatherer(groupsFilePath string) *GroupsGatherer {
	return &GroupsGatherer{groupsFilePath: groupsFilePath}
}

func (g *GroupsGatherer) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}

	groupsFile, err := os.Open(g.groupsFilePath)
	if err != nil {
		return nil, GroupsFileError.Wrap(err.Error())
	}

	defer func() {
		err := groupsFile.Close()
		if err != nil {
			log.Errorf("could not close groups file %s, error: %s", g.groupsFilePath, err)
		}
	}()

	entries, err := parseGroupsFile(groupsFile)
	if err != nil {
		return nil, GroupsFileDecodingError.Wrap(err.Error())
	}

	factValues, err := mapGroupsEntriesToFactValue(entries)
	if err != nil {
		return nil, GroupsFileDecodingError.Wrap(err.Error())
	}

	for _, requestedFact := range factsRequests {
		facts = append(facts, entities.NewFactGatheredWithRequest(requestedFact, factValues))
	}

	log.Infof("requested %s facts gathered", GroupsGathererName)

	return facts, nil
}

func parseGroupsFile(fileContent io.Reader) ([]GroupsEntry, error) {
	lineScanner := bufio.NewScanner(fileContent)
	lineScanner.Split(bufio.ScanLines)

	var entries []GroupsEntry

	for lineScanner.Scan() {
		groupsLine := lineScanner.Text()

		values := strings.Split(groupsLine, ":")

		if len(values) != 4 {
			return nil, fmt.Errorf("could not decode groups file line %s, entry are less then 4", groupsLine)
		}

		groupID, err := strconv.Atoi(values[2])
		if err != nil {
			return nil, fmt.Errorf("could not convert group id %s to integer", values[2])
		}

		groupUsers := strings.Split(values[3], ",")
		if len(groupUsers) == 1 && groupUsers[0] == "" {
			// no groups found, set the slice to empty to avoid one item with empty string as user
			groupUsers = []string{}
		}

		entries = append(entries, GroupsEntry{
			Name:  values[0],
			ID:    uint64(groupID),
			Users: groupUsers,
		})
	}

	return entries, nil
}

func mapGroupsEntriesToFactValue(entries []GroupsEntry) (entities.FactValue, error) {
	marshalled, err := json.Marshal(&entries)
	if err != nil {
		return nil, err
	}

	var unmarshalled []interface{}
	err = json.Unmarshal(marshalled, &unmarshalled)
	if err != nil {
		return nil, err
	}

	return entities.NewFactValue(unmarshalled, entities.WithStringConversion())
}
