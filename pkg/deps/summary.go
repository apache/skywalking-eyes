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
	"bytes"
	"os"
	"sort"
	"text/template"

	"github.com/apache/skywalking-eyes/pkg/header"
	"github.com/apache/skywalking-eyes/pkg/license"

	"github.com/Masterminds/sprig/v3"
)

type SummaryRenderContext struct {
	LicenseContent string                       // Current project license content
	Groups         []*SummaryRenderLicenseGroup // All dependency license groups
}

type SummaryRenderLicenseGroup struct {
	LicenseID string                  // Aggregate all same license ID dependencies
	Deps      []*SummaryRenderLicense // Same license ID dependencies
}

type SummaryRenderLicense struct {
	Name      string // Dependency name
	Version   string // Dependency version
	LicenseID string // License ID
}

func ParseTemplateFromString(templateStr string) (*template.Template, error) {
	return template.New("summary").Funcs(sprig.TxtFuncMap()).Parse(templateStr)
}
func ParseTemplate(path string) (*template.Template, error) {
	tpl, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseTemplateFromString(string(tpl))
}

// GenerateSummary generate the summary content by template, license config and dependency report
func GenerateSummary(tpl *template.Template, head *header.ConfigHeader, rep *Report) (string, error) {
	var r bytes.Buffer
	context, err := generateSummaryRenderContext(head, rep)
	if err != nil {
		return "", err
	}
	if err := tpl.Execute(&r, context); err != nil {
		return "", err
	}
	return r.String(), nil
}

func generateSummaryRenderContext(head *header.ConfigHeader, rep *Report) (*SummaryRenderContext, error) {
	// the license id of the project
	var headerContent string
	if head.License.SpdxID != "" {
		c, err := license.GetLicenseContent(head.License.SpdxID)
		if err != nil {
			return nil, err
		}
		headerContent = c
	}
	if headerContent == "" {
		headerContent = head.GetLicenseContent()
	}

	groups := make(map[string]*SummaryRenderLicenseGroup)
	for _, r := range rep.Resolved {
		group := groups[r.LicenseSpdxID]
		if group == nil {
			group = &SummaryRenderLicenseGroup{
				LicenseID: r.LicenseSpdxID,
				Deps:      make([]*SummaryRenderLicense, 0),
			}
			groups[r.LicenseSpdxID] = group
		}

		group.Deps = append(group.Deps, &SummaryRenderLicense{
			Name:      r.Dependency,
			Version:   r.Version,
			LicenseID: r.LicenseSpdxID,
		})
	}

	groupArray := make([]*SummaryRenderLicenseGroup, 0)
	for _, g := range groups {
		groupArray = append(groupArray, g)
		sort.SliceStable(g.Deps, func(i, j int) bool {
			return g.Deps[i].Name < g.Deps[j].Name
		})
	}
	sort.SliceStable(groupArray, func(i, j int) bool {
		return groupArray[i].LicenseID < groupArray[j].LicenseID
	})
	return &SummaryRenderContext{
		LicenseContent: headerContent,
		Groups:         groupArray,
	}, nil
}
