package main

import (
	"os"

	"github.com/apache/servicecomb-service-center/pkg/log"
	"github.com/apache/servicecomb-service-center/syncer/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error("Failed to execute syncer command", err)
		os.Exit(-1)
	}
}
