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
	"bytes"
	"encoding/xml"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

var unmarshalDependency = []struct {
	XMLValue string
	Expect   Dependency
	Err      error
}{
	{Expect: Dependency{}, XMLValue: ``, Err: io.EOF},
	{Expect: Dependency{}, XMLValue: `<groupId>skywalking</groupId>`, Err: xml.UnmarshalError("expected element type <dependency> but have <groupId>")},
	{Expect: Dependency{XMLName: xml.Name{Local: "dependency"}, GroupID: "skywalking", ArtifactID: "SkyWalking Eyes", Version: "0.1.0"}, XMLValue: `<dependency><groupId>skywalking</groupId><artifactId>SkyWalking Eyes</artifactId><version>0.1.0</version></dependency>`, Err: nil},
	{Expect: Dependency{XMLName: xml.Name{Local: "dependency"}, GroupID: "skywalking", ArtifactID: "SkyWalking Eyes", Version: "0.1.0"}, XMLValue: `<dependency><groupId>skywalking</groupId><artifactId>	SkyWalking Eyes	</artifactId><version>0.1.0</version></dependency>`, Err: nil},
	{Expect: Dependency{XMLName: xml.Name{Local: "dependency"}, GroupID: "skywalking", ArtifactID: "SkyWalking Eyes", Version: "0.1.0", Exclusions: &Exclusions{XMLName: xml.Name{Local: "exclusions"}, Value: []*Exclusion{{XMLName: xml.Name{Local: "exclusion"}, GroupID: "commons-logging", ArtifactID: "commons-logging"}}}}, XMLValue: `
	<dependency>
		<groupId>skywalking</groupId>
		<artifactId>	SkyWalking Eyes	</artifactId>
		<version>0.1.0</version>
		<exclusions>
			<exclusion>
				<groupId>commons-logging</groupId>
				<artifactId>commons-logging</artifactId>
			</exclusion>
		</exclusions>
	</dependency>`, Err: nil},
}

func TestUnmarshalDependency(t *testing.T) {
	for _, tt := range unmarshalDependency {
		r := bytes.NewReader([]byte(tt.XMLValue))
		dec := newXMLDecoder(r)

		v := Dependency{}
		err := dec.Decode(&v)
		require.Equal(t, tt.Err, err)
		require.Equal(t, tt.Expect, v)
	}
}

var unmarshalDependencies = []struct {
	XMLValue string
	Expect   Dependencies
	Err      error
}{
	{Expect: Dependencies{}, XMLValue: ``, Err: io.EOF},
	{Expect: Dependencies{}, XMLValue: `<dependency><groupId>skywalking</groupId></dependency>`, Err: xml.UnmarshalError("expected element type <dependencies> but have <dependency>")},
	{Expect: Dependencies{XMLName: xml.Name{Local: "dependencies"}, Value: []*Dependency{{XMLName: xml.Name{Local: "dependency"}, GroupID: "skywalking"}}}, XMLValue: `<dependencies><dependency><groupId>skywalking</groupId></dependency></dependencies>`, Err: nil},
}

func TestUnmarshalDependencies(t *testing.T) {
	for _, tt := range unmarshalDependencies {
		r := bytes.NewReader([]byte(tt.XMLValue))
		dec := newXMLDecoder(r)

		v := Dependencies{}
		err := dec.Decode(&v)
		require.Equal(t, tt.Err, err)
		require.Equal(t, tt.Expect, v)
	}
}

var unmarshalParent = []struct {
	XMLValue string
	Expect   Parent
	Err      error
}{
	{Expect: Parent{}, XMLValue: ``, Err: io.EOF},
	{Expect: Parent{}, XMLValue: `<groupId>skywalking</groupId>`, Err: xml.UnmarshalError("expected element type <parent> but have <groupId>")},
	{Expect: Parent{XMLName: xml.Name{Local: "parent"}, GroupID: "skywalking", ArtifactID: "SkyWalking Eyes"}, XMLValue: `<parent><groupId>skywalking</groupId><artifactId>SkyWalking Eyes</artifactId></parent>`, Err: nil},
}

func TestUnmarshalParent(t *testing.T) {
	for _, tt := range unmarshalParent {
		r := bytes.NewReader([]byte(tt.XMLValue))
		dec := newXMLDecoder(r)

		v := Parent{}
		err := dec.Decode(&v)
		require.Equal(t, tt.Err, err)
		require.Equal(t, tt.Expect, v)
	}
}

