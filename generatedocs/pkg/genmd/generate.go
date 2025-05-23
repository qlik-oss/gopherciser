package genmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
)

var unitTestMode = false

const (
	GeneratedFolder     = "generated"
	SessionVariableName = "sessionvariables"
	ConfigLinkName      = "Load test scenario"
	ConfigLinkFile      = "Load-test-scenario"
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
		WriteTo(io.Writer) //nolint:govet
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
		Groups     common.GroupsEntries
		Extra      map[string]common.DocEntry
	}

	ConfigSection struct {
		Data      string
		FilePath  string
		LinkTitle string
		LinkName  string
	}

	SortedDocEntryWithParams struct {
		*DocEntryWithParams
		Name string
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

func (node *DocNodeStruct) WriteTo(writer io.Writer) { //nolint:govet
	_, _ = fmt.Fprintf(writer, "%s", node.doc)
	for _, childNode := range node.children {
		childNode.WriteTo(writer)
	}
}

func (node *FoldedDocNode) WriteTo(writer io.Writer) { //nolint:govet
	_, _ = fmt.Fprint(writer, "<details>\n")
	_, _ = fmt.Fprintf(writer, "<summary>%s</summary>\n\n", node.Name)
	node.DocNode.WriteTo(writer)
	_, _ = fmt.Fprint(writer, "<hr>")
	_, _ = fmt.Fprint(writer, "</details>\n\n")
}

const (
	ExitCodeOk = iota
	ExitCodeFailedReadTemplate
	ExitCodeFailedParseTemplate
	ExitCodeFailedExecuteTemplate
	ExitCodeFailedWriteResult
	ExitCodeFailedHandleFields
	ExitCodeFailedHandleParams
	ExitCodeFailedHandleGroups
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
	if verbose {
		fmt.Printf("creating %s...\n", GeneratedFolder)
	}
	if err := createFolder(filepath.Join(wiki, GeneratedFolder), true); err != nil {
		common.Exit(err, ExitCodeFailedCreateFolder)
	}

	ungroupedActions := UngroupedActions(compiledDocs.Groups)
	if len(ungroupedActions) > 0 {
		common.Exit(fmt.Errorf("found ungrouped actions, add this to a group: %s", strings.Join(ungroupedActions, ",")), ExitCodeFailedHandleGroups)
	}

	generateWikiConfigSections(compiledDocs)
}

func generateWikiConfigSections(compiledDocs *CompiledDocs) {
	if verbose {
		fmt.Printf("creating %s.md...\n", ConfigLinkFile)
	}
	configfile, err := os.Create(fmt.Sprintf("%s/%s.md", filepath.Join(wiki, GeneratedFolder), ConfigLinkFile))
	defer func() {
		if err := configfile.Close(); err != nil {
			_, _ = os.Stderr.WriteString(err.Error())
		}
	}()
	if err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}

	configEntry := DocEntry(compiledDocs.Config["main"])
	if _, err := configfile.WriteString(configEntry.Description); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}

	if verbose {
		fmt.Println("creating config sidebar...")
	}
	configSidebar, err := os.Create(fmt.Sprintf("%s/_Sidebar.md", filepath.Join(wiki, GeneratedFolder)))
	defer func() {
		if err := configSidebar.Close(); err != nil {
			_, _ = os.Stderr.Write([]byte(err.Error()))
		}
	}()
	if err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}
	if _, err := fmt.Fprintf(configSidebar, "[Home](Home)\n\n- [%s](%s)\n\n", ConfigLinkName, ConfigLinkFile); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}

	docEntry, ok := compiledDocs.Extra[SessionVariableName]
	if !ok {
		common.Exit(fmt.Errorf("\"Extra\" section<%s> not found", SessionVariableName), ExitCodeFailedReadTemplate)
	}
	filename := fmt.Sprintf("%s/%s/%s.md", wiki, GeneratedFolder, SessionVariableName)
	if verbose {
		fmt.Printf("creating file<%s>...\n", filename)
	}
	// hack to fix login_settings link for wiki, remove if settingup.md generation is no longer used
	docEntry.Description = strings.Replace(docEntry.Description, "#login_settings", "loginSettings", 1)
	if err := os.WriteFile(filename, []byte(DocEntry(docEntry).String()), os.ModePerm); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}

	linkString := fmt.Sprintf("[%s](%s)\n\n", SessionVariableName, SessionVariableName)
	if _, err := fmt.Fprintf(configfile, "\nSome settings support the use of %s", linkString); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}
	if _, err := fmt.Fprintf(configSidebar, "	- %s", linkString); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}

	if _, err := configfile.WriteString(configEntry.Examples); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}

	if _, err := configfile.WriteString("\n\n## Sections\n\n"); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}

	configFields, err := common.Fields()
	if err != nil {
		common.Exit(err, ExitCodeFailedHandleFields)
	}
	for _, name := range sortedKeys(configFields) {
		var section ConfigSection
		switch name {
		case "scenario":
			if _, err := configSidebar.WriteString("	- [scenario](groups)\n\n"); err != nil {
				common.Exit(err, ExitCodeFailedWriteResult)
			}
			if _, err := configfile.WriteString("scenario\n\n"); err != nil {
				common.Exit(err, ExitCodeFailedWriteResult)
			}
			if verbose {
				fmt.Println("creating groups.md...")
			}
			groups := generateWikiGroups(compiledDocs)
			grouplinks, err := os.Create(fmt.Sprintf("%s/groups.md", filepath.Join(wiki, GeneratedFolder)))
			defer func() {
				if err := grouplinks.Close(); err != nil {
					_, _ = os.Stderr.Write([]byte(err.Error()))
				}
			}()
			if err != nil {
				common.Exit(err, ExitCodeFailedWriteResult)
			}

			for _, grouplink := range groups {
				if _, err := fmt.Fprintf(configSidebar, "		- %s", grouplink); err != nil {
					common.Exit(err, ExitCodeFailedWriteResult)
				}
				if _, err := grouplinks.WriteString(grouplink); err != nil {
					common.Exit(err, ExitCodeFailedWriteResult)
				}
				if _, err := fmt.Fprintf(configfile, "- %s", grouplink); err != nil {
					common.Exit(err, ExitCodeFailedWriteResult)
				}
			}
			continue
		case "scheduler":
			generateWikiSchedulers(compiledDocs, configfile, configSidebar)
			continue
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
					_, _ = fmt.Fprintf(os.Stderr, "%v", err)
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
		if _, err := fmt.Fprintf(configSidebar, "	- %s", linkString); err != nil {
			common.Exit(err, ExitCodeFailedWriteResult)
		}
	}
}

