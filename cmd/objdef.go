package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/senseobjdef"
	"github.com/spf13/cobra"
)

var (
	// Command parameters
	objDefFilePath      string
	objDefFileOverwrite bool
	objDefObjectList    []string
	objDefVerbose       bool

	// Commands
	objDefCmd = &cobra.Command{
		Use:     "objdef",
		Aliases: []string{"od"},
		Short:   "Sense object data definitions.",
		Long: `Handles sense object data definitions for gopherciser.
Use to export default values or examples or to validate custom definitions or overrides.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "failed to execute objdef command: %v\n", err)
			}
		},
	}

	generateObjDefCmd = &cobra.Command{
		Use:     "generate",
		Aliases: []string{"gen", "g"},
		Short:   "Generate object definitions from default values.",
		Long:    `Generate object definitions from default values, either a full json with all defaults or defined objects.`,
		Run: func(cmd *cobra.Command, args []string) {
			if objDefFilePath == "" {
				if err := cmd.Help(); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "failed to print objdef generate command help: %v\n", err)
				}
				return
			}

			if fileExists(objDefFilePath) && !objDefFileOverwrite {
				_, _ = fmt.Fprintf(os.Stderr, "file<%s> exists, set force flag to overwrite\n", objDefFilePath)
				return
			}

			objDefList := make(senseobjdef.ObjectDefs)
			if len(objDefObjectList) > 0 {
				missing := make([]string, 0, 1)
				for _, object := range objDefObjectList {
					if senseobjdef.DefaultObjectDefs[object] == nil {
						missing = append(missing, object)
					} else {
						objDefList[object] = senseobjdef.DefaultObjectDefs[object]
					}
				}

				if len(missing) > 0 {
					_, _ = fmt.Fprintf(os.Stderr, "no object definitions found for %v\n", missing)
				}
			} else {
				objDefList = senseobjdef.DefaultObjectDefs
			}

			jsn, err := jsonit.MarshalIndent(objDefList, "", "  ")
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "failed to generate default object definitions: %v\n", err)
				return
			}

			defFile, err := os.Create(objDefFilePath)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "failed to create file<%s>: %v\n", objDefFilePath, err)
				return
			}
			defer func() {
				if err := defFile.Close(); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "failed to close file<%s> successfully: %v\n", objDefFilePath, err)
				}
			}()

			if _, err = defFile.Write(jsn); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error while writing to file<%s>: %v\n", objDefFilePath, err)
			}
			fmt.Printf("%s written successfully.\n", objDefFilePath)
		},
	}

	validateObjDefCmd = &cobra.Command{
		Use:     "validate",
		Aliases: []string{"val", "v"},
		Short:   "Validate object definitions in file.",
		Long: `Validate object definitions from provided JSON file. Will print how many definitions were found,
it's recommended to use to -v verbose flag and verify all parameters were interpreted correctly'`,
		Run: func(cmd *cobra.Command, args []string) {
			if objDefFilePath == "" {
				if err := cmd.Help(); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "failed to print objdef validate command help: %v\n", err)
				}
				return
			}

			if !fileExists(objDefFilePath) {
				_, _ = fmt.Fprintf(os.Stderr, "file<%s> not found.\n", objDefFilePath)
				return
			}

			objDefs := make(senseobjdef.ObjectDefs)

			if err := objDefs.OverrideFromFile(objDefFilePath); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "validation failed: %v\n", err)
				return
			}

			if len(objDefs) < 1 {
				_, _ = fmt.Fprintf(os.Stderr, "no object definitions found in file<%v>\n", objDefFilePath)
				return
			}

			// print as one string through buffer to avoid text fragmentation on printout
			buf := helpers.NewBuffer()

			buf.WriteString(strconv.Itoa(len(objDefs)))
			buf.WriteString(" object definition")
			if len(objDefs) > 1 {
				buf.WriteString("s")
			}
			buf.WriteString(" found\n\n")

			if !objDefVerbose {
				fmt.Println(buf.String())
				return
			}

			for k, v := range objDefs {
				// object type
				buf.WriteString("[")
				buf.WriteString(k)
				buf.WriteString("]\n/ ")

				buf.WriteString(strconv.Itoa(len(v.Data)))
				buf.WriteString(" data constraint entr")
				if len(v.Data) == 1 {
					buf.WriteString("y.\n")
				} else {
					buf.WriteString("ies.\n")
				}

				if len(v.Data) > 0 {
					buf.WriteString("|\n")
					for _, c := range v.Data {
						constraintString(c, buf)
					}
				}

				buf.WriteString("|\n/ DataDef Type: ")
				buf.WriteString(v.DataDef.Type.String())
				buf.WriteString("\n")

				buf.WriteString("|         Path: ")
				buf.WriteString(string(v.DataDef.Path))
				buf.WriteString("\n")

				if v.Select == nil {
					buf.WriteString("| Not selectable\n")
				} else {
					buf.WriteString("|\n/ Select  Type: ")
					buf.WriteString(v.Select.Type.String())
					buf.WriteString("\n")

					buf.WriteString("|         Path: ")
					buf.WriteString(string(v.Select.Path))
					buf.WriteString("\n")
				}
				buf.WriteString("*\n\n")
			}

			fmt.Println(buf.String())
		},
	}
)

func init() {
	// add objdef command
	RootCmd.AddCommand(objDefCmd)

	// add sub commands
	objDefCmd.AddCommand(generateObjDefCmd)
	objDefCmd.AddCommand(validateObjDefCmd)

	// add "generate" parameters
	generateObjDefCmd.Flags().StringVarP(&objDefFilePath, "definitions", "d", "", "(mandatory) definitions file.")
	generateObjDefCmd.Flags().BoolVarP(&objDefFileOverwrite, "force", "f", false, "overwrite definitions file if existing.")
	generateObjDefCmd.Flags().StringSliceVarP(&objDefObjectList, "objects", "o", nil, "(optional) list of objects, defaults to all.")

	// add "validate" parameters
	validateObjDefCmd.Flags().StringVarP(&objDefFilePath, "definitions", "d", "", "(mandatory) definitions file.")
	validateObjDefCmd.Flags().BoolVarP(&objDefVerbose, "verbose", "v", false, "print summary of definitions.")
}

func fileExists(name string) bool {
	if _, err := os.Stat(name); !os.IsNotExist(err) {
		return true
	}
	return false
}

func constraintString(data senseobjdef.Data, buf *helpers.Buffer) string {
	if data.Constraint == nil {
		buf.WriteString("|   Constraint: Default\n")
	} else {
		for _, c := range data.Constraint {
			buf.WriteString("|   Constraint: ")
			buf.WriteString("[")
			buf.WriteString(string(c.Path))
			buf.WriteString("] ")
			buf.WriteString(string(c.Value))
			buf.WriteString(" Required: ")
			buf.WriteString(strconv.FormatBool(c.Required))
			buf.WriteString("\n")
		}
	}

	buf.WriteString("|     ")
	buf.WriteString(strconv.Itoa(len(data.Requests)))
	buf.WriteString(" data request")
	if len(data.Requests) != 1 {
		buf.WriteString("s")
	}

	if len(data.Requests) > 0 {
		buf.WriteString(":\n")

		for i, r := range data.Requests {
			buf.WriteString("|     [")
			buf.WriteString(strconv.Itoa(i))
			buf.WriteString("]: Type:   ")
			buf.WriteString(r.Type.String())
			buf.WriteString("\n")
			if r.Path != "" {
				buf.WriteString("|          Path:   ")
				buf.WriteString(r.Path)
				buf.WriteString("\n")
			}
			if r.Height > 0 {
				buf.WriteString("|          Height: ")
				buf.WriteString(strconv.Itoa(r.Height))
				buf.WriteString("\n")
			}
		}
	} else {
		buf.WriteString("\n")
	}

	buf.WriteString("|\n")

	return buf.String()
}
