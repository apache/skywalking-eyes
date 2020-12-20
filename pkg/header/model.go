package header

type Config struct {
	License     string   `yaml:"license"`
	Paths       []string `yaml:"paths"`
	PathsIgnore []string `yaml:"paths-ignore"`
}

type Result struct {
	Success []string `yaml:"success"`
	Failure []string `yaml:"failure"`
}
