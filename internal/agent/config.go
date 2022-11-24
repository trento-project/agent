package agent

import (
	"os"
	"path"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
	bindEnv()

	viper.SetConfigType("yaml")
	setLogLevel(viper.GetString("log-level"))
	setLogFormatter("2006-01-02 15:04:05")

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

func bindEnv() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.SetEnvPrefix("TRENTO")
	viper.AutomaticEnv() // read in environment variables that match
}

func setLogLevel(level string) {
	switch level {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.Warnln("Unrecognized minimum log level; using 'info' as default")
		log.SetLevel(log.InfoLevel)
	}
	hclog.DefaultOptions.Level = hclog.LevelFromString(level)
}

func setLogFormatter(timestampFormat string) {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = timestampFormat
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	hclog.DefaultOptions.TimeFormat = timestampFormat
}
