package agent_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trento-project/agent/internal/agent"
)

func TestAlloyConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *agent.AlloyConfig
		expectError string
	}{
		{
			name:        "missing agent ID",
			config:      &agent.AlloyConfig{},
			expectError: "agent ID is required",
		},
		{
			name: "missing prometheus URL",
			config: &agent.AlloyConfig{
				AgentID: "test-agent-id",
			},
			expectError: "prometheus URL is required",
		},
		{
			name: "missing bearer token for bearer auth",
			config: &agent.AlloyConfig{
				AgentID:       "test-agent-id",
				PrometheusURL: "https://prometheus.example.com/api/v1/write",
				AuthMethod:    agent.AuthMethodBearer,
			},
			expectError: "bearer token is required for bearer authentication",
		},
		{
			name: "missing username for basic auth",
			config: &agent.AlloyConfig{
				AgentID:       "test-agent-id",
				PrometheusURL: "https://prometheus.example.com/api/v1/write",
				AuthMethod:    agent.AuthMethodBasic,
			},
			expectError: "username is required for basic authentication",
		},
		{
			name: "missing password for basic auth",
			config: &agent.AlloyConfig{
				AgentID:       "test-agent-id",
				PrometheusURL: "https://prometheus.example.com/api/v1/write",
				AuthMethod:    agent.AuthMethodBasic,
				AuthUsername:  "user",
			},
			expectError: "password is required for basic authentication",
		},
		{
			name: "missing client cert for mtls",
			config: &agent.AlloyConfig{
				AgentID:       "test-agent-id",
				PrometheusURL: "https://prometheus.example.com/api/v1/write",
				AuthMethod:    agent.AuthMethodMTLS,
			},
			expectError: "client certificate is required for mTLS authentication",
		},
		{
			name: "missing client key for mtls",
			config: &agent.AlloyConfig{
				AgentID:       "test-agent-id",
				PrometheusURL: "https://prometheus.example.com/api/v1/write",
				AuthMethod:    agent.AuthMethodMTLS,
				TLSClientCert: "/path/to/cert",
			},
			expectError: "client key is required for mTLS authentication",
		},
		{
			name: "invalid auth method",
			config: &agent.AlloyConfig{
				AgentID:       "test-agent-id",
				PrometheusURL: "https://prometheus.example.com/api/v1/write",
				AuthMethod:    "invalid",
			},
			expectError: "invalid auth method: invalid",
		},
		{
			name: "valid config with bearer auth",
			config: &agent.AlloyConfig{
				AgentID:         "test-agent-id",
				PrometheusURL:   "https://prometheus.example.com/api/v1/write",
				AuthMethod:      agent.AuthMethodBearer,
				AuthBearerToken: "my-token",
			},
			expectError: "",
		},
		{
			name: "valid config with basic auth",
			config: &agent.AlloyConfig{
				AgentID:       "test-agent-id",
				PrometheusURL: "https://prometheus.example.com/api/v1/write",
				AuthMethod:    agent.AuthMethodBasic,
				AuthUsername:  "user",
				AuthPassword:  "password",
			},
			expectError: "",
		},
		{
			name: "valid config with mtls",
			config: &agent.AlloyConfig{
				AgentID:       "test-agent-id",
				PrometheusURL: "https://prometheus.example.com/api/v1/write",
				AuthMethod:    agent.AuthMethodMTLS,
				TLSClientCert: "/path/to/cert",
				TLSClientKey:  "/path/to/key",
			},
			expectError: "",
		},
		{
			name: "valid config with no auth",
			config: &agent.AlloyConfig{
				AgentID:       "test-agent-id",
				PrometheusURL: "https://prometheus.example.com/api/v1/write",
				AuthMethod:    agent.AuthMethodNone,
			},
			expectError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGenerateAlloyConfigBearerAuth(t *testing.T) {
	config := &agent.AlloyConfig{
		AgentID:         "test-agent-id",
		PrometheusURL:   "https://prometheus.example.com/api/v1/write",
		ScrapeInterval:  30 * time.Second,
		AuthMethod:      agent.AuthMethodBearer,
		AuthBearerToken: "my-secret-token",
	}

	var buf bytes.Buffer
	err := agent.GenerateAlloyConfig(&buf, config)
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, `Agent ID: test-agent-id`)
	assert.Contains(t, output, `replacement  = "test-agent-id"`)
	assert.Contains(t, output, `url = "https://prometheus.example.com/api/v1/write"`)
	assert.Contains(t, output, `scrape_interval = "30s"`)
	assert.Contains(t, output, `bearer_token = "my-secret-token"`)
	assert.NotContains(t, output, "basic_auth")
	assert.NotContains(t, output, "cert_file")
}

func TestGenerateAlloyConfigBasicAuth(t *testing.T) {
	config := &agent.AlloyConfig{
		AgentID:        "test-agent-id",
		PrometheusURL:  "https://prometheus.example.com/api/v1/write",
		ScrapeInterval: 15 * time.Second,
		AuthMethod:     agent.AuthMethodBasic,
		AuthUsername:   "myuser",
		AuthPassword:   "mypassword",
	}

	var buf bytes.Buffer
	err := agent.GenerateAlloyConfig(&buf, config)
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, `basic_auth {`)
	assert.Contains(t, output, `username = "myuser"`)
	assert.Contains(t, output, `password = "mypassword"`)
	assert.NotContains(t, output, "bearer_token")
}

