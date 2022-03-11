package license

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/apache/skywalking-eyes/assets"
	"github.com/apache/skywalking-eyes/internal/logger"
)

var templatesDirs = []string{
	"lcs-templates",
	// Some projects simply use the header text as their LICENSE content...
	"header-templates",
}

var dualLicensePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)This project is covered by two different licenses: (?P<license>[^.]+)`),
}

var normalizedTemplates = sync.Map{}

func init() {
	wg := sync.WaitGroup{}
	for _, dir := range templatesDirs {
		files, err := assets.AssetDir(dir)
		if err != nil {
			logger.Log.Fatalln("Failed to read license template directory:", dir, err)
		}
		wg.Add(len(files))
		for _, template := range files {
			go loadTemplate(&wg, dir, template)
		}
	}
	wg.Wait()
}

func loadTemplate(wg *sync.WaitGroup, dir string, template fs.DirEntry) {
	defer wg.Done()

	name := template.Name()
	t, err := assets.Asset(filepath.Join(dir, name))
	if err != nil {
		logger.Log.Fatalln("Failed to read license template:", dir, err)
	}
	normalizedTemplates.Store(dir+"/"+name, Normalize(string(t)))
}

// Identify identifies the Spdx ID of the given license content
func Identify(pkgPath, content string) (string, error) {
	for _, pattern := range dualLicensePatterns {
		matches := pattern.FindStringSubmatch(content)
		for i, name := range pattern.SubexpNames() {
			if name == "license" && len(matches) >= i {
				return matches[i], nil
			}
		}
	}

	content = Normalize(content)
	logger.Log.Debugf("Normalized content for %+v:\n%+v\n", pkgPath, content)

	result := make(chan string, 1)
	normalizedTemplates.Range(func(key, value interface{}) bool {
		name := key.(string)
		license := value.(string)

		// Should not use `Contains` as a root LICENSE file may include other licenses the project uses,
		// `Contains` would identify the last one license as the project's license.
		if strings.HasPrefix(content, license) {
			name = filepath.Base(name)
			result <- strings.TrimSuffix(name, filepath.Ext(name))
			return false
		}
		return true
	})
	select {
	case license := <-result:
		return license, nil
	default:
		return "", fmt.Errorf("cannot identify license content")
	}
}
