package doccompiler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
)

var UseFolderStructure = false

// ExitCodes
const (
	ExitCodeOk int = iota
	ExitCodeFailedReadParams
	ExitCodeFailedHandleAction
	ExitCodeFailedConfigFields
	ExitCodeFailedHandleConfig
	ExitCodeFailedWriteResult
	ExitCodeFailedReadGroups
	ExitCodeFailedHandleGroups
	ExitCodeFailedReadTemplate
	ExitCodeFailedParseTemplate
	ExitCodeFailedExecuteTemplate
	ExitCodeFailedCreateExtra
	ExitCodeFailedSyntaxError
	ExitCodeFailedNoDataRoot
	ExitCodeFailedListDir
)

type (
	docData struct {
		ParamMap     map[string][]string
		Groups       []common.GroupsEntry
		Actions      []string
		ActionMap    map[string]common.DocEntry
		ConfigFields []string
		ConfigMap    map[string]common.DocEntry
		Extra        []string
		ExtraMap     map[string]common.DocEntry
	}

	DocCompiler interface {
		// Compile documentation to golang represented as bytes
		Compile() []byte
		// CompileToFile compiles the data to file
		CompileToFile(file string)
		// Add documentation data from directory
		AddDataFromDir(dir string)
		// Add documentation data from variables in generated code
		AddDataFromGenerated(actions, config, extra map[string]common.DocEntry, params map[string][]string, groups []common.GroupsEntry)
	}
)

var (
	funcMap = template.FuncMap{
		"params": SortedParamsKeys,
		"join":   strings.Join,
	}
	// templateFile  string
	// Todo: Better way to do this? Using "search and replace" doesn't seem very robust.
	prepareString = strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\"", "\\\"")
)

func New() DocCompiler {
	return newData()
}

func (data *docData) Compile() []byte {
	data.sort()
	docs := generateDocs(data)
	formattedDocs, err := format.Source(docs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generated code has syntax error(s):\n  %v\n", err)
		os.Exit(ExitCodeFailedSyntaxError)
	}
	checkAndWarn(data)
	return formattedDocs
}

func (data *docData) CompileToFile(fileName string) {
	docs := data.Compile()
	if err := ioutil.WriteFile(fileName, docs, 0644); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}
	fmt.Printf("Compiled documentation to %s\n", fileName)
}

func (data *docData) AddDataFromGenerated(actions, config, extra map[string]common.DocEntry, params map[string][]string, groups []common.GroupsEntry) {
	prepareDocEntries(actions)
	prepareDocEntries(config)
	prepareDocEntries(extra)
	prepareGroupDocEntries(groups)
	data.overload(
		&docData{
			ParamMap:     params,
			Groups:       groups,
			Actions:      common.Keys(actions),
			ActionMap:    actions,
			ConfigFields: common.Keys(config),
			ConfigMap:    config,
			Extra:        common.Keys(extra),
			ExtraMap:     extra,
		},
	)

}

func (data *docData) AddDataFromDir(dataRoot string) {
	data.overload(loadData(dataRoot))
}

func newData() *docData {
	return &docData{
		ParamMap:     map[string][]string{},
		Groups:       []common.GroupsEntry{},
		Actions:      []string{},
		ActionMap:    map[string]common.DocEntry{},
		ConfigFields: []string{},
		ConfigMap:    map[string]common.DocEntry{},
		Extra:        []string{},
		ExtraMap:     map[string]common.DocEntry{},
	}
}

func (data *docData) sort() {
	sort.Slice(data.Groups, func(i, j int) bool {
		return data.Groups[i].Name < data.Groups[j].Name
	})
	sort.Strings(data.Actions)
	sort.Strings(data.ConfigFields)
	sort.Strings(data.Extra)
}

func prepareDocEntry(docEntry common.DocEntry) common.DocEntry {
	return common.DocEntry{
		Description: prepareString.Replace(docEntry.Description),
		Examples:    prepareString.Replace(docEntry.Examples),
	}

}

func prepareDocEntries(docEntries map[string]common.DocEntry) {
	for entryName, docEntry := range docEntries {
		docEntries[entryName] = prepareDocEntry(docEntry)
	}
}

func prepareGroupDocEntries(groups []common.GroupsEntry) {
	for idx, group := range groups {
		groups[idx].DocEntry = prepareDocEntry(group.DocEntry)
	}
}

func generateDocs(data *docData) []byte {
	// Create template for generating documentation.go
	documentationTemplate, err := template.New("documentationTemplate").Funcs(funcMap).Parse(TemplateStr)
	if err != nil {
		common.Exit(err, ExitCodeFailedParseTemplate)
	}

	buf := bytes.NewBuffer(nil)
	if err := documentationTemplate.Execute(buf, data); err != nil {
		common.Exit(err, ExitCodeFailedExecuteTemplate)
	}

	return buf.Bytes()
}

