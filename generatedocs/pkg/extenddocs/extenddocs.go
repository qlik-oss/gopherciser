package extenddocs

import (
	"github.com/qlik-oss/gopherciser/generatedocs/generated"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/doccompiler"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/flags"
)

func ExtendOSSDocs(dataRoot string, output string) {
	data := doccompiler.NewData()
	data.PopulateFromGenerated(generated.Actions, generated.Config, generated.Extra, generated.Params, generated.Groups)
	data.PopulateFromDataDir(flags.DataRoot())
	data.CompileToFile(flags.OutputFile())
}
