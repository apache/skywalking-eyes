package commands

import (
	"github.com/apache/skywalking-eyes/internal/logger"
	"github.com/apache/skywalking-eyes/pkg/config"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	verbosity  string
	configFile string
	Config     config.Config
)

// root represents the base command when called without any subcommands
var root = &cobra.Command{
	Use:           "license-eye command [flags]",
	Long:          "A full-featured license guard to check and fix license headers and dependencies' licenses",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level, err := logrus.ParseLevel(verbosity)
		if err != nil {
			return err
		}
		logger.Log.SetLevel(level)

		return Config.Parse(configFile)
	},
	Version: version,
}

// Execute sets flags to the root command appropriately.
// This is called by main.main(). It only needs to happen once to the root.
func Execute() error {
	root.PersistentFlags().StringVarP(&verbosity, "verbosity", "v", logrus.InfoLevel.String(), "log level (debug, info, warn, error, fatal, panic")
	root.PersistentFlags().StringVarP(&configFile, "config", "c", ".licenserc.yaml", "the config file")

	root.AddCommand(Header)
	root.AddCommand(Deps)

	return root.Execute()
}
