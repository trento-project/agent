// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

// Simplified hand-crafted subset of the SAPControl web service interface.
// The full auto-generated reference is in _generated_wsdl.go (excluded from
// compilation via //go:build ignore — it is kept for reference only).

package sapcontrolapi

import (
	"context"
	"encoding/xml"
	"fmt"
	"net"
	"net/http"
	"path"
	"time"

	"github.com/hooklift/gowsdl/soap"
)

type WebService interface {
	GetInstancePropertiesContext(ctx context.Context, request *GetInstanceProperties) (*GetInstancePropertiesResponse, error) //nolint:lll
	GetProcessListContext(ctx context.Context, request *GetProcessList) (*GetProcessListResponse, error)
	GetSystemInstanceListContext(ctx context.Context, request *GetSystemInstanceList) (*GetSystemInstanceListResponse, error) //nolint:lll
	GetVersionInfoContext(ctx context.Context, request *GetVersionInfo) (*GetVersionInfoResponse, error)
	HACheckConfigContext(ctx context.Context, request *HACheckConfig) (*HACheckConfigResponse, error)
	HAGetFailoverConfigContext(ctx context.Context, request *HAGetFailoverConfig) (*HAGetFailoverConfigResponse, error)
	StartContext(ctx context.Context, request *Start) (*StartResponse, error)
	StopContext(ctx context.Context, request *Stop) (*StopResponse, error)
	StartSystemContext(ctx context.Context, request *StartSystem) (*StartSystemResponse, error)
	StopSystemContext(ctx context.Context, request *StopSystem) (*StopSystemResponse, error)
}

type STATECOLOR string
type STATECOLOR_CODE int //nolint:revive
type HAVerificationState string
type HACheckCategory string
type StartStopOption string

//nolint:revive
const (
	STATECOLOR_GRAY   STATECOLOR = "SAPControl-GRAY"
	STATECOLOR_GREEN  STATECOLOR = "SAPControl-GREEN"
	STATECOLOR_YELLOW STATECOLOR = "SAPControl-YELLOW"
	STATECOLOR_RED    STATECOLOR = "SAPControl-RED"

	HAVerificationStateSAPControlHASUCCESS HAVerificationState = "SAPControl-HA-SUCCESS"
	HAVerificationStateSAPControlHAWARNING HAVerificationState = "SAPControl-HA-WARNING"
	HAVerificationStateSAPControlHAERROR   HAVerificationState = "SAPControl-HA-ERROR"

	HACheckCategorySAPControlSAPCONFIGURATION HACheckCategory = "SAPControl-SAP-CONFIGURATION"
	HACheckCategorySAPControlSAPSTATE         HACheckCategory = "SAPControl-SAP-STATE"
	HACheckCategorySAPControlHACONFIGURATION  HACheckCategory = "SAPControl-HA-CONFIGURATION"
	HACheckCategorySAPControlHASTATE          HACheckCategory = "SAPControl-HA-STATE"

	StartStopOptionSAPControlALLINSTANCES      StartStopOption = "SAPControl-ALL-INSTANCES"
	StartStopOptionSAPControlSCSINSTANCES      StartStopOption = "SAPControl-SCS-INSTANCES"
	StartStopOptionSAPControlDIALOGINSTANCES   StartStopOption = "SAPControl-DIALOG-INSTANCES"
	StartStopOptionSAPControlABAPINSTANCES     StartStopOption = "SAPControl-ABAP-INSTANCES"
	StartStopOptionSAPControlJ2EEINSTANCES     StartStopOption = "SAPControl-J2EE-INSTANCES"
	StartStopOptionSAPControlPRIORITYLEVEL     StartStopOption = "SAPControl-PRIORITY-LEVEL"
	StartStopOptionSAPControlTREXINSTANCES     StartStopOption = "SAPControl-TREX-INSTANCES"
	StartStopOptionSAPControlENQREPINSTANCES   StartStopOption = "SAPControl-ENQREP-INSTANCES"
	StartStopOptionSAPControlHDBINSTANCES      StartStopOption = "SAPControl-HDB-INSTANCES"
	StartStopOptionSAPControlALLNOHDBINSTANCES StartStopOption = "SAPControl-ALLNOHDB-INSTANCES"

	// NOTE: This was just copy-pasted from sap_host_exporter, not used right now
	//nolint:lll
	// see: https://github.com/SUSE/sap_host_exporter/blob/68bbf2f1b490ab0efaa2dd7b878b778f07fba2ab/lib/sapcontrol/webservice.go#L42
	STATECOLOR_CODE_GRAY   STATECOLOR_CODE = 1
	STATECOLOR_CODE_GREEN  STATECOLOR_CODE = 2
	STATECOLOR_CODE_YELLOW STATECOLOR_CODE = 3
	STATECOLOR_CODE_RED    STATECOLOR_CODE = 4
)