func TestGenerateAlloyConfigMTLS(t *testing.T) {
	config := &agent.AlloyConfig{
		AgentID:       "test-agent-id",
		PrometheusURL: "https://prometheus.example.com/api/v1/write",
		AuthMethod:    agent.AuthMethodMTLS,
		TLSCACert:     "/etc/certs/ca.crt",
		TLSClientCert: "/etc/certs/client.crt",
		TLSClientKey:  "/etc/certs/client.key",
	}

	var buf bytes.Buffer
	err := agent.GenerateAlloyConfig(&buf, config)
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, `ca_file = "/etc/certs/ca.crt"`)
	assert.Contains(t, output, `cert_file = "/etc/certs/client.crt"`)
	assert.Contains(t, output, `key_file  = "/etc/certs/client.key"`)
	assert.NotContains(t, output, "bearer_token")
	assert.NotContains(t, output, "basic_auth")
}

func TestGenerateAlloyConfigNoAuth(t *testing.T) {
	config := &agent.AlloyConfig{
		AgentID:       "test-agent-id",
		PrometheusURL: "https://prometheus.example.com/api/v1/write",
		AuthMethod:    agent.AuthMethodNone,
	}

	var buf bytes.Buffer
	err := agent.GenerateAlloyConfig(&buf, config)
	require.NoError(t, err)

	output := buf.String()

	assert.NotContains(t, output, "bearer_token")
	assert.NotContains(t, output, "basic_auth")
	assert.NotContains(t, output, "cert_file")
	assert.NotContains(t, output, "key_file")
}

func TestGenerateAlloyConfigWithTLSCACert(t *testing.T) {
	config := &agent.AlloyConfig{
		AgentID:         "test-agent-id",
		PrometheusURL:   "https://prometheus.example.com/api/v1/write",
		AuthMethod:      agent.AuthMethodBearer,
		AuthBearerToken: "token",
		TLSCACert:       "/etc/certs/ca.crt",
	}

	var buf bytes.Buffer
	err := agent.GenerateAlloyConfig(&buf, config)
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, `ca_file = "/etc/certs/ca.crt"`)
}

func TestGenerateAlloyConfigDefaultScrapeInterval(t *testing.T) {
	config := &agent.AlloyConfig{
		AgentID:        "test-agent-id",
		PrometheusURL:  "https://prometheus.example.com/api/v1/write",
		AuthMethod:     agent.AuthMethodNone,
		ScrapeInterval: 0, // zero duration should default
	}

	var buf bytes.Buffer
	err := agent.GenerateAlloyConfig(&buf, config)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `scrape_interval = "15s"`)
}

func TestGenerateAlloyConfigDefaultAuthMethod(t *testing.T) {
	config := &agent.AlloyConfig{
		AgentID:         "test-agent-id",
		PrometheusURL:   "https://prometheus.example.com/api/v1/write",
		AuthMethod:      "", // empty should default to bearer
		AuthBearerToken: "my-token",
	}

	var buf bytes.Buffer
	err := agent.GenerateAlloyConfig(&buf, config)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `bearer_token = "my-token"`)
}

func TestGenerateAlloyConfigCustomExporterName(t *testing.T) {
	config := &agent.AlloyConfig{
		AgentID:       "test-agent-id",
		PrometheusURL: "https://prometheus.example.com/api/v1/write",
		AuthMethod:    agent.AuthMethodNone,
		ExporterName:  "my_custom_exporter",
	}

	var buf bytes.Buffer
	err := agent.GenerateAlloyConfig(&buf, config)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `replacement  = "my_custom_exporter"`)
	assert.NotContains(t, output, "grafana_alloy")
}

func TestGenerateAlloyConfigContainsExpectedSections(t *testing.T) {
	config := &agent.AlloyConfig{
		AgentID:       "test-agent-id",
		PrometheusURL: "https://prometheus.example.com/api/v1/write",
		AuthMethod:    agent.AuthMethodNone,
	}

	var buf bytes.Buffer
	err := agent.GenerateAlloyConfig(&buf, config)
	require.NoError(t, err)

	output := buf.String()

	// Check for expected components
	assert.Contains(t, output, `prometheus.exporter.unix "trento_system"`)
	assert.Contains(t, output, `discovery.relabel "trento_system"`)
	assert.Contains(t, output, `prometheus.scrape "trento_system"`)
	assert.Contains(t, output, `prometheus.remote_write "trento"`)

	// Check for expected collectors
	for _, collector := range []string{"cpu", "cpufreq", "loadavg", "meminfo", "filesystem", "netdev", "uname"} {
		assert.Contains(t, output, `"`+collector+`"`)
	}

	// Check for expected labels
	assert.Contains(t, output, `target_label = "agentID"`)
	assert.Contains(t, output, `target_label = "exporter_name"`)
	assert.Contains(t, output, `replacement  = "grafana_alloy"`) // default exporter name

	// Check header
	assert.True(t, strings.HasPrefix(output, "// ============================================================================="))
	assert.Contains(t, output, "Generated by: trento-agent generate alloy")
}
