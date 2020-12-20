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
package util

import (
	"strings"
)

func InStrSliceMapKeyFunc(strs []string) func(string) bool {
	set := make(map[string]struct{})

	for _, e := range strs {
		set[e] = struct{}{}
	}

	return func(s string) bool {
		_, ok := set[s]
		return ok
	}
}

func GetFileExtension(filename string) string {
	i := strings.LastIndex(filename, ".")
	if i != -1 {
		if i+1 < len(filename) {
			return filename[i+1:]
		}
	}
	return ""
}

func CleanPathPrefixes(path string, prefixes []string) string {
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) && len(path) > 0 {
			path = path[len(prefix):]
		}
	}

	return path
}
