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
	"github.com/apache/skywalking-eyes/pkg/deps"
	"github.com/apache/skywalking-eyes/pkg/header"

	"gopkg.in/yaml.v3"
)

type V1 struct {
	Header header.ConfigHeader `yaml:"header"`
	Deps   deps.ConfigDeps     `yaml:"dependency"`
}

func ParseV1(filename string, bytes []byte) (*V1, error) {
	var config V1
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	if err := config.Header.Finalize(); err != nil {
		return nil, err
	}

	if err := config.Deps.Finalize(filename); err != nil {
		return nil, err
	}
	return &config, nil
}

func (config *V1) Headers() []*header.ConfigHeader {
	return []*header.ConfigHeader{&config.Header}
}

func (config *V1) Dependencies() *deps.ConfigDeps {
	return &config.Deps
}

type V2 struct {
	Header []*header.ConfigHeader `yaml:"header"`
	Deps   deps.ConfigDeps        `yaml:"dependency"`
}

func ParseV2(filename string, bytes []byte) (*V2, error) {
	var config V2
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		return nil, err
	}

	for _, header := range config.Header {
		if err := header.Finalize(); err != nil {
			return nil, err
		}
	}

	if err := config.Deps.Finalize(filename); err != nil {
		return nil, err
	}

	return &config, nil
}

func (config *V2) Headers() []*header.ConfigHeader {
	return config.Header
}

func (config *V2) Dependencies() *deps.ConfigDeps {
	return &config.Deps
}

type Config interface {
	Headers() []*header.ConfigHeader
	Dependencies() *deps.ConfigDeps
}

func NewConfigFromFile(filename string) (Config, error) {
	var err error
	var bytes []byte

	// attempt to read configuration from specified file
	logger.Log.Infoln("Loading configuration from file:", filename)

	if bytes, err = os.ReadFile(filename); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if os.IsNotExist(err) {
		logger.Log.Infof("Config file %s does not exist, using the default config", filename)

		if bytes, err = assets.Asset("default-config.yaml"); err != nil {
			return nil, err
		}
	}

	var config Config
	if config, err = ParseV2(filename, bytes); err == nil {
		return config, nil
	}
	if config, err = ParseV1(filename, bytes); err != nil {
		return nil, err
	}
	return config, nil
}
