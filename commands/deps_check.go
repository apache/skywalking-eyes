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
var fsfFreeOnly bool
var osiApprovedOnly bool

func init() {
	DepsCheckCommand.PersistentFlags().BoolVarP(&weakCompatible, "weak-compatible", "w", false,
		"if set to true, treat the weak-compatible licenses as compatible in dependencies check. "+
			"Note: when set to true, make sure to manually confirm that weak-compatible licenses "+
			"are used under the required conditions.")
	DepsCheckCommand.PersistentFlags().BoolVarP(&fsfFreeOnly, "fsf-free", "f", false,
		"Only consider licenses marked as FSF Free/Libre when determining compatibility. Non-FSF-free licenses are treated as incompatible.")
	DepsCheckCommand.PersistentFlags().BoolVarP(&osiApprovedOnly, "osi-approved", "o", false,
		"Only consider OSI-approved licenses when determining compatibility. Non-OSI-approved licenses are treated as incompatible.")
}

var DepsCheckCommand = &cobra.Command{
	Use:     "check",
	Aliases: []string{"c"},
	Long:    "resolves and check license compatibility in all dependencies of a module and their transitive dependencies",
	RunE: func(_ *cobra.Command, _ []string) error {
		var errors []error
		configDeps := Config.Dependencies()
		// CLI flags override to enable stricter requirements, cannot disable if enabled by config
		if configDeps != nil {
			if fsfFreeOnly {
				configDeps.RequireFSFFree = true
			}
			if osiApprovedOnly {
				configDeps.RequireOSIApproved = true
			}
		}
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
