package deps

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/license"
)

type JarResolver struct{}

func (resolver *JarResolver) CanResolve(jarFile string) bool {
	return filepath.Ext(jarFile) == ".jar"
}

func (resolver *JarResolver) Resolve(jarFile string, report *Report) error {
	state := NotFound
	if err := resolver.ResolveJar(&state, jarFile, Unknown, report); err != nil {
		dep := filepath.Base(jarFile)
		logger.Log.Warnf("Failed to resolve the license of <%s>: %v\n", dep, state.String())
		report.Skip(&Result{
			Dependency:    dep,
			LicenseSpdxID: Unknown,
		})
	}

	return nil
}

func (resolver *JarResolver) ResolveJar(state *State, jarFile, version string, report *Report) error {
	dep := filepath.Base(jarFile)

	compressedJar, err := zip.OpenReader(jarFile)
	if err != nil {
		return err
	}
	defer compressedJar.Close()

	var manifestFile *zip.File

	// traverse all files in jar
	for _, compressedFile := range compressedJar.File {
		archiveFile := compressedFile.Name
		switch {
		case reHaveManifestFile.MatchString(archiveFile):
			manifestFile = compressedFile

		case possibleLicenseFileName.MatchString(archiveFile):
			*state |= FoundLicenseInJarLicenseFile
			buf, err := resolver.ReadFileFromZip(compressedFile)
			if err != nil {
				return err
			}

			return resolver.IdentifyLicense(jarFile, dep, buf.String(), version, report)
		}
	}

	if manifestFile != nil {
		buf, err := resolver.ReadFileFromZip(manifestFile)
		if err != nil {
			return err
		}
		norm := regexp.MustCompile(`(?im)[\r\n]+ +`)
		content := norm.ReplaceAllString(buf.String(), "")

		r := reSearchLicenseInManifestFile.FindStringSubmatch(content)
		if len(r) != 0 {
			report.Resolve(&Result{
				Dependency:      dep,
				LicenseFilePath: jarFile,
				LicenseContent:  strings.TrimSpace(r[1]),
				LicenseSpdxID:   strings.TrimSpace(r[1]),
				Version:         version,
			})
			return nil
		}
	}

	return fmt.Errorf("cannot find license content")
}

func (resolver *JarResolver) ReadFileFromZip(archiveFile *zip.File) (*bytes.Buffer, error) {
	file, err := archiveFile.Open()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	w := bufio.NewWriter(buf)
	_, err = io.CopyN(w, file, int64(archiveFile.UncompressedSize64))
	if err != nil {
		return nil, err
	}

	w.Flush()
	file.Close()
	return buf, nil
}

func (resolver *JarResolver) IdentifyLicense(path, dep, content, version string, report *Report) error {
	identifier, err := license.Identify(path, content)
	if err != nil {
		return err
	}

	report.Resolve(&Result{
		Dependency:      dep,
		LicenseFilePath: path,
		LicenseContent:  content,
		LicenseSpdxID:   identifier,
		Version:         version,
	})
	return nil
}
