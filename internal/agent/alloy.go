// SPDX-FileCopyrightText: SUSE LLC
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"embed"
	"errors"
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

	DefaultAuthMethod     = AuthMethodBasic
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
		return errors.New("agent ID is required")
	}

	if c.PrometheusURL == "" {
		return errors.New("prometheus URL is required")
	}

	switch c.AuthMethod {
	case AuthMethodNone:
		// No additional validation needed
	case AuthMethodBasic:
		if c.AuthUsername == "" {
			return errors.New("username is required for basic authentication")
		}

		if c.AuthPassword == "" {
			return errors.New("password is required for basic authentication")
		}
	case AuthMethodBearer:
		if c.AuthBearerToken == "" {
			return errors.New("bearer token is required for bearer authentication")
		}
	case AuthMethodMTLS:
		if c.TLSClientCert == "" {
			return errors.New("client certificate is required for mTLS authentication")
		}

		if c.TLSClientKey == "" {
			return errors.New("client key is required for mTLS authentication")
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

	err := config.Validate()
	if err != nil {
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

	err = tmpl.Execute(w, config)
	if err != nil {
		return fmt.Errorf("failed to execute alloy template: %w", err)
	}

	return nil
}
