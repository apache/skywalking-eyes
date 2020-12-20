package header

import (
	"github.com/spf13/cobra"
)

var Header = &cobra.Command{
	Use:     "header",
	Long:    "`header` command walks the specified paths recursively and checks if the specified files have the license header in the config file.",
	Aliases: []string{"h"},
}

func init() {
	Header.AddCommand(CheckCommand)
}
