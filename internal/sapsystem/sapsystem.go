package sapsystem

import (
	"bufio"
	"crypto/md5" //nolint:gosec
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/trento-project/agent/internal/sapsystem/sapcontrol"
	"github.com/trento-project/agent/internal/utils"
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
	sapIdentifierPattern string = "^[A-Z][A-Z0-9]{2}$" // PRD, HA1, etc
	sapInstancePattern   string = "^[A-Z]+([0-9]{2})$" // HDB00, ASCS00, ERS10, etc
	sapDefaultProfile    string = "DEFAULT.PFL"
	sappfparCmd          string = "sappfpar SAPSYSTEMNAME SAPGLOBALHOST SAPFQDN SAPDBHOST dbs/hdb/dbname dbs/hdb/schema rdisp/msp/msserv rdisp/msserv_internal name=%s" //nolint:lll
)

var databaseFeatures = regexp.MustCompile("HDB.*")
var applicationFeatures = regexp.MustCompile("MESSAGESERVER.*|ENQREP|ABAP.*")
var diagnosticsAgentFeatures = regexp.MustCompile("SMDAGENT")

type SAPSystemsList []*SAPSystem
type SAPSystemsMap map[string]*SAPSystem

// A SAPSystem in this context is a SAP installation under one SID.
// It will have application or database type, mutually exclusive
// The Id parameter is not yet implemented
type SAPSystem struct {
	ID        string         `mapstructure:"id,omitempty" json:"Id"`
	SID       string         `mapstructure:"sid,omitempty"`
	Type      SystemType     `mapstructure:"type,omitempty"`
	Profile   SAPProfile     `mapstructure:"profile,omitempty"`
	Instances []*SAPInstance `mapstructure:"instances,omitempty"`
	// Only for Database type
	Databases []*DatabaseData `mapstructure:"databases,omitempty"`
	// Only for Application type
	DBAddress string `mapstructure:"db_address,omitempty"`
}

// The value is interface{} as some of the entries in the SAP profiles files and commands
// are already using "/", so the result will be a map of strings/maps
type SAPProfile map[string]interface{}
type SystemReplication map[string]interface{}
type HostConfiguration map[string]interface{}
type HdbnsutilSRstate map[string]interface{}

type SAPInstance struct {
	Name       string      `mapstructure:"name,omitempty"`
	Type       SystemType  `mapstructure:"type,omitempty"`
	Host       string      `mapstructure:"host,omitempty"`
	SAPControl *SAPControl `mapstructure:"sapcontrol,omitempty"`
	// Only for Database type
	SystemReplication SystemReplication `mapstructure:"systemreplication,omitempty"`
	HostConfiguration HostConfiguration `mapstructure:"hostconfiguration,omitempty"`
	HdbnsutilSRstate  HdbnsutilSRstate  `mapstructure:"hdbnsutilsrstate,omitempty"`
}

type SAPControl struct {
	webService sapcontrol.WebService
	Processes  []*sapcontrol.OSProcess        `mapstructure:"processes,omitempty"`
	Instances  []*sapcontrol.SAPInstance      `mapstructure:"instances,omitempty"`
	Properties []*sapcontrol.InstanceProperty `mapstructure:"properties,omitempty"`
}

type DatabaseData struct {
	Database  string `mapstructure:"database,omitempty"`
	Container string `mapstructure:"container,omitempty"`
	User      string `mapstructure:"user,omitempty"`
	Group     string `mapstructure:"group,omitempty"`
	UserID    string `mapstructure:"userid,omitempty" json:"UserId"`
	GroupID   string `mapstructure:"groupid,omitempty" json:"GroupId"`
	Host      string `mapstructure:"host,omitempty"`
	SQLPort   string `mapstructure:"sqlport,omitempty" json:"SqlPort"`
	Active    string `mapstructure:"active,omitempty"`
}

// FIXME do proper DI and fix test isolation, no globals
var newWebService = func(instanceNumber string) sapcontrol.WebService { //nolint
	return sapcontrol.NewWebService(instanceNumber)
}

//go:generate mockery --name=CustomCommand

type CustomCommand func(name string, arg ...string) *exec.Cmd

// FIXME proper DI and test
var customExecCommand CustomCommand = exec.Command //nolint

func Md5sum(data string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(data))) //nolint:gosec
}

