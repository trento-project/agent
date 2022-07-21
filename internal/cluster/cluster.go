package cluster

import (
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/trento-project/agent/internal/cloud"
	// These packages were originally imported from github.com/ClusterLabs/ha_cluster_exporter/collector/pacemaker
	// Now we mantain our own fork
	"github.com/trento-project/agent/internal/cluster/cib"
	"github.com/trento-project/agent/internal/cluster/crmmon"
)

const (
	cibAdmPath             string = "/usr/sbin/cibadmin"
	crmmonAdmPath          string = "/usr/sbin/crm_mon"
	corosyncKeyPath        string = "/etc/corosync/authkey"
	clusterNameProperty    string = "cib-bootstrap-options-cluster-name"
	stonithEnabled         string = "cib-bootstrap-options-stonith-enabled"
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
}

type Cluster struct {
	Cib      cib.Root    `mapstructure:"cib,omitempty"`
	Crmmon   crmmon.Root `mapstructure:"crmmon,omitempty"`
	SBD      SBD         `mapstructure:"sbd,omitempty"`
	ID       string      `mapstructure:"id" json:"Id"`
	Name     string      `mapstructure:"name"`
	DC       bool        `mapstructure:"dc"`
	Provider string      `mapstructure:"provider"`
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

func NewCluster() (Cluster, error) {
	return NewClusterWithDiscoveryTools(&DiscoveryTools{
		CibAdmPath:      cibAdmPath,
		CrmmonAdmPath:   crmmonAdmPath,
		CorosyncKeyPath: corosyncKeyPath,
		SBDPath:         SBDPath,
		SBDConfigPath:   SBDConfigPath,
	})
}

func NewClusterWithDiscoveryTools(discoveryTools *DiscoveryTools) (Cluster, error) {
	var cluster = Cluster{
		Cib:      cib.Root{},    //nolint
		Crmmon:   crmmon.Root{}, //nolint
		SBD:      SBD{},         //nolint
		ID:       "",
		Name:     clusterNameProperty,
		DC:       false,
		Provider: "",
	}

	cibParser := cib.NewCibAdminParser(discoveryTools.CibAdmPath)

	cibConfig, err := cibParser.Parse()
	if err != nil {
		return cluster, err
	}

	cluster.Cib = cibConfig

	crmmonParser := crmmon.NewCrmMonParser(discoveryTools.CrmmonAdmPath)

	crmmonConfig, err := crmmonParser.Parse()
	if err != nil {
		return cluster, err
	}

	cluster.Crmmon = crmmonConfig

	// Set MD5-hashed key based on the corosync auth key
	cluster.ID, err = getCorosyncAuthkeyMd5(discoveryTools.CorosyncKeyPath)
	if err != nil {
		return cluster, err
	}

	cluster.Name = getName(cluster)

	if cluster.IsFencingSBD() {
		sbdData, err := NewSBD(cluster.ID, discoveryTools.SBDPath, discoveryTools.SBDConfigPath)
		if err != nil {
			return cluster, err
		}

		cluster.SBD = sbdData
	}

	cluster.DC = isDC(&cluster)

	provider, _ := cloud.IdentifyCloudProvider()
	cluster.Provider = provider

	return cluster, nil
}

func getCorosyncAuthkeyMd5(corosyncKeyPath string) (string, error) {
	kp, err := Md5sumFile(corosyncKeyPath)
	return kp, err
}

func getName(c Cluster) string {
	// Handle not named clusters
	for _, prop := range c.Cib.Configuration.CrmConfig.ClusterProperties {
		if prop.ID == clusterNameProperty {
			return prop.Value
		}
	}

	return ""
}

func isDC(c *Cluster) bool {
	host, _ := os.Hostname()

	for _, nodes := range c.Crmmon.Nodes {
		if nodes.Name == host {
			return nodes.DC
		}
	}

	return false
}

func (c *Cluster) IsFencingEnabled() bool {
	for _, prop := range c.Cib.Configuration.CrmConfig.ClusterProperties {
		if prop.ID == stonithEnabled {
			b, err := strconv.ParseBool(prop.Value)
			if err != nil {
				return false
			}
			return b
		}
	}

	return false
}

func (c *Cluster) FencingResourceExists() bool {
	f := c.FencingType()

	return f != stonithResourceMissing
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
