// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package sapcontrolapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hooklift/gowsdl/soap"
	"github.com/stretchr/testify/require"
)

// requestRecorder captures the last request received by the fake SOAP server, populated
// only once the call under test has actually run.
type requestRecorder struct {
	Request *http.Request
}

// newTestWebService starts a fake SOAP server that always answers with the given
// SOAP body (already wrapped in an envelope by soapEnvelope) and returns a
// webService pointed at it, along with a recorder that captures the last received
// request (valid only after the call under test has been made).
func newTestWebService(t *testing.T, responseBody string) (WebService, *requestRecorder) {
	t.Helper()

	recorder := &requestRecorder{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder.Request = r
		w.Header().Set("Content-Type", "text/xml; charset=utf-8")
		fmt.Fprint(w, responseBody)
	}))
	t.Cleanup(server.Close)

	client := soap.NewClient(server.URL, soap.WithHTTPClient(server.Client()))
	return &webService{client: client}, recorder
}

func soapEnvelope(body string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/">
  <SOAP-ENV:Body>` + body + `</SOAP-ENV:Body>
</SOAP-ENV:Envelope>`
}

func TestGetInstancePropertiesContext(t *testing.T) {
	body := `<ns1:GetInstancePropertiesResponse xmlns:ns1="urn:SAPControl">
		<properties>
			<item>
				<property>SAPSYSTEMNAME</property>
				<propertytype>string</propertytype>
				<value>DEV</value>
			</item>
		</properties>
	</ns1:GetInstancePropertiesResponse>`

	ws, _ := newTestWebService(t, soapEnvelope(body))

	resp, err := ws.GetInstancePropertiesContext(context.Background(), new(GetInstanceProperties))
	require.NoError(t, err)
	require.Len(t, resp.Properties, 1)
	require.Equal(t, "SAPSYSTEMNAME", resp.Properties[0].Property)
	require.Equal(t, "string", resp.Properties[0].Propertytype)
	require.Equal(t, "DEV", resp.Properties[0].Value)
}

func TestGetProcessListContext(t *testing.T) {
	body := `<ns1:GetProcessListResponse xmlns:ns1="urn:SAPControl">
		<process>
			<item>
				<name>msg_server</name>
				<description>MessageServer</description>
				<dispstatus>SAPControl-GREEN</dispstatus>
				<textstatus>Running</textstatus>
				<starttime>20240101 00:00:00</starttime>
				<elapsedtime>1:00:00</elapsedtime>
				<pid>123</pid>
			</item>
		</process>
	</ns1:GetProcessListResponse>`

	ws, _ := newTestWebService(t, soapEnvelope(body))

	resp, err := ws.GetProcessListContext(context.Background(), new(GetProcessList))
	require.NoError(t, err)
	require.Len(t, resp.Processes, 1)
	process := resp.Processes[0]
	require.Equal(t, "msg_server", process.Name)
	require.Equal(t, "MessageServer", process.Description)
	require.Equal(t, STATECOLOR_GREEN, process.Dispstatus)
	require.Equal(t, "Running", process.Textstatus)
	require.Equal(t, int32(123), process.Pid)
}

func TestGetSystemInstanceListContext(t *testing.T) {
	body := `<ns1:GetSystemInstanceListResponse xmlns:ns1="urn:SAPControl">
		<instance>
			<item>
				<hostname>host</hostname>
				<instanceNr>0</instanceNr>
				<httpPort>50013</httpPort>
				<httpsPort>50014</httpsPort>
				<startPriority>0.3</startPriority>
				<features>MESSAGESERVER|ENQUE</features>
				<dispstatus>SAPControl-GREEN</dispstatus>
			</item>
		</instance>
	</ns1:GetSystemInstanceListResponse>`

	ws, _ := newTestWebService(t, soapEnvelope(body))

	resp, err := ws.GetSystemInstanceListContext(context.Background(), new(GetSystemInstanceList))
	require.NoError(t, err)
	require.Len(t, resp.Instances, 1)
	instance := resp.Instances[0]
	require.Equal(t, "host", instance.Hostname)
	require.Equal(t, int32(50013), instance.HttpPort)
	require.Equal(t, int32(50014), instance.HttpsPort)
	require.Equal(t, "MESSAGESERVER|ENQUE", instance.Features)
	require.Equal(t, STATECOLOR_GREEN, instance.Dispstatus)
}

func TestGetVersionInfoContext(t *testing.T) {
	body := `<ns1:GetVersionInfoResponse xmlns:ns1="urn:SAPControl">
		<version>
			<item>
				<Filename>disp+work</Filename>
				<VersionInfo>disp+work information</VersionInfo>
				<Time>Jan 1 2024 00:00:00</Time>
			</item>
		</version>
	</ns1:GetVersionInfoResponse>`

	ws, _ := newTestWebService(t, soapEnvelope(body))

	resp, err := ws.GetVersionInfoContext(context.Background(), new(GetVersionInfo))
	require.NoError(t, err)
	require.Len(t, resp.InstanceVersions, 1)
	require.Equal(t, "disp+work", resp.InstanceVersions[0].Filename)
	require.Equal(t, "disp+work information", resp.InstanceVersions[0].VersionInfo)
}

// TestHACheckConfigContext exercises the exact response type whose CallContext call used
// to pass `&response` (a **HACheckConfigResponse) instead of `response`. It asserts that the
// fields are actually populated on the value returned to the caller.
func TestHACheckConfigContext(t *testing.T) {
	body := `<ns1:HACheckConfigResponse xmlns:ns1="urn:SAPControl">
		<check>
			<item>
				<state>SAPControl-HA-SUCCESS</state>
				<category>SAPControl-SAP-CONFIGURATION</category>
				<description>Basic check</description>
				<comment>All good</comment>
			</item>
		</check>
	</ns1:HACheckConfigResponse>`

	ws, _ := newTestWebService(t, soapEnvelope(body))

	resp, err := ws.HACheckConfigContext(context.Background(), new(HACheckConfig))
	require.NoError(t, err)
	require.Len(t, resp.Checks, 1)
	check := resp.Checks[0]
	require.NotNil(t, check.State)
	require.Equal(t, HAVerificationStateSAPControlHASUCCESS, *check.State)
	require.NotNil(t, check.Category)
	require.Equal(t, HACheckCategorySAPControlSAPCONFIGURATION, *check.Category)
	require.Equal(t, "Basic check", check.Description)
	require.Equal(t, "All good", check.Comment)
}

// TestHAGetFailoverConfigContext exercises the other response type whose CallContext call
// used to pass `&response` (a **HAGetFailoverConfigResponse) instead of `response`.
func TestHAGetFailoverConfigContext(t *testing.T) {
	body := `<ns1:HAGetFailoverConfigResponse xmlns:ns1="urn:SAPControl">
		<HAActive>true</HAActive>
		<HAProductVersion>1.0</HAProductVersion>
		<HASAPInterfaceVersion>2.0</HASAPInterfaceVersion>
		<HADocumentation>some documentation</HADocumentation>
		<HAActiveNode>node1</HAActiveNode>
		<HANodes>
			<item>node1</item>
			<item>node2</item>
		</HANodes>
	</ns1:HAGetFailoverConfigResponse>`

	ws, _ := newTestWebService(t, soapEnvelope(body))

	resp, err := ws.HAGetFailoverConfigContext(context.Background(), new(HAGetFailoverConfig))
	require.NoError(t, err)
	require.True(t, resp.HAActive)
	require.Equal(t, "1.0", resp.HAProductVersion)
	require.Equal(t, "2.0", resp.HASAPInterfaceVersion)
	require.Equal(t, "some documentation", resp.HADocumentation)
	require.Equal(t, "node1", resp.HAActiveNode)
	require.NotNil(t, resp.HANodes)
	require.Equal(t, []string{"node1", "node2"}, *resp.HANodes)
}

func TestStartContext(t *testing.T) {
	body := `<ns1:StartResponse xmlns:ns1="urn:SAPControl"></ns1:StartResponse>`
	ws, recorder := newTestWebService(t, soapEnvelope(body))

	_, err := ws.StartContext(context.Background(), &Start{Runlevel: "3"})
	require.NoError(t, err)
	require.Equal(t, "''", recorder.Request.Header.Get("SOAPAction"))
}

func TestStopContext(t *testing.T) {
	body := `<ns1:StopResponse xmlns:ns1="urn:SAPControl"></ns1:StopResponse>`
	ws, _ := newTestWebService(t, soapEnvelope(body))

	_, err := ws.StopContext(context.Background(), &Stop{Softtimeout: 30})
	require.NoError(t, err)
}

func TestStartSystemContext(t *testing.T) {
	body := `<ns1:StartSystemResponse xmlns:ns1="urn:SAPControl"></ns1:StartSystemResponse>`
	ws, _ := newTestWebService(t, soapEnvelope(body))

	_, err := ws.StartSystemContext(context.Background(), new(StartSystem))
	require.NoError(t, err)
}

func TestStopSystemContext(t *testing.T) {
	body := `<ns1:StopSystemResponse xmlns:ns1="urn:SAPControl"></ns1:StopSystemResponse>`
	ws, _ := newTestWebService(t, soapEnvelope(body))

	_, err := ws.StopSystemContext(context.Background(), new(StopSystem))
	require.NoError(t, err)
}

func TestWebServiceHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	client := soap.NewClient(server.URL, soap.WithHTTPClient(server.Client()))
	ws := &webService{client: client}

	_, err := ws.GetProcessListContext(context.Background(), new(GetProcessList))
	require.Error(t, err)
}
