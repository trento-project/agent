package sapsystem

import (
	"bufio"
	"context"
	"crypto/md5" //nolint:gosec
	"fmt"
	"io"
	"net"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"log/slog"

	"github.com/hashicorp/go-envparse"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"github.com/trento-project/agent/internal/core/sapsystem/sapcontrolapi"
	"github.com/trento-project/agent/pkg/utils"
)

type SystemType int

const (
	Unknown SystemType = iota
	Database
	Application
	DiagnosticsAgent
)

const (
	sapInstallationPath  string = "/usr/sap"
	sapMntPath           string = "/sapmnt"
	sapIdentifierPattern string = "^[A-Z][A-Z0-9]{2}$" // PRD, HA1, etc
	sapInstancePattern   string = "^[A-Z]+([0-9]{2})$" // HDB00, ASCS00, ERS10, etc
	sapProfilePattern    string = "^(DEFAULT\\.PFL|[^.]*)$"
	SapDefaultProfile    string = "DEFAULT.PFL"
	sappfparCmd          string = "sappfpar SAPSYSTEMNAME SAPGLOBALHOST SAPFQDN SAPDBHOST dbs/hdb/dbname dbs/hdb/schema rdisp/msp/msserv rdisp/msserv_internal name=%s" //nolint:lll
)

var (
	sapIdentifierPatternCompiled = regexp.MustCompile(sapIdentifierPattern)
	sapInstancePatternCompiled   = regexp.MustCompile(sapInstancePattern)
	sapProfilePatternCompiled    = regexp.MustCompile(sapProfilePattern)
)

type SAPSystemsList []*SAPSystem
type SAPSystemsMap map[string]*SAPSystem

// A SAPSystem in this context is a SAP installation under one SID.
// It will have application or database type, mutually exclusive
// The Id parameter is not yet implemented
type SAPSystem struct {
	ID        string `json:"Id"`
	SID       string
	Type      SystemType
	Profile   SAPProfile
	Instances []*SAPInstance
	// Only for Database type
	Databases []*DatabaseData
	// Only for Application type
	DBAddress string
}

type SAPProfile map[string]string

type DatabaseData struct {
	Database  string
	Container string
	User      string
	Group     string
	UserID    string `json:"UserId"`
	GroupID   string `json:"GroupId"`
	Host      string
	SQLPort   string `json:"SqlPort"`
	Active    string
}

func Md5sum(data string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(data))) //nolint:gosec
}

func NewDefaultSAPSystemsList(ctx context.Context) (SAPSystemsList, error) {
	return NewSAPSystemsList(
		ctx,
		afero.NewOsFs(),
		utils.Executor{},
		sapcontrolapi.WebServiceUnix{},
	)
}

func NewSAPSystemsList(
	ctx context.Context,
	fs afero.Fs,
	executor utils.CommandExecutor,
	webService sapcontrolapi.WebServiceConnector,
) (SAPSystemsList, error) {
	var systems = SAPSystemsList{}

	systemPaths, err := FindSystems(fs)
	if err != nil {
		return systems, errors.Wrap(err, "Error walking the path")
	}

	// Find systems
	for _, sysPath := range systemPaths {
		system, err := NewSAPSystem(ctx, fs, executor, webService, sysPath)
		if err != nil {
			slog.Error("Error discovering a SAP system", "error", err.Error())
			continue
		}
		systems = append(systems, system)
	}

	return systems, nil
}

func (sl SAPSystemsList) GetSIDsString() string {
	sidString := []string{}

	for _, system := range sl {
		sidString = append(sidString, system.SID)
	}

	return strings.Join(sidString, ",")
}