func mergeGroups(baseGroups []common.GroupsEntry, newGroups []common.GroupsEntry) []common.GroupsEntry {
	// init new group lookup table
	newGroupMap := make(map[string]common.GroupsEntry, len(newGroups))
	for _, g := range newGroups {
		newGroupMap[g.Name] = g
	}

	// init return value
	mergedGroups := make([]common.GroupsEntry, 0, len(baseGroups)+len(newGroups))

	// merge groups existing in base
	for _, baseGroup := range baseGroups {
		if newGroup, existInBase := newGroupMap[baseGroup.Name]; existInBase {
			// mark new group as merged by deleting it from lookup table
			delete(newGroupMap, baseGroup.Name)

			// override string fields
			if newGroup.Description != "" {
				baseGroup.Description = newGroup.Description
			}
			if newGroup.Examples != "" {
				baseGroup.Examples = newGroup.Examples
			}
			if newGroup.Title != "" {
				baseGroup.Title = newGroup.Title
			}

			//append actions
			baseGroup.Actions = append(baseGroup.Actions, newGroup.Actions...)
		}
		mergedGroups = append(mergedGroups, baseGroup)
	}

	// append unmerged groups
	// slice newGroups is iterated to preserve order
	for _, g := range newGroups {
		if _, ok := newGroupMap[g.Name]; ok {
			mergedGroups = append(mergedGroups, g)
		}
	}

	return mergedGroups
}

func overloadDocMap(baseMap, newMap map[string]common.DocEntry, baseNames *[]string, newNames []string) {
	if baseNames != nil {
		for _, newName := range newNames {
			if _, ok := baseMap[newName]; !ok {
				*baseNames = append(*baseNames, newName)
			}
		}
	}
	for newName, newDocEntry := range newMap {
		if newDocEntry.Description == "" {
			newDocEntry.Description = baseMap[newName].Description
		}
		if newDocEntry.Examples == "" {
			newDocEntry.Examples = baseMap[newName].Examples
		}
		baseMap[newName] = newDocEntry
	}
}

// overload assumes data, newData and their members are initialized
func (baseData *docData) overload(newData *docData) {
	// overload parameters
	for docKey, paramInfo := range newData.ParamMap {
		baseData.ParamMap[docKey] = paramInfo
	}

	// overload groups
	baseData.Groups = mergeGroups(baseData.Groups, newData.Groups)

	// overload actions
	overloadDocMap(baseData.ActionMap, newData.ActionMap, &baseData.Actions, newData.Actions)

	// overload config
	overloadDocMap(baseData.ConfigMap, newData.ConfigMap, &baseData.ConfigFields, newData.ConfigFields)

	// overload extra
	overloadDocMap(baseData.ExtraMap, newData.ExtraMap, &baseData.Extra, newData.Extra)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func subDirs(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil && !os.IsNotExist(err) {
		common.Exit(err, ExitCodeFailedListDir)
	}

	dirs := []string{}
	for _, f := range files {
		if f.IsDir() {
			dirs = append(dirs, f.Name())
		}
	}
	return dirs
}

func populateDocMap(dataRoot, subDir string, docMap map[string]common.DocEntry, entryNames *[]string) {
	*entryNames = subDirs(fmt.Sprintf("%s/%s", dataRoot, subDir))
	for _, entryName := range *entryNames {
		docEntry, err := CreateDocEntry(dataRoot, subDir, entryName)
		if err != nil {
			common.Exit(err, ExitCodeFailedHandleAction)
		}
		docMap[entryName] = docEntry
	}
}

func loadData(dataRoot string) *docData {
	data := newData()

	// Get parameters
	if err := ReadAndUnmarshal(fmt.Sprintf("%s/params.json", dataRoot), &data.ParamMap); err != nil {
		common.Exit(err, ExitCodeFailedReadParams)
	}

	// Get Groups
	var groups []common.GroupsEntry
	if err := ReadAndUnmarshal(fmt.Sprintf("%s/groups/groups.json", dataRoot), &groups); err != nil {
		common.Exit(err, ExitCodeFailedReadGroups)
	}

	for _, group := range groups {
		var err error
		group.DocEntry, err = CreateDocEntry(dataRoot, "groups", group.Name)
		if err != nil {
			common.Exit(err, ExitCodeFailedHandleGroups)
		}
		data.Groups = append(data.Groups, group)
	}

	populateDocMap(dataRoot, "actions", data.ActionMap, &data.Actions)
	populateDocMap(dataRoot, "config", data.ConfigMap, &data.ConfigFields)
	populateDocMap(dataRoot, "extra", data.ExtraMap, &data.Extra)

	return data

}

// ReadAndUnmarshal file to object
func ReadAndUnmarshal(filename string, output interface{}) error {
	fileData, err := common.ReadFile(filename)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(fileData, output); err != nil {
		return err
	}

	return nil
}

// CreateDocEntry create DocEntry using files in sub folder
func CreateDocEntry(dataRoot string, subFolders ...string) (common.DocEntry, error) {
	var docEntry common.DocEntry
	var err error

	docEntry.Description, err = GetMarkDownFile(dataRoot, subFolders, "description.md")
	if err != nil {
		return docEntry, err
	}

	docEntry.Examples, err = GetMarkDownFile(dataRoot, subFolders, "examples.md")
	if err != nil {
		return docEntry, err
	}

	return docEntry, nil
}

// GetMarkDownFile read markdown file into memory and do necessary escaping
func GetMarkDownFile(dataRoot string, subFolders []string, file string) (string, error) {
	subPath := strings.Join(subFolders, "/")
	filepath := fmt.Sprintf("%s/%s/%s", dataRoot, subPath, file)

	if exist, err := FileExists(filepath); err != nil {
		return "", err
	} else if !exist {
		// _, _ = os.Stderr.WriteString(fmt.Sprintf("Warning: %s does not have a %s file\n", subPath, file))
		return "", nil
	}

	markdown, err := common.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return prepareString.Replace(string(markdown)), nil
}

// FileExists check if file exists
func FileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
