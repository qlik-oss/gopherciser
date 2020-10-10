package main

import (
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompiler"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompilerflag"
)

func main() {
	compiler := doccompiler.New()
	compiler.AddDataFromDir(doccompilerflag.DataRoot())
	compiler.CompileToFile(doccompilerflag.OutputFile())
}