func NewSAPSystem(
	ctx context.Context,
	fs afero.Fs,
	executor utils.CommandExecutor,
	webService sapcontrolapi.WebServiceConnector,
	sysPath string,
) (*SAPSystem, error) {

	var systemType SystemType
	instances := []*SAPInstance{}

	sid := filepath.Base(sysPath)
	profilePath := getProfilePath(sysPath)
	profile, err := GetProfileData(fs, profilePath)
	if err != nil {
		slog.Error("Error getting SAP profile data", "error", err.Error())
		return nil, err
	}

	instPaths, err := FindInstances(fs, sysPath)
	if err != nil {
		slog.Error("Error finding SAP instances", "error", err.Error())
		return nil, err
	}

	for _, instPath := range instPaths {
		webService := webService.New(instPath[1])
		instance, err := NewSAPInstance(ctx, webService, executor, fs)
		if err != nil {
			slog.Error("Error discovering a SAP instance", "error", err.Error())
			continue
		}

		systemType = instance.Type
		instances = append(instances, instance)
	}

	systemID, err := detectSystemID(fs, executor, systemType, sid)
	if err != nil {
		slog.Error("Error detecting system ID", "error", err.Error())
		return nil, err
	}

	system := &SAPSystem{
		ID:        systemID,
		SID:       sid,
		Type:      systemType,
		Profile:   profile,
		Instances: instances,
		Databases: nil,
		DBAddress: "",
	}

	if systemType == Database {
		databaseList, err := GetDatabases(fs, sid)
		if err != nil {
			slog.Error("Error getting the database list", "error", err.Error())
		} else {
			system.Databases = databaseList
		}
	} else if systemType == Application {
		addr, err := system.GetDBAddress()
		if err != nil {
			slog.Error("Error getting the database address", "error", err.Error())
		} else {
			system.DBAddress = addr
		}
	}

	return system, nil
}

func (system *SAPSystem) GetDBAddress() (string, error) {
	sapdbhost, found := system.Profile["SAPDBHOST"]
	if !found {
		return "", fmt.Errorf("SAPDBHOST field not found in the SAP profile")
	}

	addrList, err := net.LookupIP(sapdbhost)
	if err != nil {
		return "", fmt.Errorf("could not resolve \"%s\" hostname", sapdbhost)
	}

	// Get 1st IPv4 address
	for _, addr := range addrList {
		addrStr := addr.String()
		ip := net.ParseIP(addrStr)
		if ip.To4() != nil {
			return addrStr, nil
		}
	}

	return "", fmt.Errorf("could not get any IPv4 address")
}

// FindSystems returns the installed SAP systems in the /usr/sap folder
// It returns the list of found SAP system paths
func FindSystems(fs afero.Fs) ([]string, error) {
	var systems = []string{}

	exists, _ := afero.DirExists(fs, sapInstallationPath)
	if !exists {
		slog.Info("SAP installation not found")
		return systems, nil
	}

	files, err := afero.ReadDir(fs, sapInstallationPath)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if sapIdentifierPatternCompiled.MatchString(f.Name()) {
			slog.Info("New SAP system installation found", "name", f.Name())
			systems = append(systems, path.Join(sapInstallationPath, f.Name()))
		}
	}

	return systems, nil
}

// FindInstances returns the installed SAP instances in the /usr/sap/${SID} folder
// It returns a list with [instanceName, instanceNumber] combination
func FindInstances(fs afero.Fs, sapPath string) ([][]string, error) {
	var instances = [][]string{}

	files, err := afero.ReadDir(fs, sapPath)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		for _, matches := range sapInstancePatternCompiled.FindAllStringSubmatch(f.Name(), -1) {
			slog.Info("New SAP instance installation found", "name", matches[0])
			instances = append(instances, matches)
		}
	}

	return instances, nil
}

// FindProfiles returns the latest profile file names in the /sapmnt/${SID}/folder folder
func FindProfiles(fs afero.Fs, sid string) ([]string, error) {
	var profiles = []string{}

	files, err := afero.ReadDir(fs, path.Join(sapMntPath, sid, "profile"))
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if sapProfilePatternCompiled.MatchString(f.Name()) {
			slog.Info("New SAP profile found", "name", f.Name())
			profiles = append(profiles, f.Name())
		}
	}

	return profiles, nil
}

