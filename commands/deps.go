package commands

import (
	"github.com/spf13/cobra"
)

var Deps = &cobra.Command{
	Use:     "dependency",
	Aliases: []string{"d", "deps", "dep", "dependencies"},
	Short:   "Dependencies related commands; e.g. check, etc.",
	Long:    "deps command checks all dependencies of a module and their transitive dependencies.",
}

func init() {
	Deps.AddCommand(DepsResolveCommand)
	Deps.AddCommand(DepsCheckCommand)
}
