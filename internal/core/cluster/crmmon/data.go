package crmmon

// *** crm_mon XML unserialization structures

import (
	"encoding/xml"
)

type Root struct {
	Version string `xml:"version,attr"`
	Summary struct {
		Nodes struct {
			Number int `xml:"number,attr"`
		} `xml:"nodes_configured"`
		LastChange struct {
			Time string `xml:"time,attr"`
		} `xml:"last_change"`
		Resources struct {
			Number   int `xml:"number,attr"`
			Disabled int `xml:"disabled,attr"`
			Blocked  int `xml:"blocked,attr"`
		} `xml:"resources_configured"`
		ClusterOptions struct {
			StonithEnabled bool `xml:"stonith-enabled,attr"`
		} `xml:"cluster_options"`
	} `xml:"summary"`
	Nodes          []Node `xml:"nodes>node"`
	NodeAttributes struct {
		Nodes []struct {
			Name       string `xml:"name,attr"`
			Attributes []struct {
				Name  string `xml:"name,attr"`
				Value string `xml:"value,attr"`
			} `xml:"attribute"`
		} `xml:"node"`
	} `xml:"node_attributes"`
	NodeHistory struct {
		Nodes []struct {
			Name            string `xml:"name,attr"`
			ResourceHistory []struct {
				Name               string `xml:"id,attr" json:"Name"`
				MigrationThreshold int    `xml:"migration-threshold,attr"`
				FailCount          int    `xml:"fail-count,attr"`
			} `xml:"resource_history"`
		} `xml:"node"`
	} `xml:"node_history"`
	Resources []Resource `xml:"resources>resource"`
	Clones    []Clone    `xml:"resources>clone"`
	Groups    []Group    `xml:"resources>group"`
}

type Node struct {
	Name             string `xml:"name,attr"`
	ID               string `xml:"id,attr" json:"Id"`
	Online           bool   `xml:"online,attr"`
	Standby          bool   `xml:"standby,attr"`
	StandbyOnFail    bool   `xml:"standby_onfail,attr"`
	Maintenance      bool   `xml:"maintenance,attr"`
	Pending          bool   `xml:"pending,attr"`
	Unclean          bool   `xml:"unclean,attr"`
	Shutdown         bool   `xml:"shutdown,attr"`
	ExpectedUp       bool   `xml:"expected_up,attr"`
	DC               bool   `xml:"is_dc,attr"`
	ResourcesRunning int    `xml:"resources_running,attr"`
	Type             string `xml:"type,attr"`
}

type Resource struct {
	ID             string `xml:"id,attr" json:"Id"`
	Agent          string `xml:"resource_agent,attr"`
	Role           string `xml:"role,attr"`
	Active         bool   `xml:"active,attr"`
	Orphaned       bool   `xml:"orphaned,attr"`
	Blocked        bool   `xml:"blocked,attr"`
	Managed        bool   `xml:"managed,attr"`
	Failed         bool   `xml:"failed,attr"`
	FailureIgnored bool   `xml:"failure_ignored,attr"`
	NodesRunningOn int    `xml:"nodes_running_on,attr"`
	Node           *struct {
		Name   string `xml:"name,attr"`
		ID     string `xml:"id,attr" json:"Id"`
		Cached bool   `xml:"cached,attr"`
	} `xml:"node,omitempty"`
}

type Clone struct {
	ID             string     `xml:"id,attr" json:"Id"`
	MultiState     bool       `xml:"multi_state,attr"`
	Managed        bool       `xml:"managed,attr"`
	Failed         bool       `xml:"failed,attr"`
	FailureIgnored bool       `xml:"failure_ignored,attr"`
	Unique         bool       `xml:"unique,attr"`
	Resources      []Resource `xml:"resource"`
}

type Group struct {
	ID        string     `xml:"id,attr" json:"Id"`
	Managed   bool       `xml:"managed,attr"`
	Resources []Resource `xml:"resource"`
}

// UnmarshalXML of Group to set Managed field default value to true
func (g *Group) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type resultGroup Group // new type to prevent recursion
	item := resultGroup{
		Managed: true,
	}
	if err := d.DecodeElement(&item, &start); err != nil {
		return err
	}
	*g = (Group)(item)
	return nil
}
