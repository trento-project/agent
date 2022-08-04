package internal

import (
	"os"
	"path"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/trento-project/agent/internal/utils"
)

// InitConfig intializes the config for the application
// If no config file is provided with the --config flag
// it will look for a config in following locations:
//
// ${context} being one of the supported components: agent|web|runner
//
// Order represents priority
// /etc/trento/${context}.yaml
// /usr/etc/trento/${context}.yaml
// $HOME/.config/trento/${context}.yaml
func InitConfig(configName string) error {
	BindEnv()

	viper.SetConfigType("yaml")
	utils.SetLogLevel(viper.GetString("log-level"))
	utils.SetLogFormatter("2006-01-02 15:04:05")

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
		log.Infof("Using config file: %s", configFile)
	}

	return nil
}

func BindEnv() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.SetEnvPrefix("TRENTO")
	viper.AutomaticEnv() // read in environment variables that match
}
