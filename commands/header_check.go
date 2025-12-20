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
	"os"

	"github.com/apache/skywalking-eyes/pkg/header"
	"github.com/apache/skywalking-eyes/pkg/logger"
	"github.com/apache/skywalking-eyes/pkg/review"

	"github.com/spf13/cobra"
)

var CheckCommand = &cobra.Command{
	Use:     "check",
	Aliases: []string{"c"},
	Long:    "check command walks the specified paths recursively and checks if the specified files have the license header in the config file.",
	RunE: func(_ *cobra.Command, args []string) error {
		hasErrors := false
		for _, h := range Config.Headers() {
			var result header.Result

			if len(args) > 0 {
				logger.Log.Debugln("Overriding paths with command line args.")
				h.Paths = args
			}

			if err := header.Check(h, &result); err != nil {
				return err
			}

			logger.Log.Infoln(result.String())

			writeSummaryQuietly(&result)

			if result.HasFailure() {
				if err := review.Header(&result, h); err != nil {
					logger.Log.Warnln("Failed to create review comments", err)
				}
				hasErrors = true
				logger.Log.Error(result.Error())
			}
		}
		if hasErrors {
			return fmt.Errorf("one or more files does not have a valid license header")
		}
		return nil
	},
}

func writeSummaryQuietly(result *header.Result) {
	if summaryFileName := os.Getenv("GITHUB_STEP_SUMMARY"); summaryFileName != "" {
		if summaryFile, err := os.OpenFile(summaryFileName, os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
			defer summaryFile.Close()
			_, _ = summaryFile.WriteString("# License Eye Summary\n")
			_, _ = summaryFile.WriteString(result.String())
			if result.HasFailure() {
				_, _ = summaryFile.WriteString(", the following files are lack of license headers:\n")
				for _, failure := range result.Failure {
					_, _ = fmt.Fprintf(summaryFile, "- %s\n", failure)
				}
			}
		}
	}
}
