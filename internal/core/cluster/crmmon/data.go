// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package crmmon

// *** crm_mon XML unserialization structures

import (
	"encoding/xml"
)

// ClusterOptions holds cluster-level fencing configuration parsed from crm_mon XML.
// Pacemaker 3.0.2+ emits both FencingEnabled (new) and StonithEnabled (deprecated) simultaneously.
// Use IsFencingEnabled() to get the resolved value with correct precedence.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/crm_mon-2.42.rng
type ClusterOptions struct {
	StonithEnabled   bool `xml:"stonith-enabled,attr"`    // deprecated in Pacemaker 3.0.2+
	FencingEnabled   bool `xml:"fencing-enabled,attr"`    // new in Pacemaker 3.0.2+
	StonithTimeoutMs int  `xml:"stonith-timeout-ms,attr"` // deprecated in Pacemaker 3.0.2+
	FencingTimeoutMs int  `xml:"fencing-timeout-ms,attr"` // new in Pacemaker 3.0.2+
}

// IsFencingEnabled returns true if fencing is enabled, preferring the new fencing-enabled
// attribute and falling back to the deprecated stonith-enabled attribute.
func (co ClusterOptions) IsFencingEnabled() bool {
	return co.FencingEnabled || co.StonithEnabled
}

// Root is the top-level crm_mon XML output structure.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/crm_mon-2.42.rng
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
		ClusterOptions ClusterOptions `xml:"cluster_options"`
	} `xml:"summary"`
	Nodes []Node `xml:"nodes>node"`
	// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/node-attrs-2.8.rng
	NodeAttributes struct {
		Nodes []struct {
			Name       string `xml:"name,attr"`
			Attributes []struct {
				Name  string `xml:"name,attr"`
				Value string `xml:"value,attr"`
			} `xml:"attribute"`
		} `xml:"node"`
	} `xml:"node_attributes"`
	// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/node-history-2.41.rng
	NodeHistory struct {
		Nodes []struct {
			Name            string `xml:"name,attr"`
			ResourceHistory []struct {
				Name               string `xml:"id,attr" json:"Name"`
				MigrationThreshold int    `xml:"migration-threshold,attr"`
				FailCount          int    `xml:"fail-count,attr"`
				Orphan             bool   `xml:"orphan,attr"`  // deprecated in Pacemaker 3.0.2+
				Removed            bool   `xml:"removed,attr"` // new in Pacemaker 3.0.2+
			} `xml:"resource_history"`
		} `xml:"node"`
	} `xml:"node_history"`
	Resources []Resource `xml:"resources>resource"`
	Clones    []Clone    `xml:"resources>clone"`
	Groups    []Group    `xml:"resources>group"`
}

// Node represents a cluster node element from crm_mon XML output.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/nodes-2.41.rng
type Node struct {
	Name             string `xml:"name,attr"`
	ID               string `xml:"id,attr" json:"Id"` //nolint
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

// Resource, Clone, Group, and Bundle represent cluster resource elements from crm_mon XML output.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/resources-2.41.rng
type Resource struct {
	ID             string `xml:"id,attr" json:"Id"` //nolint
	Agent          string `xml:"resource_agent,attr"`
	Role           string `xml:"role,attr"`
	Active         bool   `xml:"active,attr"`
	Orphaned       bool   `xml:"orphaned,attr"` // deprecated in Pacemaker 3.0.2+
	Removed        bool   `xml:"removed,attr"`  // new in Pacemaker 3.0.2+
	Blocked        bool   `xml:"blocked,attr"`
	Managed        bool   `xml:"managed,attr"`
	Failed         bool   `xml:"failed,attr"`
	FailureIgnored bool   `xml:"failure_ignored,attr"`
	NodesRunningOn int    `xml:"nodes_running_on,attr"`
	// TODO: Schema allows zeroOrMore node elements, but in practice a primitive resource runs on at most one node at a time.
	// Changing this to a slice would require updating downstream consumers (e.g. trento-web) that expect a single "Node" object.
	Node *struct {
		Name   string `xml:"name,attr"`
		ID     string `xml:"id,attr" json:"Id"` //nolint
		Cached bool   `xml:"cached,attr"`
	} `xml:"node,omitempty"`
}

// Clone represents a clone resource (including promotable/multi-state clones) in crm_mon XML output.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/resources-2.41.rng
type Clone struct {
	ID             string     `xml:"id,attr" json:"Id"` //nolint
	MultiState     bool       `xml:"multi_state,attr"`
	Managed        bool       `xml:"managed,attr"`
	Failed         bool       `xml:"failed,attr"`
	FailureIgnored bool       `xml:"failure_ignored,attr"`
	Unique         bool       `xml:"unique,attr"`
	Resources      []Resource `xml:"resource"`
}

// Group represents a resource group in crm_mon XML output.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/resources-2.41.rng
type Group struct {
	ID        string     `xml:"id,attr" json:"Id"` //nolint
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
