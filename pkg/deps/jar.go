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

package deps

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/license"
)

type JarResolver struct{}

func (resolver *JarResolver) CanResolve(jarFile string) bool {
	return filepath.Ext(jarFile) == ".jar"
}

func (resolver *JarResolver) Resolve(jarFile string, licenses []*ConfigDepLicense, report *Report) error {
	state := NotFound
	if err := resolver.ResolveJar(&state, jarFile, Unknown, report); err != nil {
		dep := filepath.Base(jarFile)
		logger.Log.Warnf("Failed to resolve the license of <%s>: %v\n", dep, state.String())
		report.Skip(&Result{
			Dependency:    dep,
			LicenseSpdxID: Unknown,
		})
	}

	return nil
}

func (resolver *JarResolver) ResolveJar(state *State, jarFile, version string, report *Report) error {
	dep := filepath.Base(jarFile)

	compressedJar, err := zip.OpenReader(jarFile)
	if err != nil {
		return err
	}
	defer compressedJar.Close()

	var manifestFile *zip.File

	// traverse all files in jar
	for _, compressedFile := range compressedJar.File {
		archiveFile := compressedFile.Name
		switch {
		case reHaveManifestFile.MatchString(archiveFile):
			manifestFile = compressedFile

		case possibleLicenseFileName.MatchString(archiveFile):
			*state |= FoundLicenseInJarLicenseFile
			buf, err := resolver.ReadFileFromZip(compressedFile)
			if err != nil {
				return err
			}

			return resolver.IdentifyLicense(jarFile, dep, buf.String(), version, nil, report)
		}
	}

	if manifestFile != nil {
		buf, err := resolver.ReadFileFromZip(manifestFile)
		if err != nil {
			return err
		}
		norm := regexp.MustCompile(`(?im)[\r\n]+ +`)
		content := norm.ReplaceAllString(buf.String(), "")

		r := reSearchLicenseInManifestFile.FindStringSubmatch(content)
		if len(r) != 0 {
			report.Resolve(&Result{
				Dependency:      dep,
				LicenseFilePath: jarFile,
				LicenseContent:  strings.TrimSpace(r[1]),
				LicenseSpdxID:   strings.TrimSpace(r[1]),
				Version:         version,
			})
			return nil
		}
	}

	return fmt.Errorf("cannot find license content")
}

func (resolver *JarResolver) ReadFileFromZip(archiveFile *zip.File) (*bytes.Buffer, error) {
	file, err := archiveFile.Open()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	w := bufio.NewWriter(buf)
	_, err = io.CopyN(w, file, int64(archiveFile.UncompressedSize64))
	if err != nil {
		return nil, err
	}

	w.Flush()
	file.Close()
	return buf, nil
}

func (resolver *JarResolver) IdentifyLicense(path, dep, content, version string, declareLicense *ConfigDepLicense, report *Report) error {
	identifier, err := license.Identify(path, content)
	if err != nil {
		if declareLicense == nil {
			return err
		}
		identifier = declareLicense.License
	}

	report.Resolve(&Result{
		Dependency:      dep,
		LicenseFilePath: path,
		LicenseContent:  content,
		LicenseSpdxID:   identifier,
		Version:         version,
	})
	return nil
}
