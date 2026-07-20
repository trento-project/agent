// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package crmmon

// *** crm_mon XML unserialization structures

import (
	"encoding/xml"
	"math"
	"strconv"
)

// ClusterOptions holds cluster-level fencing configuration parsed from crm_mon XML.
// Pacemaker 3.0.2+ emits both FencingEnabled (new) and StonithEnabled (deprecated) simultaneously.
// Use IsFencingEnabled() to get the resolved value with correct precedence.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/crm_mon-2.42.rng
type ClusterOptions struct {
	StonithEnabled         bool   `xml:"stonith-enabled,attr"`    // deprecated in Pacemaker 3.0.2+
	FencingEnabled         bool   `xml:"fencing-enabled,attr"`    // new in Pacemaker 3.0.2+
	StonithTimeoutMs       int    `xml:"stonith-timeout-ms,attr"` // deprecated in Pacemaker 3.0.2+
	FencingTimeoutMs       int    `xml:"fencing-timeout-ms,attr"` // new in Pacemaker 3.0.2+
	SymmetricCluster       bool   `xml:"symmetric-cluster,attr"`
	NoQuorumPolicy         string `xml:"no-quorum-policy,attr"`
	MaintenanceMode        bool   `xml:"maintenance-mode,attr"`
	StopAllResources       bool   `xml:"stop-all-resources,attr"`
	PriorityFencingDelayMs int    `xml:"priority-fencing-delay-ms,attr"`
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
	// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/pacemakerd-health-2.25.rng
	PacemakerdHealth struct {
		SysFrom     string `xml:"sys_from,attr"`
		State       string `xml:"state,attr"`
		LastUpdated string `xml:"last_updated,attr"`
	} `xml:"pacemakerd"`
	// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/crm_mon-2.42.rng
	Summary struct {
		Stack struct {
			Type            string `xml:"type,attr"`
			PacemakerdState string `xml:"pacemakerd-state,attr"`
		} `xml:"stack"`
		CurrentDc struct {
			Present      bool   `xml:"present,attr"`
			Version      string `xml:"version,attr"`
			Name         string `xml:"name,attr"`
			ID           string `xml:"id,attr" json:"Id"` //nolint
			WithQuorum   bool   `xml:"with_quorum,attr"`
			MixedVersion bool   `xml:"mixed_version,attr"`
		} `xml:"current_dc"`
		LastUpdate struct {
			Time   string `xml:"time,attr"`
			Origin string `xml:"origin,attr"`
		} `xml:"last_update"`
		Nodes struct {
			Number int `xml:"number,attr"`
		} `xml:"nodes_configured"`
		LastChange struct {
			Time   string `xml:"time,attr"`
			User   string `xml:"user,attr"`
			Client string `xml:"client,attr"`
			Origin string `xml:"origin,attr"`
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
				Name     string `xml:"name,attr"`
				Value    string `xml:"value,attr"`
				Expected *int   `xml:"expected,attr"` // optional; nil means absent
			} `xml:"attribute"`
		} `xml:"node"`
	} `xml:"node_attributes"`
	// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/node-history-2.41.rng
	NodeHistory struct {
		Nodes []struct {
			Name            string `xml:"name,attr"`
			ResourceHistory []struct {
				Name               string    `xml:"id,attr" json:"Name"`
				MigrationThreshold int       `xml:"migration-threshold,attr"`
				FailCount          FailCount `xml:"fail-count,attr"` // "INFINITY" maps to math.MaxInt32
				Orphan             bool      `xml:"orphan,attr"`     // deprecated in Pacemaker 3.0.2+
				Removed            bool      `xml:"removed,attr"`    // new in Pacemaker 3.0.2+
				LastFailure        string    `xml:"last-failure,attr"`
				OperationHistory   []struct {
					Call         string `xml:"call,attr"`
					Task         string `xml:"task,attr"`
					Interval     string `xml:"interval,attr"`
					LastRcChange string `xml:"last-rc-change,attr"`
					LastRun      string `xml:"last-run,attr"`
					ExecTime     string `xml:"exec-time,attr"`
					QueueTime    string `xml:"queue-time,attr"`
					RC           int    `xml:"rc,attr"`
					RCText       string `xml:"rc_text,attr"`
				} `xml:"operation_history"`
			} `xml:"resource_history"`
		} `xml:"node"`
	} `xml:"node_history"`
	// The schema is permissive, but Pacemaker doesn't usually emit that nesting.
	Resources []Resource `xml:"resources>resource"`
	Clones    []Clone    `xml:"resources>clone"`
	Groups    []Group    `xml:"resources>group"`
	Bundles   []Bundle   `xml:"resources>bundle"`
	// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/failure-2.8.rng
	Failures []struct {
		OpKey        string `xml:"op_key,attr"`       // mutually exclusive with ID
		ID           string `xml:"id,attr" json:"Id"` //nolint
		Node         string `xml:"node,attr"`
		ExitStatus   string `xml:"exitstatus,attr"`
		ExitReason   string `xml:"exitreason,attr"`
		ExitCode     int    `xml:"exitcode,attr"`
		Call         int    `xml:"call,attr"`
		Status       string `xml:"status,attr"`
		LastRcChange string `xml:"last-rc-change,attr"`
		Queued       int    `xml:"queued,attr"`
		Exec         int    `xml:"exec,attr"`
		Interval     int    `xml:"interval,attr"`
		Task         string `xml:"task,attr"`
	} `xml:"failures>failure"`
	// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/fence-event-2.15.rng
	FenceHistory struct {
		Status      int `xml:"status,attr"`
		FenceEvents []struct {
			Status         string `xml:"status,attr"` // failed, success, or pending
			ExtendedStatus string `xml:"extended-status,attr"`
			ExitReason     string `xml:"exit-reason,attr"`
			Delegate       string `xml:"delegate,attr"`
			Action         string `xml:"action,attr"`
			Target         string `xml:"target,attr"`
			Client         string `xml:"client,attr"`
			Origin         string `xml:"origin,attr"`
			Completed      string `xml:"completed,attr"`
		} `xml:"fence_event"`
	} `xml:"fence_history"`
	// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/ticket-2.35.rng
	Tickets []struct {
		ID          string `xml:"id,attr" json:"Id"` //nolint
		Status      string `xml:"status,attr"`       // granted or revoked
		Standby     bool   `xml:"standby,attr"`
		LastGranted string `xml:"last-granted,attr"`
		Attribute   struct {
			Name  string `xml:"name,attr"`
			Value string `xml:"value,attr"`
		} `xml:"attribute"`
		// Constraints lists the resources associated with this ticket (geo-cluster setups).
		Constraints []struct {
			ID           string `xml:"id,attr" json:"Id"` //nolint
			Resource     string `xml:"rsc,attr"`
			ResourceRole string `xml:"rsc-role,attr"`
			TicketID     string `xml:"ticket,attr"`
			LossPolicy   string `xml:"loss-policy,attr"` // stop, demote, fence, or freeze
			ResourceSets []struct {
				ID           string `xml:"id,attr" json:"Id"` //nolint
				IDRef        string `xml:"id-ref,attr"`
				Sequential   bool   `xml:"sequential,attr"`
				RequireAll   bool   `xml:"require-all,attr"`
				Ordering     string `xml:"ordering,attr"` // group or listed
				Action       string `xml:"action,attr"`   // start, promote, demote, or stop
				Role         string `xml:"role,attr"`
				Score        string `xml:"score,attr"` // integer or INFINITY/+INFINITY/-INFINITY
				Kind         string `xml:"kind,attr"`  // Optional, Mandatory, or Serialize
				ResourceRefs []struct {
					ID string `xml:"id,attr" json:"Id"` //nolint
				} `xml:"resource_ref"`
			} `xml:"resource_set"`
		} `xml:"constraints>rsc_ticket"`
	} `xml:"tickets>ticket"`
	Bans []struct {
		ID           string `xml:"id,attr" json:"Id"` //nolint
		Resource     string `xml:"resource,attr"`
		Node         string `xml:"node,attr"`
		Weight       int    `xml:"weight,attr"`
		PromotedOnly bool   `xml:"promoted-only,attr"`
		MasterOnly   bool   `xml:"master_only,attr"` // deprecated duplicate of PromotedOnly
	} `xml:"bans>ban"`
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
	Health           string `xml:"health,attr"`
	FeatureSet       string `xml:"feature_set,attr"`
	ContainerID      string `xml:"container_id,attr"`
	IDAsResource     string `xml:"id_as_resource,attr"`
	// The schema is permissive, but Pacemaker doesn't usually emit that nesting.
	Resources []Resource `xml:"resource"`
	Clones    []Clone    `xml:"clone"`
	Groups    []Group    `xml:"group"`
	Bundles   []Bundle   `xml:"bundle"`
}

