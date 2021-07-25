package deps

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/apache/skywalking-eyes/license-eye/internal/logger"
	"github.com/apache/skywalking-eyes/license-eye/pkg/license"
)

func init() {
	var err error
	_, err = exec.Command("mvn", "--version").Output()
	if err != nil {
		return
	}

	Resolvers = append(Resolvers, new(MavenPomResolver))
}

// TempDirGenerator Create and destroy one temporary directory
type TempDirGenerator interface {
	Create() (string, error)
	Destroy() error
}

type tempDir struct {
	dir string
}

func (t *tempDir) Create() (string, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	t.dir = tmpDir
	return tmpDir, nil
}

func (t *tempDir) Destroy() error {
	if t.dir == "" {
		return fmt.Errorf("the temporary directory does not exist")
	}
	return os.RemoveAll(t.dir)
}

func NewTempDirGenerator() TempDirGenerator {
	return new(tempDir)
}

var possiblePomFileName = regexp.MustCompile(`(?i)^pom\.xml|.+\.pom$`)

type MavenPomResolver struct{}

func (resolver *MavenPomResolver) CanResolve(file string) bool {
	base := filepath.Base(file)
	logger.Log.Debugln("Base name:", base)
	return possiblePomFileName.MatchString(base)
}

// Resolve resolves licenses of all dependencies declared in the pom.xml file.
func (resolver *MavenPomResolver) Resolve(mavenPomFile string, report *Report) (err error) {
	dir, base := filepath.Split(mavenPomFile)
	if err := os.Chdir(dir); err != nil {
		return err
	}

	tempDirGenerator := NewTempDirGenerator()
	dependenciesDir, err := tempDirGenerator.Create()
	if err != nil {
		return err
	}
	defer tempDirGenerator.Destroy()

	// mvn dependency:copy-dependencies -DoutputDirectory=lib
	_, err = exec.Command("mvn", "dependency:copy-dependencies", "-f", base, fmt.Sprintf("-DoutputDirectory=%s", dependenciesDir)).Output()
	if err != nil {
		return err
	}

	jarFiles, err := ioutil.ReadDir(dependenciesDir)
	if err != nil {
		return err
	}

	logger.Log.Debugln("jars size:", len(jarFiles))

	if err = resolver.ResolveJarFiles(dependenciesDir, jarFiles, report); err != nil {
		return err
	}

	return nil
}

// ResolveJarFiles resolves the licenses of the given packages.
func (resolver *MavenPomResolver) ResolveJarFiles(dir string, jarFiles []fs.FileInfo, report *Report) (err error) {
	for _, jar := range jarFiles {
		dependencyPath := filepath.Join(dir, jar.Name())
		err = resolver.ResolveJarLicense(dependencyPath, jar, report)
		if err != nil {
			logger.Log.Warnf("Failed to resolve the license of <%s>: %v\n", filepath.Base(dependencyPath), err)
			report.Skip(&Result{
				Dependency:    filepath.Base(dependencyPath),
				LicenseSpdxID: Unknown,
			})
		}
	}
	return nil
}

var possibleLicenseFileNameInJar = regexp.MustCompile(`(?i)^(\S*)?LICEN[SC]E(\S*\.txt)?$`)

func (resolver *MavenPomResolver) ResolveJarLicense(dependencyPath string, jar fs.FileInfo, report *Report) (err error) {
	compressedJar, err := zip.OpenReader(dependencyPath)
	if err != nil {
		return err
	}
	defer compressedJar.Close()

	for _, compressedFile := range compressedJar.File {
		if possibleLicenseFileNameInJar.MatchString(compressedFile.Name) {
			file, err := compressedFile.Open()
			if err != nil {
				return err
			}

			buf := bytes.NewBuffer(nil)
			w := bufio.NewWriter(buf)
			_, err = io.Copy(w, file)
			if err != nil {
				return err
			}
			w.Flush()
			content := buf.String()
			file.Close()

			licenseFilePath := filepath.Join(dependencyPath, compressedFile.Name)
			identifier, err := license.Identify(dependencyPath, string(content))
			if err != nil {
				return err
			}

			report.Resolve(&Result{
				Dependency:      filepath.Base(dependencyPath),
				LicenseFilePath: licenseFilePath,
				LicenseContent:  content,
				LicenseSpdxID:   identifier,
			})
			return nil
		}
	}
	return nil
}
