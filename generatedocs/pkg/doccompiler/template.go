package doccompiler

import (
	"sort"
	"strings"
	"text/template"
)

var FuncMap = template.FuncMap{
	"params": SortedParamsKeys,
	"join":   strings.Join,
}

// SortedParamsKeys returns map keys as a sorted slice
func SortedParamsKeys(paramsMap map[string][]string) []string {
	params := make([]string, 0, len(paramsMap))
	for param := range paramsMap {
		params = append(params, param)
	}
	sort.Strings(params)
	return params
}

// TemplateStr used to generate in memory documentation golang package
const TemplateStr = `package generated

/*
	This file has been generated, do not edit the file directly.

	Generate with go run ./generatedocs/compile/main.go or by running go generate in gopherciser root project.
*/

import "github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
{{with $data := .}}
var (
    {{/* Loop over action slice instead of map in order to keep order consistent */}}
    Actions = map[string]common.DocEntry{ {{range $action := $data.Actions}}{{with $actionEntry := index $data.ActionMap $action}}
        "{{$action}}": {
            Description: "{{$actionEntry.Description}}",
            Examples: "{{$actionEntry.Examples}}",
        },{{end}}{{end}}
    }

    Schedulers = map[string]common.DocEntry{ {{range $scheduler := $data.Schedulers}}{{with $schedulerEntry := index $data.SchedulerMap $scheduler}}
        "{{$scheduler}}": {
            Description: "{{$schedulerEntry.Description}}",
            Examples: "{{$schedulerEntry.Examples}}",
        },{{end}}{{end}}
    }

    Params = map[string][]string{ {{range $param := params $data.ParamMap}}
        "{{.}}": { "{{join (index $data.ParamMap $param) "\",\""}}"  },  {{end}}
    }
    {{/* Loop over config fields slice instead of map in order to keep order consistent */}}
    Config = map[string]common.DocEntry{ {{range $field := $data.ConfigFields}}{{with $configEntry := index $data.ConfigMap $field}}
        "{{$field}}" : {
            Description: "{{$configEntry.Description}}",
            Examples: "{{$configEntry.Examples}}",
        },{{end}}{{end}}
    }

    Groups = []common.GroupsEntry{ {{range $group := $data.Groups}}
            {
                Name: "{{$group.Name}}",
                Title: "{{$group.Title}}",
                Actions: []string{ "{{join $group.Actions "\",\""}}" },
                DocEntry: common.DocEntry{
                    Description: "{{$group.Description}}",
                    Examples: "{{$group.Examples}}",
                },
            },{{end}}
    }

    Extra = map[string]common.DocEntry{ {{range $extra := $data.Extra}}{{with $extraEntry := index $data.ExtraMap $extra}}
        "{{$extra}}": {
            Description: "{{$extraEntry.Description}}",
            Examples: "{{$extraEntry.Examples}}",
        },{{end}}{{end}}
    }
){{end}}


`
