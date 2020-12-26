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
//
package header

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
	"github.com/apache/skywalking-eyes/license-eye/pkg"
	"github.com/apache/skywalking-eyes/license-eye/pkg/config"
	"github.com/apache/skywalking-eyes/license-eye/pkg/header"
)

var FixCommand = &cobra.Command{
	Use:     "fix",
	Aliases: []string{"f"},
	Long:    "fix command walks the specified paths recursively and fix the license header if the specified files don't have the license header.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var config config.Config
		var result pkg.Result

		if err := config.Parse(cfgFile); err != nil {
			return err
		}

		if err := header.Check(&config.Header, &result); err != nil {
			return err
		}

		var errors []string
		for _, file := range result.Failure {
			if err := header.Fix(file, &config.Header, &result); err != nil {
				errors = append(errors, err.Error())
			}
		}

		logger.Log.Infoln(result.String())

		if len(errors) > 0 {
			return fmt.Errorf(strings.Join(errors, "\n"))
		}

		return nil
	},
}
