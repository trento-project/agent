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
	Configuration struct {
		// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/nvset-3.9.rng
		CrmConfig struct {
			ClusterProperties []Attribute `xml:"cluster_property_set>nvpair"`
		} `xml:"crm_config"`
		// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/nodes-3.9.rng
		Nodes []struct {
			ID                 string      `xml:"id,attr" json:"Id"`
			Uname              string      `xml:"uname,attr"`
			InstanceAttributes []Attribute `xml:"instance_attributes>nvpair"`
		} `xml:"nodes>node"`
		// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/resources-3.9.rng
		Resources struct {
			Primitives []Primitive `xml:"primitive"`
			Masters    []Clone     `xml:"master"`
			Clones     []Clone     `xml:"clone"`
			Groups     []Group     `xml:"group"`
		} `xml:"resources"`
		// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/constraints-3.9.rng
		Constraints struct {
			RscLocations []struct {
				ID       string `xml:"id,attr" json:"Id"`
				Node     string `xml:"node,attr"`
				Resource string `xml:"rsc,attr"`
				Role     string `xml:"role,attr"`
				Score    string `xml:"score,attr"` // integer or INFINITY/+INFINITY/-INFINITY
			} `xml:"rsc_location"`
		} `xml:"constraints"`
	} `xml:"configuration"`
}

// Attribute is an nvpair element.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/nvset-3.9.rng
type Attribute struct {
	ID    string `xml:"id,attr" json:"Id"`
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// Primitive is a simple (non-grouped, non-cloned) resource or resource template in the CIB.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/resources-3.9.rng
type Primitive struct {
	ID                 string      `xml:"id,attr" json:"Id"`
	Class              string      `xml:"class,attr"`
	Type               string      `xml:"type,attr"`
	Provider           string      `xml:"provider,attr"`
	InstanceAttributes []Attribute `xml:"instance_attributes>nvpair"`
	MetaAttributes     []Attribute `xml:"meta_attributes>nvpair"`
	Operations         []struct {
		ID   string `xml:"id,attr" json:"Id"`
		Name string `xml:"name,attr"`
		Role string `xml:"role,attr"`
		// TODO: interval and timeout are time based vars. We should in future parse them correctly instead of string
		Interval string `xml:"interval,attr"`
		Timeout  string `xml:"timeout,attr"`
	} `xml:"operations>op"`
}

// Clone is a cloned or promotable-clone resource. Pacemaker 2.x used <master> for promotable clones;
// Pacemaker 3.x dropped <master> and uses <clone> with promotable="true" in meta_attributes instead.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/resources-3.9.rng
type Clone struct {
	ID             string      `xml:"id,attr" json:"Id"`
	MetaAttributes []Attribute `xml:"meta_attributes>nvpair"`
	Primitive      Primitive   `xml:"primitive"`
}

// Group is a set of primitives that start and stop together in order.
// Schema: https://github.com/ClusterLabs/pacemaker/blob/main/xml/resources-3.9.rng
type Group struct {
	ID         string      `xml:"id,attr" json:"Id"`
	Primitives []Primitive `xml:"primitive"`
}
