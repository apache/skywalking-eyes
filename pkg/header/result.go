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

package header

import (
	"fmt"
	"strings"
	"sync"
)

type Result struct {
	mu      sync.Mutex
	Success []string
	Failure []string
	Ignored []string
	Fixed   []string
}

func (result *Result) Fail(file string) {
	result.mu.Lock()
	result.Failure = append(result.Failure, file)
	result.mu.Unlock()
}

func (result *Result) Succeed(file string) {
	result.mu.Lock()
	result.Success = append(result.Success, file)
	result.mu.Unlock()
}

func (result *Result) Ignore(file string) {
	result.mu.Lock()
	result.Ignored = append(result.Ignored, file)
	result.mu.Unlock()
}

func (result *Result) Fix(file string) {
	result.mu.Lock()
	result.Fixed = append(result.Fixed, file)
	result.mu.Unlock()
}

func (result *Result) HasFailure() bool {
	result.mu.Lock()
	has := len(result.Failure) > 0
	result.mu.Unlock()
	return has
}

func (result *Result) Error() error {
	result.mu.Lock()
	msg := fmt.Errorf(
		"the following files don't have a valid license header: \n%v",
		strings.Join(result.Failure, "\n"),
	)
	result.mu.Unlock()
	return msg
}

func (result *Result) String() string {
	result.mu.Lock()
	s := fmt.Sprintf(
		"Totally checked %d files, valid: %d, invalid: %d, ignored: %d, fixed: %d",
		len(result.Success)+len(result.Failure)+len(result.Ignored),
		len(result.Success),
		len(result.Failure),
		len(result.Ignored),
		len(result.Fixed),
	)
	result.mu.Unlock()
	return s
}
