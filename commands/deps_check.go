// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/deps"
)

var weakCompatible bool

func init() {
	DepsCheckCommand.PersistentFlags().BoolVarP(&weakCompatible, "weak-compatible", "w", false,
		"if set to true, treat the weak-compatible licenses as compatible in dependencies check. "+
			"Note: when set to true, make sure to manually confirm that weak-compatible licenses "+
			"are used under the required conditions.")
}

var DepsCheckCommand = &cobra.Command{
	Use:     "check",
	Aliases: []string{"c"},
	Long:    "resolves and check license compatibility in all dependencies of a module and their transitive dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		var errors []error
		configDeps := Config.Dependencies()
		for _, header := range Config.Headers() {
			if err := deps.Check(header.License.SpdxID, configDeps, weakCompatible); err != nil {
				errors = append(errors, err)
			}
		}
		if len(errors) > 0 {
			for _, err := range errors {
				logger.Log.Error(err)
			}
			return fmt.Errorf("one or more errors occurred checking license compatibility")
		}
		return nil
	},
}
