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
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/apache/skywalking-eyes/pkg/header"
	"github.com/apache/skywalking-eyes/pkg/logger"
)

var DiffCommand = &cobra.Command{
	Use:     "diff [paths...]",
	Aliases: []string{"d"},
	Long: "diff command walks the specified paths recursively and shows where the " +
		"license headers of the invalid files differ from the license header in the " +
		"config file, to help understand why the check command fails. " +
		"Accepts files, directories, and glob patterns. " +
		"If no paths are specified, checks the current directory " +
		"recursively as defined in the config file. " +
		"The texts are compared in the same normalized forms that the check command " +
		"compares (comment markers stripped, whitespace flattened, case-insensitive, etc.), " +
		"so every difference shown is a real cause of the check failure: " +
		"[-text-] is expected by the configured license but missing in the file, " +
		"{+text+} is in the file but not expected by the configured license.",
	RunE: func(_ *cobra.Command, args []string) error {
		hasErrors := false
		var errors []string
		for _, h := range Config.Headers() {
			var result header.Result

			if len(args) > 0 {
				logger.Log.Debugln("Overriding paths with command line args.")
				h.Paths = args
			}

			if err := header.Check(h, &result); err != nil {
				return err
			}

			sort.Strings(result.Failure)
			for _, file := range result.Failure {
				diff, err := header.DiffFile(file, h)
				if err != nil {
					errors = append(errors, err.Error())
					continue
				}
				if diff == "" {
					continue
				}
				fmt.Printf("%v:\n\t%v\n", file, diff)
			}

			logger.Log.Infoln(result.String())

			if result.HasFailure() {
				hasErrors = true
			}
		}
		if len(errors) > 0 {
			return fmt.Errorf("%s", strings.Join(errors, "\n"))
		}
		if hasErrors {
			return fmt.Errorf("one or more files does not have a valid license header")
		}
		return nil
	},
}
