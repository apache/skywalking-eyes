// Copyright © 2020 Hoshea Jiang <hoshea@apache.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package header

import (
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"license-checker/internal/logger"
	"license-checker/pkg/header"
	"strings"
)

var (
	// cfgFile is the config path to the config file of header command.
	cfgFile string
)

var CheckCommand = &cobra.Command{
	Use:     "check",
	Long:    "`check` command walks the specified paths recursively and checks if the specified files have the license header in the config file.",
	Aliases: []string{"c"},
	RunE: func(cmd *cobra.Command, args []string) error {
		var config header.Config
		if err := loadConfig(&config); err != nil {
			return err
		}
		return header.Check(&config)
	},
}

func init() {
	CheckCommand.Flags().StringVarP(&cfgFile, "config", "c", ".licenserc.yaml", "the config file")
}

// loadConfig reads and parses the header check configurations in config file.
func loadConfig(config *header.Config) error {
	logger.Log.Infoln("Loading configuration from file:", cfgFile)

	bytes, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return err
	}

	if err = yaml.Unmarshal(bytes, config); err != nil {
		return err
	}

	var lines []string
	for _, line := range strings.Split(config.License, "\n") {
		if len(line) > 0 {
			lines = append(lines, strings.Trim(line, header.CommentChars))
		}
	}
	config.License = strings.Join(lines, " ")

	logger.Log.Infoln("License header is:", config.License)

	if len(config.Paths) == 0 {
		config.Paths = []string{"**"}
	}

	return nil
}
