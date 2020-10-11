package genmd

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
)

type (
	Data struct {
		*CompiledDocs
		ActionFields map[string]interface{}
		ConfigFields map[string]interface{}
	}

	CompiledDocs struct {
		Actions map[string]common.DocEntry
		Params  map[string][]string
		Config  map[string]common.DocEntry
		Groups  []common.GroupsEntry
		Extra   map[string]common.DocEntry
	}
)

const (
	ExitCodeOk = iota
	ExitCodeFailedReadTemplate
	ExitCodeFailedParseTemplate
	ExitCodeFailedExecuteTemplate
	ExitCodeFailedWriteResult
	ExitCodeFailedHandleFields
	ExitCodeFailedHandleParams
)

var (
	compiledDocs *CompiledDocs
	output       string
)

const (
	defaultIndent = "  "
)

func GenerateMarkdown(docs *CompiledDocs) {
	handleFlags()
	mdBytes := generateFromCompiled(docs)
	if err := ioutil.WriteFile(output, mdBytes, 0644); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}
	fmt.Printf("Generated markdown documentation to output<%s>\n", output)
}

func handleFlags() {
	flagHelp := flag.Bool("help", false, "shows help")
	flag.StringVar(&output, "output", "generatedocs/generated/settingup.md", "path to output file")

	flag.Parse()

	if *flagHelp {
		flag.PrintDefaults()
		os.Exit(ExitCodeOk)
	}
}

func generateFromCompiled(docs *CompiledDocs) []byte {
	compiledDocs = docs
	data := &Data{
		CompiledDocs: docs,
	}

	// Create template for generating settingup.md
	documentationTemplate, err := template.New("documentationTemplate").Funcs(funcMap).Parse(templateString)
	if err != nil {
		common.Exit(err, ExitCodeFailedParseTemplate)
	}

	data.ConfigFields, err = common.Fields()
	if err != nil {
		common.Exit(err, ExitCodeFailedHandleFields)
	}

	data.ActionFields = common.Actions()

	buf := bytes.NewBuffer(nil)
	if err := documentationTemplate.Execute(buf, data); err != nil {
		common.Exit(err, ExitCodeFailedExecuteTemplate)
	}
	return buf.Bytes()
}