type GetInstanceProperties struct {
	XMLName xml.Name `xml:"urn:SAPControl GetInstanceProperties"`
}

type GetProcessList struct {
	XMLName xml.Name `xml:"urn:SAPControl GetProcessList"`
}

type GetProcessListResponse struct {
	XMLName   xml.Name     `xml:"urn:SAPControl GetProcessListResponse"`
	Processes []*OSProcess `json:"process>item,omitempty"               xml:"process>item,omitempty"`
}
type GetInstancePropertiesResponse struct {
	XMLName    xml.Name            `xml:"urn:SAPControl GetInstancePropertiesResponse"`
	Properties []*InstanceProperty `json:"properties>item,omitempty"                   xml:"properties>item,omitempty"`
}

type GetSystemInstanceList struct {
	XMLName xml.Name `xml:"urn:SAPControl GetSystemInstanceList"`
	Timeout int32    `json:"timeout,omitempty"                   xml:"timeout,omitempty"`
}

type GetSystemInstanceListResponse struct {
	XMLName   xml.Name       `xml:"urn:SAPControl GetSystemInstanceListResponse"`
	Instances []*SAPInstance `json:"instance>item,omitempty"                     xml:"instance>item,omitempty"`
}

type GetVersionInfo struct {
	XMLName xml.Name `xml:"urn:SAPControl GetVersionInfo"`
}

type GetVersionInfoResponse struct {
	XMLName          xml.Name       `xml:"urn:SAPControl GetVersionInfoResponse"`
	InstanceVersions []*VersionInfo `json:"version>item,omitempty"               xml:"version>item,omitempty"`
}

type HACheckConfig struct {
	XMLName xml.Name `xml:"urn:SAPControl HACheckConfig"`
}

type HACheckConfigResponse struct {
	XMLName xml.Name   `xml:"urn:SAPControl HACheckConfigResponse"`
	Checks  []*HACheck `json:"check>item,omitempty"                xml:"check>item,omitempty"`
}

type HAGetFailoverConfig struct {
	XMLName xml.Name `xml:"urn:SAPControl HAGetFailoverConfig"`
}

type HAGetFailoverConfigResponse struct {
	XMLName               xml.Name  `xml:"urn:SAPControl HAGetFailoverConfigResponse"`
	HAActive              bool      `json:"HAActive,omitempty"                        xml:"HAActive,omitempty"`
	HAProductVersion      string    `json:"HAProductVersion,omitempty"                xml:"HAProductVersion,omitempty"`
	HASAPInterfaceVersion string    `json:"HASAPInterfaceVersion,omitempty"           xml:"HASAPInterfaceVersion,omitempty"` //nolint:lll
	HADocumentation       string    `json:"HADocumentation,omitempty"                 xml:"HADocumentation,omitempty"`
	HAActiveNode          string    `json:"HAActiveNode,omitempty"                    xml:"HAActiveNode,omitempty"`
	HANodes               *[]string `json:"HANodes>item,omitempty"                    xml:"HANodes>item,omitempty"`
}

