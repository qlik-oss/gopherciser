package extenddocs

import (
	"github.com/qlik-oss/gopherciser/generatedocs/generated"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompiler"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/flags"
)

func ExtendOSSDocs(dataRoot string, output string) {
	compiler := doccompiler.New()
	compiler.AddDataFromGenerated(generated.Actions, generated.Config, generated.Extra, generated.Params, generated.Groups)
	compiler.AddDataFromDir(flags.DataRoot())
	compiler.CompileToFile(flags.OutputFile())
}