// GetProfileData returns the SAP profile file content
func GetProfileData(fs afero.Fs, profilePath string) (map[string]string, error) {
	profileFile, err := fs.Open(profilePath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open profile file")
	}

	defer func() {
		err := profileFile.Close()
		if err != nil {
			slog.Error("Error closing profile file", "error", err.Error())
		}
	}()

	profile, err := envparse.Parse(profileFile)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse profile file")
	}

	return profile, nil
}

// The content type of the databases.lst looks like
// # DATABASE:CONTAINER:USER:GROUP:USERID:GROUPID:HOST:SQLPORT:ACTIVE
// PRD::::::hana02:30015:yes
// DEV::::::hana02:30044:yes
func GetDatabases(fs afero.Fs, sid string) ([]*DatabaseData, error) {
	databasesListPath := fmt.Sprintf(
		"/usr/sap/%s/SYS/global/hdb/mdc/databases.lst", sid)
	databasesListFile, err := fs.Open(databasesListPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open the databases list file")
	}

	defer databasesListFile.Close()

	databaseScanner := bufio.NewScanner(databasesListFile)
	databaseList := make([]*DatabaseData, 0)

	for databaseScanner.Scan() {
		line := databaseScanner.Text()
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		data := strings.Split(line, ":")
		if len(data) != 9 {
			continue
		}

		databaseEntry := &DatabaseData{
			Database:  data[0],
			Container: data[1],
			User:      data[2],
			Group:     data[3],
			UserID:    data[4],
			GroupID:   data[5],
			Host:      data[6],
			SQLPort:   data[7],
			Active:    data[8],
		}

		databaseList = append(databaseList, databaseEntry)
	}

	return databaseList, nil
}

func getProfilePath(sysPath string) string {
	return path.Join(sysPath, "SYS", "profile", SapDefaultProfile)
}

func detectSystemID(fs afero.Fs, executor utils.CommandExecutor, sType SystemType, sid string) (string, error) {
	switch sType {
	case Database:
		return getUniqueIDHana(fs, sid)
	case Application:
		return getUniqueIDApplication(executor, sid)
	case DiagnosticsAgent:
		return getUniqueIDDiagnostics(fs)
	case Unknown:
		fallthrough
	default:
		return "-", nil
	}
}

func getUniqueIDHana(fs afero.Fs, sid string) (string, error) {
	nameserverConfigPath := fmt.Sprintf(
		"/usr/sap/%s/SYS/global/hdb/custom/config/nameserver.ini", sid)
	nameserver, err := fs.Open(nameserverConfigPath)
	if err != nil {
		return "", errors.Wrap(err, "could not open the nameserver configuration file")
	}

	defer nameserver.Close()

	nameserverRaw, err := io.ReadAll(nameserver)

	if err != nil {
		return "", errors.Wrap(err, "could not read the nameserver configuration file")
	}

	configMap := utils.FindMatches(`([\w\/]+)\s=\s(.+)`, nameserverRaw)
	hanaID, found := configMap["id"]
	if !found {
		return "", fmt.Errorf("could not find the landscape id in the configuration file")
	}

	hanaIDMd5 := Md5sum(fmt.Sprintf("%v", hanaID))
	return hanaIDMd5, nil
}

func getUniqueIDApplication(executor utils.CommandExecutor, sid string) (string, error) {
	user := fmt.Sprintf("%sadm", strings.ToLower(sid))
	cmd := fmt.Sprintf(sappfparCmd, sid)
	sappfpar, err := executor.Exec("/usr/bin/su", "-lc", cmd, user)
	if err != nil {
		return "", fmt.Errorf("error running sappfpar command with sid %s", sid)
	}

	appIDMd5 := Md5sum(string(sappfpar))
	return appIDMd5, nil
}

func getUniqueIDDiagnostics(fs afero.Fs) (string, error) {
	machineIDBytes, err := afero.ReadFile(fs, "/etc/machine-id")

	if err != nil {
		return "", err
	}

	machineID := strings.TrimSpace(string(machineIDBytes))
	id := Md5sum(machineID)
	return id, nil
}