type OSProcess struct {
	Name        string     `json:"name,omitempty"        xml:"name,omitempty"`
	Description string     `json:"description,omitempty" xml:"description,omitempty"`
	Dispstatus  STATECOLOR `json:"dispstatus,omitempty"  xml:"dispstatus,omitempty"`
	Textstatus  string     `json:"textstatus,omitempty"  xml:"textstatus,omitempty"`
	Starttime   string     `json:"starttime,omitempty"   xml:"starttime,omitempty"`
	Elapsedtime string     `json:"elapsedtime,omitempty" xml:"elapsedtime,omitempty"`
	Pid         int32      `json:"pid,omitempty"         xml:"pid,omitempty"`
}

type InstanceProperty struct {
	Property     string `json:"property,omitempty"     xml:"property,omitempty"`
	Propertytype string `json:"propertytype,omitempty" xml:"propertytype,omitempty"`
	Value        string `json:"value,omitempty"        xml:"value,omitempty"`
}

type SAPInstance struct {
	Hostname      string     `json:"hostname,omitempty"      xml:"hostname,omitempty"`
	InstanceNr    int32      `json:"instanceNr"              xml:"instanceNr,omitempty"` //nolint:revive
	HttpPort      int32      `json:"httpPort,omitempty"      xml:"httpPort,omitempty"`   //nolint:revive
	HttpsPort     int32      `json:"httpsPort,omitempty"     xml:"httpsPort,omitempty"`  //nolint:revive
	StartPriority string     `json:"startPriority,omitempty" xml:"startPriority,omitempty"`
	Features      string     `json:"features,omitempty"      xml:"features,omitempty"`
	Dispstatus    STATECOLOR `json:"dispstatus,omitempty"    xml:"dispstatus,omitempty"`
	// Added manually as a virtual field to identify if the instance belongs to
	// the currently discovered instance
	CurrentInstance bool `json:"currentInstance"`
}

type VersionInfo struct {
	Filename    string `json:"Filename,omitempty"    xml:"Filename,omitempty"`
	VersionInfo string `json:"VersionInfo,omitempty" xml:"VersionInfo,omitempty"`
	Time        string `json:"Time,omitempty"        xml:"Time,omitempty"`
}

// InstanceVersionInfo is the name used in the generated WSDL code for the same type.
type InstanceVersionInfo = VersionInfo

type HACheck struct {
	State       *HAVerificationState `json:"state,omitempty"       xml:"state,omitempty"`
	Category    *HACheckCategory     `json:"category,omitempty"    xml:"category,omitempty"`
	Description string               `json:"description,omitempty" xml:"description,omitempty"`
	Comment     string               `json:"comment,omitempty"     xml:"comment,omitempty"`
}

type Start struct {
	XMLName  xml.Name `xml:"urn:SAPControl Start"`
	Runlevel string   `json:"runlevel,omitempty"  xml:"runlevel,omitempty"`
}

type StartResponse struct {
}

type Stop struct {
	XMLName      xml.Name `xml:"urn:SAPControl Stop"`
	Softtimeout  int32    `json:"softtimeout,omitempty"  xml:"softtimeout,omitempty"`
	IsSystemStop int32    `json:"IsSystemStop,omitempty" xml:"IsSystemStop,omitempty"`
}

type StopResponse struct {
}

type StartSystem struct {
	XMLName       xml.Name         `xml:"urn:SAPControl StartSystem"`
	Options       *StartStopOption `json:"options,omitempty"         xml:"options,omitempty"`
	Prioritylevel string           `json:"prioritylevel,omitempty"   xml:"prioritylevel,omitempty"`
	Waittimeout   int32            `json:"waittimeout,omitempty"     xml:"waittimeout,omitempty"`
	Runlevel      string           `json:"runlevel,omitempty"        xml:"runlevel,omitempty"`
}

type StartSystemResponse struct {
}

type StopSystem struct {
	XMLName       xml.Name         `xml:"urn:SAPControl StopSystem"`
	Options       *StartStopOption `json:"options,omitempty"        xml:"options,omitempty"`
	Prioritylevel string           `json:"prioritylevel,omitempty"  xml:"prioritylevel,omitempty"`
	Softtimeout   int32            `json:"softtimeout,omitempty"    xml:"softtimeout,omitempty"`
	Waittimeout   int32            `json:"waittimeout,omitempty"    xml:"waittimeout,omitempty"`
}

