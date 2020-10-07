package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	doccompiler "github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompiler"
)

// ExitCodes
const (
	ExitCodeOk                = 0
	ExitCodeFailedWriteResult = 100
)

var (
	dataRootParam string
	dataRoots     []string
	output        string
	// templateFile  string
)

func main() {
	handleFlags()
	generatedDocs := doccompiler.Compile(dataRoots...)
	if err := ioutil.WriteFile(output, generatedDocs, 0644); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(ExitCodeFailedWriteResult)
	}
	fmt.Printf("Compiled data<%s> to output<%s>\n", dataRootParam, output)
}

func handleFlags() {
	flagHelp := flag.Bool("help", false, "shows help")
	flag.StringVar(&dataRootParam, "data", "generatedocs/data", "a comma separated list of paths to data folders")
	// flag.StringVar(&templateFile, "template", "generatedocs/compile/templates/documentation.template", "path to template file")
	flag.StringVar(&output, "output", "generatedocs/generated/documentation.go", "path to generated code file")

	flag.Parse()

	if *flagHelp {
		flag.PrintDefaults()
		os.Exit(ExitCodeOk)
	}

	dataRoots = strings.Split(dataRootParam, ",")
}
