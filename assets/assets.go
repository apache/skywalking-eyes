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

package assets

import (
	"embed"
	"io/fs"
	"path/filepath"
)

//go:embed *
var assets embed.FS

func FS() fs.FS {
	return assets
}

func Asset(file string) ([]byte, error) {
	return assets.ReadFile(filepath.ToSlash(file))
}

func AssetDir(dir string) ([]fs.DirEntry, error) {
	return assets.ReadDir(filepath.ToSlash(dir))
}
