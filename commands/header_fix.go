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
	"strings"

	"github.com/spf13/cobra"

	"github.com/apache/skywalking-eyes/pkg/header"
	"github.com/apache/skywalking-eyes/pkg/logger"
)

var FixCommand = &cobra.Command{
	Use:     "fix",
	Aliases: []string{"f"},
	Long:    "fix command walks the specified paths recursively and fix the license header if the specified files don't have the license header.",
	RunE: func(_ *cobra.Command, args []string) error {
		var errors []string
		for _, h := range Config.Headers() {
			var result header.Result
			var files []string

			if len(args) > 0 {
				files = args
			} else if err := header.Check(h, &result); err != nil {
				return err
			} else {
				files = result.Failure
			}

			for _, file := range files {
				if err := header.Fix(file, h, &result); err != nil {
					errors = append(errors, err.Error())
				}
			}

			logger.Log.Infoln(result.String())
		}
		if len(errors) > 0 {
			return fmt.Errorf("%s", strings.Join(errors, "\n"))
		}
		return nil
	},
}
