package cmd

import (
	"io/ioutil"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	jsonit  = jsoniter.ConfigCompatibleWithStandardLibrary
)

// AddAllSharedParameters add shared parameters to command
func AddAllSharedParameters(cmd *cobra.Command) {
	AddConfigParameter(cmd)
}

// AddConfigParameter add config file parameter to command
func AddConfigParameter(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&cfgFile, "config", "c", "", `Scenario config file.`)
}

func unmarshalConfigFile() (*config.Config, error) {
	if cfgFile == "" {
		return nil, errors.Errorf("No config file defined")
	}

	cfgJSON, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading config from file<%s>", cfgFile)
	}

	var cfg config.Config
	if err = jsonit.Unmarshal(cfgJSON, &cfg); err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal config from json")
	}

	return &cfg, nil
}
