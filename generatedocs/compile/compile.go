package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
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
	dataRoot, output string
	// Todo: Better way to do this? Using "search and replace" doesn't seem very robust.
	prepareString = strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\"", "\\\"")
)

func main() {
	handleFlags()
	compile()
}

func handleFlags() {
	flagHelp := flag.Bool("help", false, "shows help")
	flag.StringVar(&dataRoot, "data", "generatedocs/data", "path to data folder")
	flag.StringVar(&output, "output", "generatedocs/generated/documentation.go", "path to generated code file")

	flag.Parse()

	if *flagHelp {
		flag.PrintDefaults()
		os.Exit(ExitCodeOk)
	}
}

func compile() {
	var data Data

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
		group.DocEntry, err = CreateGroupsDocEntry(group.Name)
		if err != nil {
			common.Exit(err, ExitCodeFailedHandleGroups)
		}
		data.Groups = append(data.Groups, group)
	}

	// Get all actions
	data.Actions = common.ActionStrings()
	data.ActionMap = make(map[string]common.DocEntry, len(data.Actions))
	for _, action := range data.Actions {
		actionDocEntry, err := CreateActionDocEntry(action)
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
		configDocEntry, err := CreateConfigDocEntry(field)
		if err != nil {
			common.Exit(err, ExitCodeFailedHandleConfig)
		}
		data.ConfigMap[field] = configDocEntry
	}

	// Walk "extra" folder and add things outside normal structure
	if err := CreateExtraDocEntries(&data); err != nil {
		common.Exit(err, ExitCodeFailedCreateExtra)
	}

	// Create template for generating documentation.go
	docTemplateFile, err := common.ReadFile(fmt.Sprintf("%s/documentation.template", dataRoot))
	if err != nil {
		common.Exit(err, ExitCodeFailedReadTemplate)
	}
	documentationTemplate, err := template.New("documentationTemplate").Funcs(funcMap).Parse(string(docTemplateFile))
	if err != nil {
		common.Exit(err, ExitCodeFailedParseTemplate)
	}

	buf := bytes.NewBuffer(nil)
	if err := documentationTemplate.Execute(buf, data); err != nil {
		common.Exit(err, ExitCodeFailedExecuteTemplate)
	}

	if err := ioutil.WriteFile(output, buf.Bytes(), 0644); err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
		os.Exit(ExitCodeFailedWriteResult)
	}

	fmt.Printf("Compiled data<%s> to output<%s>\n", dataRoot, output)
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
func CreateActionDocEntry(action string) (common.DocEntry, error) {
	return CreateDocEntry([]string{"actions", action})
}

// CreateConfigDocEntry create DocEntry from config sub directory
func CreateConfigDocEntry(field string) (common.DocEntry, error) {
	return CreateDocEntry([]string{"config", field})
}

// CreateGroupsDocEntry create DocEntry from groups sub directory
func CreateGroupsDocEntry(group string) (common.DocEntry, error) {
	return CreateDocEntry([]string{"groups", group})
}

// CreateExtraDocEntries create DocEntries for sub folders to "extra" folder
func CreateExtraDocEntries(data *Data) error {
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
		data.ExtraMap[fi.Name()], err = CreateDocEntry([]string{"extra", fi.Name()})
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateDocEntry create DocEntry using files in sub folder
func CreateDocEntry(subFolders []string) (common.DocEntry, error) {
	var docEntry common.DocEntry
	var err error

	docEntry.Description, err = GetMarkDownFile(subFolders, "description.md")
	if err != nil {
		return docEntry, err
	}

	docEntry.Examples, err = GetMarkDownFile(subFolders, "examples.md")
	if err != nil {
		return docEntry, err
	}

	return docEntry, nil
}

// GetMarkDownFile read markdown file into memory and do necessary escaping
func GetMarkDownFile(subFolders []string, file string) (string, error) {
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
