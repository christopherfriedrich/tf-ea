package tfea

import "github.com/spf13/cobra"

var inspectCmd = &cobra.Command{
	Use: "inspect",
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