type StopSystemResponse struct {
}

type webService struct {
	client *soap.Client
}

type WebServiceConnector interface {
	New(instanceNumber string) WebService
}

type WebServiceUnix struct{}

func (w WebServiceUnix) New(instanceNumber string) WebService {
	return NewWebServiceUnix(instanceNumber)
}

func NewWebServiceUnix(instNumber string) WebService {
	socket := path.Join("/tmp", fmt.Sprintf(".sapstream5%s13", instNumber))

	udsClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				d := net.Dialer{}

				return d.DialContext(ctx, "unix", socket)
			},
		},
	}

	// The url used here is just phony:
	// we need a well formed url to create the instance but the above DialContext function won't actually use it.
	client := soap.NewClient("http://unix", soap.WithHTTPClient(udsClient))

	return &webService{
		client: client,
	}
}

// GetInstanceProperties returns a list of available instance features and information how to get it.
func (service *webService) GetInstancePropertiesContext(
	ctx context.Context,
	request *GetInstanceProperties,
) (*GetInstancePropertiesResponse, error) {
	response := new(GetInstancePropertiesResponse)

	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetProcessList returns a list of all processes directly started by the webservice
// according to the SAP start profile.
func (service *webService) GetProcessListContext(
	ctx context.Context,
	request *GetProcessList,
) (*GetProcessListResponse, error) {
	response := new(GetProcessListResponse)

	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetSystemInstanceList returns a list of all processes directly started by the webservice
// according to the SAP start profile.
func (service *webService) GetSystemInstanceListContext(
	ctx context.Context,
	request *GetSystemInstanceList,
) (*GetSystemInstanceListResponse, error) {
	response := new(GetSystemInstanceListResponse)

	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetVersionInfo returns a list version information for the most important files of the instance.
func (service *webService) GetVersionInfoContext(
	ctx context.Context,
	request *GetVersionInfo,
) (*GetVersionInfoResponse, error) {
	response := new(GetVersionInfoResponse)

	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// HACheckConfig checks high availability configuration and status of the system.
func (service *webService) HACheckConfigContext(
	ctx context.Context,
	request *HACheckConfig,
) (*HACheckConfigResponse, error) {
	response := new(HACheckConfigResponse)

	err := service.client.CallContext(ctx, "''", request, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// HAGetFailoverConfig returns HA failover third party information.
func (service *webService) HAGetFailoverConfigContext(
	ctx context.Context,
	request *HAGetFailoverConfig,
) (*HAGetFailoverConfigResponse, error) {
	response := new(HAGetFailoverConfigResponse)

	err := service.client.CallContext(ctx, "''", request, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// StartContext starts a SAP instance.
func (service *webService) StartContext(
	ctx context.Context,
	request *Start,
) (*StartResponse, error) {
	response := new(StartResponse)

	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// StopContext stops a SAP instance.
func (service *webService) StopContext(
	ctx context.Context,
	request *Stop,
) (*StopResponse, error) {
	response := new(StopResponse)

	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// StartSystemContext starts a SAP system.
func (service *webService) StartSystemContext(
	ctx context.Context,
	request *StartSystem,
) (*StartSystemResponse, error) {
	response := new(StartSystemResponse)

	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// StopSystemContext stops a SAP system.
func (service *webService) StopSystemContext(
	ctx context.Context,
	request *StopSystem,
) (*StopSystemResponse, error) {
	response := new(StopSystemResponse)

	err := service.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func DispstatusCodeFromStr(state STATECOLOR) STATECOLOR_CODE {
	return map[STATECOLOR]STATECOLOR_CODE{
		STATECOLOR_GRAY:   STATECOLOR_CODE_GRAY,
		STATECOLOR_GREEN:  STATECOLOR_CODE_GREEN,
		STATECOLOR_YELLOW: STATECOLOR_CODE_YELLOW,
		STATECOLOR_RED:    STATECOLOR_CODE_RED,
	}[state]
}
