package main

import (
	"fmt"

	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompiler"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompilerflag"
)

func main() {
	compiler := doccompiler.New()
	compiler.AddDataFromDir(doccompilerflag.DataRoot())
	if err := compiler.CompileToFile(doccompilerflag.OutputFile()); err != nil {
		fmt.Printf("Errors:\n%v\n", err)
		common.Exit(fmt.Errorf("incomplete documentation"), doccompiler.ExitCodeUndocumented)
	}
}
