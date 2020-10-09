package flags

import (
	"flag"
	"os"
)

var (
	dataRoot string
	output   string
)

func init() {
	handleFlags()
}

func DataRoot() string {
	return dataRoot
}

func OutputFile() string {
	return output
}

func handleFlags() {
	flagHelp := flag.Bool("help", false, "shows help")
	flag.StringVar(&dataRoot, "data", "generatedocs/data", "paths to data folder")
	flag.StringVar(&output, "output", "generatedocs/generated/documentation.go", "path to generated code file")

	flag.Parse()

	if *flagHelp {
		flag.PrintDefaults()
		os.Exit(0)
	}
}
