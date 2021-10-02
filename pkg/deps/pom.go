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
	"sort"
	"strings"

	"golang.org/x/net/html/charset"
)

func NewPomFile() *PomFile {
	return &PomFile{
		Header: xml.Header,
	}
}

type xmlDecoder struct {
	*xml.Decoder
}

func (dec xmlDecoder) Token() (xml.Token, error) {
	t, err := dec.Decoder.Token()
	if data, ok := t.(xml.CharData); ok {
		t = xml.CharData(bytes.TrimSpace(data))
	}
	return t, err
}

func newXMLDecoder(r io.Reader) *xml.Decoder {
	rawDec := xml.NewDecoder(r)
	dec := xml.NewTokenDecoder(xmlDecoder{rawDec})
	dec.CharsetReader = charset.NewReaderLabel
	return dec
}

func DecodePomFile(r io.Reader) (*PomFile, error) {
	dec := newXMLDecoder(r)

	pomFile := NewPomFile()

	var start xml.StartElement

	buf := bytes.NewBuffer(nil)
loop:
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		switch tok := tok.(type) {
		case xml.Comment:
			buf.Write(tok.Copy())

		case xml.StartElement:
			start = tok
			break loop
		}
	}
	pomFile.HeaderComment = strings.TrimSpace(buf.String())

	err := dec.DecodeElement(pomFile, &start)
	if err != nil {
		return nil, err
	}

	return pomFile, nil
}

type PomFile struct {
	Header         string         `xml:"-"`
	HeaderComment  string         `xml:"-"`
	XMLName        xml.Name       `xml:"project"`
	NameSpace      NameSpace      `xml:"xmlns,attr"`
	SchemeInstance SchemeInstance `xml:"xsi,attr"`
	SchemaLocation SchemaLocation `xml:"schemaLocation,attr"`
	ModelVersion   ModelVersion   `xml:"modelVersion"`

	// The Basics
	GroupID              string                `xml:"groupId,omitempty"`
	ArtifactID           string                `xml:"artifactId"`
	Version              string                `xml:"version,omitempty"`
	Packaging            string                `xml:"packaging,omitempty"`
	Dependencies         *Dependencies         `xml:"dependencies,omitempty"`
	Parent               *Parent               `xml:"parent,omitempty"`
	DependencyManagement *DependencyManagement `xml:"dependencyManagement,omitempty"`
	Modules              *Modules              `xml:"modules,omitempty"`
	Properties           *Properties           `xml:"properties,omitempty"`

	// More Project Information
	Licenses *Licenses `xml:"licenses,omitempty"`

	// Environment Settings
	Profiles *Profiles `xml:"profiles,omitempty"`
}

func (pom *PomFile) Encode() []byte {
	w := bytes.NewBuffer(nil)
	err := pom.EncodeToWriter(w)
	if err != nil {
		return nil
	}
	return w.Bytes()
}

func (pom *PomFile) EncodeToWriter(w io.Writer) error {
	enc := xml.NewEncoder(w)
	_, err := w.Write([]byte(pom.Header))
	if err != nil {
		return err
	}
	if len(pom.HeaderComment) > 0 {
		err = enc.EncodeToken(xml.Comment(pom.HeaderComment))
		if err != nil {
			return err
		}
	}
	return enc.Encode(pom)
}

type NameSpace string

func (ns *NameSpace) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: xml.Name{Space: "", Local: "xmlns"}, Value: "http://maven.apache.org/POM/4.0.0"}, nil
}

type SchemeInstance string

func (si *SchemeInstance) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: xml.Name{Space: "", Local: "xmlns:xsi"}, Value: "http://www.w3.org/2001/XMLSchema-instance"}, nil
}

type SchemaLocation string

func (sl *SchemaLocation) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: xml.Name{Space: "", Local: "xsi:schemaLocation"},
		Value: "http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd"}, nil
}

type ModelVersion string

func (mv *ModelVersion) MarshalText() (text []byte, err error) {
	return []byte("4.0.0"), nil
}

