package deps

import (
	"os"
	"path/filepath"
)

type ConfigDeps struct {
	Files []string `yaml:"files"`
}

func (config *ConfigDeps) Finalize(configFile string) error {
	configFileAbsPath, err := filepath.Abs(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for i, file := range config.Files {
		config.Files[i] = filepath.Join(filepath.Dir(configFileAbsPath), file)
	}

	return nil
}
