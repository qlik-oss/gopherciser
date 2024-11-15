package genmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
)

var unitTestMode = false

type (
	DocNodeStruct struct {
		doc      fmt.Stringer
		children []DocNode
	}

	FoldedDocNode struct {
		Name string
		DocNode
	}

	DocEntry common.DocEntry

	DocNode interface {
		WriteTo(io.Writer)
		AddChild(DocNode)
		Children() []DocNode
	}

	EmptyDocEntry struct{}

	DocEntryWithParams struct {
		DocEntry
		Params string
	}

	CompiledDocs struct {
		Actions    map[string]common.DocEntry
		Schedulers map[string]common.DocEntry
		Params     map[string][]string
		Config     map[string]common.DocEntry
		Groups     []common.GroupsEntry
		Extra      map[string]common.DocEntry
	}
)

func NewDocNode(doc fmt.Stringer) DocNode {
	return &DocNodeStruct{
		doc:      doc,
		children: []DocNode{},
	}
}

func NewFoldedDocNode(foldStr string, doc fmt.Stringer) DocNode {
	return &FoldedDocNode{foldStr, NewDocNode(doc)}
}

func (node *DocNodeStruct) AddChild(child DocNode) {
	node.children = append(node.children, child)
}

func (node *DocNodeStruct) Children() []DocNode {
	return node.children
}

func (doc DocEntry) String() string {
	return fmt.Sprintf("%s\n%s\n", doc.Description, doc.Examples)
}

func (doc EmptyDocEntry) String() string {
	return ""
}

func (doc DocEntryWithParams) String() string {
	return fmt.Sprintf("%s\n%s\n%s\n", doc.Description, doc.Params, doc.Examples)
}

func (node *DocNodeStruct) WriteTo(writer io.Writer) {
	fmt.Fprintf(writer, "%s", node.doc)
	for _, childNode := range node.children {
		childNode.WriteTo(writer)
	}
}

func (node *FoldedDocNode) WriteTo(writer io.Writer) {
	fmt.Fprint(writer, "<details>\n")
	fmt.Fprintf(writer, "<summary>%s</summary>\n\n", node.Name)
	node.DocNode.WriteTo(writer)
	fmt.Fprint(writer, "<hr>")
	fmt.Fprint(writer, "</details>\n\n")
}

const (
	ExitCodeOk = iota
	ExitCodeFailedReadTemplate
	ExitCodeFailedParseTemplate
	ExitCodeFailedExecuteTemplate
	ExitCodeFailedWriteResult
	ExitCodeFailedHandleFields
	ExitCodeFailedHandleParams
)

func GenerateMarkdown(docs *CompiledDocs) {
	handleFlags()
	mdBytes := generateFromCompiled(docs)
	if err := os.WriteFile(output, mdBytes, 0644); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}
	fmt.Printf("Generated markdown documentation to output<%s>\n", output)
}

func generateFromCompiled(compiledDocs *CompiledDocs) []byte {
	main := compiledDocs.Config["main"]
	mainNode := NewDocNode(DocEntry(main))
	addConfigFields(mainNode, compiledDocs)

	var buf bytes.Buffer
	mainNode.WriteTo(&buf)
	return buf.Bytes()
}

func addActions(node DocNode, compiledDocs *CompiledDocs, actions []string, actionSettings map[string]interface{}) {
	for _, action := range actions {
		compiledEntry, ok := compiledDocs.Actions[action]
		if !ok {
			compiledEntry.Description = "*Missing description*\n"
		}
		actionParams := actionSettings[action]
		if actionParams == nil {
			os.Stderr.WriteString(fmt.Sprintf("%s gives nil actionparams, skipping...\n", action))
			continue
		}
		actionEntry := &DocEntryWithParams{
			DocEntry: DocEntry(compiledEntry),
			Params:   MarkdownParams(actionParams, compiledDocs.Params),
		}
		newNode := NewFoldedDocNode(action, actionEntry)
		node.AddChild(newNode)
	}
}

func addGroups(node DocNode, compiledDocs *CompiledDocs) {
	actionSettings := common.Actions()
	for _, group := range compiledDocs.Groups {
		groupNode := NewFoldedDocNode(group.Title, DocEntry(group.DocEntry))
		node.AddChild(groupNode)
		addActions(groupNode, compiledDocs, group.Actions, actionSettings)
	}
	ungroupedActions := UngroupedActions(compiledDocs.Groups)
	if unitTestMode {
		ungroupedActionsSkipNew := []string{}
		for _, action := range ungroupedActions {
			if _, ok := compiledDocs.Actions[action]; ok {
				ungroupedActionsSkipNew = append(ungroupedActionsSkipNew, action)
			}
		}
		ungroupedActions = ungroupedActionsSkipNew
	}
	if len(ungroupedActions) > 0 {
		ungroupedGroup := NewFoldedDocNode("Ungrouped actions", EmptyDocEntry{})
		node.AddChild(ungroupedGroup)
		addActions(ungroupedGroup, compiledDocs, ungroupedActions, actionSettings)
	}

}

func addSchedulers(node DocNode, compiledDocs *CompiledDocs) {
	schedulerSettings := common.Schedulers()
	schedulers := make([]string, 0, len(schedulerSettings))
	for sched := range schedulerSettings {
		schedulers = append(schedulers, sched)
	}
	sort.Strings(schedulers)
	for _, sched := range schedulers {
		compiledEntry, ok := compiledDocs.Schedulers[sched]
		if !ok {
			compiledEntry.Description = "*Missing description*\n"
		}
		schedParams := schedulerSettings[sched]
		if schedParams == nil {
			os.Stderr.WriteString(fmt.Sprintf("%s gives nil schedparams, skipping...\n", sched))
			continue
		}
		schedEntry := &DocEntryWithParams{
			DocEntry: DocEntry(compiledEntry),
			Params:   MarkdownParams(schedParams, compiledDocs.Params),
		}
		newNode := NewFoldedDocNode(sched, schedEntry)
		node.AddChild(newNode)
	}
}

func addExtra(node DocNode, compiledDocs *CompiledDocs, name string) {
	docEntry, ok := compiledDocs.Extra[name]
	if ok {
		node.AddChild(NewDocNode(DocEntry(docEntry)))
	}
}

func addConfigFields(node DocNode, compiledDocs *CompiledDocs) {
	configFields, err := common.Fields()
	if err != nil {
		common.Exit(err, ExitCodeFailedHandleFields)
	}
	for _, name := range sortedKeys(configFields) {
		configStruct := configFields[name]
		fieldEntry := &DocEntryWithParams{}
		fieldEntry.DocEntry = DocEntry(compiledDocs.Config[name])
		fieldEntry.Params = MarkdownParams(configStruct, compiledDocs.Params)
		newNode := NewFoldedDocNode(name, fieldEntry)
		node.AddChild(newNode)
		if name == "scheduler" {
			addSchedulers(newNode, compiledDocs)
		}
		if name == "scenario" {
			addGroups(newNode, compiledDocs)
			addExtra(newNode, compiledDocs, "sessionvariables")
		}
	}
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
