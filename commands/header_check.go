package commands

import (
	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/header"
	"github.com/apache/skywalking-eyes/pkg/review"

	"github.com/spf13/cobra"
)

var CheckCommand = &cobra.Command{
	Use:     "check",
	Aliases: []string{"c"},
	Long:    "check command walks the specified paths recursively and checks if the specified files have the license header in the config file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var result header.Result

		if len(args) > 0 {
			logger.Log.Debugln("Overriding paths with command line args.")
			Config.Header.Paths = args
		}

		if err := header.Check(&Config.Header, &result); err != nil {
			return err
		}

		logger.Log.Infoln(result.String())

		if result.HasFailure() {
			if err := review.Header(&result, &Config); err != nil {
				logger.Log.Warnln("Failed to create review comments", err)
			}
			return result.Error()
		}

		return nil
	},
}