var unmarshalDependencyManagement = []struct {
	XMLValue string
	Expect   DependencyManagement
	Err      error
}{
	{Expect: DependencyManagement{}, XMLValue: ``, Err: io.EOF},
	{Expect: DependencyManagement{}, XMLValue: `<groupId>skywalking</groupId>`, Err: xml.UnmarshalError("expected element type <dependencyManagement> but have <groupId>")},
	{Expect: DependencyManagement{XMLName: xml.Name{Local: "dependencyManagement"}, Value: Dependencies{XMLName: xml.Name{Local: "dependencies"}, Value: []*Dependency{{XMLName: xml.Name{Local: "dependency"}, GroupID: "skywalking1"}, {XMLName: xml.Name{Local: "dependency"}, GroupID: "skywalking2"}}}}, XMLValue: `<dependencyManagement><dependencies><dependency><groupId>skywalking1</groupId></dependency><dependency><groupId>skywalking2</groupId></dependency></dependencies></dependencyManagement>`, Err: nil},
}

func TestUnmarshalDependencyManagement(t *testing.T) {
	for _, tt := range unmarshalDependencyManagement {
		r := bytes.NewReader([]byte(tt.XMLValue))
		dec := newXMLDecoder(r)

		v := DependencyManagement{}
		err := dec.Decode(&v)
		require.Equal(t, tt.Err, err)
		require.Equal(t, tt.Expect, v)
	}
}

var unmarshalModule = []struct {
	XMLValue string
	Expect   Module
	Err      error
}{
	{Expect: Module{Value: ""}, XMLValue: ``, Err: io.EOF},
	{Expect: Module{Value: ""}, XMLValue: `		`, Err: io.EOF},
	{Expect: Module{Value: ""}, XMLValue: `<nomodule>hello world</nomodule>`, Err: xml.UnmarshalError("expected element type <module> but have <nomodule>")},
	{Expect: Module{XMLName: xml.Name{Local: "module"}, Value: "hello world"}, XMLValue: `<module>hello world</module>`, Err: nil},
}

func TestUnmarshalModule(t *testing.T) {
	for _, tt := range unmarshalModule {
		r := bytes.NewReader([]byte(tt.XMLValue))
		dec := newXMLDecoder(r)

		v := Module{}
		err := dec.Decode(&v)
		require.Equal(t, tt.Err, err)
		require.Equal(t, tt.Expect, v)
	}
}

var unmarshalModules = []struct {
	XMLValue string
	Expect   Modules
	Err      error
}{
	{
		Expect:   Modules{},
		XMLValue: ``,
		Err:      io.EOF,
	},
	{
		Expect:   Modules{},
		XMLValue: `<module>hello world</module>`,
		Err:      xml.UnmarshalError("expected element type <modules> but have <module>"),
	},
	{
		Expect: Modules{
			XMLName: xml.Name{Local: "modules"},
			Values: []Module{
				{xml.Name{Local: "module"}, "hello world"},
			},
		},
		XMLValue: `<modules><module>hello world</module></modules>`,
		Err:      nil,
	},
	{
		Expect: Modules{
			XMLName: xml.Name{Local: "modules"},
			Values: []Module{
				{xml.Name{Local: "module"}, "module1"},
				{xml.Name{Local: "module"}, "module2"},
				{xml.Name{Local: "module"}, "module3"},
			},
		},
		XMLValue: `<modules>
		<module>module1</module>
		<module>module2</module>
		<module>module3</module>
		</modules>`,
		Err: nil,
	},
}

func TestUnmarshalModules(t *testing.T) {
	for _, tt := range unmarshalModules {
		r := bytes.NewReader([]byte(tt.XMLValue))
		dec := newXMLDecoder(r)

		v := Modules{}
		err := dec.Decode(&v)
		require.Equal(t, tt.Err, err)
		require.Equal(t, tt.Expect, v)
	}
}

var unmarshalProperties = []struct {
	XMLValue string
	Expect   Properties
	Err      error
}{
	{Expect: Properties{}, XMLValue: ``, Err: io.EOF},
	{Expect: Properties{}, XMLValue: `<groupId>skywalking</groupId>`, Err: xml.UnmarshalError("expected element type <properties> but have <groupId>")},
	{Expect: Properties{m: map[string]string{"compiler.version": "0.1.0", "project.build.sourceEncoding": "UTF-8"}}, XMLValue: `<properties><compiler.version>0.1.0</compiler.version><project.build.sourceEncoding>UTF-8</project.build.sourceEncoding></properties>`, Err: nil},
}

func TestUnmarshalProperties(t *testing.T) {
	for _, tt := range unmarshalProperties {
		r := bytes.NewReader([]byte(tt.XMLValue))
		dec := newXMLDecoder(r)

		v := Properties{}
		err := dec.Decode(&v)
		require.EqualValues(t, tt.Err, err)
		require.EqualValues(t, tt.Expect, v)
	}
}

