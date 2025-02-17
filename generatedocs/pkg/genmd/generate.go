package genmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
)

var unitTestMode = false

const (
	GeneratedFolder     = "generated"
	SessionVariableName = "sessionvariables"
)

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

	ConfigSection struct {
		Data      string
		FilePath  string
		LinkTitle string
		LinkName  string
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
	ExitCodeFailedCreateFolder
	ExitCodeFailedDeleteFolder
	ExitCodeFailedDeleteFile
)

func GenerateMarkdown(docs *CompiledDocs) {
	handleFlags()
	if wiki == "" && output == "" {
		_, _ = os.Stderr.WriteString("must defined at least one of --wiki or --output")
		return
	}
	if output != "" {
		mdBytes := generateFullMarkdownFromCompiled(docs)
		if err := os.WriteFile(output, mdBytes, 0644); err != nil {
			common.Exit(err, ExitCodeFailedWriteResult)
		}
		fmt.Printf("Generated markdown documentation to output<%s>\n", output)
	}
	if wiki != "" {
		if err := os.RemoveAll(filepath.Join(wiki, GeneratedFolder)); err != nil {
			common.Exit(err, ExitCodeFailedDeleteFolder)
		}
		if err := os.Remove(fmt.Sprintf("%s/_Sidebar.md", wiki)); err != nil && !errors.Is(err, os.ErrNotExist) {
			common.Exit(err, ExitCodeFailedDeleteFile)
		}
		generateWikiFromCompiled(docs)
	}
}

func generateFullMarkdownFromCompiled(compiledDocs *CompiledDocs) []byte {
	main := compiledDocs.Config["main"]
	mainNode := NewDocNode(DocEntry(main))
	addConfigFields(mainNode, compiledDocs)

	var buf bytes.Buffer
	mainNode.WriteTo(&buf)
	return buf.Bytes()
}

func generateWikiFromCompiled(compiledDocs *CompiledDocs) {
	// TODO warning (error?)for ungrouped

	if err := createFolder(filepath.Join(wiki, GeneratedFolder), true); err != nil {
		common.Exit(err, ExitCodeFailedCreateFolder)
	}
	if verbose {
		fmt.Println("creating groups sidebar...")
	}

	generateWikiConfigSections(compiledDocs)
}

func generateWikiConfigSections(compiledDocs *CompiledDocs) {
	configfile, err := os.Create(fmt.Sprintf("%s/config.md", filepath.Join(wiki, GeneratedFolder)))
	defer func() {
		if err := configfile.Close(); err != nil {
			_, _ = os.Stderr.WriteString(err.Error())
		}
	}()
	if err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}

	configSidebar, err := os.Create(fmt.Sprintf("%s/_Sidebar.md", filepath.Join(wiki, GeneratedFolder)))
	defer func() {
		if err := configSidebar.Close(); err != nil {
			os.Stderr.Write([]byte(err.Error()))
		}
	}()
	if err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}
	if _, err := configSidebar.WriteString("[Home](home)\n\n- [Config](config)\n\n"); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}

	configFields, err := common.Fields()
	if err != nil {
		common.Exit(err, ExitCodeFailedHandleFields)
	}
	configFields[SessionVariableName] = struct{}{} // TODO adding here for now, but figure out best placement in link structure
	for _, name := range sortedKeys(configFields) {
		var section ConfigSection
		switch name {
		case SessionVariableName:
			// TODO remove extra expander in session variable section
			docEntry, ok := compiledDocs.Extra[SessionVariableName]
			if !ok {
				common.Exit(fmt.Errorf("\"Extra\" section<%s> not found", SessionVariableName), ExitCodeFailedReadTemplate)
			}
			section = ConfigSection{
				Data:      DocEntry(docEntry).String(),
				FilePath:  fmt.Sprintf("%s/%s/%s.md", wiki, GeneratedFolder, SessionVariableName),
				LinkTitle: SessionVariableName,
				LinkName:  SessionVariableName,
			}
		case "scenario":
			// action groups
			if verbose {
				fmt.Println("creating groups.md...")
			}
			if _, err := configSidebar.WriteString("	- [Action groups](groups)\n\n"); err != nil {
				common.Exit(err, ExitCodeFailedWriteResult)
			}
			groups := generateWikiGroups(compiledDocs)
			grouplinks, err := os.Create(fmt.Sprintf("%s/groups.md", filepath.Join(wiki, GeneratedFolder)))
			defer func() {
				if err := grouplinks.Close(); err != nil {
					os.Stderr.Write([]byte(err.Error()))
				}
			}()
			if err != nil {
				common.Exit(err, ExitCodeFailedWriteResult)
			}

			for name, title := range groups {
				groupslink := fmt.Sprintf("[%s](%s)\n\n", title, name)
				if _, err := configSidebar.WriteString(fmt.Sprintf("		- %s", groupslink)); err != nil {
					common.Exit(err, ExitCodeFailedWriteResult)
				}
				if _, err := grouplinks.WriteString(groupslink); err != nil {
					common.Exit(err, ExitCodeFailedWriteResult)
				}
			}

			section = ConfigSection{
				LinkTitle: "Scenario actions",
				LinkName:  "groups",
			}
		case "scheduler":
			// addSchedulers(newNode, compiledDocs)
			continue // TODO needs special handling
		default:
			fieldEntry := &DocEntryWithParams{
				DocEntry: DocEntry(compiledDocs.Config[name]),
				Params:   MarkdownParams(configFields[name], compiledDocs.Params),
			}
			section = ConfigSection{
				Data:      fieldEntry.String(),
				FilePath:  fmt.Sprintf("%s/%s.md", filepath.Join(wiki, GeneratedFolder), name),
				LinkTitle: name,
				LinkName:  name,
			}
		}

		if section.FilePath != "" {
			if verbose {
				fmt.Printf("creating file<%s>...\n", section.FilePath)
			}
			sectionfile, err := os.Create(section.FilePath)
			defer func() {
				if err := sectionfile.Close(); err != nil {
					_, _ = os.Stderr.WriteString(fmt.Sprintf("%v", err))
				}
			}()
			if err != nil {
				common.Exit(err, ExitCodeFailedWriteResult)
			}
			if _, err := sectionfile.WriteString(section.Data); err != nil {
				common.Exit(err, ExitCodeFailedWriteResult)
			}
		}

		linkString := fmt.Sprintf("[%s](%s)\n\n", section.LinkTitle, section.LinkName)
		if _, err := configfile.WriteString(linkString); err != nil {
			common.Exit(err, ExitCodeFailedWriteResult)
		}
		if _, err := configSidebar.WriteString(fmt.Sprintf("	- %s", linkString)); err != nil {
			common.Exit(err, ExitCodeFailedWriteResult)
		}
	}
}

