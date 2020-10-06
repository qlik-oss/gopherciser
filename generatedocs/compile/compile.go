package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/qlik-oss/gopherciser/generatedocs/common"
)

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
	dataRootParam string
	dataRoots     []string
	output        string
	templateFile  string
	// Todo: Better way to do this? Using "search and replace" doesn't seem very robust.
	prepareString = strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\"", "\\\"")
)

func main() {
	handleFlags()
	generatedDocs := compile(templateFile, dataRoots...)
	if err := ioutil.WriteFile(output, generatedDocs, 0644); err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		os.Exit(ExitCodeFailedWriteResult)
	}
	fmt.Printf("Compiled data<%s> to output<%s>\n", dataRootParam, output)
}

func compile(templatePath string, dataRoots ...string) []byte {
	if len(dataRoots) == 0 {
		fmt.Fprint(os.Stderr, "no data roots passed")
		os.Exit(ExitCodeFailedNoDataRoot)
	}
	data := loadData(dataRoots[0])
	for _, dataRoot := range dataRoots[1:] {
		data.overload(loadData(dataRoot))

	}
	tmplBytes, err := common.ReadFile(templatePath)
	if err != nil {
		common.Exit(err, ExitCodeFailedReadTemplate)
	}
	docs := generateDocs(tmplBytes, data)
	formattedDocs, err := format.Source(docs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generated code has syntax error(s):\n%v\n", err)
		os.Exit(ExitCodeFailedSyntaxError)
	}
	return formattedDocs
}

func handleFlags() {
	flagHelp := flag.Bool("help", false, "shows help")
	flag.StringVar(&dataRootParam, "data", "generatedocs/data", "a comma separated list of paths to data folders")
	flag.StringVar(&templateFile, "template", "generatedocs/data/documentation.template", "path to template file")
	flag.StringVar(&output, "output", "generatedocs/generated/documentation.go", "path to generated code file")

	flag.Parse()

	if *flagHelp {
		flag.PrintDefaults()
		os.Exit(ExitCodeOk)
	}

	dataRoots = strings.Split(dataRootParam, ",")
}

func generateDocs(tmpl []byte, data *Data) []byte {
	// Create template for generating documentation.go
	documentationTemplate, err := template.New("documentationTemplate").Funcs(funcMap).Parse(string(tmpl))
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
			if baseGroup.Description != "" {
				baseGroup.Description = newGroup.Description
			}
			if baseGroup.Examples != "" {
				baseGroup.Examples = newGroup.Examples
			}
			if baseGroup.Title != "" {
				baseGroup.Title = newGroup.Title
			}

			//append actions
			baseGroup.Actions = append(baseGroup.Actions, newGroup.Actions...)
		}
		mergedGroups = append(mergedGroups, baseGroup)
	}

	// append unmerged groups
	for _, g := range newGroupMap {
		mergedGroups = append(mergedGroups, g)
	}

	return mergedGroups
}

func overloadDocMap(baseMap, newMap map[string]common.DocEntry) {
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
	baseData.Actions = append(baseData.Actions, newData.Actions...)
	overloadDocMap(baseData.ActionMap, newData.ActionMap)

	// overload config
	baseData.ConfigFields = append(baseData.ConfigFields, newData.ConfigFields...)
	overloadDocMap(baseData.ConfigMap, newData.ConfigMap)

	// overload extra
	baseData.Extra = append(baseData.Extra, newData.Extra...)
	overloadDocMap(baseData.ExtraMap, newData.ExtraMap)
}