var unmarshalLicenses = []struct {
	XMLValue string
	Expect   Licenses
	Err      error
}{
	{Expect: Licenses{}, XMLValue: ``, Err: io.EOF},
	{Expect: Licenses{}, XMLValue: `<groupId>skywalking</groupId>`, Err: xml.UnmarshalError("expected element type <licenses> but have <groupId>")},
	{Expect: Licenses{XMLName: xml.Name{Local: "licenses"}, Values: []License{{XMLName: xml.Name{Local: "license"}, Name: "The Apache Software License, Version 1.0", URL: "http://www.apache.org/licenses/LICENSE-1.0.txt"}, {XMLName: xml.Name{Local: "license"}, Name: "The Apache Software License, Version 2.0", URL: "http://www.apache.org/licenses/LICENSE-2.0.txt"}}}, XMLValue: `<licenses><license><name>The Apache Software License, Version 1.0</name><url>http://www.apache.org/licenses/LICENSE-1.0.txt</url></license><license><name>The Apache Software License, Version 2.0</name><url>http://www.apache.org/licenses/LICENSE-2.0.txt</url></license></licenses>`, Err: nil},
}

func TestUnmarshalLicenses(t *testing.T) {
	for _, tt := range unmarshalLicenses {
		r := bytes.NewReader([]byte(tt.XMLValue))
		dec := newXMLDecoder(r)

		v := Licenses{}
		err := dec.Decode(&v)
		require.Equal(t, tt.Err, err)
		require.Equal(t, tt.Expect, v)
	}
}

var unmarshalProfiles = []struct {
	XMLValue string
	Expect   Profiles
	Err      error
}{
	{Expect: Profiles{}, XMLValue: ``, Err: io.EOF},
	{Expect: Profiles{}, XMLValue: `<groupId>skywalking</groupId>`, Err: xml.UnmarshalError("expected element type <profiles> but have <groupId>")},
}

func TestUnmarshalProfiles(t *testing.T) {
	for _, tt := range unmarshalProfiles {
		r := bytes.NewReader([]byte(tt.XMLValue))
		dec := newXMLDecoder(r)

		v := Profiles{}
		err := dec.Decode(&v)
		require.Equal(t, tt.Err, err)
		require.Equal(t, tt.Expect, v)
	}
}