func generateWikiGroups(compiledDocs *CompiledDocs) map[string]string {
	groups := make(map[string]string)
	for _, group := range compiledDocs.Groups {
		if verbose {
			fmt.Printf("Generating wiki actions for GROUP %s...\n", group.Name)
		}
		if err := createFolder(filepath.Join(wiki, GeneratedFolder, group.Name), false); err != nil {
			common.Exit(err, ExitCodeFailedCreateFolder)
		}
		groups[group.Name] = group.Title
		generateWikiGroup(compiledDocs, group)
	}
	return groups
}

func generateWikiGroup(compiledDocs *CompiledDocs, group common.GroupsEntry) {
	file := fmt.Sprintf("%s/%s.md", filepath.Join(wiki, GeneratedFolder, group.Name), group.Name)
	if verbose {
		fmt.Printf("creating file<%s>...\n", file)
	}
	if err := os.WriteFile(file, []byte(DocEntry(group.DocEntry).String()), os.ModePerm); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}
	actionsSidebar, err := os.Create(fmt.Sprintf("%s/_Sidebar.md", filepath.Join(wiki, GeneratedFolder, group.Name)))
	defer func() {
		if err := actionsSidebar.Close(); err != nil {
			os.Stderr.Write([]byte(err.Error()))
		}
	}()
	if err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}

	if _, err := actionsSidebar.WriteString(fmt.Sprintf("[Home](home)\n\n- [Config](config)\n\n	- [Action Groups](groups)\n\n		- [%s](%s)\n\n", group.Title, group.Name)); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}

	for _, action := range group.Actions {
		actionEntry := createActionEntry(compiledDocs, common.Actions(), action)
		if actionEntry == nil {
			continue
		}
		file = fmt.Sprintf("%s/%s.md", filepath.Join(wiki, GeneratedFolder, group.Name), action)
		if verbose {
			fmt.Printf("creating file<%s>...\n", file)
		}
		if err := os.WriteFile(file, []byte(actionEntry.String()), os.ModePerm); err != nil {
			common.Exit(err, ExitCodeFailedWriteResult)
		}
		if _, err := actionsSidebar.WriteString(fmt.Sprintf("			- [%s](%s)\n\n", action, action)); err != nil {
			common.Exit(err, ExitCodeFailedWriteResult)
		}
	}
}

func createFolder(path string, footer bool) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	if footer {
		if err := os.WriteFile(fmt.Sprintf("%s/_Footer.md", path), []byte("The file has been generated, do not edit manually\n"), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

func addActions(node DocNode, compiledDocs *CompiledDocs, actions []string, actionSettings map[string]interface{}) {
	for _, action := range actions {
		actionEntry := createActionEntry(compiledDocs, actionSettings, action)
		if actionEntry == nil {
			continue
		}
		newNode := NewFoldedDocNode(action, actionEntry)
		node.AddChild(newNode)
	}
}

func createActionEntry(compiledDocs *CompiledDocs, actionSettings map[string]interface{}, action string) *DocEntryWithParams {
	compiledEntry, ok := compiledDocs.Actions[action]
	if !ok {
		compiledEntry.Description = "*Missing description*\n"
	}
	actionParams := actionSettings[action]
	if actionParams == nil {
		os.Stderr.WriteString(fmt.Sprintf("%s gives nil actionparams, skipping...\n", action))
		return nil
	}
	return &DocEntryWithParams{
		DocEntry: DocEntry(compiledEntry),
		Params:   MarkdownParams(actionParams, compiledDocs.Params),
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
