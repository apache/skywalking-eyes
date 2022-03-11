package deps

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/apache/skywalking-eyes/assets"
	"github.com/apache/skywalking-eyes/internal/logger"
)

type compatibilityMatrix struct {
	Compatible   []string `yaml:"compatible"`
	Incompatible []string `yaml:"incompatible"`
}

var matrices = make(map[string]compatibilityMatrix)

func init() {
	dir := "compatibility"
	files, err := assets.AssetDir(dir)
	if err != nil {
		logger.Log.Fatalln("Failed to list assets/compatibility directory:", err)
	}
	for _, file := range files {
		name := file.Name()
		matrix := compatibilityMatrix{}
		if bytes, err := assets.Asset(filepath.Join(dir, name)); err != nil {
			logger.Log.Fatalln("Failed to read compatibility file:", name, err)
		} else if err := yaml.Unmarshal(bytes, &matrix); err != nil {
			logger.Log.Fatalln("Failed to unmarshal compatibility file:", file, err)
		}
		matrices[strings.TrimSuffix(name, filepath.Ext(name))] = matrix
	}
}

func Check(mainLicenseSpdxID string, config *ConfigDeps) error {
	report := Report{}
	if err := Resolve(config, &report); err != nil {
		return nil
	}

	matrix := matrices[mainLicenseSpdxID]
	var incompatibleResults []*Result
	for _, result := range append(report.Resolved, report.Skipped...) {
		compare := func(list []string) bool {
			for _, com := range list {
				if result.LicenseSpdxID == com {
					return true
				}
			}
			return false
		}
		if compatible := compare(matrix.Compatible); compatible {
			continue
		}
		if incompatible := compare(matrix.Incompatible); incompatible {
			incompatibleResults = append(incompatibleResults, result)
		}
	}

	if len(incompatibleResults) > 0 {
		str := ""
		for _, r := range incompatibleResults {
			str += fmt.Sprintf("\nLicense: %v Dependency: %v", r.LicenseSpdxID, r.Dependency)
		}
		return fmt.Errorf("the following licenses are incompatible with the main license: %v %v", mainLicenseSpdxID, str)
	}

	return nil
}
