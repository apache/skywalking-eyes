// Copyright © 2020 Hoshea Jiang <hoshea@apache.org>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package fix

import (
	"io/ioutil"
	"license-checker/internal/logger"
	"license-checker/pkg/header"
	"os"
	"strings"
)

func GoLang(file string, config *header.Config, result *header.Result) error {
	var r header.Result
	if err := header.CheckFile(file, config, &r); err != nil || !r.HasFailure() {
		logger.Log.Warnln("Try to fix a valid file, returning:", file)
		return err
	}

	stat, err := os.Stat(file)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	lines := "// " + strings.Join(strings.Split(config.License, "\n"), "\n// ") + "\n"

	if err := ioutil.WriteFile(file, append([]byte(lines), content...), stat.Mode()); err != nil {
		return err
	}

	result.Fix(file)

	return nil
}
