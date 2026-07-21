// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package cib

/*
The Cluster Information Base (Root)
is an XML representation of the cluster's configuration
and the state of all nodes and resources.
The Root manager (pacemaker-based) keeps the Root synchronized across the cluster, and handles requests to modify it.

https://clusterlabs.org/pacemaker/doc/en-US/Pacemaker/2.0/html-single/Pacemaker_Administration/index.html

Schema (cib element attributes):
  https://github.com/ClusterLabs/pacemaker/blob/main/xml/cib-1.2.rng
  https://github.com/ClusterLabs/pacemaker/blob/main/xml/cib/versions.rng
*/

type Root struct {
	// <cib> element attributes
	ValidateWith    string `xml:"validate-with,attr"`
	AdminEpoch      int    `xml:"admin_epoch,attr"`
	Epoch           int    `xml:"epoch,attr"`
	NumUpdates      int    `xml:"num_updates,attr"`
	CRMFeatureSet   string `xml:"crm_feature_set,attr"`
	HaveQuorum      bool   `xml:"have-quorum,attr"`
	DCUuid          string `xml:"dc-uuid,attr"`
	CibLastWritten  string `xml:"cib-last-written,attr"`
	UpdateOrigin    string `xml:"update-origin,attr"`
	UpdateClient    string `xml:"update-client,attr"`
	UpdateUser      string `xml:"update-user,attr"`
	NoQuorumPanic   bool   `xml:"no-quorum-panic,attr"`
	ExecutionDate   string `xml:"execution-date,attr"`
	RemoteTLSPort   int    `xml:"remote-tls-port,attr"`
	RemoteClearPort int    `xml:"remote-clear-port,attr"` // deprecated since Pacemaker 3.0.2
	Configuration   struct {
		// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/nvset-3.9.rng
		CrmConfig struct {
			ClusterProperties []Attribute `xml:"cluster_property_set>nvpair"`
		} `xml:"crm_config"`
		// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/nodes-3.9.rng
		Nodes []struct {
			ID                 string      `xml:"id,attr" json:"Id"`
			Uname              string      `xml:"uname,attr"`
			Type               string      `xml:"type,attr"`
			Description        string      `xml:"description,attr"`
			Score              string      `xml:"score,attr"`
			InstanceAttributes []Attribute `xml:"instance_attributes>nvpair"`
			Utilization        []Attribute `xml:"utilization>nvpair"`
		} `xml:"nodes>node"`
		// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/resources-3.9.rng
		Resources struct {
			Primitives []Primitive `xml:"primitive"`
			Templates  []Primitive `xml:"template"`
			// Masters parses the legacy <master> element (pacemaker 2.x).
			// Pacemaker 3.x dropped <master>: promotable clones are expressed as
			// <clone> with promotable="true" in meta_attributes, so this will be
			// empty on pacemaker 3.x clusters. See Clones.
			Masters []Clone `xml:"master"`
			// Clones parses all <clone> elements, including promotable clones on
			// pacemaker 3.x (those have promotable="true" in their MetaAttributes).
			Clones  []Clone `xml:"clone"`
			Groups  []Group `xml:"group"`
			Bundles []struct {
				ID                 string      `xml:"id,attr" json:"Id"`
				Description        string      `xml:"description,attr"`
				MetaAttributes     []Attribute `xml:"meta_attributes>nvpair"`
				InstanceAttributes []Attribute `xml:"instance_attributes>nvpair"`
				// Docker, Rkt, and Podman use pointers: all three share the same struct type BundleContainer, so
				// a nil check is the only way to determine which container type is actually configured.
				Docker  *BundleContainer `xml:"docker"`
				Rkt     *BundleContainer `xml:"rkt"`
				Podman  *BundleContainer `xml:"podman"`
				Network struct {
					IPRangeStart  string `xml:"ip-range-start,attr"`
					AddHost       bool   `xml:"add-host,attr"`
					ControlPort   int    `xml:"control-port,attr"`
					HostInterface string `xml:"host-interface,attr"`
					HostNetmask   int    `xml:"host-netmask,attr"`
					PortMappings  []struct {
						ID           string `xml:"id,attr" json:"Id"`
						Port         int    `xml:"port,attr"`
						InternalPort int    `xml:"internal-port,attr"`
						Range        string `xml:"range,attr"`
					} `xml:"port-mapping"`
				} `xml:"network"`
				Storage struct {
					StorageMappings []struct {
						ID            string `xml:"id,attr" json:"Id"`
						SourceDir     string `xml:"source-dir,attr"`
						SourceDirRoot string `xml:"source-dir-root,attr"`
						TargetDir     string `xml:"target-dir,attr"`
						Options       string `xml:"options,attr"`
					} `xml:"storage-mapping"`
				} `xml:"storage"`
				Primitive Primitive `xml:"primitive"`
			} `xml:"bundle"`
		} `xml:"resources"`
		// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/constraints-3.9.rng
		Constraints struct {
			RscLocations []struct {
				ID                string        `xml:"id,attr" json:"Id"`
				Node              string        `xml:"node,attr"`
				Resource          string        `xml:"rsc,attr"`
				RscPattern        string        `xml:"rsc-pattern,attr"`
				Role              string        `xml:"role,attr"`
				Score             string        `xml:"score,attr"` // integer or INFINITY/+INFINITY/-INFINITY
				ResourceDiscovery string        `xml:"resource-discovery,attr"`
				Rules             []Rule        `xml:"rule"`
				ResourceSets      []ResourceSet `xml:"resource_set"`
				Lifetime          Lifetime      `xml:"lifetime"`
			} `xml:"rsc_location"`
			RscColocations []struct {
				ID               string        `xml:"id,attr" json:"Id"`
				Score            string        `xml:"score,attr"` // integer or INFINITY/+INFINITY/-INFINITY
				Influence        string        `xml:"influence,attr"`
				Resource         string        `xml:"rsc,attr"`
				WithResource     string        `xml:"with-rsc,attr"`
				NodeAttribute    string        `xml:"node-attribute,attr"`
				ResourceRole     string        `xml:"rsc-role,attr"`
				WithResourceRole string        `xml:"with-rsc-role,attr"`
				ResourceSets     []ResourceSet `xml:"resource_set"`
				Lifetime         Lifetime      `xml:"lifetime"`
			} `xml:"rsc_colocation"`
			RscOrders []struct {
				ID           string        `xml:"id,attr" json:"Id"`
				Score        string        `xml:"score,attr"`
				Kind         string        `xml:"kind,attr"` // Optional, Mandatory, or Serialize
				Symmetrical  bool          `xml:"symmetrical,attr"`
				RequireAll   bool          `xml:"require-all,attr"`
				First        string        `xml:"first,attr"`
				FirstAction  string        `xml:"first-action,attr"` // start, promote, demote, or stop
				Then         string        `xml:"then,attr"`
				ThenAction   string        `xml:"then-action,attr"` // start, promote, demote, or stop
				ResourceSets []ResourceSet `xml:"resource_set"`
				Lifetime     Lifetime      `xml:"lifetime"`
			} `xml:"rsc_order"`
			RscTickets []struct {
				ID           string        `xml:"id,attr" json:"Id"`
				Resource     string        `xml:"rsc,attr"`
				ResourceRole string        `xml:"rsc-role,attr"`
				Ticket       string        `xml:"ticket,attr"`
				LossPolicy   string        `xml:"loss-policy,attr"` // stop, demote, fence, or freeze
				ResourceSets []ResourceSet `xml:"resource_set"`
			} `xml:"rsc_ticket"`
		} `xml:"constraints"`
		// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/nvset-3.9.rng
		RscDefaults struct {
			MetaAttributes []Attribute `xml:"meta_attributes>nvpair"`
		} `xml:"rsc_defaults"`
		// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/nvset-3.9.rng
		OpDefaults struct {
			MetaAttributes []Attribute `xml:"meta_attributes>nvpair"`
		} `xml:"op_defaults"`
	} `xml:"configuration"`
}

