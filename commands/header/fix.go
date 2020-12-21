// Copyright © 2020 Hoshea Jiang <hoshea@apache.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package header

import (
	"github.com/spf13/cobra"
	"license-checker/internal/logger"
	"license-checker/pkg/header"
	"license-checker/pkg/header/fix"
	"strings"
)

var FixCommand = &cobra.Command{
	Use:     "fix",
	Aliases: []string{"f"},
	Long:    "`fix` command walks the specified paths recursively and fix the license header if the specified files don't have the license header in the config file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var config header.Config
		var result header.Result

		if err := config.Parse(cfgFile); err != nil {
			return err
		}

		if err := header.Check(&config, &result); err != nil {
			return err
		}

		for _, file := range result.Failure {
			if strings.HasSuffix(file, ".go") {
				logger.Log.Infoln("Fixing file:", file)
				if err := fix.GoLang(file, &config, &result); err != nil {
					return err
				}
			}
		}

		return nil
	},
}
