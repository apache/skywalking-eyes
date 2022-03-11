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

package deps_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/deps"
)

func TestCanResolveJarFile(t *testing.T) {
	resolver := new(deps.JarResolver)
	for _, test := range []struct {
		fileName string
		exp      bool
	}{
		{"1.jar", true},
		{"/tmp/1.jar", true},
		{"1.jar2", false},
		{"protobuf-java-3.13.0.jar", true},
		{"slf4j-api-1.7.25.jar", true},
	} {
		b := resolver.CanResolve(test.fileName)
		if b != test.exp {
			t.Errorf("JarResolver.CanResolve(\"%v\") = %v, want %v", test.fileName, b, test.exp)
		}
	}
}

func copyJars(t *testing.T, pomFile, content string) ([]string, error) {
	dir := filepath.Dir(pomFile)

	if err := os.Chdir(dir); err != nil {
		return nil, err
	}

	if err := dumpPomFile(pomFile, content); err != nil {
		return nil, err
	}

	if _, err := exec.Command("mvn", "dependency:copy-dependencies", "-DoutputDirectory=./lib", "-DincludeScope=runtime").Output(); err != nil {
		return nil, err
	}

	jars := []string{}
	files, err := ioutil.ReadDir(filepath.Join(dir, "lib"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			jars = append(jars, filepath.Join(dir, "lib", file.Name()))
		}
	}

	return jars, nil
}

func TestResolveJar(t *testing.T) {
	if _, err := exec.Command("mvn", "--version").Output(); err != nil {
		logger.Log.Warnf("Failed to find mvn, the test `TestResolveJar` was skipped")
		return
	}

	resolver := new(deps.JarResolver)

	path, err := tmpDir()
	if err != nil {
		t.Error(err)
		return
	}
	defer destroyTmpDir(t, path)

	pomFile := filepath.Join(path, "pom.xml")

	for _, test := range []struct {
		pomContent string
		cnt        int
	}{
		{`<?xml version="1.0" encoding="UTF-8"?>
	<project xmlns="http://maven.apache.org/POM/4.0.0"
		xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
		<modelVersion>4.0.0</modelVersion>
	
		<groupId>apache</groupId>
		<artifactId>skywalking-eyes</artifactId>
		<version>1.0</version>
	
		<dependencies>
			<!-- https://mvnrepository.com/artifact/junit/junit -->
			<dependency>
				<groupId>junit</groupId>
				<artifactId>junit</artifactId>
				<version>4.12</version>
			</dependency>
			<!-- https://mvnrepository.com/artifact/commons-logging/commons-logging -->
			<dependency>
				<groupId>commons-logging</groupId>
				<artifactId>commons-logging</artifactId>
				<version>1.2</version>
			</dependency>
			<!-- https://mvnrepository.com/artifact/org.apache.commons/commons-math3 -->
			<dependency>
				<groupId>org.apache.commons</groupId>
				<artifactId>commons-math3</artifactId>
				<version>3.6.1</version>
			</dependency>
		</dependencies>
	</project>`, 4},
	} {
		jars, err := copyJars(t, pomFile, test.pomContent)
		if err != nil {
			t.Error(err)
			return
		}

		report := deps.Report{}
		for _, jar := range jars {
			if resolver.CanResolve(jar) {
				if err := resolver.Resolve(jar, &report); err != nil {
					t.Error(err)
					return
				}

			}
		}
		if len(report.Resolved)+len(report.Skipped) != test.cnt {
			t.Errorf("the expected number of jar packages is: %d, but actually: %d. result:\n%v", test.cnt, len(report.Resolved)+len(report.Skipped), report.String())
		}
		fmt.Println(report.String())
	}
}