func generateWikiGroups(compiledDocs *CompiledDocs) []string {
	// make sure generated same order every time
	sort.Sort(compiledDocs.Groups)
	groups := make([]string, len(compiledDocs.Groups))

	for i := 0; i < len(compiledDocs.Groups); i++ {
		group := compiledDocs.Groups[i]
		groups[i] = fmt.Sprintf("[%s](%s)\n\n", group.Title, group.Name)
		if verbose {
			fmt.Printf("Generating wiki actions for GROUP %s...\n", group.Name)
		}
		if err := createFolder(filepath.Join(wiki, GeneratedFolder, group.Name), false); err != nil {
			common.Exit(err, ExitCodeFailedCreateFolder)
		}
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
			_, _ = os.Stderr.Write([]byte(err.Error()))
		}
	}()
	if err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}

	if _, err := fmt.Fprintf(actionsSidebar, "[Home](Home)\n\n- [%s](%s)\n\n	- [Action Groups](groups)\n\n		- [%s](%s)\n\n", ConfigLinkName, ConfigLinkFile, group.Title, group.Name); err != nil {
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
		if _, err := fmt.Fprintf(actionsSidebar, "			- [%s](%s)\n\n", action, action); err != nil {
			common.Exit(err, ExitCodeFailedWriteResult)
		}
	}
}

func generateWikiSchedulers(compiledDocs *CompiledDocs, configFile, sidebar *os.File) {
	if _, err := sidebar.WriteString("	- Schedulers\n\n"); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}
	if _, err := configFile.WriteString("scheduler\n\n"); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}
	for _, entry := range createSchedulerEntrys(compiledDocs) {
		if verbose {
			fmt.Printf("Generating wiki scheduler entry for %s...\n", entry.Name)
		}
		file := fmt.Sprintf("%s/%s.md", filepath.Join(wiki, GeneratedFolder), entry.Name)
		if verbose {
			fmt.Printf("creating file<%s>...\n", file)
		}
		if err := os.WriteFile(file, []byte(entry.String()), os.ModePerm); err != nil {
			common.Exit(err, ExitCodeFailedWriteResult)
		}
		if _, err := fmt.Fprintf(sidebar, "		- [%s](%s)\n\n", entry.Name, entry.Name); err != nil {
			common.Exit(err, ExitCodeFailedWriteResult)
		}
		if _, err := fmt.Fprintf(configFile, "- [%s](%s)\n\n", entry.Name, entry.Name); err != nil {
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
		if err := os.WriteFile(fmt.Sprintf("%s/_Footer.md", path), []byte("This file has been automatically generated, do not edit manually\n"), os.ModePerm); err != nil {
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
		fmt.Fprintf(os.Stderr, "%s gives nil actionparams, skipping...\n", action)
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
	entries := createSchedulerEntrys(compiledDocs)
	for _, entry := range entries {
		newNode := NewFoldedDocNode(entry.Name, entry.DocEntryWithParams)
		node.AddChild(newNode)
	}
}

func createSchedulerEntrys(compiledDocs *CompiledDocs) []SortedDocEntryWithParams {
	schedulerSettings := common.Schedulers()

	// sort
	schedulers := make([]string, 0, len(schedulerSettings))
	for sched := range schedulerSettings {
		schedulers = append(schedulers, sched)
	}
	sort.Strings(schedulers)

	schedulerEntries := make([]SortedDocEntryWithParams, 0, len(schedulers))
	for _, sched := range schedulers {
		compiledEntry, ok := compiledDocs.Schedulers[sched]
		if !ok {
			compiledEntry.Description = "*Missing description*\n"
		}
		schedParams := schedulerSettings[sched]
		if schedParams == nil {
			fmt.Fprintf(os.Stderr, "%s gives nil schedparams, skipping...\n", sched)
			continue
		}
		schedulerEntries = append(schedulerEntries, SortedDocEntryWithParams{
			Name: sched,
			DocEntryWithParams: &DocEntryWithParams{
				DocEntry: DocEntry(compiledEntry),
				Params:   MarkdownParams(schedParams, compiledDocs.Params),
			},
		})
	}
	return schedulerEntries
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
