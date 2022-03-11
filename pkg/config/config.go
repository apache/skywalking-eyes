package config

import (
	"os"

	"github.com/apache/skywalking-eyes/assets"
	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/deps"
	"github.com/apache/skywalking-eyes/pkg/header"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Header header.ConfigHeader `yaml:"header"`
	Deps   deps.ConfigDeps     `yaml:"dependency"`
}

// Parse reads and parses the header check configurations in config file.
func (config *Config) Parse(file string) (err error) {
	var bytes []byte

	// attempt to read configuration from specified file
	logger.Log.Infoln("Loading configuration from file:", file)

	if bytes, err = os.ReadFile(file); err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		logger.Log.Infof("Config file %s does not exist, using the default config", file)

		if bytes, err = assets.Asset("default-config.yaml"); err != nil {
			return err
		}
	}

	if err := yaml.Unmarshal(bytes, config); err != nil {
		return err
	}

	if err := config.Header.Finalize(); err != nil {
		return err
	}

	return config.Deps.Finalize(file)
}
