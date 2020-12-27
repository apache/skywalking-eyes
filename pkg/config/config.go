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
package config

import (
	"io/ioutil"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
	"github.com/apache/skywalking-eyes/license-eye/pkg/deps"
	"github.com/apache/skywalking-eyes/license-eye/pkg/header"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Header header.ConfigHeader `yaml:"header"`
	Deps   deps.ConfigDeps     `yaml:"dependency"`
}

// Parse reads and parses the header check configurations in config file.
func (config *Config) Parse(file string) error {
	logger.Log.Infoln("Loading configuration from file:", file)

	if bytes, err := ioutil.ReadFile(file); err != nil {
		return err
	} else if err := yaml.Unmarshal(bytes, config); err != nil {
		return err
	}

	if err := config.Header.Finalize(); err != nil {
		return err
	}

	if err := config.Deps.Finalize(file); err != nil {
		return err
	}

	return nil
}
