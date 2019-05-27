package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "syncer",
	Short: "Syncer is a synchronization tool for multiple service centers",
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		log.Println(err)
	}
	return err
}
