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
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/apache/skywalking-eyes/assets"

	"github.com/spf13/cobra"

	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/deps"
)

var (
	outDir         string
	licensePath    string
	summaryTplPath string
	summaryTpl     *template.Template
)

func init() {
	DepsResolveCommand.PersistentFlags().StringVarP(&outDir, "output", "o", "",
		"the directory to output the resolved dependencies' licenses, if not set the dependencies' licenses won't be saved")
	DepsResolveCommand.PersistentFlags().StringVarP(&summaryTplPath, "summary", "s", "",
		"the template file to write the summary of dependencies' licenses, a new file named \"LICENSE\" will be "+
			"created in the same directory as the template file, to save the final summary.")
	DepsResolveCommand.PersistentFlags().StringVarP(&licensePath, "license", "l", "",
		"the path to the LICENSE file to be generated. The default summary format will be used if summary template file is not specified")
}

var fileNamePattern = regexp.MustCompile(`[^a-zA-Z0-9\\.\-]`)

var DepsResolveCommand = &cobra.Command{
	Use:     "resolve",
	Aliases: []string{"r"},
	Long:    "resolves all dependencies of a module and their transitive dependencies",
	PreRunE: func(_ *cobra.Command, _ []string) error {
		if outDir != "" {
			absPath, err := filepath.Abs(outDir)
			if err != nil {
				return err
			}
			outDir = absPath
			if err := os.MkdirAll(outDir, 0o700); err != nil && !os.IsExist(err) {
				return err
			}
		}
		if summaryTplPath != "" {
			absPath, err := filepath.Abs(summaryTplPath)
			if err != nil {
				return err
			}
			summaryTplPath = absPath
			tpl, err := deps.ParseTemplate(os.DirFS(filepath.Dir(absPath)), filepath.Base(absPath))
			if err != nil {
				return err
			}
			summaryTpl = tpl
		}
		if licensePath != "" {
			absPath, err := filepath.Abs(licensePath)
			if err != nil {
				return err
			}
			licensePath = absPath
			if err := os.MkdirAll(filepath.Dir(outDir), 0o700); err != nil && !os.IsExist(err) {
				return err
			}

			if summaryTpl == nil {
				tpl, err := deps.ParseTemplate(assets.FS(), "default-license.tpl")
				if err != nil {
					return err
				}
				summaryTpl = tpl
			}
		}
		return nil
	},
	RunE: func(_ *cobra.Command, _ []string) error {
		report := deps.Report{}

		configDeps := Config.Dependencies()
		if err := deps.Resolve(configDeps, &report); err != nil {
			return err
		}

		if summaryTpl != nil {
			if err := writeSummary(&report, licensePath); err != nil {
				return err
			}
		}

		if outDir != "" {
			for _, result := range report.Resolved {
				writeLicense(result)
			}
		}

		fmt.Println(report.String())

		if skipped := len(report.Skipped); skipped > 0 {
			pkgs := make([]string, skipped)
			for i, s := range report.Skipped {
				pkgs[i] = s.Dependency
			}
			return fmt.Errorf(
				"failed to identify the licenses of following packages (%d):\n%s",
				len(pkgs), strings.Join(pkgs, "\n"),
			)
		}

		return nil
	},
}

func writeLicense(result *deps.Result) {
	filename := string(fileNamePattern.ReplaceAll([]byte(result.Dependency), []byte("-")))
	filename = filepath.Join(outDir, "license-"+filename+".txt")
	if _, err := os.Stat(filename); err == nil {
		logger.Log.Debugf("File already exists, skipping: %s", filename)
		return
	}
	file, err := os.Create(filename)
	if err != nil {
		logger.Log.Errorf("failed to create license file %v: %v", filename, err)
		return
	}
	defer func(file *os.File) { _ = file.Close() }(file)
	_, err = file.WriteString(result.LicenseContent)
	if err != nil {
		logger.Log.Errorf("failed to write license file, %v: %v", filename, err)
		return
	}
}

func writeSummary(rep *deps.Report, path string) error {
	if path == "" {
		path = filepath.Join(filepath.Dir(summaryTplPath), "LICENSE")
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	headers := Config.Headers()
	if len(headers) > 1 {
		return fmt.Errorf("unable to write summary as multiple licenses were provided in configuration")
	}
	summary, err := deps.GenerateSummary(summaryTpl, headers[0], rep)
	if err != nil {
		return err
	}
	_, err = file.WriteString(summary)
	return err
}
