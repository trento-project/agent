// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package cib

/*
The Cluster Information Base (Root)
is an XML representation of the cluster’s configuration
and the state of all nodes and resources.
The Root manager (pacemaker-based) keeps the Root synchronized across the cluster, and handles requests to modify it.

https://clusterlabs.org/pacemaker/doc/en-US/Pacemaker/2.0/html-single/Pacemaker_Administration/index.html

*/

type Root struct {
	Configuration struct {
		CrmConfig struct {
			ClusterProperties []Attribute `xml:"cluster_property_set>nvpair"`
		} `xml:"crm_config"`
		Nodes []struct {
			ID                 string      `xml:"id,attr" json:"Id"`
			Uname              string      `xml:"uname,attr"`
			InstanceAttributes []Attribute `xml:"instance_attributes>nvpair"`
		} `xml:"nodes>node"`
		Resources struct {
			Primitives []Primitive `xml:"primitive"`
			Masters    []Clone     `xml:"master"`
			Clones     []Clone     `xml:"clone"`
			Groups     []Group     `xml:"group"`
		} `xml:"resources"`
		Constraints struct {
			RscLocations []struct {
				ID       string `xml:"id,attr" json:"Id"`
				Node     string `xml:"node,attr"`
				Resource string `xml:"rsc,attr"`
				Role     string `xml:"role,attr"`
				Score    string `xml:"score,attr"`
			} `xml:"rsc_location"`
		} `xml:"constraints"`
	} `xml:"configuration"`
}

type Attribute struct {
	ID    string `xml:"id,attr" json:"Id"`
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

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
		// todo: interval and timeout are time based vars. We should in future parse them correctly instead of string
		Interval string `xml:"interval,attr"`
		Timeout  string `xml:"timeout,attr"`
	} `xml:"operations>op"`
}

type Clone struct {
	ID             string      `xml:"id,attr" json:"Id"`
	MetaAttributes []Attribute `xml:"meta_attributes>nvpair"`
	Primitive      Primitive   `xml:"primitive"`
}

type Group struct {
	ID         string      `xml:"id,attr" json:"Id"`
	Primitives []Primitive `xml:"primitive"`
}
