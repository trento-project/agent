package cluster

import (
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"io"
	"log/slog"
	"os"
	"strings"

	// These packages were originally imported from github.com/ClusterLabs/ha_cluster_exporter/collector/pacemaker
	// Now we mantain our own fork

	"github.com/trento-project/agent/internal/core/cloud"
	"github.com/trento-project/agent/internal/core/cluster/cib"
	"github.com/trento-project/agent/internal/core/cluster/crmmon"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	cibAdmPath             string = "/usr/sbin/cibadmin"
	crmmonAdmPath          string = "/usr/sbin/crm_mon"
	corosyncKeyPath        string = "/etc/corosync/authkey"
	clusterNameProperty    string = "cib-bootstrap-options-cluster-name"
	stonithResourceMissing string = "notconfigured"
	stonithAgent           string = "stonith:"
	sbdFencingAgentName    string = "external/sbd"
)

type DiscoveryTools struct {
	CibAdmPath      string
	CrmmonAdmPath   string
	CorosyncKeyPath string
	SBDPath         string
	SBDConfigPath   string
	CommandExecutor utils.CommandExecutor
}

type Cluster struct {
	Cib      cib.Root
	Crmmon   crmmon.Root
	SBD      SBD
	ID       string `json:"Id"`
	Name     string
	DC       bool
	Provider string
}

func Md5sumFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New() //nolint:gosec
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func NewCluster() (*Cluster, error) {
	return NewClusterWithDiscoveryTools(&DiscoveryTools{
		CibAdmPath:      cibAdmPath,
		CrmmonAdmPath:   crmmonAdmPath,
		CorosyncKeyPath: corosyncKeyPath,
		SBDPath:         SBDPath,
		SBDConfigPath:   SBDConfigPath,
		CommandExecutor: utils.Executor{},
	})
}

func NewClusterWithDiscoveryTools(discoveryTools *DiscoveryTools) (*Cluster, error) {
	return makeOnlineHostPayload(discoveryTools)
}

func makeOnlineHostPayload(discoveryTools *DiscoveryTools) (*Cluster, error) {
	cibParser := cib.NewCibAdminParser(discoveryTools.CibAdmPath)

	cibConfig, err := cibParser.Parse()
	if err != nil {
		return nil, err
	}

	var cluster = &Cluster{
		Cib:      cib.Root{},    //nolint
		Crmmon:   crmmon.Root{}, //nolint
		SBD:      SBD{},         //nolint
		ID:       "",
		Name:     clusterNameProperty,
		DC:       false,
		Provider: "",
	}

	cluster.Cib = cibConfig

	crmmonParser := crmmon.NewCrmMonParser(discoveryTools.CrmmonAdmPath)

	crmmonConfig, err := crmmonParser.Parse()
	if err != nil {
		return nil, err
	}

	cluster.Crmmon = crmmonConfig

	// Set MD5-hashed key based on the corosync auth key
	cluster.ID, err = getCorosyncAuthkeyMd5(discoveryTools.CorosyncKeyPath)
	if err != nil {
		return nil, err
	}

	cluster.Name = getName(cluster)

	sbdData, err := NewSBD(discoveryTools.CommandExecutor, discoveryTools.SBDPath, discoveryTools.SBDConfigPath)
	if err != nil && cluster.IsFencingSBD() {
		slog.Error("Error discovering SBD data", "error", err)
	}

	cluster.SBD = sbdData

	cluster.DC = cluster.IsDC()

	cloudIdentifier := cloud.NewIdentifier(discoveryTools.CommandExecutor)
	provider, err := cloudIdentifier.IdentifyCloudProvider()
	if err != nil {
		slog.Warn("Cloud provider not identified", "error", err)
	}
	cluster.Provider = provider

	return cluster, nil
}

func getCorosyncAuthkeyMd5(corosyncKeyPath string) (string, error) {
	kp, err := Md5sumFile(corosyncKeyPath)
	return kp, err
}

func getName(c *Cluster) string {
	// Handle not named clusters
	for _, prop := range c.Cib.Configuration.CrmConfig.ClusterProperties {
		if prop.ID == clusterNameProperty {
			return prop.Value
		}
	}

	return ""
}

func (c *Cluster) IsDC() bool {
	host, _ := os.Hostname()

	for _, nodes := range c.Crmmon.Nodes {
		if nodes.Name == host {
			return nodes.DC
		}
	}

	return false
}

func (c *Cluster) FencingType() string {
	for _, resource := range c.Crmmon.Resources {
		if strings.HasPrefix(resource.Agent, stonithAgent) {
			return strings.Split(resource.Agent, ":")[1]
		}
	}
	return stonithResourceMissing
}

func (c *Cluster) IsFencingSBD() bool {
	f := c.FencingType()

	return f == sbdFencingAgentName
}
