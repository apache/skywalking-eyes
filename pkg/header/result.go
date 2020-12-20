// Copyright © 2020 Hoshea Jiang <hoshea@apache.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package header

import (
	"fmt"
	"strings"
)

type Result struct {
	Success []string
	Failure []string
	Ignored []string
}

func (result *Result) Fail(file string) {
	result.Failure = append(result.Failure, file)
}

func (result *Result) Succeed(file string) {
	result.Success = append(result.Success, file)
}

func (result *Result) Ignore(file string) {
	result.Ignored = append(result.Ignored, file)
}

func (result *Result) HasFailure() bool {
	return len(result.Failure) > 0
}

func (result *Result) Error() error {
	return fmt.Errorf(
		"The following files don't have a valid license header: \n%v",
		strings.Join(result.Failure, "\n"),
	)
}