var marshalTests = []struct {
	Name      string
	Value     interface{}
	ExpectXML string
	Err       error
}{
	{Name: "Dependency", Value: Dependency{}, ExpectXML: `<dependency></dependency>`, Err: nil},
	{Name: "Dependency", Value: Dependency{GroupID: "skywalking"}, ExpectXML: `<dependency><groupId>skywalking</groupId></dependency>`, Err: nil},
	{Name: "Dependency", Value: Dependency{GroupID: "skywalking", Exclusions: &Exclusions{XMLName: xml.Name{Local: "exclusions"}, Value: []*Exclusion{{XMLName: xml.Name{Local: "exclusion"}, GroupID: "commons-logging", ArtifactID: "commons-logging"}}}}, ExpectXML: `<dependency><groupId>skywalking</groupId><exclusions><exclusion><groupId>commons-logging</groupId><artifactId>commons-logging</artifactId></exclusion></exclusions></dependency>`, Err: nil},
	{Name: "Dependency", Value: Dependency{GroupID: "skywalking", ArtifactID: "SkyWalking Eyes", Version: "0.1.0"}, ExpectXML: `<dependency><groupId>skywalking</groupId><artifactId>SkyWalking Eyes</artifactId><version>0.1.0</version></dependency>`, Err: nil},

	{Name: "Dependencies", Value: Dependencies{}, ExpectXML: `<dependencies></dependencies>`, Err: nil},
	{Name: "Dependencies", Value: Dependencies{xml.Name{}, []*Dependency{{GroupID: "skywalking"}}}, ExpectXML: `<dependencies><dependency><groupId>skywalking</groupId></dependency></dependencies>`, Err: nil},
	{Name: "Dependencies", Value: Dependencies{xml.Name{}, []*Dependency{{GroupID: "skywalking1"}, {GroupID: "skywalking2"}}}, ExpectXML: `<dependencies><dependency><groupId>skywalking1</groupId></dependency><dependency><groupId>skywalking2</groupId></dependency></dependencies>`, Err: nil},

	{Name: "Parent", Value: Parent{}, ExpectXML: `<parent><artifactId></artifactId></parent>`, Err: nil},
	{Name: "Parent", Value: Parent{GroupID: "skywalking", ArtifactID: "SkyWalking Eyes"}, ExpectXML: `<parent><groupId>skywalking</groupId><artifactId>SkyWalking Eyes</artifactId></parent>`, Err: nil},

	{Name: "DependencyManagement", Value: DependencyManagement{}, ExpectXML: `<dependencyManagement><dependencies></dependencies></dependencyManagement>`, Err: nil},
	{Name: "DependencyManagement", Value: DependencyManagement{Value: Dependencies{Value: []*Dependency{{GroupID: "skywalking"}}}}, ExpectXML: `<dependencyManagement><dependencies><dependency><groupId>skywalking</groupId></dependency></dependencies></dependencyManagement>`, Err: nil},
	{Name: "DependencyManagement", Value: DependencyManagement{Value: Dependencies{Value: []*Dependency{{GroupID: "skywalking1"}, {GroupID: "skywalking2"}}}}, ExpectXML: `<dependencyManagement><dependencies><dependency><groupId>skywalking1</groupId></dependency><dependency><groupId>skywalking2</groupId></dependency></dependencies></dependencyManagement>`, Err: nil},

	{Name: "Module", Value: Module{Value: ""}, ExpectXML: `<module></module>`, Err: nil},
	{Name: "Module", Value: Module{Value: "hello world"}, ExpectXML: `<module>hello world</module>`, Err: nil},

	{Name: "Modules", Value: Modules{}, ExpectXML: `<modules></modules>`, Err: nil},
	{Name: "Modules", Value: Modules{Values: []Module{{xml.Name{}, ""}}}, ExpectXML: `<modules><module></module></modules>`, Err: nil},
	{Name: "Modules", Value: Modules{Values: []Module{{xml.Name{}, "hello"}}}, ExpectXML: `<modules><module>hello</module></modules>`, Err: nil},
	{Name: "Modules", Value: Modules{Values: []Module{{xml.Name{}, "module1"}, {xml.Name{}, "module2"}}}, ExpectXML: `<modules><module>module1</module><module>module2</module></modules>`, Err: nil},

	{Name: "Properties", Value: Properties{}, ExpectXML: ``, Err: nil},
	{Name: "Properties", Value: Properties{m: map[string]string{"compiler.version": "0.1.0"}}, ExpectXML: `<properties><compiler.version>0.1.0</compiler.version></properties>`, Err: nil},
	{Name: "Properties", Value: Properties{m: map[string]string{"compiler.version": "0.1.0", "project.build.sourceEncoding": "UTF-8"}}, ExpectXML: `<properties><compiler.version>0.1.0</compiler.version><project.build.sourceEncoding>UTF-8</project.build.sourceEncoding></properties>`, Err: nil},

	{Name: "Licenses", Value: Licenses{}, ExpectXML: `<licenses></licenses>`, Err: nil},
	{Name: "Licenses", Value: Licenses{Values: []License{{Name: "The Apache Software License, Version 2.0", URL: "http://www.apache.org/licenses/LICENSE-2.0.txt"}}}, ExpectXML: `<licenses><license><name>The Apache Software License, Version 2.0</name><url>http://www.apache.org/licenses/LICENSE-2.0.txt</url></license></licenses>`, Err: nil},
	{Name: "Licenses", Value: Licenses{Values: []License{{Name: "The Apache Software License, Version 1.0", URL: "http://www.apache.org/licenses/LICENSE-1.0.txt"}, {Name: "The Apache Software License, Version 2.0", URL: "http://www.apache.org/licenses/LICENSE-2.0.txt"}}}, ExpectXML: `<licenses><license><name>The Apache Software License, Version 1.0</name><url>http://www.apache.org/licenses/LICENSE-1.0.txt</url></license><license><name>The Apache Software License, Version 2.0</name><url>http://www.apache.org/licenses/LICENSE-2.0.txt</url></license></licenses>`, Err: nil},

	{Name: "Profiles", Value: Profiles{}, ExpectXML: `<profiles></profiles>`, Err: nil},
	{Name: "Profiles", Value: Profiles{Values: []Profile{{Modules: Modules{Values: []Module{{xml.Name{}, "module1"}, {xml.Name{}, "module2"}}}}}}, ExpectXML: `<profiles><profile><modules><module>module1</module><module>module2</module></modules></profile></profiles>`, Err: nil},
	{Name: "Profiles", Value: Profiles{Values: []Profile{{Modules: Modules{Values: []Module{{xml.Name{}, "module1"}, {xml.Name{}, "module2"}}}}, {Modules: Modules{Values: []Module{{xml.Name{}, "module3"}, {xml.Name{}, "module4"}}}}}}, ExpectXML: `<profiles><profile><modules><module>module1</module><module>module2</module></modules></profile><profile><modules><module>module3</module><module>module4</module></modules></profile></profiles>`, Err: nil},
}

func TestMarshal(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	enc := xml.NewEncoder(buf)
	for _, tt := range marshalTests {
		t.Run(tt.Name, func(t *testing.T) {
			buf.Reset()
			err := enc.Encode(tt.Value)
			require.Equal(t, tt.Err, err)
			require.Equal(t, tt.ExpectXML, buf.String())

			buf.Reset()
			err = enc.Encode(&tt.Value)
			require.Equal(t, tt.Err, err)
			require.Equal(t, tt.ExpectXML, buf.String())
		})
	}
}
