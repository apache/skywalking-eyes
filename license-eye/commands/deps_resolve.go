//
// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
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
	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
	"github.com/apache/skywalking-eyes/license-eye/pkg/deps"
	"github.com/spf13/cobra"
)

var ResolveCommand = &cobra.Command{
	Use:     "resolve",
	Aliases: []string{"r"},
	Long:    "resolves all dependencies of a go.mod file and their transitive dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		report := deps.Report{}

		if err := deps.Resolve(&Config.Deps, &report); err != nil {
			return err
		}

		for _, result := range report.Resolved {
			logger.Log.Debugln("Pkg: ", result.Dependency, " License:", result.LicenseSpdxID)
		}

		logger.Log.Debugln("Skipped:", len(report.Skipped))

		return nil
	},
}