func loadData(dataRoot string) *Data {
	data := &Data{}

	// Get parameters
	data.ParamMap = make(map[string][]string)
	if err := ReadAndUnmarshal(fmt.Sprintf("%s/params.json", dataRoot), &data.ParamMap); err != nil {
		common.Exit(err, ExitCodeFailedReadParams)
	}

	// Get Groups
	var groups []common.GroupsEntry
	if err := ReadAndUnmarshal(fmt.Sprintf("%s/groups/groups.json", dataRoot), &groups); err != nil {
		common.Exit(err, ExitCodeFailedReadGroups)
	}

	data.Groups = make([]common.GroupsEntry, 0, len(groups))
	for _, group := range groups {
		var err error
		group.DocEntry, err = CreateGroupsDocEntry(dataRoot, group.Name)
		if err != nil {
			common.Exit(err, ExitCodeFailedHandleGroups)
		}
		data.Groups = append(data.Groups, group)
	}

	// Get all actions
	data.Actions = common.ActionStrings()
	data.ActionMap = make(map[string]common.DocEntry, len(data.Actions))
	for _, action := range data.Actions {
		actionDocEntry, err := CreateActionDocEntry(dataRoot, action)
		if err != nil {
			common.Exit(err, ExitCodeFailedHandleAction)
		}

		data.ActionMap[action] = actionDocEntry
	}

	// Get all config fields
	fields, err := common.FieldsString()
	if err != nil {
		common.Exit(err, ExitCodeFailedConfigFields)
	}
	data.ConfigFields = fields

	// Add documentation wrapping entire document as "main" entry into config map
	data.ConfigFields = append(data.ConfigFields, "main")

	data.ConfigMap = make(map[string]common.DocEntry, len(data.ConfigFields))
	for _, field := range data.ConfigFields {
		configDocEntry, err := CreateConfigDocEntry(dataRoot, field)
		if err != nil {
			common.Exit(err, ExitCodeFailedHandleConfig)
		}
		data.ConfigMap[field] = configDocEntry
	}

	// Walk "extra" folder and add things outside normal structure
	if err := CreateExtraDocEntries(dataRoot, data); err != nil {
		common.Exit(err, ExitCodeFailedCreateExtra)
	}

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

// CreateActionDocEntry create DocEntry from actions sub directory
func CreateActionDocEntry(dataRoot, action string) (common.DocEntry, error) {
	return CreateDocEntry(dataRoot, []string{"actions", action})
}

// CreateConfigDocEntry create DocEntry from config sub directory
func CreateConfigDocEntry(dataRoot string, field string) (common.DocEntry, error) {
	return CreateDocEntry(dataRoot, []string{"config", field})
}

// CreateGroupsDocEntry create DocEntry from groups sub directory
func CreateGroupsDocEntry(dataRoot string, group string) (common.DocEntry, error) {
	return CreateDocEntry(dataRoot, []string{"groups", group})
}

// CreateExtraDocEntries create DocEntries for sub folders to "extra" folder
func CreateExtraDocEntries(dataRoot string, data *Data) error {
	dataDir, err := os.Open(fmt.Sprintf("%s/extra", dataRoot))
	if err != nil {
		return err
	}

	// Read all the files in the dataRoot/extra directory
	fileInfos, err := dataDir.Readdir(-1)
	_ = dataDir.Close()
	if err != nil {
		return err
	}

	data.ExtraMap = make(map[string]common.DocEntry)

	for _, fi := range fileInfos {
		if !fi.IsDir() {
			continue
		}
		data.Extra = append(data.Extra, fi.Name())
		data.ExtraMap[fi.Name()], err = CreateDocEntry(dataRoot, []string{"extra", fi.Name()})
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateDocEntry create DocEntry using files in sub folder
func CreateDocEntry(dataRoot string, subFolders []string) (common.DocEntry, error) {
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
		_, _ = os.Stderr.WriteString(fmt.Sprintf("Warning: %s does not have a %s file\n", subPath, file))
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

// SortedParamsKeys returns map keys as a sorted slice
func SortedParamsKeys(paramsMap map[string][]string) []string {
	params := make([]string, 0, len(paramsMap))
	for param := range paramsMap {
		params = append(params, param)
	}
	sort.Strings(params)
	return params
}
