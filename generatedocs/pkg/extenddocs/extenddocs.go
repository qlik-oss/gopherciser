package extenddocs

import (
	"github.com/qlik-oss/gopherciser/generatedocs/generated"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompiler"
)

func ExtendOSSDocs() {
	compiler := doccompiler.New()
	compiler.AddDataFromGenerated(generated.Actions, generated.Config, generated.Extra, generated.Params, generated.Groups)
	compiler.AddDataFromDir(doccompilerflag.DataRoot())
	compiler.CompileToFile(doccompilerflag.OutputFile())
}
