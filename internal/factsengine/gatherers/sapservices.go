package gatherers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/trento-project/agent/pkg/factsengine/entities"
)

type SapServicesStartupKind string

const (
	SapServicesSystemdStartup  SapServicesStartupKind = "systemctl"
	SapServicesSapstartStartup SapServicesStartupKind = "sapstartsrv"
	sapServicesDefaultPath                            = "/usr/sap/sapservices"
	SapServicesGathererName                           = "sapservices"
)

// nolint:gochecknoglobals
var (
	SapServicesParsingError = entities.FactGatheringError{
		Type:    "sap-services-parsing-error",
		Message: "error parsing the sap services file",
	}
	SapServicesFileError = entities.FactGatheringError{
		Type:    "sap-services-parsing-error",
		Message: "error reading the sap services file",
	}
	SapstartSIDExtractionPattern = regexp.MustCompile(`(?s)pf=([^[:space:]]+)/(.*?)_.*_.*`)
	SystemdSIDExtractionPattern  = regexp.MustCompile(`(?s)start SAP(.*?)_.*`)
)

type SapServicesEntry struct {
	SID     string                 `json:"sid"`
	Kind    SapServicesStartupKind `json:"kind"`
	Content string                 `json:"content"`
}

func systemdStartup(sapServicesContent string) bool {
	return strings.Contains(sapServicesContent, "systemctl")
}

func sapstartStartup(sapServicesContent string) bool {
	return strings.Contains(sapServicesContent, "sapstartsrv")
}

func extractSIDFromSystemdService(sapServicesContent string) string {
	matches := SystemdSIDExtractionPattern.FindStringSubmatch(sapServicesContent)
	if len(matches) != 2 {
		return ""
	}
	return matches[1]
}

func extractSIDFromSapstartService(sapServicesContent string) string {
	matches := SapstartSIDExtractionPattern.FindStringSubmatch(sapServicesContent)
	if len(matches) != 3 {
		return ""
	}
	return matches[2]
}

type SapServices struct {
	fs               afero.Fs
	servicesFilePath string
}

func NewSapServicesGatherer(servicesFilePath string, fs afero.Fs) *SapServices {
	return &SapServices{
		servicesFilePath: servicesFilePath,
		fs:               fs,
	}
}

func NewDefaultSapServicesGatherer() *SapServices {
	return &SapServices{servicesFilePath: sapServicesDefaultPath, fs: afero.NewOsFs()}
}

func (s *SapServices) Gather(factsRequests []entities.FactRequest) ([]entities.Fact, error) {
	facts := []entities.Fact{}
	log.Infof("Starting %s facts gathering process", SapServicesGathererName)

	entries, err := s.getSapServicesFileEntries()
	if err != nil {
		return nil, err
	}

	factValues, err := convertSapServicesEntriesToFactValue(entries)
	if err != nil {
		return nil, SapServicesParsingError.Wrap(err.Error())
	}

	for _, requestedFact := range factsRequests {
		fact := entities.NewFactGatheredWithRequest(requestedFact, factValues)

		facts = append(facts, fact)
	}

	log.Infof("Requested %s facts gathered", SapServicesGathererName)
	return facts, nil
}

func (s *SapServices) getSapServicesFileEntries() ([]SapServicesEntry, error) {
	f, err := s.fs.Open(s.servicesFilePath)
	if err != nil {
		return nil, SapServicesFileError.Wrap(err.Error())
	}

	defer func() {
		err := f.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	fileScanner := bufio.NewScanner(f)
	fileScanner.Split(bufio.ScanLines)

	var entries []SapServicesEntry

	for fileScanner.Scan() {
		scannedLine := fileScanner.Text()
		if strings.HasPrefix(scannedLine, "#") || strings.HasPrefix(scannedLine, "//") || scannedLine == "" {
			continue
		}

		var kind SapServicesStartupKind
		var sid string

		if systemdStartup(scannedLine) {
			kind = SapServicesSystemdStartup
			extractedSID := extractSIDFromSystemdService(scannedLine)
			if extractedSID == "" {
				return nil, SapServicesParsingError.Wrap(
					fmt.Sprintf("could not extract sid from systemd sap services entry: %s", scannedLine),
				)
			}
			sid = extractedSID
		}

		if sapstartStartup(scannedLine) {
			kind = SapServicesSapstartStartup
			extractedSID := extractSIDFromSapstartService(scannedLine)
			if extractedSID == "" {
				return nil, SapServicesParsingError.Wrap(
					fmt.Sprintf("could not extract sid from sapstartsrv sap services entry: %s", scannedLine),
				)
			}
			sid = extractedSID
		}

		if kind == "" {
			// the line is not a recognized entry
			continue
		}

		entry := SapServicesEntry{
			SID:     sid,
			Kind:    kind,
			Content: scannedLine,
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

func convertSapServicesEntriesToFactValue(entries []SapServicesEntry) (entities.FactValue, error) {
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
