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
package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	headercommand "license-checker/commands/header"
	"license-checker/internal/logger"
)

var (
	verbosity string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "license-checker command [flags]",
	Long:          "license-checker walks the specified path recursively and checks if the specified files have the license header in the config file.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if level, err := logrus.ParseLevel(verbosity); err != nil {
			return err
		} else {
			logger.Log.SetLevel(level)
		}
		return nil
	},
}

// Execute sets flags to the root command appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	rootCmd.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", logrus.InfoLevel.String(), "Log level (debug, info, warn, error, fatal, panic")

	rootCmd.AddCommand(headercommand.Header)

	return rootCmd.Execute()
}
