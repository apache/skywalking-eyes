package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/header"
)

var FixCommand = &cobra.Command{
	Use:     "fix",
	Aliases: []string{"f"},
	Long:    "fix command walks the specified paths recursively and fix the license header if the specified files don't have the license header.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var result header.Result
		var errors []string
		var files []string

		if len(args) > 0 {
			files = args
		} else if err := header.Check(&Config.Header, &result); err != nil {
			return err
		} else {
			files = result.Failure
		}

		for _, file := range files {
			if err := header.Fix(file, &Config.Header, &result); err != nil {
				errors = append(errors, err.Error())
			}
		}

		logger.Log.Infoln(result.String())

		if len(errors) > 0 {
			return fmt.Errorf(strings.Join(errors, "\n"))
		}

		return nil
	},
}
