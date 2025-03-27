// generated
// nolint
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

//go:generate mockery --all

type WebService interface {
	GetInstanceProperties(ctx context.Context) (*GetInstancePropertiesResponse, error)
	GetProcessList(ctx context.Context) (*GetProcessListResponse, error)
	GetSystemInstanceList(ctx context.Context) (*GetSystemInstanceListResponse, error)
	GetVersionInfo(ctx context.Context) (*GetVersionInfoResponse, error)
	HACheckConfig(ctx context.Context) (*HACheckConfigResponse, error)
	HAGetFailoverConfig(ctx context.Context) (*HAGetFailoverConfigResponse, error)
}

type STATECOLOR string
type STATECOLOR_CODE int
type HAVerificationState string
type HACheckCategory string

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

	// NOTE: This was just copy-pasted from sap_host_exporter, not used right now
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
	Processes []*OSProcess `xml:"process>item,omitempty" json:"process>item,omitempty"`
}
type GetInstancePropertiesResponse struct {
	XMLName    xml.Name            `xml:"urn:SAPControl GetInstancePropertiesResponse"`
	Properties []*InstanceProperty `xml:"properties>item,omitempty" json:"properties>item,omitempty"`
}

type GetSystemInstanceList struct {
	XMLName xml.Name `xml:"urn:SAPControl GetSystemInstanceList"`
	Timeout int32    `xml:"timeout,omitempty" json:"timeout,omitempty"`
}

type GetSystemInstanceListResponse struct {
	XMLName   xml.Name       `xml:"urn:SAPControl GetSystemInstanceListResponse"`
	Instances []*SAPInstance `xml:"instance>item,omitempty" json:"instance>item,omitempty"`
}

type GetVersionInfo struct {
	XMLName xml.Name `xml:"urn:SAPControl GetVersionInfo"`
}

type GetVersionInfoResponse struct {
	XMLName          xml.Name       `xml:"urn:SAPControl GetVersionInfoResponse"`
	InstanceVersions []*VersionInfo `xml:"version>item,omitempty" json:"version>item,omitempty"`
}

type HACheckConfig struct {
	XMLName xml.Name `xml:"urn:SAPControl HACheckConfig"`
}

type HACheckConfigResponse struct {
	XMLName xml.Name   `xml:"urn:SAPControl HACheckConfigResponse"`
	Checks  []*HACheck `xml:"check>item,omitempty" json:"check>item,omitempty"`
}

type HAGetFailoverConfig struct {
	XMLName xml.Name `xml:"urn:SAPControl HAGetFailoverConfig"`
}

type HAGetFailoverConfigResponse struct {
	XMLName               xml.Name  `xml:"urn:SAPControl HAGetFailoverConfigResponse"`
	HAActive              bool      `xml:"HAActive,omitempty" json:"HAActive,omitempty"`
	HAProductVersion      string    `xml:"HAProductVersion,omitempty" json:"HAProductVersion,omitempty"`
	HASAPInterfaceVersion string    `xml:"HASAPInterfaceVersion,omitempty" json:"HASAPInterfaceVersion,omitempty"`
	HADocumentation       string    `xml:"HADocumentation,omitempty" json:"HADocumentation,omitempty"`
	HAActiveNode          string    `xml:"HAActiveNode,omitempty" json:"HAActiveNode,omitempty"`
	HANodes               *[]string `xml:"HANodes>item,omitempty" json:"HANodes>item,omitempty"`
}

type OSProcess struct {
	Name        string     `xml:"name,omitempty" json:"name,omitempty"`
	Description string     `xml:"description,omitempty" json:"description,omitempty"`
	Dispstatus  STATECOLOR `xml:"dispstatus,omitempty" json:"dispstatus,omitempty"`
	Textstatus  string     `xml:"textstatus,omitempty" json:"textstatus,omitempty"`
	Starttime   string     `xml:"starttime,omitempty" json:"starttime,omitempty"`
	Elapsedtime string     `xml:"elapsedtime,omitempty" json:"elapsedtime,omitempty"`
	Pid         int32      `xml:"pid,omitempty" json:"pid,omitempty"`
}

type InstanceProperty struct {
	Property     string `xml:"property,omitempty" json:"property,omitempty"`
	Propertytype string `xml:"propertytype,omitempty" json:"propertytype,omitempty"`
	Value        string `xml:"value,omitempty" json:"value,omitempty"`
}

type SAPInstance struct {
	Hostname      string     `xml:"hostname,omitempty" json:"hostname,omitempty"`
	InstanceNr    int32      `xml:"instanceNr,omitempty" json:"instanceNr"`
	HttpPort      int32      `xml:"httpPort,omitempty" json:"httpPort,omitempty"`
	HttpsPort     int32      `xml:"httpsPort,omitempty" json:"httpsPort,omitempty"`
	StartPriority string     `xml:"startPriority,omitempty" json:"startPriority,omitempty"`
	Features      string     `xml:"features,omitempty" json:"features,omitempty"`
	Dispstatus    STATECOLOR `xml:"dispstatus,omitempty" json:"dispstatus,omitempty"`
	// Added manually as a virtual field
	RunningLocally bool `json:"runningLocally"`
}

type VersionInfo struct {
	Filename    string `xml:"Filename,omitempty" json:"Filename,omitempty"`
	VersionInfo string `xml:"VersionInfo,omitempty" json:"VersionInfo,omitempty"`
	Time        string `xml:"Time,omitempty" json:"Time,omitempty"`
}

type HACheck struct {
	State       *HAVerificationState `xml:"state,omitempty" json:"state,omitempty"`
	Category    *HACheckCategory     `xml:"category,omitempty" json:"category,omitempty"`
	Description string               `xml:"description,omitempty" json:"description,omitempty"`
	Comment     string               `xml:"comment,omitempty" json:"comment,omitempty"`
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
func (s *webService) GetInstanceProperties(ctx context.Context) (*GetInstancePropertiesResponse, error) {
	request := &GetInstanceProperties{}
	response := &GetInstancePropertiesResponse{}
	err := s.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetProcessList returns a list of all processes directly started by the webservice
// according to the SAP start profile.
func (s *webService) GetProcessList(ctx context.Context) (*GetProcessListResponse, error) {
	request := &GetProcessList{}
	response := &GetProcessListResponse{}
	err := s.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetSystemInstanceList returns a list of all processes directly started by the webservice
// according to the SAP start profile.
func (s *webService) GetSystemInstanceList(ctx context.Context) (*GetSystemInstanceListResponse, error) {
	request := &GetSystemInstanceList{}
	response := &GetSystemInstanceListResponse{}
	err := s.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// GetVersionInfo returns a list version information for the most important files of the instance
func (s *webService) GetVersionInfo(ctx context.Context) (*GetVersionInfoResponse, error) {
	request := &GetVersionInfo{}
	response := &GetVersionInfoResponse{}
	err := s.client.CallContext(ctx, "''", request, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// HACheckConfig checks high availability configurration and status of the system
func (s *webService) HACheckConfig(ctx context.Context) (*HACheckConfigResponse, error) {
	request := &HACheckConfig{}
	response := &HACheckConfigResponse{}
	err := s.client.CallContext(ctx, "''", request, &response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// HAGetFailoverConfig returns HA failover third party information
func (s *webService) HAGetFailoverConfig(ctx context.Context) (*HAGetFailoverConfigResponse, error) {
	request := &HAGetFailoverConfig{}
	response := &HAGetFailoverConfigResponse{}
	err := s.client.CallContext(ctx, "''", request, &response)
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
