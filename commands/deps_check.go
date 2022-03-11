package commands

import (
	"github.com/spf13/cobra"

	"github.com/apache/skywalking-eyes/pkg/deps"
)

var DepsCheckCommand = &cobra.Command{
	Use:     "check",
	Aliases: []string{"c"},
	Long:    "resolves and check license compatibility in all dependencies of a module and their transitive dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		return deps.Check(Config.Header.License.SpdxID, &Config.Deps)
	},
}