type Dependencies struct {
	XMLName xml.Name      `xml:"dependencies"`
	Value   []*Dependency `xml:"dependency,omitempty"`
}

type Dependency struct {
	XMLName    xml.Name    `xml:"dependency"`
	GroupID    string      `xml:"groupId,omitempty"`
	ArtifactID string      `xml:"artifactId,omitempty"`
	Version    string      `xml:"version,omitempty"`
	Classifier string      `xml:"classifier,omitempty"`
	Type       string      `xml:"type,omitempty"`
	Scope      string      `xml:"scope,omitempty"`
	SystemPath string      `xml:"systemPath,omitempty"`
	Optional   bool        `xml:"optional,omitempty"`
	Exclusions *Exclusions `xml:"exclusions,omitempty"`
}

type Exclusions struct {
	XMLName xml.Name     `xml:"exclusions"`
	Value   []*Exclusion `xml:"exclusion,omitempty"`
}

type Exclusion struct {
	XMLName    xml.Name `xml:"exclusion"`
	GroupID    string   `xml:"groupId,omitempty"`
	ArtifactID string   `xml:"artifactId,omitempty"`
}

type Parent struct {
	XMLName      xml.Name `xml:"parent"`
	GroupID      string   `xml:"groupId,omitempty"`
	ArtifactID   string   `xml:"artifactId"`
	Version      string   `xml:"version,omitempty"`
	RelativePath string   `xml:"relativePath,omitempty"`
}

type DependencyManagement struct {
	XMLName      xml.Name     `xml:"dependencyManagement"`
	Dependencies Dependencies `xml:"dependencies,omitempty"`
}

type Modules struct {
	XMLName xml.Name `xml:"modules"`
	Values  []Module `xml:"module,omitempty"`
}

type Module struct {
	XMLName xml.Name `xml:"module,omitempty"`
	Value   string   `xml:",chardata"`
}

type Licenses struct {
	XMLName xml.Name  `xml:"licenses"`
	Values  []License `xml:"license"`
}

type License struct {
	XMLName      xml.Name `xml:"license"`
	Name         string   `xml:"name,omitempty"`
	URL          string   `xml:"url,omitempty"`
	Distribution string   `xml:"distribution,omitempty"`
	Comments     string   `xml:"comments,omitempty"`
}

type Properties struct {
	XMLName xml.Name          `xml:"properties"`
	m       map[string]string `xml:"-"`
}

const (
	propertiesName = "properties"
)

func (p *Properties) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if start.Name.Local != propertiesName {
		return xml.UnmarshalError("expected element type <" + propertiesName + "> but have <" + start.Name.Local + ">")
	}

	p.m = map[string]string{}
	var key string
	var value string
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if tokenType, ok := token.(xml.StartElement); ok {
			key = tokenType.Name.Local
			err := d.DecodeElement(&value, &start)
			if err != nil {
				return err
			}
			p.m[key] = value
		}
	}
	return nil
}

func (p Properties) MarshalXML(e *xml.Encoder, start xml.StartElement) (err error) {
	if len(p.m) == 0 {
		return nil
	}

	start.Name.Local = "properties"
	err = e.EncodeToken(start)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(p.m))

	for k := range p.m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	for _, k := range keys {
		v := p.m[k]
		s := xml.StartElement{Name: xml.Name{Local: k}}
		err = e.EncodeToken(s)
		if err != nil {
			return err
		}

		err = e.EncodeToken(xml.CharData(v))
		if err != nil {
			return err
		}

		end := s.End()
		err = e.EncodeToken(end)
		if err != nil {
			return err
		}
	}

	end := start.End()
	return e.EncodeToken(end)
}

type Profiles struct {
	XMLName xml.Name  `xml:"profiles"`
	Values  []Profile `xml:"profile,omitempty"`
}

type Profile struct {
	XMLName xml.Name `xml:"profile"`
	ID      string   `xml:"id,omitempty"`
	Modules Modules  `xml:"modules,omitempty"`
}
