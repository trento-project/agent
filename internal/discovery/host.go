package discovery

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/Showmax/go-fqdn"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	log "github.com/sirupsen/logrus"
	"github.com/trento-project/agent/internal/core/hosts"
	"github.com/trento-project/agent/internal/discovery/collector"
	"github.com/trento-project/agent/version"
)

const HostDiscoveryID string = "host_discovery"
const HostDiscoveryMinPeriod time.Duration = 1 * time.Second

const NodeExporterName string = "node_exporter"
const NodeExporterPort int = 9100

type PrometheusTargets map[string]string

type HostDiscovery struct {
	id                string
	collectorClient   collector.Client
	host              string
	prometheusTargets PrometheusTargets
	interval          time.Duration
}

func NewHostDiscovery(
	collectorClient collector.Client,
	hostname string,
	prometheusTargets PrometheusTargets,
	config DiscoveriesConfig,
) Discovery {
	return HostDiscovery{
		id:                HostDiscoveryID,
		collectorClient:   collectorClient,
		host:              hostname,
		prometheusTargets: prometheusTargets,
		interval:          config.DiscoveriesPeriodsConfig.Host,
	}
}

func (d HostDiscovery) GetID() string {
	return d.id
}

func (d HostDiscovery) GetInterval() time.Duration {
	return d.interval
}

// Execute one iteration of a discovery and publish to the collector
func (d HostDiscovery) Discover(ctx context.Context) (string, error) {
	ipAddresses, netmasks, err := getNetworksData()
	if err != nil {
		return "", err
	}

	updatedPrometheusTargets := updatePrometheusTargets(d.prometheusTargets, ipAddresses)

	host := hosts.DiscoveredHost{
		OSVersion:                getOSVersion(),
		HostIPAddresses:          ipAddresses,
		Netmasks:                 netmasks,
		HostName:                 d.host,
		CPUCount:                 getLogicalCPUs(),
		SocketCount:              getCPUSocketCount(),
		TotalMemoryMB:            getTotalMemoryMB(),
		AgentVersion:             version.Version,
		InstallationSource:       version.InstallationSource,
		FullyQualifiedDomainName: getHostFQDN(),
		PrometheusTargets:        updatedPrometheusTargets,
	}

	err = d.collectorClient.Publish(ctx, d.id, host)
	if err != nil {
		log.Debugf("Error while sending host discovery to data collector: %s", err)
		return "", err
	}

	return fmt.Sprintf("Host with name: %s successfully discovered", d.host), nil
}

func getNetworksData() ([]string, []int, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, err
	}

	ipAddrList := make([]string, 0)
	netmasks := make([]int, 0)

	for _, inter := range interfaces {
		addrs, err := inter.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			cidrAddr, ip, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			ipAddrList = append(ipAddrList, cidrAddr.String())
			mask, _ := ip.Mask.Size()
			netmasks = append(netmasks, mask)
		}
	}

	return ipAddrList, netmasks, nil
}

func updatePrometheusTargets(targets PrometheusTargets, ipAddresses []string) PrometheusTargets {
	// Return exporter details if they are given by the user
	nodeExporterTarget, ok := targets[NodeExporterName]
	if ok && nodeExporterTarget != "" {
		return PrometheusTargets{
			NodeExporterName: nodeExporterTarget,
		}
	}

	// Fallback to lowest IP address value to replace empty exporter targets
	ips := make([]net.IP, 0, len(ipAddresses))
	for _, ip := range ipAddresses {
		parsedIp := net.ParseIP(ip)
		if parsedIp.To4() != nil && !parsedIp.IsLoopback() {
			ips = append(ips, parsedIp)
		}

	}

	sort.Slice(ips, func(i, j int) bool {
		return bytes.Compare(ips[i], ips[j]) < 0
	})

	return PrometheusTargets{
		NodeExporterName: fmt.Sprintf("%s:%d", ips[0], NodeExporterPort),
	}
}

func getHostFQDN() *string {

	fqdn, err := fqdn.FqdnHostname()
	if err != nil {
		log.Errorf("could not get the fully qualified domain name of the machine")
	}

	if len(fqdn) == 0 {
		return nil
	}

	return &fqdn
}

func getOSVersion() string {
	infoStat, err := host.Info()
	if err != nil {
		log.Errorf("Error while getting host info: %s", err)
	}
	return infoStat.PlatformVersion
}

func getTotalMemoryMB() int {
	v, err := mem.VirtualMemory()
	if err != nil {
		log.Errorf("Error while getting memory info: %s", err)
	}
	return int(v.Total) / 1024 / 1024
}

func getLogicalCPUs() int {
	logical, err := cpu.Counts(true)
	if err != nil {
		log.Errorf("Error while getting logical CPU count: %s", err)
	}
	return logical
}

func getCPUSocketCount() int {
	info, err := cpu.Info()

	if err != nil {
		log.Errorf("Error while getting CPU info: %s", err)
		return 0
	}

	// Get the last CPU info and get the physical ID of it
	lastCPUInfo := info[len(info)-1]

	physicalID, err := strconv.Atoi(lastCPUInfo.PhysicalID)

	if err != nil {
		log.Errorf("Unable to convert CPU socket count: %s", err)
		return 0
	}

	// Increase by one as physicalIDs start in zero
	return physicalID + 1
}
