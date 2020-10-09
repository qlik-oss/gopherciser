package main

import (
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompiler"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/flags"
)

func main() {
	compiler := doccompiler.New()
	compiler.AddDataFromDir(flags.DataRoot())
	compiler.CompileToFile(flags.OutputFile())
}
