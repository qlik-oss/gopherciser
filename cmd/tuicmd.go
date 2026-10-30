package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/qlik-oss/gopherciser/tui"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(tuiCmd)
	tuiCmd.Flags().StringVarP(&tuiDir, "dir", "d", "", "directory to browse for scripts (default: current directory).")
	tuiCmd.Flags().StringVarP(&tuiScript, "script", "c", "", "script to open directly instead of showing the script list.")
}

var tuiDir, tuiScript string

var tuiCmd = &cobra.Command{
	Use:     "tui [dir]",
	Aliases: []string{"ui"},
	Args:    cobra.MaximumNArgs(1),
	Short:   "browse, edit and execute scenario scripts in a terminal UI",
	Long: `Start a terminal UI listing all gopherciser scripts
found in a directory given with --dir or as argument (default: current
directory). Scenarios can be edited action by action and executed with live log
output.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := tuiDir
		if tuiScript != "" && dir == "" && len(args) == 0 {
			dir = filepath.Dir(tuiScript)
		}
		if dir == "" && len(args) > 0 {
			dir = args[0]
		}
		if dir == "" {
			dir = "."
		}

		absDir, err := filepath.Abs(dir)
		if err != nil {
			return fmt.Errorf("failed to resolve dir<%s>: %w", dir, err)
		}
		dir = absDir
		info, err := os.Stat(dir)
		if err != nil {
			return fmt.Errorf("failed to read dir<%s>: %w", dir, err)
		}
		if !info.IsDir() {
			// a script was pointed out, browse its directory
			dir = filepath.Dir(dir)
		}

		app := tui.NewApp(dir)
		if tuiScript != "" {
			if err := app.OpenScript(tuiScript); err != nil {
				return err
			}
		}
		return app.Run()
	},
}
