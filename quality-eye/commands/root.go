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
//
package commands

import (
	"github.com/spf13/cobra"

	"github.com/apache/skywalking-eyes/quality-eye/commands/cleanup"
	"github.com/apache/skywalking-eyes/quality-eye/commands/run"
	"github.com/apache/skywalking-eyes/quality-eye/commands/setup"
	"github.com/apache/skywalking-eyes/quality-eye/commands/trigger"
	"github.com/apache/skywalking-eyes/quality-eye/commands/verify"
)

// Root represents the base command when called without any subcommands
var Root = &cobra.Command{
	Use:     "quality-eye command [flags]",
	Short:   "The next generation End-to-End Testing framework",
	Version: version,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	Root.AddCommand(run.Run)
	Root.AddCommand(setup.Setup)
	Root.AddCommand(trigger.Trigger)
	Root.AddCommand(verify.Verify)
	Root.AddCommand(cleanup.Cleanup)

	return Root.Execute()
}
