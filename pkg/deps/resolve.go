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
package deps

import (
	"fmt"
)

type Resolver interface {
	CanResolve(string) bool
	Resolve(string, *Report) error
}

var Resolvers = []Resolver{
	new(GoModResolver),
}

func Resolve(config *ConfigDeps, report *Report) error {
	for _, file := range config.Files {
		for _, resolver := range Resolvers {
			if !resolver.CanResolve(file) {
				continue
			}
			if err := resolver.Resolve(file, report); err != nil {
				return err
			}
			return nil
		}
		return fmt.Errorf("unable to find a resolver to resolve dependency declaration file: %v", file)
	}

	return nil
}
