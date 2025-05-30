package agent_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/trento-project/agent/cmd"
)

type ConfigTestSuite struct {
	suite.Suite
	cmd *cobra.Command
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (suite *ConfigTestSuite) SetupTest() {
	os.Clearenv()

	cmd := cmd.NewRootCmd()

	for _, command := range cmd.Commands() {
		command.Run = func(_ *cobra.Command, _ []string) {
			// do nothing
		}
	}

	var b bytes.Buffer
	cmd.SetOut(&b)

	suite.cmd = cmd
}

func (suite *ConfigTestSuite) TearDownTest() {
	viper.Reset()
}

func (suite *ConfigTestSuite) TestLoadingDefaultLogLevel() {
	suite.cmd.SetArgs([]string{
		"start",
	})
	_ = suite.cmd.Execute()

	defaultLogLevel := viper.GetString("log-level")
	suite.Equal("info", defaultLogLevel)
}

func (suite *ConfigTestSuite) TestOverridesLogLevelFromArgs() {
	defaultLogLevel := viper.GetString("log-level")
	suite.Equal("info", defaultLogLevel)

	suite.cmd.SetArgs([]string{
		"--log-level",
		"error",
		"start",
	})

	_ = suite.cmd.Execute()

	overriddenLogLevel := viper.GetString("log-level")
	suite.Equal("error", overriddenLogLevel)
}

func (suite *ConfigTestSuite) TestLoadingCorrectLevelFromConfigFile() {
	defaultLogLevel := viper.GetString("log-level")
	suite.Equal("info", defaultLogLevel)

	suite.cmd.SetArgs([]string{
		"--config",
		"../../test/fixtures/config/agent.yaml",
		"start",
	})

	_ = suite.cmd.Execute()

	overriddenLogLevel := viper.GetString("log-level")
	suite.Equal("warning", overriddenLogLevel)
}
