package assets

import (
	"embed"
	"io/fs"
	"path/filepath"
)

//go:embed *
var assets embed.FS

func Asset(file string) ([]byte, error) {
	return assets.ReadFile(filepath.ToSlash(file))
}

func AssetDir(dir string) ([]fs.DirEntry, error) {
	return assets.ReadDir(filepath.ToSlash(dir))
}
