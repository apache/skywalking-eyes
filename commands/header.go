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
	"github.com/spf13/cobra"

	"github.com/apache/skywalking-eyes/pkg/header"
	"github.com/apache/skywalking-eyes/pkg/logger"
)

var Header = &cobra.Command{
	Use:     "header",
	Aliases: []string{"h"},
	Short:   "License header related commands; e.g. check, fix, etc.",
	Long: "`header` command walks the specified paths and checks if the specified " +
		"files have the license header in the config file. " +
		"Accepts files, directories, and glob patterns.",
}

func init() {
	Header.AddCommand(CheckCommand)
	Header.AddCommand(FixCommand)
	Header.AddCommand(DiffCommand)
}

// checkHeader checks the files that the header config covers. Paths given on
// the command line select the files to check instead of the config's paths, but
// the header still checks only the selected files that its own paths and
// paths-ignore cover, so with multiple headers each file is checked only against
// the headers that cover it.
func checkHeader(h *header.ConfigHeader, args []string, result *header.Result) error {
	if len(args) == 0 {
		return header.Check(h, result)
	}
	logger.Log.Debugln("Checking the paths given on the command line.")
	return header.CheckPaths(h, args, result)
}
