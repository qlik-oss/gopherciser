package genmd

import (
	"strings"
	"text/template"
)

var funcMap = template.FuncMap{
	"join":      strings.Join,
	"params":    handleParams,
	"ungrouped": UngroupedActions,
}

const templateString = `{{with $data := .}}{{(index $data.Config "main").Description}}
{{range $field, $obj := $data.ConfigFields}}{{if ne $field "scenario"}}{{with $configEntry := index $data.Config $field}}<details>
<summary>{{$field}}</summary>

{{$configEntry.Description}}
{{params $obj}}
{{$configEntry.Examples}}
</details>{{end}}{{end}}{{end}}<details>
<summary>scenario</summary>

{{with $scenarioEntry := index $data.Config "scenario"}}{{$scenarioEntry.Description}}
{{params (index $data.ConfigFields "scenario")}}{{$scenarioEntry.Examples}}
{{range $group := $data.Groups}}<details>
<summary>{{$group.Title}}</summary>

{{$group.DocEntry.Description}}
{{$group.DocEntry.Examples}}
{{range $action := $group.Actions}}<details>
<summary>{{$action}}</summary>
{{with $actionEntry := index $data.Actions $action}}
{{$actionEntry.Description}}
{{with $params := params (index $data.ActionFields $action)}}### Settings

{{$params}}{{end}}
{{$actionEntry.Examples}}
</details>{{end}}{{end}}
</details>{{end}}{{with $groups := ungrouped}}
<details>
<summary>Ungrouped actions</summary>
{{range $action := $groups}}{{with $actionEntry := index $data.Actions $action}}
{{$actionEntry.Description}}
{{params (index $data.ActionFields $action)}}
{{$actionEntry.Examples}}
</details>{{end}}{{end}}{{end}}
{{(index $data.Extra "sessionvariables").Description}}
</details>{{end}}
{{(index $data.Config "main").Examples}}{{end}}`
