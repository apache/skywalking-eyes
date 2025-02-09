// gitignore package provides functions to load global and system gitignore file patterns,
// it's a portion of https://github.com/go-git/go-git/blob/main/plumbing/format/gitignore/dir.go
// to include global ignore files ~/.gitignore and ~/.config/git/ignore, remove this until
// https://github.com/go-git/go-git/issues/1210 is fixed.
package gitignore

import (
	"bufio"
	"os"
	"os/user"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	gconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

const (
	commentPrefix = "#"
	coreSection   = "core"
	excludesfile  = "excludesfile"
)

func readIgnoreFile(fs billy.Filesystem, path []string, ignoreFile string) (ps []gitignore.Pattern, err error) {
	ignoreFile, _ = replaceTildeWithHome(ignoreFile)

	f, err := fs.Open(fs.Join(append(path, ignoreFile)...))
	if err == nil {
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			s := scanner.Text()
			if !strings.HasPrefix(s, commentPrefix) && strings.TrimSpace(s) != "" {
				ps = append(ps, gitignore.ParsePattern(s, path))
			}
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	return
}

func loadPatterns(cfg *gconfig.Config) (ps []gitignore.Pattern, err error) {
	s := cfg.Raw.Section(coreSection)
	efo := s.Options.Get(excludesfile)
	if efo == "" {
		return nil, nil
	}

	ps, err = readIgnoreFile(osfs.New(""), nil, efo)
	if os.IsNotExist(err) {
		return nil, nil
	}

	return
}

// LoadGlobalPatterns loads the global gitignore patterns.
func LoadGlobalPatterns() ([]gitignore.Pattern, error) {
	cfg, err := gconfig.LoadConfig(gconfig.GlobalScope)
	if err != nil {
		return nil, err
	}

	return loadPatterns(cfg)
}

// LoadSystemPatterns loads the system gitignore patterns.
func LoadSystemPatterns() ([]gitignore.Pattern, error) {
	cfg, err := gconfig.LoadConfig(gconfig.SystemScope)
	if err != nil {
		return nil, err
	}

	return loadPatterns(cfg)
}

// LoadGlobalIgnoreFile loads the global gitignore files, ~/.gitignore and ~/.config/git/ignore.
func LoadGlobalIgnoreFile() ([]gitignore.Pattern, error) {
	ps, _ := readIgnoreFile(osfs.New(""), nil, "~/.gitignore")
	ps2, _ := readIgnoreFile(osfs.New(""), nil, "~/.config/git/ignore")
	return append(ps, ps2...), nil
}

func replaceTildeWithHome(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		firstSlash := strings.Index(path, "/")
		if firstSlash == 1 {
			home, err := os.UserHomeDir()
			if err != nil {
				return path, err
			}
			return strings.Replace(path, "~", home, 1), nil
		} else if firstSlash > 1 {
			username := path[1:firstSlash]
			userAccount, err := user.Lookup(username)
			if err != nil {
				return path, err
			}
			return strings.Replace(path, path[:firstSlash], userAccount.HomeDir, 1), nil
		}
	}

	return path, nil
}
