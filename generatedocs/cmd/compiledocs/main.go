package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	doccompiler "github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompiler"
)

// ExitCodes
const (
	ExitCodeOk                = 0
	ExitCodeFailedWriteResult = 100
)

var (
	dataRoot string
	output   string
)

func main() {
	handleFlags()
	data := doccompiler.NewData()
	data.PopulateFromDataDir(dataRoot)
	generatedDocs := data.Compile()
	if err := ioutil.WriteFile(output, generatedDocs, 0644); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(ExitCodeFailedWriteResult)
	}
	fmt.Printf("Compiled data<%s> to output<%s>\n", dataRoot, output)
}

func handleFlags() {
	flagHelp := flag.Bool("help", false, "shows help")
	flag.StringVar(&dataRoot, "data", "generatedocs/data", "paths to data folder")
	flag.StringVar(&output, "output", "generatedocs/generated/documentation.go", "path to generated code file")

	flag.Parse()

	if *flagHelp {
		flag.PrintDefaults()
		os.Exit(ExitCodeOk)
	}
}
