package cmd

import (
	"fmt"
	"os"

	"github.com/qlik-oss/gopherciser/version"
	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any sub-commands
var RootCmd = &cobra.Command{
	Use:   "gopherciser",
	Short: "User simulation tool",
	Long:  `Tool to simulate user actions in a sense application`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if version.Version != "" {
		_, _ = fmt.Fprintf(os.Stderr, "Version: %v\n", version.Version)
	}

	if err := RootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func init() {}
