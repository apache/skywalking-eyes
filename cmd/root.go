/*
Copyright © 2020 Hoshea Jiang <hoshea@apache.org>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"license-checker/util"
	"os"
	"path/filepath"
	"strings"
)

var cfgFile string
var checkPath string
var loose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "license-checker [flags]",
	Long: `license-checker walks the specified path recursively and checks 
if the specified files have the license header in the config file.`,
	Run: func(cmd *cobra.Command, args []string) {
		config, err := LoadConfig()
		if err != nil {
			fmt.Println(err)
		}

		if err := Walk(checkPath, config); err != nil {
			fmt.Println(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", ".licenserc.json", "the config file")
	rootCmd.PersistentFlags().StringVarP(&checkPath, "path", "p", "", "the path to check (required)")
	rootCmd.PersistentFlags().BoolVarP(&loose, "loose", "l", false, "loose mode")
	if err := rootCmd.MarkPersistentFlagRequired("path"); err != nil {
		fmt.Println(err)
	}
}

type excludeConfig struct {
	Files       []string `json:"files"`
	Extensions  []string `json:"extensions"`
	Directories []string `json:"directories"`
}

type Config struct {
	LicenseStrict []string      `json:"licenseStrict"`
	LicenseLoose  []string      `json:"licenseLoose"`
	TargetFiles   []string      `json:"targetFiles"`
	Exclude       excludeConfig `json:"exclude"`
}

type Result struct {
	Success []string `json:"success"`
	Failure []string `json:"failure"`
}

// LoadConfig reads in config file.
func LoadConfig() (*Config, error) {
	var config Config
	bytes, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func Walk(p string, cfg *Config) error {
	var license []string
	if loose {
		license = cfg.LicenseStrict
	} else {
		license = cfg.LicenseLoose
	}

	inExcludeDir := util.InStrSliceMapKeyFunc(cfg.Exclude.Directories)
	inExcludeExt := util.InStrSliceMapKeyFunc(cfg.Exclude.Extensions)
	inExcludeFiles := util.InStrSliceMapKeyFunc(cfg.Exclude.Files)

	err := filepath.Walk(p, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err) // can't walk here,
			return nil       // but continue walking elsewhere
		}

		if fi.IsDir() {
			if inExcludeDir(fi.Name()) {
				return filepath.SkipDir
			}
		} else {
			ext := util.GetFileExtension(fi.Name())
			if inExcludeFiles(fi.Name()) || inExcludeExt(ext) {
				return nil
			}

			// TODO: open the file and check
			fmt.Println(path)

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains()
			}

		}

		return nil
	})

	return err
}
