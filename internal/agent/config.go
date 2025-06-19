package agent

import (
	"log/slog"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// InitConfig intializes the config for the application
// If no config file is provided with the --config flag
// it will look for a config in following locations:
//
// Order represents priority
// /etc/trento/agent.yaml
// /usr/etc/trento/agent.yaml
// $HOME/.config/trento/agent.yaml
func InitConfig(configName string) error {
	bindEnv()

	viper.SetConfigType("yaml")

	cfgFile := viper.GetString("config")
	if cfgFile != "" {
		_, err := os.Stat(cfgFile)

		if err != nil {
			// if a config file has been explicitly provided by --config flag,
			// then we should break if that file does not exist
			return errors.Wrapf(err, "cannot load configuration file: %s", cfgFile)
		}

		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()

		if err != nil {
			return err
		}

		// if no configuration file was explicitly provided,
		// we should look for a config in the expected locations
		viper.AddConfigPath("/etc/trento/")
		viper.AddConfigPath("/usr/etc/trento/")
		viper.AddConfigPath(path.Join(home, ".config", "trento"))
		viper.SetConfigName(configName)
	}

	err := viper.ReadInConfig()

	if err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return err
		}
	}

	if configFile := viper.ConfigFileUsed(); configFile != "" {
		slog.Info("Using config file", "file", configFile)
	}

	return nil
}

func bindEnv() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.SetEnvPrefix("TRENTO")
	viper.AutomaticEnv() // read in environment variables that match
}
