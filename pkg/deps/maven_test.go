package deps_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/apache/skywalking-eyes/license-eye/pkg/deps"
)

func TestCanResolve(t *testing.T) {
	resolver := new(deps.MavenPomResolver)
	for _, test := range []struct {
		fileName string
		exp      bool
	}{
		{"pom.xml", true},
		{"POM.XML", true},
		{"log4j-1.2.12.pom", true},
		{".pom", false},
	} {
		b := resolver.CanResolve(test.fileName)
		if b != test.exp {
			t.Errorf("MavenPomResolver.CanResolve(\"%v\") = %v, want %v", test.fileName, b, test.exp)
		}
	}
}

func TestResolve(t *testing.T) {
	resolver := new(deps.MavenPomResolver)
	for _, test := range []struct {
		pomFile string
	}{
		{"../../test/testdata/pom.xml"},
	} {
		if resolver.CanResolve(test.pomFile) {
			report := deps.Report{}
			if err := resolver.Resolve(test.pomFile, &report); err != nil {
				t.Error(err)
			}

			fmt.Println(report.String())

			if skipped := len(report.Skipped); skipped > 0 {
				pkgs := make([]string, skipped)
				for i, s := range report.Skipped {
					pkgs[i] = s.Dependency
				}
				t.Errorf(
					"failed to identify the licenses of following packages (%d):\n%s",
					len(pkgs), strings.Join(pkgs, "\n"),
				)
			}
		}
	}
}
