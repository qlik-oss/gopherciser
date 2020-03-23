package cmd

import (
	"fmt"

	"github.com/qlik-oss/gopherciser/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"ver"},
	Short:   "Print the version information of Gopherciser",
	Long:    `Print the version information of Gopherciser`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf(`Gopherciser Branch : %s
Gopherciser Revision : %s
Gopherciser Version : %s
Gopherciser BuildTime : %s
`,
			version.Branch, version.Revision, version.Version, version.BuildTime)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