// Attribute is an nvpair element.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/nvset-3.9.rng
type Attribute struct {
	ID    string `xml:"id,attr" json:"Id"`
	IDRef string `xml:"id-ref,attr"`
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// ResourceSet is used in set-based constraints (location, colocation, ordering, ticket).
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/constraints-3.9.rng
type ResourceSet struct {
	ID           string `xml:"id,attr" json:"Id"`
	IDRef        string `xml:"id-ref,attr"`
	Sequential   bool   `xml:"sequential,attr"`
	RequireAll   bool   `xml:"require-all,attr"`
	Ordering     string `xml:"ordering,attr"` // group or listed
	Action       string `xml:"action,attr"`   // start, promote, demote, or stop
	Role         string `xml:"role,attr"`
	Score        string `xml:"score,attr"` // integer or INFINITY/+INFINITY/-INFINITY
	Kind         string `xml:"kind,attr"`  // Optional, Mandatory, or Serialize
	ResourceRefs []struct {
		ID string `xml:"id,attr" json:"Id"`
	} `xml:"resource_ref"`
}

// Lifetime bounds a constraint to a time window.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/constraints-3.9.rng
type Lifetime struct {
	Rules []Rule `xml:"rule"`
}

// Rule is a conditional expression used in location constraints and lifetime elements.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/rule-3.9.rng
type Rule struct {
	ID             string `xml:"id,attr" json:"Id"`
	IDRef          string `xml:"id-ref,attr"`
	Role           string `xml:"role,attr"`  // Promoted, Unpromoted, Started, or Stopped
	Score          string `xml:"score,attr"` // integer or INFINITY; mutually exclusive with ScoreAttribute
	ScoreAttribute string `xml:"score-attribute,attr"`
	BooleanOp      string `xml:"boolean-op,attr"` // "and" or "or"
	Expressions    []struct {
		ID          string `xml:"id,attr" json:"Id"`
		Attribute   string `xml:"attribute,attr"`
		Operation   string `xml:"operation,attr"` // lt, gt, lte, gte, eq, ne, defined, not_defined
		Value       string `xml:"value,attr"`
		ValueSource string `xml:"value-source,attr"` // literal, param, or meta
		Type        string `xml:"type,attr"`         // string, integer, number, or version
	} `xml:"expression"`
	RscExpressions []struct {
		ID       string `xml:"id,attr" json:"Id"`
		Class    string `xml:"class,attr"`
		Provider string `xml:"provider,attr"`
		Type     string `xml:"type,attr"`
	} `xml:"rsc_expression"`
	OpExpressions []struct {
		ID       string `xml:"id,attr" json:"Id"`
		Name     string `xml:"name,attr"`
		Interval string `xml:"interval,attr"`
	} `xml:"op_expression"`
	DateExpressions []struct {
		ID        string `xml:"id,attr" json:"Id"`
		Operation string `xml:"operation,attr"` // in_range, date, gt, or lt
		Start     string `xml:"start,attr"`
		End       string `xml:"end,attr"`
		DateSpec  struct {
			ID        string `xml:"id,attr" json:"Id"`
			Hours     string `xml:"hours,attr"`
			Minutes   string `xml:"minutes,attr"`
			Seconds   string `xml:"seconds,attr"`
			Monthdays string `xml:"monthdays,attr"`
			Weekdays  string `xml:"weekdays,attr"`
			Yeardays  string `xml:"yeardays,attr"`
			Yearsdays string `xml:"yearsdays,attr"`
			Months    string `xml:"months,attr"`
			Weeks     string `xml:"weeks,attr"`
			Years     string `xml:"years,attr"`
			Weekyears string `xml:"weekyears,attr"`
			Moon      string `xml:"moon,attr"`
		} `xml:"date_spec"`
		Duration struct {
			ID        string `xml:"id,attr" json:"Id"`
			Hours     string `xml:"hours,attr"`
			Minutes   string `xml:"minutes,attr"`
			Seconds   string `xml:"seconds,attr"`
			Days      string `xml:"days,attr"`
			Weeks     string `xml:"weeks,attr"`
			Months    string `xml:"months,attr"`
			Monthdays string `xml:"monthdays,attr"`
			Weekdays  string `xml:"weekdays,attr"`
			Years     string `xml:"years,attr"`
			Yearsdays string `xml:"yearsdays,attr"`
			Weekyears string `xml:"weekyears,attr"`
			Moon      string `xml:"moon,attr"`
		} `xml:"duration"` // used for operation="in_range"
	} `xml:"date_expression"`
	Rules []Rule `xml:"rule"` // nested rules
}

// Primitive is a simple (non-grouped, non-cloned) resource or resource template in the CIB.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/resources-3.9.rng
type Primitive struct {
	ID                 string      `xml:"id,attr" json:"Id"`
	Class              string      `xml:"class,attr"`
	Type               string      `xml:"type,attr"`
	Provider           string      `xml:"provider,attr"`
	Template           string      `xml:"template,attr"`
	Description        string      `xml:"description,attr"`
	InstanceAttributes []Attribute `xml:"instance_attributes>nvpair"`
	MetaAttributes     []Attribute `xml:"meta_attributes>nvpair"`
	Utilization        []Attribute `xml:"utilization>nvpair"`
	Operations         []struct {
		ID                 string      `xml:"id,attr" json:"Id"`
		Name               string      `xml:"name,attr"`
		Description        string      `xml:"description,attr"`
		Role               string      `xml:"role,attr"`
		Enabled            bool        `xml:"enabled,attr"`
		RecordPending      bool        `xml:"record-pending,attr"`
		OnFail             string      `xml:"on-fail,attr"`
		Interval           string      `xml:"interval,attr"`
		Timeout            string      `xml:"timeout,attr"`
		StartDelay         string      `xml:"start-delay,attr"`     // mutually exclusive with IntervalOrigin
		IntervalOrigin     string      `xml:"interval-origin,attr"` // mutually exclusive with StartDelay
		MetaAttributes     []Attribute `xml:"meta_attributes>nvpair"`
		InstanceAttributes []Attribute `xml:"instance_attributes>nvpair"`
	} `xml:"operations>op"`
}

// Clone is a cloned or promotable-clone resource. Pacemaker 2.x used <master> for promotable clones;
// Pacemaker 3.x dropped <master> and uses <clone> with promotable="true" in meta_attributes instead.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/resources-3.9.rng
type Clone struct {
	ID                 string      `xml:"id,attr" json:"Id"`
	Description        string      `xml:"description,attr"`
	MetaAttributes     []Attribute `xml:"meta_attributes>nvpair"`
	InstanceAttributes []Attribute `xml:"instance_attributes>nvpair"`
	Primitive          Primitive   `xml:"primitive"`
	Group              Group       `xml:"group"`
}

// BundleContainer holds the attributes common to <docker>, <rkt>, and <podman> sub-elements.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/resources-3.9.rng
type BundleContainer struct {
	Image           string `xml:"image,attr"`
	Replicas        int    `xml:"replicas,attr"`
	ReplicasPerHost int    `xml:"replicas-per-host,attr"`
	PromotedMax     int    `xml:"promoted-max,attr"`
	Masters         int    `xml:"masters,attr"` // deprecated alias for promoted-max (Pacemaker 2.x)
	RunCommand      string `xml:"run-command,attr"`
	Network         string `xml:"network,attr"`
	Options         string `xml:"options,attr"`
}

// Group is a set of primitives that start and stop together in order.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/resources-3.9.rng
type Group struct {
	ID                 string      `xml:"id,attr" json:"Id"`
	Description        string      `xml:"description,attr"`
	MetaAttributes     []Attribute `xml:"meta_attributes>nvpair"`
	InstanceAttributes []Attribute `xml:"instance_attributes>nvpair"`
	Primitives         []Primitive `xml:"primitive"`
}
