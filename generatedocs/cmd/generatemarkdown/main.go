package main

import (
	"github.com/qlik-oss/gopherciser/generatedocs/generated"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/genmd"
)

func main() {
	genmd.GenerateMarkdown(&genmd.CompiledDocs{
		Actions:    generated.Actions,
		Schedulers: generated.Schedulers,
		Params:     generated.Params,
		Config:     generated.Config,
		Groups:     generated.Groups,
		Extra:      generated.Extra,
	})
}
