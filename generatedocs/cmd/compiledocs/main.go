package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompiler"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompilerflag"
)

func main() {
	compiler := doccompiler.New()
	compiler.AddDataFromDir(doccompilerflag.DataRoot())
	errs := compiler.CompileToFile(doccompilerflag.OutputFile())
	for _, err := range errs {
		fmt.Println("Error: ", err)
	}
	if len(errs) > 0 {
		common.Exit(errors.Errorf("Incomplete documentation"), doccompiler.ExitCodeUndocumented)
	}
}
