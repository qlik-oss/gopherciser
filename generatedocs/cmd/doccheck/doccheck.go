package main

import (
	"fmt"

	"github.com/qlik-oss/gopherciser/generatedocs/generated"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
)

func main() {
	configFields, err := common.FieldsString()
	if err != nil {
		common.Exit(err, 1)
	}
	configFields = append(configFields, "main")
	check("config field", generated.Config, configFields)

	actions := common.ActionStrings()
	check("action", generated.Actions, actions)

	// checkActionGroups(generated.Groups, generated.Actions)

}

func check(name string, doc map[string]common.DocEntry, docExpected []string) {
	warnNotDocumented := func(field string) {
		fmt.Printf("WARNING: %s \"%s\" documentation is not generated\n", name, field)
	}
	warnNoDescr := func(field string) {
		fmt.Printf("WARNING: %s \"%s\" documentation has no description\n", name, field)
	}
	warnNoExamples := func(field string) {
		fmt.Printf("WARNING: %s \"%s\" documentation has no examples\n", name, field)
	}

	for _, field := range docExpected {
		if docEntry, ok := doc[field]; !ok {
			warnNotDocumented(field)
		} else {
			if docEntry.Description == "" {
				warnNoDescr(field)
			}
			if docEntry.Examples == "" {
				warnNoExamples(field)
			}

		}
	}

}

// func checkActionGroups(groups common.GroupsEntry, actions map[string]common.DocEntry)
