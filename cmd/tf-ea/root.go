package tfea

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tf-ea",
	Short: "tf-ea - a simple CLI to transform terraform files to archimate models",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error while exexcuting the CLI '%s'", err)
		os.Exit(1)
	}
}
