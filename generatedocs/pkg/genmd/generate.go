package genmd

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"sort"

	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
)

type (
	DocNode struct {
		Name     string
		Doc      fmt.Stringer
		Children []*DocNode
	}

	DocEntry common.DocEntry

	EmptyDocEntry struct{}

	DocEntryWithParams struct {
		DocEntry
		Params string
	}

	CompiledDocs struct {
		Actions map[string]common.DocEntry
		Params  map[string][]string
		Config  map[string]common.DocEntry
		Groups  []common.GroupsEntry
		Extra   map[string]common.DocEntry
	}
)

func NewDocNode(name string, doc fmt.Stringer) *DocNode {
	return &DocNode{
		Name:     name,
		Doc:      doc,
		Children: []*DocNode{},
	}
}

func (node *DocNode) AddChild(child *DocNode) {
	node.Children = append(node.Children, child)
}
func (doc DocEntry) String() string {
	return fmt.Sprintf("%s\n%s", doc.Description, doc.Examples)
}

func (doc EmptyDocEntry) String() string {
	return ""
}

func (doc DocEntryWithParams) String() string {
	return fmt.Sprintf("%s\n%s\n%s", doc.Description, doc.Params, doc.Examples)
}

func (node *DocNode) WriteTo(writer io.StringWriter) {
	_, _ = writer.WriteString(node.Doc.String())
	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Name < node.Children[j].Name
	})
	for _, childNode := range node.Children {
		_, _ = writer.WriteString(fmt.Sprintf("\n<details>\n<summary>%s</summary>\n\n", childNode.Name))
		childNode.WriteTo(writer)
		_, _ = writer.WriteString("\n</details>\n")
	}
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

var (
	compiledDocsGlobal *CompiledDocs
)

func GenerateMarkdown(docs *CompiledDocs) {
	handleFlags()
	mdBytes := generateFromCompiled(docs)
	if err := ioutil.WriteFile(output, mdBytes, 0644); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}
	fmt.Printf("Generated markdown documentation to output<%s>\n", output)
}

func generateFromCompiled(compiledDocs *CompiledDocs) []byte {
	compiledDocsGlobal = compiledDocs

	main := compiledDocs.Config["main"]
	mainNode := NewDocNode("main", DocEntry(main))
	mainNode.addConfigFields(compiledDocs)

	var buf bytes.Buffer
	mainNode.WriteTo(&buf)
	return buf.Bytes()
}

func (node *DocNode) addActions(compiledDocs *CompiledDocs, actions []string, actionSettigns map[string]interface{}) {
	for _, action := range actions {
		compiledEntry, ok := compiledDocs.Actions[action]
		if !ok {
			compiledEntry.Description = "*Missing description*\n"
		}
		actionParams := actionSettigns[action]
		actionEntry := &DocEntryWithParams{
			DocEntry: DocEntry(compiledEntry),
			Params:   MarkdownParams(actionParams),
		}
		newNode := NewDocNode(action, actionEntry)
		node.AddChild(newNode)
	}
}

func (node *DocNode) addGroups(compiledDocs *CompiledDocs) {
	actionSettigns := common.Actions()
	for _, group := range compiledDocs.Groups {
		groupNode := NewDocNode(group.Title, DocEntry(group.DocEntry))
		node.AddChild(groupNode)
		groupNode.addActions(compiledDocs, group.Actions, actionSettigns)
	}
	ungroupedActions := UngroupedActions(compiledDocs.Groups)
	if len(ungroupedActions) > 0 {
		ungroupedGroup := NewDocNode("Ungrouped actions", EmptyDocEntry{})
		node.AddChild(ungroupedGroup)
		ungroupedGroup.addActions(compiledDocs, ungroupedActions, actionSettigns)
	}

}

func (node *DocNode) addExtra(compiledDocs *CompiledDocs, name string, title string) {
	docEntry, ok := compiledDocs.Extra[name]
	if ok {
		node.AddChild(NewDocNode(title, DocEntry(docEntry)))
	}
}

func (node *DocNode) addConfigFields(compiledDocs *CompiledDocs) {
	configFields, err := common.Fields()
	if err != nil {
		common.Exit(err, ExitCodeFailedHandleFields)
	}
	for name, configStruct := range configFields {
		fieldEntry := &DocEntryWithParams{}
		fieldEntry.DocEntry = DocEntry(compiledDocs.Config[name])
		fieldEntry.Params = MarkdownParams(configStruct)
		newNode := NewDocNode(name, fieldEntry)
		node.AddChild(newNode)
		if name == "scenario" {
			newNode.addGroups(compiledDocs)
			newNode.addExtra(compiledDocs, "sessionvariables", "Session Variables")
		}
	}
}
