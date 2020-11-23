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

var (
	// cfgFile is the path to the config file.
	cfgFile string
	// checkPath is the path to check license.
	checkPath string
	// loose is flag to enable loose mode.
	loose bool
	// verbose is flag to enable verbose mode.
	verbose bool
)

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

		res, err := WalkAndCheck(checkPath, config)
		if err != nil {
			fmt.Println(err)
		}
		printResult(res)
	},
}

// Execute sets flags to the root command appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", ".licenserc.json", "the config file")
	rootCmd.PersistentFlags().StringVarP(&checkPath, "path", "p", ".", "the path to check")
	rootCmd.PersistentFlags().BoolVarP(&loose, "loose", "l", false, "loose mode")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
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

// WalkAndCheck traverses the path p and check every target file's license.
func WalkAndCheck(p string, cfg *Config) (*Result, error) {
	var license []string
	if loose {
		license = cfg.LicenseLoose
	} else {
		license = cfg.LicenseStrict
	}

	inExcludeDir := util.InStrSliceMapKeyFunc(cfg.Exclude.Directories)
	inExcludeExt := util.InStrSliceMapKeyFunc(cfg.Exclude.Extensions)
	inExcludeFiles := util.InStrSliceMapKeyFunc(cfg.Exclude.Files)

	var result Result
	err := filepath.Walk(p, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err) // can't walk here,
			return nil       // but continue walking elsewhere
		}

		if fi.IsDir() {
			if inExcludeDir(fi.Name()) ||
				inExcludeDir(util.CleanPathPrefixes(path, []string{p, string(os.PathSeparator)})) {
				return filepath.SkipDir
			}
		} else {
			ext := util.GetFileExtension(fi.Name())
			if inExcludeFiles(fi.Name()) || inExcludeExt(ext) {
				return nil
			}

			ok, err := CheckLicense(path, license)
			if err != nil {
				return err
			}

			if ok {
				result.Success = append(result.Success, fmt.Sprintf("[Pass]: %s", path))
			} else {
				result.Failure = append(result.Failure, fmt.Sprintf("[No Specified License]: %s", path))
			}
		}

		return nil
	})

	return &result, err
}

// CheckLicense checks license of single file.
func CheckLicense(filePath string, license []string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	index := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() && index < len(license) {
		line := scanner.Text()
		if strings.Contains(line, license[index]) {
			index++
		}
	}

	if index != len(license) {
		return false, nil
	}
	return true, nil
}

// printResult prints license check result.
func printResult(r *Result) {
	if verbose {
		for _, s := range r.Success {
			fmt.Println(s)
		}
	}

	for _, s := range r.Failure {
		fmt.Println(s)
	}

	fmt.Printf("Total check %d files, success: %d, failure: %d\n",
		len(r.Success)+len(r.Failure), len(r.Success), len(r.Failure))
}
