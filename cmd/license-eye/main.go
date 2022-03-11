package main

import (
	"os"

	"github.com/apache/skywalking-eyes/commands"
	"github.com/apache/skywalking-eyes/internal/logger"
)

func main() {
	if err := commands.Execute(); err != nil {
		logger.Log.Errorln(err)
		os.Exit(1)
	}
}
