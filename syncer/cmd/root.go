package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "syncer",
	Short: "Syncer is a synchronization tool for mutilple service centers",
	Long: "Syncer is a synchronization tool for mutilple service centers, supported: ServiceComb service-center, " +
		"going to be supported: SpringCloud Eureka, Istio, K8S",
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		log.Println(err)
	}
	return err
}
