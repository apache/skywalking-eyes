package header

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/comments"
)

// Fix adds the configured license header to the given file.
func Fix(file string, config *ConfigHeader, result *Result) error {
	var r Result
	if err := CheckFile(file, config, &r); err != nil || !r.HasFailure() {
		logger.Log.Warnln("Try to fix a valid file, do nothing:", file)
		return err
	}

	style := comments.FileCommentStyle(file)

	if style == nil {
		return fmt.Errorf("unsupported file: %v", file)
	}

	return InsertComment(file, style, config, result)
}

func InsertComment(file string, style *comments.CommentStyle, config *ConfigHeader, result *Result) error {
	stat, err := os.Stat(file)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	licenseHeader, err := GenerateLicenseHeader(style, config)
	if err != nil {
		return err
	}

	content = rewriteContent(style, content, licenseHeader)

	if err := os.WriteFile(file, content, stat.Mode()); err != nil {
		return err
	}

	result.Fix(file)

	return nil
}

func rewriteContent(style *comments.CommentStyle, content []byte, licenseHeader string) []byte {
	if style.After == "" {
		return append([]byte(licenseHeader), content...)
	}

	content = []byte(strings.TrimLeft(string(content), " \n"))
	afterPattern := regexp.MustCompile(style.After)
	location := afterPattern.FindIndex(content)
	if location == nil || len(location) != 2 {
		if style.EnsureAfter != "" {
			return append([]byte(style.EnsureAfter+"\n"+licenseHeader+style.EnsureBefore), content...)
		}
		return append([]byte(licenseHeader), content...)
	}

	// if files do not have an empty line at the end, the content slice index given
	//  at index location[1]+1 could be out of range
	startIdx := math.Min(float64(location[1]+1), float64(len(content)))
	return append(content[0:location[1]],
		append(append([]byte("\n"), []byte(licenseHeader)...), content[int64(startIdx):]...)...,
	)
}

func GenerateLicenseHeader(style *comments.CommentStyle, config *ConfigHeader) (string, error) {
	if err := style.Validate(); err != nil {
		return "", err
	}

	content := config.GetLicenseContent()
	// trailing leading and trailing newlines
	content = strings.TrimSpace(content)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = fmt.Sprintf("%v %v", style.Middle, line)
		} else {
			lines[i] = style.Middle
		}
	}

	if style.Start != style.Middle {
		lines = append([]string{style.Start}, lines...)
	}

	if style.End != style.Middle {
		lines = append(lines, style.End)
	}

	return strings.Join(lines, "\n") + "\n\n", nil
}
