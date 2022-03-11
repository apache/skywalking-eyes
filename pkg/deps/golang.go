package deps

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/build"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/license"

	"golang.org/x/tools/go/packages"
)

type GoModResolver struct {
	Resolver
}

func (resolver *GoModResolver) CanResolve(file string) bool {
	base := filepath.Base(file)
	logger.Log.Debugln("Base name:", base)
	return base == "go.mod"
}

// Resolve resolves licenses of all dependencies declared in the go.mod file.
func (resolver *GoModResolver) Resolve(goModFile string, report *Report) error {
	if err := os.Chdir(filepath.Dir(goModFile)); err != nil {
		return err
	}

	goModDownload := exec.Command("go", "mod", "download")
	logger.Log.Debugf("Run command: %v, please wait", goModDownload.String())
	goModDownload.Stdout = os.Stdout
	goModDownload.Stderr = os.Stderr
	if err := goModDownload.Run(); err != nil {
		return err
	}

	output, err := exec.Command("go", "mod", "download", "-json").Output()
	if err != nil {
		return err
	}

	modules := make([]*packages.Module, 0)
	decoder := json.NewDecoder(bytes.NewReader(output))
	for {
		var m packages.Module
		if err := decoder.Decode(&m); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		modules = append(modules, &m)
	}

	logger.Log.Debugln("Module size:", len(modules))

	return resolver.ResolvePackages(modules, report)
}

// ResolvePackages resolves the licenses of the given packages.
func (resolver *GoModResolver) ResolvePackages(modules []*packages.Module, report *Report) error {
	for _, module := range modules {
		err := resolver.ResolvePackageLicense(module, report)
		if err != nil {
			logger.Log.Warnf("Failed to resolve the license of <%s>: %v\n", module.Path, err)
			report.Skip(&Result{
				Dependency:    module.Path,
				LicenseSpdxID: Unknown,
				Version:       module.Version,
			})
		}
	}
	return nil
}

var possibleLicenseFileName = regexp.MustCompile(`(?i)^LICENSE|LICENCE(\.txt)?|COPYING(\.txt)?$`)

func (resolver *GoModResolver) ResolvePackageLicense(module *packages.Module, report *Report) error {
	dir := module.Dir

	for {
		logger.Log.Debugf("Directory of %+v is %+v", module.Path, dir)
		files, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, info := range files {
			if !possibleLicenseFileName.MatchString(info.Name()) {
				continue
			}
			licenseFilePath := filepath.Join(dir, info.Name())
			content, err := os.ReadFile(licenseFilePath)
			if err != nil {
				return err
			}
			identifier, err := license.Identify(module.Path, string(content))
			if err != nil {
				return err
			}
			report.Resolve(&Result{
				Dependency:      module.Path,
				LicenseFilePath: licenseFilePath,
				LicenseContent:  string(content),
				LicenseSpdxID:   identifier,
				Version:         module.Version,
			})
			return nil
		}
		if resolver.shouldStopAt(dir, module.Dir) {
			break
		}
		dir = filepath.Dir(dir)
	}
	return fmt.Errorf("cannot find license file")
}

func (resolver *GoModResolver) shouldStopAt(dir, moduleDir string) bool {
	return dir == moduleDir || dir == build.Default.GOPATH
}
