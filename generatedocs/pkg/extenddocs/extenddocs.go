package extenddocs

import (
	"github.com/qlik-oss/gopherciser/generatedocs/generated"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompiler"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompilerflag"
)

func ExtendOSSDocs() error {
	compiler := doccompiler.New()
	compiler.AddDataFromGenerated(generated.Actions, generated.Schedulers, generated.Config, generated.Extra, generated.Params, generated.Groups)
	compiler.AddDataFromDir(doccompilerflag.DataRoot())
	return compiler.CompileToFile(doccompilerflag.OutputFile())
}
