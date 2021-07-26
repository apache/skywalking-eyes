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
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
	"github.com/apache/skywalking-eyes/license-eye/pkg/license"
)

var maven string

func init() {
	err := CheckMVN()
	if err != nil {
		logger.Log.Errorln("an error occurred when checking maven tool:", err)
	}
}

// CheckMVN use mvn by default, use mvn if mvnw is not found
func CheckMVN() error {
	var err error

	logger.Log.Debugln("searching mvnw ...")
	_, err = exec.Command("./mvnw", "--version").Output()
	if err == nil {
		maven = "./mvnw"
		logger.Log.Debugln("use mvnw")
		return nil
	}

	logger.Log.Debugln("mvnw is not found, searching mvn ...")
	_, err = exec.Command("mvn", "--version").Output()
	if err == nil {
		maven = "mvn"
		logger.Log.Debugln("use mvn")
		return nil
	}

	return fmt.Errorf("neither found mvnw nor mvn")
}

// TempDirGenerator Create and destroy one temporary directory
type TempDirGenerator interface {
	Create() (string, error)
	Destroy()
}

func NewTempDirGenerator() TempDirGenerator {
	return new(tempDir)
}

// tempDir an implementation of the TempDirGenerator
type tempDir struct {
	dir string
}

func (t *tempDir) Create() (string, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	t.dir = tmpDir
	return tmpDir, nil
}

func (t *tempDir) Destroy() {
	if t.dir == "" {
		logger.Log.Errorf("the temporary directory does not exist")
		return
	}

	err := os.RemoveAll(t.dir)
	if err != nil {
		logger.Log.Errorln(err)
	}
}

var possiblePomFileName = regexp.MustCompile(`(?i)^pom\.xml|.+\.pom$`)

type MavenPomResolver struct{}

// CanResolve determine whether the file can be resolve by name of the file
func (resolver *MavenPomResolver) CanResolve(mavenPomFile string) bool {
	if maven == "" {
		return false
	}

	// switch to the directory where the file is located for searching mvnw
	dir, base := resolver.Split(mavenPomFile)
	if err := os.Chdir(dir); err != nil {
		logger.Log.Errorf("an error occurred when entering dir <%s> to parser file <%s>:%v\n", dir, base, err)
		return false
	}

	logger.Log.Debugln("Base name:", base)
	return possiblePomFileName.MatchString(base)
}

// Split a simple wraper of filepath.Split
func (resolver *MavenPomResolver) Split(path string) (dir, file string) {
	dir, file = filepath.Split(path)
	if dir == "" {
		dir = "./"
	}
	return
}

// Resolve resolves licenses of all dependencies declared in the pom.xml file.
func (resolver *MavenPomResolver) Resolve(mavenPomFile string, report *Report) error {
	dir, base := resolver.Split(mavenPomFile)
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("an error occurred when entering dir <%s> to parser file <%s>:%v", dir, base, err)
	}

	tempDirGenerator := NewTempDirGenerator()
	dependenciesDir, err := tempDirGenerator.Create()
	if err != nil {
		return fmt.Errorf("an error occurred when create temporary dir: %v", err)
	}
	defer tempDirGenerator.Destroy()

	cmd := exec.Command(maven, "dependency:copy-dependencies", "-f", base,
		fmt.Sprintf("-DoutputDirectory=%s", dependenciesDir), "-DincludeScope=runtime")
	_, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("an error occurred when execute maven command 「%v」: %v", cmd.String(), err)
	}

	jarFiles, err := ioutil.ReadDir(dependenciesDir)
	if err != nil {
		return err
	}

	logger.Log.Debugln("jars size:", len(jarFiles))

	if err := resolver.ResolveJarFiles(dependenciesDir, jarFiles, report); err != nil {
		return err
	}

	return nil
}

// ResolveJarFiles resolves the licenses of the given packages.
func (resolver *MavenPomResolver) ResolveJarFiles(dir string, jarFiles []fs.FileInfo, report *Report) (err error) {
	for _, jar := range jarFiles {
		dependencyPath := filepath.Join(dir, jar.Name())
		err = resolver.ResolveJarLicense(dependencyPath, jar, report)
		if err != nil {
			logger.Log.Warnf("Failed to resolve the license of <%s>: %v\n", filepath.Base(dependencyPath), err)
			report.Skip(&Result{
				Dependency:    filepath.Base(dependencyPath),
				LicenseSpdxID: Unknown,
			})
		}
	}
	return nil
}

var possibleLicenseFileNameInJar = regexp.MustCompile(`(?i)^(\S*)?LICEN[SC]E(\S*\.txt)?$`)

// ResolveJarLicense search license file of the jar package, then identify it
func (resolver *MavenPomResolver) ResolveJarLicense(dependencyPath string, jar fs.FileInfo, report *Report) (err error) {
	compressedJar, err := zip.OpenReader(dependencyPath)
	if err != nil {
		return err
	}
	defer compressedJar.Close()

	// traverse all files in jar
	for _, compressedFile := range compressedJar.File {
		archiveFile := compressedFile.Name
		if !possibleLicenseFileNameInJar.MatchString(archiveFile) {
			continue
		}

		file, err := compressedFile.Open()
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(nil)
		w := bufio.NewWriter(buf)
		_, err = io.CopyN(w, file, int64(compressedFile.UncompressedSize64))
		if err != nil {
			return err
		}

		w.Flush()
		content := buf.String()
		file.Close()

		identifier, err := license.Identify(dependencyPath, content)
		if err != nil {
			return err
		}

		report.Resolve(&Result{
			Dependency:      filepath.Base(dependencyPath),
			LicenseFilePath: archiveFile,
			LicenseContent:  content,
			LicenseSpdxID:   identifier,
		})
		return nil
	}
	return fmt.Errorf("cannot find license file")
}