func NewSAPSystemsList() (SAPSystemsList, error) {
	var systems = SAPSystemsList{}

	appFS := afero.NewOsFs()
	systemPaths, err := findSystems(appFS)
	if err != nil {
		return systems, errors.Wrap(err, "Error walking the path")
	}

	// Find systems
	for _, sysPath := range systemPaths {
		system, err := NewSAPSystem(appFS, sysPath)
		if err != nil {
			log.Printf("Error discovering a SAP system: %s", err)
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

// FIXME Declarative initialization of SAPSystem
func NewSAPSystem(fs afero.Fs, sysPath string) (*SAPSystem, error) {
	system := &SAPSystem{ //nolint
		SID: sysPath[strings.LastIndex(sysPath, "/")+1:],
	}

	profilePath := getProfilePath(sysPath)
	profile, err := getProfileData(fs, profilePath)
	if err != nil {
		log.Print(err.Error())
		return system, err
	}
	system.Profile = profile

	instPaths, err := findInstances(fs, sysPath)
	if err != nil {
		log.Print(err.Error())
		return system, err
	}

	// Find instances
	for _, instPath := range instPaths {
		webService := newWebService(instPath[1])
		instance, err := NewSAPInstance(webService)
		if err != nil {
			log.Printf("Error discovering a SAP instance: %s", err)
			continue
		}

		system.Type = instance.Type
		system.Instances = append(system.Instances, instance)
	}

	// FIXME default type
	switch system.Type { //nolint
	case Database:
		databaseList, err := getDatabases(fs, system.SID)
		if err != nil {
			log.Printf("Error getting the database list: %s", err)
		} else {
			system.Databases = databaseList
		}
	case Application:
		addr, err := getDBAddress(system)
		if err != nil {
			log.Printf("Error getting the database address: %s", err)
		} else {
			system.DBAddress = addr
		}
	}

	system, err = setSystemID(fs, system)
	if err != nil {
		return system, err
	}

	return system, nil
}

// Find the installed SAP instances in the /usr/sap folder
// It returns a list of paths where SAP system is found
func findSystems(fs afero.Fs) ([]string, error) {
	var systems = []string{}

	exists, _ := afero.DirExists(fs, sapInstallationPath)
	if !exists {
		log.Print("SAP installation not found")
		return systems, nil
	}

	files, err := afero.ReadDir(fs, sapInstallationPath)
	if err != nil {
		return nil, err
	}

	reSAPIdentifier := regexp.MustCompile(sapIdentifierPattern)

	for _, f := range files {
		if reSAPIdentifier.MatchString(f.Name()) {
			log.Printf("New SAP system installation found: %s", f.Name())
			systems = append(systems, path.Join(sapInstallationPath, f.Name()))
		}
	}

	return systems, nil
}

// Find the installed SAP instances in the /usr/sap/${SID} folder
func findInstances(fs afero.Fs, sapPath string) ([][]string, error) {
	var instances = [][]string{}
	reSAPInstancer := regexp.MustCompile(sapInstancePattern)

	files, err := afero.ReadDir(fs, sapPath)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		for _, matches := range reSAPInstancer.FindAllStringSubmatch(f.Name(), -1) {
			log.Printf("New SAP instance installation found: %s", matches[0])
			instances = append(instances, matches)
		}
	}

	return instances, nil
}

func getProfilePath(sysPath string) string {
	return path.Join(sysPath, "SYS", "profile", sapDefaultProfile)
}

// Get SAP profile file content
func getProfileData(fs afero.Fs, profilePath string) (map[string]interface{}, error) {
	profile, err := fs.Open(profilePath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open profile file")
	}

	defer profile.Close()

	profileRaw, err := ioutil.ReadAll(profile)

	if err != nil {
		return nil, errors.Wrap(err, "could not read profile file")
	}

	configMap := utils.FindMatches(`([\w\/]+)\s=\s(.+)`, profileRaw)

	return configMap, nil
}

func getDBAddress(system *SAPSystem) (string, error) {
	sapdbhost, found := system.Profile["SAPDBHOST"]
	if !found {
		return "", fmt.Errorf("SAPDBHOST field not found in the SAP profile")
	}

	addrList, err := net.LookupIP(sapdbhost.(string))
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

func setSystemID(fs afero.Fs, system *SAPSystem) (*SAPSystem, error) {
	// Set system ID
	var err error
	var id string

	switch system.Type {
	case Database:
		id, err = getUniqueIDHana(fs, system.SID)
	case Application:
		id, err = getUniqueIDApplication(system.SID)
	case DiagnosticsAgent:
		id, err = getUniqueIDDiagnostics(fs)
	case Unknown:
		id = "-"
	}

	system.ID = id
	return system, err
}

func getUniqueIDHana(fs afero.Fs, sid string) (string, error) {
	nameserverConfigPath := fmt.Sprintf(
		"/usr/sap/%s/SYS/global/hdb/custom/config/nameserver.ini", sid)
	nameserver, err := fs.Open(nameserverConfigPath)
	if err != nil {
		return "", errors.Wrap(err, "could not open the nameserver configuration file")
	}

	defer nameserver.Close()

	nameserverRaw, err := ioutil.ReadAll(nameserver)

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

func getUniqueIDApplication(sid string) (string, error) {
	user := fmt.Sprintf("%sadm", strings.ToLower(sid))
	cmd := fmt.Sprintf(sappfparCmd, sid)
	sappfpar, err := customExecCommand("su", "-lc", cmd, user).Output()
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

// The content type of the databases.lst looks like
// # DATABASE:CONTAINER:USER:GROUP:USERID:GROUPID:HOST:SQLPORT:ACTIVE
// PRD::::::hana02:30015:yes
// DEV::::::hana02:30044:yes
func getDatabases(fs afero.Fs, sid string) ([]*DatabaseData, error) {
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

// FIXME declarative build of SapInstance
func NewSAPInstance(w sapcontrol.WebService) (*SAPInstance, error) {
	host, _ := os.Hostname()
	var sapInstance = &SAPInstance{ //nolint
		Host: host,
	}

	scontrol, err := NewSAPControl(w)
	if err != nil {
		return sapInstance, err
	}

	sapInstance.SAPControl = scontrol

	instanceName, err := sapInstance.SAPControl.findProperty("INSTANCE_NAME")
	if err != nil {
		return sapInstance, err
	}
	sapInstance.Name = instanceName

	instanceType, err := detectType(sapInstance.SAPControl)
	if err != nil {
		return sapInstance, err
	}
	sapInstance.Type = instanceType

	if sapInstance.Type == Database {
		sid, _ := sapInstance.SAPControl.findProperty("SAPSYSTEMNAME")
		sapInstance.SystemReplication = systemReplicationStatus(sid, sapInstance.Name)
		sapInstance.HostConfiguration = landscapeHostConfiguration(sid, sapInstance.Name)
		sapInstance.HdbnsutilSRstate = hdbnsutilSrstate(sid, sapInstance.Name)
	}

	return sapInstance, nil
}

func detectType(sapControl *SAPControl) (SystemType, error) {
	sapLocalhost, err := sapControl.findProperty("SAPLOCALHOST")
	if err != nil {
		return Unknown, err
	}

	for _, instance := range sapControl.Instances {
		if instance.Hostname == sapLocalhost {
			switch {
			case databaseFeatures.MatchString(instance.Features):
				return Database, nil
			case applicationFeatures.MatchString(instance.Features):
				return Application, nil
			case diagnosticsAgentFeatures.MatchString(instance.Features):
				return DiagnosticsAgent, nil
			default:
				return Unknown, nil
			}
		}
	}

	return Unknown, nil
}

func runPythonSupport(sid, instance, script string) map[string]interface{} {
	user := fmt.Sprintf("%sadm", strings.ToLower(sid))
	cmdPath := path.Join(sapInstallationPath, sid, instance, "exe/python_support", script)
	cmd := fmt.Sprintf("python %s --sapcontrol=1", cmdPath)
	// Even with a error return code, some data is available
	srData, _ := customExecCommand("su", "-lc", cmd, user).Output()

	dataMap := utils.FindMatches(`(\S+)=(.*)`, srData)

	return dataMap
}

func systemReplicationStatus(sid, instance string) map[string]interface{} {
	return runPythonSupport(sid, instance, "systemReplicationStatus.py")
}

func landscapeHostConfiguration(sid, instance string) map[string]interface{} {
	return runPythonSupport(sid, instance, "landscapeHostConfiguration.py")
}

func hdbnsutilSrstate(sid, instance string) map[string]interface{} {
	user := fmt.Sprintf("%sadm", strings.ToLower(sid))
	cmdPath := path.Join(sapInstallationPath, sid, instance, "exe", "hdbnsutil")
	cmd := fmt.Sprintf("%s -sr_state -sapcontrol=1", cmdPath)
	srData, _ := customExecCommand("su", "-lc", cmd, user).Output()
	dataMap := utils.FindMatches(`(.+)=(.*)`, srData)
	return dataMap
}

// FIXME declarative build of SAPControl
func NewSAPControl(w sapcontrol.WebService) (*SAPControl, error) {
	var scontrol = &SAPControl{ //nolint
		webService: w,
	}

	properties, err := scontrol.webService.GetInstanceProperties()
	if err != nil {
		return scontrol, errors.Wrap(err, "SAPControl web service error")
	}

	scontrol.Properties = append(scontrol.Properties, properties.Properties...)

	processes, err := scontrol.webService.GetProcessList()
	if err != nil {
		return scontrol, errors.Wrap(err, "SAPControl web service error")
	}

	scontrol.Processes = append(scontrol.Processes, processes.Processes...)

	instances, err := scontrol.webService.GetSystemInstanceList()
	if err != nil {
		return scontrol, errors.Wrap(err, "SAPControl web service error")
	}

	scontrol.Instances = append(scontrol.Instances, instances.Instances...)

	return scontrol, nil
}

func (s *SAPControl) findProperty(key string) (string, error) {
	for _, item := range s.Properties {
		if item.Property == key {
			return item.Value, nil
		}
	}

	return "", fmt.Errorf("Property %s not found", key)
}