// Resource, Clone, Group, and Bundle represent cluster resource elements from crm_mon XML output.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/resources-2.41.rng
type Resource struct {
	ID             string `xml:"id,attr" json:"Id"` //nolint
	Agent          string `xml:"resource_agent,attr"`
	Role           string `xml:"role,attr"`
	TargetRole     string `xml:"target_role,attr"`
	Description    string `xml:"description,attr"`
	Active         bool   `xml:"active,attr"`
	Orphaned       bool   `xml:"orphaned,attr"` // deprecated in Pacemaker 3.0.2+
	Removed        bool   `xml:"removed,attr"`  // new in Pacemaker 3.0.2+
	Blocked        bool   `xml:"blocked,attr"`
	Managed        bool   `xml:"managed,attr"`
	Maintenance    bool   `xml:"maintenance,attr"`
	Failed         bool   `xml:"failed,attr"`
	FailureIgnored bool   `xml:"failure_ignored,attr"`
	NodesRunningOn int    `xml:"nodes_running_on,attr"`
	Pending        string `xml:"pending,attr"`
	LockedTo       string `xml:"locked_to,attr"`
	// TODO: Schema allows zeroOrMore node elements, but in practice a primitive resource runs on a single node.
	// Changing this to a slice would require updating downstream consumers (e.g. trento-web).
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
	Description    string     `xml:"description,attr"`
	MultiState     bool       `xml:"multi_state,attr"`
	Managed        bool       `xml:"managed,attr"`
	Disabled       bool       `xml:"disabled,attr"`
	Maintenance    bool       `xml:"maintenance,attr"`
	Failed         bool       `xml:"failed,attr"`
	FailureIgnored bool       `xml:"failure_ignored,attr"`
	Unique         bool       `xml:"unique,attr"`
	TargetRole     string     `xml:"target_role,attr"`
	Resources      []Resource `xml:"resource"`
	// The schema is permissive, but Pacemaker doesn't usually emit that nesting.
	Clones  []Clone  `xml:"clone"`
	Groups  []Group  `xml:"group"`
	Bundles []Bundle `xml:"bundle"`
}

