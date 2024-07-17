package discovery

import (
	"context"
	"fmt"
	"net"
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

type HostDiscovery struct {
	id              string
	collectorClient collector.Client
	host            string
	interval        time.Duration
}

func NewHostDiscovery(
	collectorClient collector.Client,
	hostname string,
	config DiscoveriesConfig,
) Discovery {
	return HostDiscovery{
		id:              HostDiscoveryID,
		collectorClient: collectorClient,
		host:            hostname,
		interval:        config.DiscoveriesPeriodsConfig.Host,
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
	networkInterfaces, ipAddresses, err := getNetworks()
	if err != nil {
		return "", err
	}

	host := hosts.DiscoveredHost{
		OSVersion:                getOSVersion(),
		HostIPAddresses:          ipAddresses,
		NetworkInterfaces:        networkInterfaces,
		HostName:                 d.host,
		CPUCount:                 getLogicalCPUs(),
		SocketCount:              getCPUSocketCount(),
		TotalMemoryMB:            getTotalMemoryMB(),
		AgentVersion:             version.Version,
		InstallationSource:       version.InstallationSource,
		FullyQualifiedDomainName: getHostFQDN(),
	}

	err = d.collectorClient.Publish(ctx, d.id, host)
	if err != nil {
		log.Debugf("Error while sending host discovery to data collector: %s", err)
		return "", err
	}

	return fmt.Sprintf("Host with name: %s successfully discovered", d.host), nil
}

// getNetworks still returns ip addresses list in 2nd return value
// for backward compatibility
func getNetworks() ([]hosts.NetworkInterface, []string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, err
	}

	ipAddrList := make([]string, 0)
	networkInterfaces := make([]hosts.NetworkInterface, 0)

	for _, inter := range interfaces {
		addrs, err := inter.Addrs()
		if err != nil {
			continue
		}

		newInterface := hosts.NetworkInterface{
			Index:     inter.Index,
			Name:      inter.Name,
			Addresses: make([]hosts.Address, 0),
		}

		for _, addr := range addrs {
			cidrAddr, ip, err := net.ParseCIDR(addr.String())
			if err != nil {
				continue
			}

			ipAddrList = append(ipAddrList, cidrAddr.String())
			mask, _ := ip.Mask.Size()
			newInterface.Addresses = append(newInterface.Addresses, hosts.Address{
				Address: cidrAddr.String(),
				Netmask: mask,
			})

		}

		networkInterfaces = append(networkInterfaces, newInterface)
	}

	return networkInterfaces, ipAddrList, nil
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
