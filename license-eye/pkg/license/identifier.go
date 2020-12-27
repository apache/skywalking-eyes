package license

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/apache/skywalking-eyes/license-eye/assets"
)

const templatesDir = "assets/lcs-templates"

// Identify identifies the Spdx ID of the given license content
func Identify(content string) (string, error) {
	content = Normalize(content)

	templates, err := assets.AssetDir(templatesDir)
	if err != nil {
		return "", err
	}

	for _, template := range templates {
		t, err := assets.Asset(filepath.Join(templatesDir, template))
		if err != nil {
			return "", err
		}
		license := string(t)
		license = Normalize(license)
		if license == content {
			return strings.TrimSuffix(template, filepath.Ext(template)), nil
		}
	}

	return "", fmt.Errorf("cannot identify license content")
}
