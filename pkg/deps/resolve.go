package deps

import (
	"fmt"
)

type Resolver interface {
	CanResolve(string) bool
	Resolve(string, *Report) error
}

var Resolvers = []Resolver{
	new(GoModResolver),
	new(NpmResolver),
	new(MavenPomResolver),
}

func Resolve(config *ConfigDeps, report *Report) error {
resolveFile:
	for _, file := range config.Files {
		for _, resolver := range Resolvers {
			if !resolver.CanResolve(file) {
				continue
			}
			if err := resolver.Resolve(file, report); err != nil {
				return err
			}
			continue resolveFile
		}
		return fmt.Errorf("unable to find a resolver to resolve dependency declaration file: %v", file)
	}

	return nil
}
