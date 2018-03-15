package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{Use: "Kubetool"}

// Execute Viper Command
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Func to init Cobra Flag and bind flag
func init() {
	cobra.OnInitialize()
}
