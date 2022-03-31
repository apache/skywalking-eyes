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

package config

import (
	"os"

	"github.com/apache/skywalking-eyes/assets"
	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/comments"
	"github.com/apache/skywalking-eyes/pkg/deps"
	"github.com/apache/skywalking-eyes/pkg/header"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Header    header.ConfigHeader          `yaml:"header"`
	Deps      deps.ConfigDeps              `yaml:"dependency"`
	Languages map[string]comments.Language `yaml:"language"`
}

// Parse reads and parses the header check configurations in config file.
func (config *Config) Parse(file string) (err error) {
	var bytes []byte

	// attempt to read configuration from specified file
	logger.Log.Infoln("Loading configuration from file:", file)

	if bytes, err = os.ReadFile(file); err != nil && !os.IsNotExist(err) {
		return err
	}

	if os.IsNotExist(err) {
		logger.Log.Infof("Config file %s does not exist, using the default config", file)

		if bytes, err = assets.Asset("default-config.yaml"); err != nil {
			return err
		}
	}

	if err := yaml.Unmarshal(bytes, config); err != nil {
		return err
	}

	if err := config.Header.Finalize(); err != nil {
		return err
	}

	return config.Deps.Finalize(file)
}
