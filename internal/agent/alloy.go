package agent

import (
	"embed"
	"fmt"
	"io"
	"text/template"
	"time"
)

//go:embed templates/trento.alloy.tmpl
var alloyTemplateFS embed.FS

const (
	alloyTemplatePath = "templates/trento.alloy.tmpl"

	AuthMethodNone   = "none"
	AuthMethodBasic  = "basic"
	AuthMethodBearer = "bearer"
	AuthMethodMTLS   = "mtls"

	DefaultAuthMethod     = AuthMethodBearer
	DefaultExporterName   = "grafana_alloy"
	DefaultScrapeInterval = 15 * time.Second
)

type AlloyConfig struct {
	AgentID         string
	PrometheusURL   string
	ScrapeInterval  time.Duration
	ExporterName    string
	AuthMethod      string
	AuthUsername    string
	AuthPassword    string
	AuthBearerToken string
	TLSCACert       string
	TLSClientCert   string
	TLSClientKey    string
}

func (c *AlloyConfig) Validate() error {
	if c.AgentID == "" {
		return fmt.Errorf("agent ID is required")
	}

	if c.PrometheusURL == "" {
		return fmt.Errorf("prometheus URL is required")
	}

	switch c.AuthMethod {
	case AuthMethodNone:
		// No additional validation needed
	case AuthMethodBasic:
		if c.AuthUsername == "" {
			return fmt.Errorf("username is required for basic authentication")
		}
		if c.AuthPassword == "" {
			return fmt.Errorf("password is required for basic authentication")
		}
	case AuthMethodBearer:
		if c.AuthBearerToken == "" {
			return fmt.Errorf("bearer token is required for bearer authentication")
		}
	case AuthMethodMTLS:
		if c.TLSClientCert == "" {
			return fmt.Errorf("client certificate is required for mTLS authentication")
		}
		if c.TLSClientKey == "" {
			return fmt.Errorf("client key is required for mTLS authentication")
		}
	default:
		return fmt.Errorf("invalid auth method: %s (valid values: none, basic, bearer, mtls)", c.AuthMethod)
	}

	return nil
}

func GenerateAlloyConfig(w io.Writer, config *AlloyConfig) error {
	// Apply defaults before validation
	if config.ScrapeInterval == 0 {
		config.ScrapeInterval = DefaultScrapeInterval
	}
	if config.ExporterName == "" {
		config.ExporterName = DefaultExporterName
	}
	if config.AuthMethod == "" {
		config.AuthMethod = DefaultAuthMethod
	}

	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid alloy configuration: %w", err)
	}

	tmplContent, err := alloyTemplateFS.ReadFile(alloyTemplatePath)
	if err != nil {
		return fmt.Errorf("failed to read alloy template: %w", err)
	}

	tmpl, err := template.New("alloy").Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("failed to parse alloy template: %w", err)
	}

	if err := tmpl.Execute(w, config); err != nil {
		return fmt.Errorf("failed to execute alloy template: %w", err)
	}

	return nil
}
