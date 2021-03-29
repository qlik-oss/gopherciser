package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/config"
	"github.com/spf13/cobra"
)

var (
	scriptOverwrite bool

	// structure command parameters
	outputFolder string
	includeRaw   bool
)

func init() {
	RootCmd.AddCommand(scriptCmd)
	scriptCmd.AddCommand(templateCmd)

	// validate sub command
	scriptCmd.AddCommand(validateCmd)
	AddAllSharedParameters(validateCmd)

	// connect sub command
	scriptCmd.AddCommand(testConnectionCmd)
	AddAllSharedParameters(testConnectionCmd)

	// template sub command
	AddConfigParameter(templateCmd)
	templateCmd.Flags().BoolVarP(&scriptOverwrite, "force", "f", false, "overwrite existing script file")

	// structure sub command
	scriptCmd.AddCommand(structureCmd)
	AddAllSharedParameters(structureCmd)
	AddLoggingParameters(structureCmd)
	structureCmd.Flags().StringVarP(&outputFolder, "output", "o", "", "script output folder. Defaults to working folder")
	structureCmd.Flags().BoolVarP(&includeRaw, "raw", "r", false, "include raw properties in structure")
}

// scriptCmd represents the script command
var scriptCmd = &cobra.Command{
	Use:     "script",
	Aliases: []string{"s"},
	Short:   "Set of commands to handle scenario scripts.",
	Long:    `Set of commands to handle scenario scripts.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := cmd.Help(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to execute script command: %v\n", err)
		}
	},
}

// templateCmd represents the template command
var templateCmd = &cobra.Command{
	Use:     "template",
	Aliases: []string{"tmpl", "t"},
	Short:   "Create script template file",
	Long:    `Create script template file`,
	Run: func(cmd *cobra.Command, args []string) {
		if cfgFile == "" {
			cfgFile = "template.json"
		}

		if fileExists(cfgFile) && !scriptOverwrite {
			_, _ = fmt.Fprintf(os.Stderr, "file<%s> exists, use -f to overwrite\n", cfgFile)
			return
		}

		cfg, err := config.NewExampleConfig()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "creating example config:\n%v\n", err)
		}

		jsonCfg, err := jsonit.MarshalIndent(cfg, "", "  ")
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error marshling template config:\n%v\n", err)
			return
		}

		scenarioFile, err := os.Create(cfgFile)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to create file<%s>: %v\n", cfgFile, err)
			return
		}
		defer func() {
			if err := scenarioFile.Close(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "failed to close file<%s> successfully: %v\n", cfgFile, err)
			}
		}()
		if _, err = scenarioFile.Write(jsonCfg); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error while writing to file<%s>: %v\n", cfgFile, err)
		}
		fmt.Printf("%s written successfully.\n", cfgFile)
	},
}

var validateCmd = &cobra.Command{
	Use:     "validate",
	Aliases: []string{"v"},
	Short:   "validate a scenario config file",
	Long:    `validate a scenario config file`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := unmarshalConfigFile()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
			return
		}

		if err := validateConfigAndPrintWarnings(cfg); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
			return
		}

		_, _ = os.Stderr.WriteString("Config Valid!\n")
	},
}

var testConnectionCmd = &cobra.Command{
	Use:     "connect",
	Aliases: []string{"c"},
	Short:   "test connection",
	Long:    `test connection using settings provided by the config file`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := unmarshalConfigFile()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
			os.Exit(ExitCodeJSONParseError)
		}

		if err = cfg.TestConnection(context.Background()); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(ExitCodeConnectionError)
		}

		_, _ = os.Stderr.WriteString("Connection Successful!\n")
	},
}

var structureCmd = &cobra.Command{
	Use:     "structure",
	Aliases: []string{"s"},
	Short:   "Get app structure",
	Long: `Get app structure using settings provided by the config file.
Will save one .structure file per app in script in the folder defined by output parameter.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Read object definition overrides and additions
		if err := ReadObjectDefinitions(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "ObjectDefError: %s\n", err)
			os.Exit(ExitCodeObjectDefError)
		}

		cfg, err := unmarshalConfigFile()
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
			os.Exit(ExitCodeJSONParseError)
		}

		// Override output folder
		cfg.Settings.OutputsSettings.Dir = outputFolder

		// Override log settings in case of parameters being set
		if err := ConfigOverrideLogSettings(cfg); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
			os.Exit(ExitCodeLogFormatError)
		}

		err = cfg.GetAppStructures(context.Background(), includeRaw)
		switch err.(type) {
		case appstructure.AppStructureNoScenarioActionsError:
			// Not an error but print info
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		case nil:
		default:
			_, _ = fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
			os.Exit(ExitCodeAppStructure)
		}
	},
}

func validateConfigAndPrintWarnings(cfg *config.Config) error {
	err := cfg.Validate()
	if err != nil {
		return err
	}

	warningsCount := len(cfg.ValidationWarnings)
	if warningsCount < 1 {
		return nil
	}

	_, _ = fmt.Fprintf(os.Stderr, "%d script validation warnings:\n", warningsCount)
	for i, warning := range cfg.ValidationWarnings {
		_, _ = fmt.Fprintf(os.Stderr, "%d. %s\n", i+1, warning)
		if i == 9 {
			_, _ = fmt.Fprintf(os.Stderr, "...(%d) additional warnings\n", warningsCount-i+1)
			return nil
		}
	}
	return nil
}
