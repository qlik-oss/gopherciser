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
	Data struct {
		ParamMap     map[string][]string
		Groups       []common.GroupsEntry
		Actions      []string
		ActionMap    map[string]common.DocEntry
		ConfigFields []string
		ConfigMap    map[string]common.DocEntry
		Extra        []string
		ExtraMap     map[string]common.DocEntry
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

func (data *Data) sort() {
	sort.Slice(data.Groups, func(i, j int) bool {
		return data.Groups[i].Name < data.Groups[j].Name
	})
	sort.Strings(data.Actions)
	sort.Strings(data.ConfigFields)
	sort.Strings(data.Extra)
}

func (data *Data) Compile() []byte {
	data.sort()
	docs := generateDocs(data)
	formattedDocs, err := format.Source(docs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generated code has syntax error(s):\n  %v\n", err)
		os.Exit(ExitCodeFailedSyntaxError)
	}
	return formattedDocs
}

func (data *Data) CompileToFile(fileName string) {
	docs := data.Compile()
	if err := ioutil.WriteFile(fileName, docs, 0644); err != nil {
		common.Exit(err, ExitCodeFailedWriteResult)
	}
	fmt.Printf("Compiled to %s\n", fileName)
}

func NewData() *Data {
	return &Data{
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

func (data *Data) PopulateFromGenerated(actions, config, extra map[string]common.DocEntry, params map[string][]string, groups []common.GroupsEntry) {
	prepareDocEntries(actions)
	prepareDocEntries(config)
	prepareDocEntries(extra)
	prepareGroupDocEntries(groups)
	data.overload(
		&Data{
			ParamMap:     params,
			Groups:       groups,
			Actions:      keys(actions),
			ActionMap:    actions,
			ConfigFields: keys(config),
			ConfigMap:    config,
			Extra:        keys(extra),
			ExtraMap:     extra,
		},
	)

}

func (data *Data) PopulateFromDataDir(dataRoot string) {
	data.overload(loadData(dataRoot))
}

func keys(m map[string]common.DocEntry) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func generateDocs(data *Data) []byte {
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
		*baseNames = append(*baseNames, newNames...)
	}
	for k, v := range newMap {
		baseMap[k] = v
	}
}

// overload assumes data, newData and their members are initialized
func (baseData *Data) overload(newData *Data) {
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

// func groupNames(groups []common.GroupsEntry) []string {
// 	names := make([]string, 0, len(groups))
// 	for _, group := range groups {
// 		names = append(names, group.Name)
// 	}
// 	return names
// }

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

func loadData(dataRoot string) *Data {
	data := NewData()

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

	// if UseFolderStructure {
	// 	data.ConfigFields = subdirs(dataRoot + "/config")
	// } else {
	// 	var err error
	// 	// Get all config fields
	// 	data.ConfigFields, err = common.FieldsString()
	// 	if err != nil {
	// 		common.Exit(err, ExitCodeFailedConfigFields)
	// 	}
	// 	// Add documentation wrapping entire document as "main" entry into config map
	// 	data.ConfigFields = append(data.ConfigFields, "main")
	// }
	// sort.Strings(data.ConfigFields)

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
