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
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/apache/skywalking-eyes/pkg/deps"
)

func TestCanResolvePomFile(t *testing.T) {
	resolver := new(deps.MavenPomResolver)
	for _, test := range []struct {
		fileName string
		exp      bool
	}{
		{"pom.xml", true},
		{"POM.XML", false},
		{"log4j-1.2.12.pom", false},
		{".pom", false},
	} {
		b := resolver.CanResolve(test.fileName)
		if b != test.exp {
			t.Errorf("MavenPomResolver.CanResolve(\"%v\") = %v, want %v", test.fileName, b, test.exp)
		}
	}
}

func writeFile(fileName, content string) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	write := bufio.NewWriter(file)
	_, err = write.WriteString(content)
	if err != nil {
		return err
	}

	_ = write.Flush()
	return nil
}

func ensureDir(dirName string) error {
	return os.MkdirAll(dirName, 0777)
}

//go:embed testdata/maven/*
var testAssets embed.FS

func TestResolveMaven(t *testing.T) {
	resolver := new(deps.MavenPomResolver)
	tempDir := t.TempDir()
	base := "testdata/maven"

	fs.WalkDir(testAssets, base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		filename := filepath.Join(tempDir, strings.Replace(path, base, "", 1))
		if err := ensureDir(filepath.Dir(filename)); err != nil {
			return err
		}

		content, err := testAssets.ReadFile(path)
		if err != nil {
			t.Error(err)
		}
		writeFile(filename, string(content))

		return nil
	})

	pomFile := filepath.Join(tempDir, "pom.xml")

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
			<!-- https://mvnrepository.com/artifact/org.apache.skywalking/skywalking-sharing-server-plugin -->
			<dependency>
				<groupId>org.apache.skywalking</groupId>
				<artifactId>skywalking-sharing-server-plugin</artifactId>
				<version>8.6.0</version>
			</dependency>
			<dependency>
				<groupId>com.fasterxml.jackson.datatype</groupId>
				<artifactId>jackson-datatype-jsr310</artifactId>
				<version>2.13.3</version>
			</dependency>
		</dependencies>
	</project>`, 110},
	} {
		_ = writeFile(pomFile, test.pomContent)

		config := deps.ConfigDeps{}
		config.Finalize("")

		if resolver.CanResolve(pomFile) {
			report := deps.Report{}
			if err := resolver.Resolve(pomFile, &config, &report); err != nil {
				t.Error(err)
				return
			}

			if len(report.Resolved)+len(report.Skipped) != test.cnt {
				t.Errorf("the expected number of jar packages is: %d, but actually: %d. result:\n%v", test.cnt, len(report.Resolved)+len(report.Skipped), report.String())
			}
			fmt.Println(report.String())
		}
	}
}
