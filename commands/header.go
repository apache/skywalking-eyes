package commands

import (
	"github.com/spf13/cobra"
)

var Header = &cobra.Command{
	Use:     "header",
	Aliases: []string{"h"},
	Short:   "License header related commands; e.g. check, fix, etc.",
	Long:    "`header` command walks the specified paths recursively and checks if the specified files have the license header in the config file.",
}

func init() {
	Header.AddCommand(CheckCommand)
	Header.AddCommand(FixCommand)
}
