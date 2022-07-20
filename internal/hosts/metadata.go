package hosts

type Metadata struct {
	Cluster         string `mapstructure:"ha-cluster,omitempty"`
	ClusterID       string `mapstructure:"ha-cluster-id,omitempty" json:"ClusterId"`
	SAPSystems      string `mapstructure:"sap-systems,omitempty"`
	SAPSystemsID    string `mapstructure:"sap-systems-id,omitempty" json:"SAPSystemsId"`
	SAPSystemsType  string `mapstructure:"sap-systems-type,omitempty"`
	CloudProvider   string `mapstructure:"cloud-provider,omitempty"`
	HostIPAddresses string `mapstructure:"host-ip-addresses,omitempty" json:"HostIpAddresses"`
	AgentVersion    string `mapstructure:"agent-version,omitempty"`
}