// Group represents a resource group in crm_mon XML output.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/resources-2.41.rng
type Group struct {
	ID              string     `xml:"id,attr" json:"Id"` //nolint
	Description     string     `xml:"description,attr"`
	Managed         bool       `xml:"managed,attr"`
	Resources       []Resource `xml:"resource"`
	NumberResources int        `xml:"number_resources,attr"`
	Disabled        bool       `xml:"disabled,attr"`
	Maintenance     bool       `xml:"maintenance,attr"`
	// The schema is permissive, but Pacemaker doesn't usually emit that nesting.
	Clones  []Clone  `xml:"clone"`
	Groups  []Group  `xml:"group"`
	Bundles []Bundle `xml:"bundle"`
}

// Bundle is a crm_mon container bundle resource (docker/rkt/podman).
// Note: structurally different from the CIB bundle in the cib package.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/api/resources-2.41.rng
type Bundle struct {
	ID          string `xml:"id,attr" json:"Id"` //nolint
	Type        string `xml:"type,attr"`         // docker, rkt, or podman
	Image       string `xml:"image,attr"`
	Unique      bool   `xml:"unique,attr"`
	Maintenance bool   `xml:"maintenance,attr"`
	Description string `xml:"description,attr"`
	Managed     bool   `xml:"managed,attr"`
	Failed      bool   `xml:"failed,attr"`
	Replicas    []struct {
		ID        int        `xml:"id,attr" json:"Id"` //nolint
		Resources []Resource `xml:"resource"`
	} `xml:"replica"`
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

// FailCount is an int that also handles the special "INFINITY" value that
// Pacemaker emits for resources that are permanently banned. Absent attributes
// unmarshal to 0 (Go's zero value), matching the old int behavior.
type FailCount int

func (fc *FailCount) UnmarshalXMLAttr(attr xml.Attr) error {
	switch attr.Value {
	case "INFINITY", "+INFINITY":
		*fc = FailCount(math.MaxInt32)
	case "-INFINITY":
		*fc = FailCount(math.MinInt32)
	default:
		n, err := strconv.Atoi(attr.Value)
		if err != nil {
			return err
		}
		*fc = FailCount(n)
	}
	return nil
}
