package cluster

import (
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	// These packages were originally imported from github.com/ClusterLabs/ha_cluster_exporter/collector/pacemaker
	// Now we mantain our own fork

	"github.com/trento-project/agent/internal/core/cloud"
	"github.com/trento-project/agent/internal/core/cluster/cib"
	"github.com/trento-project/agent/internal/core/cluster/corosync"
	"github.com/trento-project/agent/internal/core/cluster/crmmon"
	"github.com/trento-project/agent/internal/core/cluster/systemctl"
	"github.com/trento-project/agent/pkg/utils"
)

const (
	cibAdmPath             string = "/usr/sbin/cibadmin"
	crmmonAdmPath          string = "/usr/sbin/crm_mon"
	corosyncConfigPath     string = "/etc/corosync/corosync.conf"
	corosyncKeyPath        string = "/etc/corosync/authkey"
	clusterNameProperty    string = "cib-bootstrap-options-cluster-name"
	stonithResourceMissing string = "notconfigured"
	stonithAgent           string = "stonith:"
	sbdFencingAgentName    string = "external/sbd"
)

type DiscoveryTools struct {
	CibAdmPath         string
	CrmmonAdmPath      string
	CorosyncKeyPath    string
	CorosyncConfigPath string
	SBDPath            string
	SBDConfigPath      string
	CommandExecutor    utils.CommandExecutor
}

type ClusterBase struct {
	ID   string
	Name string
}

type Cluster struct {
	Cib      cib.Root
	Crmmon   crmmon.Root
	ID       string `json:"Id"`
	Name     string
	SBD      SBD
	DC       bool
	Provider string
	Online   bool
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
		CibAdmPath:         cibAdmPath,
		CrmmonAdmPath:      crmmonAdmPath,
		CorosyncConfigPath: corosyncConfigPath,
		CorosyncKeyPath:    corosyncKeyPath,
		SBDPath:            SBDPath,
		SBDConfigPath:      SBDConfigPath,
		CommandExecutor:    utils.Executor{},
	})
}

func NewClusterWithDiscoveryTools(discoveryTools *DiscoveryTools) (*Cluster, error) {
	detectedCluster, found, err := detectCluster(discoveryTools)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	isOnline, err := isHostOnline(discoveryTools)
	if err != nil {
		return nil, fmt.Errorf("error checking if host is online: %w", err)
	}

	if !isOnline {
		return makeOfflineHostPayload(detectedCluster)
	}
	return makeOnlineHostPayload(detectedCluster, discoveryTools)
}

func detectCluster(discoveryTools *DiscoveryTools) (ClusterBase, bool, error) {
	noCluster := ClusterBase{}

	for _, filepath := range []string{
		discoveryTools.CorosyncKeyPath,
		discoveryTools.CorosyncConfigPath} {

		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			return noCluster, false, nil
		} else if err != nil {
			return noCluster, err, false
		}
	}

	id, err := getCorosyncAuthkeyMd5(discoveryTools.CorosyncKeyPath)
	if err != nil {
		return noCluster, err, false
	}

	name, err := getCorosyncClusterName(discoveryTools.CorosyncConfigPath)
	if err != nil {
		return noCluster, err, false
	}

	return ClusterBase{
		ID:   id,
		Name: name,
	}, nil, true

}

func isHostOnline(discoveryTools *DiscoveryTools) (bool, error) {
	systemctl := systemctl.NewSystemctl(discoveryTools.CommandExecutor)

	for _, service := range []string{"corosync", "pacemaker"} {
		active := systemctl.IsActive(service)
		if !active {
			slog.Warn("Service is not active", "service", service)
			return false, nil
		}
	}

	return true, nil

}

func makeOfflineHostPayload(detectedCluster ClusterBase) (*Cluster, error) {
	return &Cluster{
		ID:     detectedCluster.ID,
		Name:   detectedCluster.Name,
		Online: false,
	}, nil
}

func makeOnlineHostPayload(detectedCluster ClusterBase, discoveryTools *DiscoveryTools) (*Cluster, error) {
	cibParser := cib.NewCibAdminParser(discoveryTools.CibAdmPath)

	cibConfig, err := cibParser.Parse()
	if err != nil {
		return nil, err
	}

	var cluster = &Cluster{
		Cib:      cib.Root{},    //nolint
		Crmmon:   crmmon.Root{}, //nolint
		SBD:      SBD{},         //nolint
		ID:       detectedCluster.ID,
		Name:     detectedCluster.Name,
		DC:       false,
		Provider: "",
		Online:   true,
	}

	cluster.Cib = cibConfig

	crmmonParser := crmmon.NewCrmMonParser(discoveryTools.CrmmonAdmPath)

	crmmonConfig, err := crmmonParser.Parse()
	if err != nil {
		return nil, err
	}

	cluster.Crmmon = crmmonConfig

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

func getCorosyncClusterName(corosyncConfigPath string) (string, error) {
	parser := corosync.NewCorosyncParser(corosyncConfigPath)
	corosyncConf, err := parser.Parse()
	if err != nil {
		return "", fmt.Errorf("error parsing corosync.conf: %w", err)
	}
	name, ok := corosyncConf.Totem["cluster_name"].(string)

	if !ok {
		return "", fmt.Errorf("cluster_name not found or not a string in corosync.conf")
	}
	return name, nil
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
