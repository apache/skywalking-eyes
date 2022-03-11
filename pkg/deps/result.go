package deps

import (
	"fmt"
	"math"
	"strings"
)

type SpdxID string

const (
	Unknown string = "Unknown"
)

// Result is a single item that represents a resolved dependency license.
type Result struct {
	Dependency      string
	LicenseFilePath string
	LicenseContent  string
	LicenseSpdxID   string
	ResolveErrors   []error
	Version         string
}

// Report is a collection of resolved Result.
type Report struct {
	Resolved []*Result
	Skipped  []*Result
}

// Resolve marks the dependency's license is resolved.
func (report *Report) Resolve(result *Result) {
	report.Resolved = append(report.Resolved, result)
}

// Skip marks the dependency's license is skipped for some reasons.
func (report *Report) Skip(result *Result) {
	report.Skipped = append(report.Skipped, result)
}

func (report *Report) String() string {
	dWidth, lWidth, vWidth := .0, .0, .0
	for _, r := range report.Skipped {
		dWidth = math.Max(float64(len(r.Dependency)), dWidth)
		lWidth = math.Max(float64(len(r.LicenseSpdxID)), lWidth)
		vWidth = math.Max(float64(len(r.Version)), vWidth)
	}
	for _, r := range report.Resolved {
		dWidth = math.Max(float64(len(r.Dependency)), dWidth)
		lWidth = math.Max(float64(len(r.LicenseSpdxID)), lWidth)
		vWidth = math.Max(float64(len(r.Version)), vWidth)
	}

	rowTemplate := fmt.Sprintf("%%-%dv | %%%dv | %%%dv\n", int(dWidth), int(lWidth), int(vWidth))
	s := fmt.Sprintf(rowTemplate, "Dependency", "License", "Version")
	s += fmt.Sprintf(rowTemplate, strings.Repeat("-", int(dWidth)), strings.Repeat("-", int(lWidth)), strings.Repeat("-", int(vWidth)))
	for _, r := range report.Resolved {
		s += fmt.Sprintf(rowTemplate, r.Dependency, r.LicenseSpdxID, r.Version)
	}
	for _, r := range report.Skipped {
		s += fmt.Sprintf(rowTemplate, r.Dependency, Unknown, r.Version)
	}

	return s
}
