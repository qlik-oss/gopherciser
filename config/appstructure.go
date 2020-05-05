package config

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/scheduler"
	"github.com/qlik-oss/gopherciser/senseobjdef"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	getAppStructureSettings struct {
		IncludeRaw bool        `json:"includeRaw,omitempty"`
		Summary    SummaryType `json:"summary,omitempty"`
	}

	// MetaDef meta information for Library objects such as dimension and measure
	MetaDef struct {
		// Title of library item
		Title string `json:"title,omitempty"`
		// Description of library item
		Description string `json:"description,omitempty"`
		// Tags of  of library item
		Tags []string `json:"tags,omitempty"`
	}

	// AppObjectDef title and ID of a Sense object
	AppObjectDef struct {
		// Id of object
		Id string `json:"id"`
		// Type of Sense object
		Type string `json:"type"`
	}

	AppStructureMeasureMeta struct {
		// Meta information, only included for library items
		Meta *MetaDef `json:"meta,omitempty"`
		// LibraryId connects measure to separately defined measure
		LibraryId string `json:"libraryId,omitempty"`
		// Label of on measure
		Label string `json:"label,omitempty"`
		// Def the actual measure definition
		Def string `json:"def,omitempty"`
	}

	AppStructureDimensionMeta struct {
		// Meta information, only included for library items
		Meta *MetaDef `json:"meta,omitempty"`
		// LibraryId connects dimension to separately defined dimension
		LibraryId string `json:"libraryId,omitempty"`
		// LabelExpression optional parameter with label expression
		LabelExpression string `json:"labelExpression,omitempty"`
		// Defs definitions of dimension
		Defs []string `json:"defs,omitempty"`
		// Labels of dimension
		Labels []string `json:"labels,omitempty"`
	}

	// AppStructureObject sense object structure
	AppStructureObject struct {
		AppObjectDef
		MetaDef
		// RawProperties of Sense object
		RawProperties json.RawMessage `json:"rawProperties,omitempty"`
		// RawProperties of extended Sense object
		RawExtendedProperties json.RawMessage `json:"rawExtendedProperties,omitempty"`
		// Children to the sense object
		Children map[string]string `json:"children,omitempty"`
		// Selectable true if select can be done in object
		Selectable bool `json:"selectable"`
		// Dimensions meta information of dimensions defined in object
		Dimensions []AppStructureDimensionMeta `json:"dimensions,omitempty"`
		// Measures meta information of measures defined in object
		Measures []AppStructureMeasureMeta `json:"measures,omitempty"`
		// ExtendsId ID of linked object
		ExtendsId string `json:"extendsId,omitempty"`
		// Visualization visualization of object, if exists
		Visualization string `json:"visualization,omitempty"`
	}

	// AppStructureAppMeta meta information about the app
	AppStructureAppMeta struct {
		// Title of the app
		Title string `json:"title"`
		// Guid of the app
		Guid string `json:"guid"`
	}

	// AppStructureBookmark list of bookmarks in the app
	AppStructureBookmark struct {
		// ID of bookmark
		ID string `json:"id"`
		// Title of bookmark
		Title string `json:"title"`
		// Description of bookmark
		Description string `json:"description"`
		// SheetId connected sheet ID, null if none
		SheetId *string `json:"sheetId,omitempty"`
		// SelectionFields fields bookmark would select in
		SelectionFields string `json:"selectionFields"`
	}

	// AppStructureReport reports warnings and fetched objects for app structure
	AppStructureReport struct {
		warnings []string
		objects  []AppObjectDef

		warningsLock sync.Mutex
		objectsLock  sync.Mutex
	}

	// AppStructure of Sense app
	AppStructure struct {
		AppMeta AppStructureAppMeta `json:"meta"`
		// Objects in Sense app
		Objects map[string]AppStructureObject `json:"objects"`
		// Bookmarks list of bookmarks in the app
		Bookmarks map[string]AppStructureBookmark `json:"bookmarks"`

		logEntry      *logger.LogEntry
		report        AppStructureReport
		structureLock sync.Mutex
	}

	ObjectType                      int
	AppStructureObjectNotFoundError string
)

const (
	ObjectTypeDefault ObjectType = iota
	ObjectTypeDimension
	ObjectTypeMeasure
	ObjectTypeBookmark
	ObjectTypeMasterObject
	ObjectTypeAutoChart
	ObjectSheet
	ObjectLoadModel
	ObjectAppprops
)

var (
	ObjectTypeEnumMap, _ = enummap.NewEnumMap(map[string]int{
		"dimension":    int(ObjectTypeDimension),
		"measure":      int(ObjectTypeMeasure),
		"bookmark":     int(ObjectTypeBookmark),
		"masterobject": int(ObjectTypeMasterObject),
		"auto-chart":   int(ObjectTypeAutoChart),
		"sheet":        int(ObjectSheet),
		"loadmodel":    int(ObjectLoadModel),
		"appprops":     int(ObjectAppprops),
	})
)

// Error object was not found in app structure
func (err AppStructureObjectNotFoundError) Error() string {
	return string(err)
}

func (cfg *Config) getAppStructureScenario(includeRaw bool, summary SummaryType) []scenario.Action {
	return evaluateActionList(cfg.Scenario, includeRaw, summary)
}

func evaluateActionList(actions []scenario.Action, includeRaw bool, summary SummaryType) []scenario.Action {
	appStructureScenario := make([]scenario.Action, 0, len(actions))
	for _, act := range actions {
		info, subActions := act.AppStructureAction()
		if info != nil {
			if info.Include {
				appStructureScenario = append(appStructureScenario, act)
			}
			if info.IsAppAction {
				appStructureScenario = append(appStructureScenario, scenario.Action{
					ActionCore: scenario.ActionCore{
						Type:  "getappstructure",
						Label: "Get app structure",
					},
					Settings: &getAppStructureSettings{
						IncludeRaw: includeRaw,
						Summary:    summary,
					},
				})

			}
			if len(subActions) > 0 {
				appStructureScenario = append(appStructureScenario, evaluateActionList(subActions, includeRaw, summary)...)
			}
		}
	}
	return appStructureScenario
}

// GetAppStructures for all apps in scenario
func (cfg *Config) GetAppStructures(ctx context.Context, includeRaw bool) error {
	// find all auth and actions
	appStructureScenario := cfg.getAppStructureScenario(includeRaw, cfg.Settings.LogSettings.getSummaryType())
	if len(appStructureScenario) < 1 {
		return errors.New("no applicable actions in scenario") // Todo warn only? Only include in summary?
	}

	// Replace scheduler with 1 iteration 1 user simple scheduler
	cfg.Scheduler = &scheduler.SimpleScheduler{
		Scheduler: scheduler.Scheduler{
			SchedType: scheduler.SchedSimple,
		},
		Settings: scheduler.SimpleSchedSettings{
			ExecutionTime:   -1,
			Iterations:      1,
			ConcurrentUsers: 1,
			RampupDelay:     1.0,
		},
	}

	if err := cfg.Scheduler.Validate(); err != nil {
		return errors.WithStack(err)
	}

	// Setup outputs folder
	outputsDir, err := setupOutputs(cfg.Settings.OutputsSettings)
	if err != nil {
		return errors.WithStack(err)
	}

	logSettings := cfg.Settings.LogSettings

	fileName := cfg.Settings.LogSettings.FileName.String()
	ext := filepath.Ext(fileName)
	appStructureLogPath := fmt.Sprintf("%s-appstructure%s", strings.TrimSuffix(fileName, ext), ext)
	stmpl, err := session.NewSyncedTemplate(appStructureLogPath)
	if err != nil {
		return errors.WithStack(err)
	}
	logSettings.FileName = *stmpl

	log, err := setupLogging(ctx, logSettings, nil, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	timeout := time.Duration(cfg.Settings.Timeout) * time.Second
	if err := cfg.Scheduler.Execute(ctx, log, timeout, appStructureScenario, outputsDir, cfg.LoginSettings, &cfg.ConnectionSettings); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// Validate implements ActionSettings interface
func (settings *getAppStructureSettings) Validate() error {
	return nil
}

// Execute implements ActionSettings interface
func (settings *getAppStructureSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment"))
		return
	}

	app := sessionState.Connection.Sense().CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense app"))
		return
	}

	var allInfos []*enigma.NxInfo

	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		var err error
		allInfos, err = app.Doc.GetAllInfos(ctx)
		return err
	}); err != nil {
		actionState.AddErrors(err)
		return
	}

	appStructure := &AppStructure{
		AppMeta: AppStructureAppMeta{
			Title: app.Layout.Title,
			Guid:  app.GUID,
		},
		logEntry: sessionState.LogEntry,
	}

	for _, info := range allInfos {
		if info == nil {
			continue
		}
		if err := appStructure.getStructureForObjectAsync(sessionState, actionState, app, info.Id, info.Type, settings.IncludeRaw); err != nil {
			actionState.AddErrors(err)
			return
		}
	}

	if sessionState.Wait(actionState) {
		return // An error occurred
	}

	raw, err := json.MarshalIndent(appStructure, "", "  ")
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "error marshaling app structure"))
		return
	}

	outputDir := sessionState.OutputsDir
	if outputDir != "" && outputDir[len(outputDir)-1:] != "/" {
		outputDir += "/"
	}

	structureFile, err := os.Create(fmt.Sprintf("%s%s.structure", outputDir, app.GUID))
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to create structure file"))
		return
	}
	defer func() {
		if err := structureFile.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to close file<%v> successfully: %v\n", structureFile, err)
		}
	}()

	if _, err = structureFile.Write(raw); err != nil {
		actionState.AddErrors(errors.Wrap(err, "error while writing to structure file"))
		return
	}

	appStructure.printSummary(settings.Summary)
}

func (structure *AppStructure) printSummary(summary SummaryType) {
	if structure == nil || summary == SummaryTypeNone {
		return
	}

	buf := helpers.NewBuffer()
	//switch summary {
	//case SummaryTypeSimple:
	//// todo files + warns
	//case SummaryTypeExtended:
	//// todo files + warns + object count
	//case SummaryTypeFull:
	//	// todo files + warns + object specific
	//}

	if len(structure.report.warnings) > 0 {
		buf.WriteString(fmt.Sprintf("%d Warnings found:\n", len(structure.report.warnings)))
	}

	for _, warning := range structure.report.warnings {
		buf.WriteString(warning)
		buf.WriteString("\n")
	}

	buf.WriteTo(ansiWriter)
}

func (structure *AppStructure) getStructureForObjectAsync(sessionState *session.State, actionState *action.State, app *senseobjects.App, id, typ string, includeRaw bool) error {
	if structure == nil {
		return errors.New("appStructure object is nil")
	}

	sessionState.QueueRequest(func(ctx context.Context) error {
		obj := AppStructureObject{
			AppObjectDef: AppObjectDef{
				Id:   id,
				Type: typ,
			},
		}

		objectType := ObjectTypeDefault
		if oType, err := ObjectTypeEnumMap.Int(typ); err == nil {
			objectType = ObjectType(oType)
		}

		// handle some special types
		switch objectType {
		case ObjectTypeDimension:
			if err := structure.handleDimension(ctx, app, id, &obj); err != nil {
				return errors.WithStack(err)
			}
		case ObjectTypeMeasure:
			if err := structure.handleMeasure(ctx, app, id, &obj); err != nil {
				return errors.WithStack(err)
			}
		case ObjectTypeBookmark:
			if err := structure.handleBookmark(ctx, app, id); err != nil {
				return errors.WithStack(err)
			}
		case ObjectTypeAutoChart:
			if err := structure.handleAutoChart(ctx, app, id, &obj); err != nil {
				return errors.WithStack(err)
			}
		default:
			if err := structure.handleDefaultObject(ctx, app, id, typ, &obj); err != nil {
				return errors.WithStack(err)
			}
		}

		if !includeRaw {
			// Remove raw properties from structure output
			obj.RawProperties = nil
			obj.RawExtendedProperties = nil
		}

		return nil
	}, actionState, true, "")

	return nil
}

// AddObject to structure
func (structure *AppStructure) AddObject(obj AppStructureObject) {
	structure.structureLock.Lock()
	defer structure.structureLock.Unlock()
	if structure.Objects == nil {
		structure.Objects = make(map[string]AppStructureObject)
	}
	structure.Objects[obj.Id] = obj
}

// AddBookmark to structure
func (structure *AppStructure) AddBookmark(bookmark AppStructureBookmark) {
	structure.structureLock.Lock()
	defer structure.structureLock.Unlock()
	if structure.Bookmarks == nil {
		structure.Bookmarks = make(map[string]AppStructureBookmark)
	}
	structure.Bookmarks[bookmark.ID] = bookmark
}

// GetSelectables get selectable objects from app structure
func (structure *AppStructure) GetSelectables(rooObject string) ([]AppStructureObject, error) {
	rootObj, ok := structure.Objects[rooObject]
	if !ok {
		return nil, AppStructureObjectNotFoundError(rooObject)
	}

	return structure.addSelectableChildren(rootObj), nil
}

func (structure *AppStructure) addSelectableChildren(obj AppStructureObject) []AppStructureObject {
	selectables := make([]AppStructureObject, 0, 1)
	if obj.Selectable {
		selectables = append(selectables, obj)
	}

	for id := range obj.Children {
		child, ok := structure.Objects[id]
		if !ok {
			continue
		}

		selectableChildren := structure.addSelectableChildren(child)
		selectables = append(selectables, selectableChildren...)
	}
	return selectables
}

func (structure *AppStructure) warn(warning string) {
	structure.report.AddWarning(warning)
	if structure.logEntry != nil {
		structure.logEntry.Log(logger.WarningLevel, warning)
	}
}

func (structure *AppStructure) handleDefaultObject(ctx context.Context, app *senseobjects.App, id, typ string, obj *AppStructureObject) error {
	genObj, err := app.Doc.GetObject(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	obj.RawProperties, err = genObj.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	// Lookup and set ExtendsID
	extendsIdPath := senseobjdef.NewDataPath("/qExtendsId")
	rawExtendsID, _ := extendsIdPath.Lookup(obj.RawProperties)
	_ = jsonit.Unmarshal(rawExtendsID, &obj.ExtendsId)

	if obj.ExtendsId != "" {
		extendedObject, err := app.Doc.GetObject(ctx, obj.ExtendsId)
		if err != nil {
			return errors.WithStack(err)
		}
		obj.RawExtendedProperties, err = extendedObject.GetPropertiesRaw(ctx)
		if err != nil {
			return errors.WithStack(err)
		}
		if err := handleChildren(ctx, extendedObject, obj); err != nil {
			return errors.WithStack(err)
		}
	} else {
		if err := handleChildren(ctx, genObj, obj); err != nil {
			return errors.WithStack(err)
		}
	}

	return errors.WithStack(structure.handleObject(typ, obj))
}

func (structure *AppStructure) handleAutoChart(ctx context.Context, app *senseobjects.App, id string, obj *AppStructureObject) error {
	genObj, err := app.Doc.GetObject(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}

	autoChartProperties, err := genObj.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	generatedPropertiesPath := senseobjdef.NewDataPath("/qUndoExclude/generated")
	obj.RawProperties, _ = generatedPropertiesPath.Lookup(autoChartProperties)
	obj.RawExtendedProperties = autoChartProperties

	if err := handleChildren(ctx, genObj, obj); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(structure.handleObject(ObjectTypeEnumMap.StringDefault(int(ObjectTypeAutoChart), "auto-chart"), obj))
}

func handleChildren(ctx context.Context, genObj *enigma.GenericObject, obj *AppStructureObject) error {
	childInfos, err := genObj.GetChildInfos(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, child := range childInfos {
		if child == nil {
			continue
		}
		if obj.Children == nil {
			obj.Children = make(map[string]string)
		}
		obj.Children[child.Id] = child.Type
	}

	return nil
}

func (structure *AppStructure) handleObject(typ string, obj *AppStructureObject) error {
	// Lookup and set Visualization
	visualizationPath := senseobjdef.NewDataPath("/visualization")
	rawVisualization, _ := visualizationPath.Lookup(obj.RawProperties)
	_ = jsonit.Unmarshal(rawVisualization, &obj.Visualization)

	vis := obj.Visualization
	if vis == "" {
		vis = typ
	}

	properties := obj.RawProperties
	if obj.ExtendsId != "" && obj.RawExtendedProperties != nil {
		properties = obj.RawExtendedProperties
	}

	metaDef := senseobjdef.NewDataPath("/qMetaDef")
	rawMetaDef, _ := metaDef.Lookup(properties)
	_ = jsonit.Unmarshal(rawMetaDef, &obj.MetaDef)

	enumTyp, _ := ObjectTypeEnumMap.Int(typ) // 0 will be default in case of "error" == ObjectTypeDefault

	// Should we look for measures and dimensions?
	switch ObjectType(enumTyp) {
	case ObjectSheet, ObjectAppprops, ObjectLoadModel:
		// Known object which does not have measures and dimensions
		return nil
	}

	def, err := senseobjdef.GetObjectDef(vis)
	if err != nil {
		switch errors.Cause(err).(type) {
		case senseobjdef.NoDefError:
			structure.warn(fmt.Sprintf("Object type<%s> not supported", vis))
			return nil
		default:
			return errors.WithStack(err)
		}
	}

	// Set selectable flag
	obj.Selectable = def.Select != nil

	var dimensionsPath, measuresPath, listObjectPath string
	if obj.Selectable {
		// Hyper cube dimensions and measures
		dimensionsPath = fmt.Sprintf("%s/qDimensions", def.Select.Path)
		measuresPath = fmt.Sprintf("%s/qMeasures", def.Select.Path)

		// Handle list object
		listObjectPath = def.Select.Path
		if !strings.HasSuffix(listObjectPath, "/qListObjectDef") {
			listObjectPath = fmt.Sprintf("%s/qListObjectDef", def.Select.Path)
		}
	} else {
		// Guess data paths using DataDef.Path as base

		// Hyper cube dimensions and measures
		dimensionsPath = fmt.Sprintf("%sDef/qDimensions", def.DataDef.Path)
		measuresPath = fmt.Sprintf("%sDef/qMeasures", def.DataDef.Path)

		// Handle list object
		listObjectPath = fmt.Sprintf("%sDef", def.DataDef.Path)
		if !strings.HasSuffix(listObjectPath, "/qListObjectDef") {
			listObjectPath = fmt.Sprintf("%s/qListObjectDef", def.DataDef.Path)
		}
	}

	dimensions := senseobjdef.NewDataPath(dimensionsPath)
	measures := senseobjdef.NewDataPath(measuresPath)
	listObjects := senseobjdef.NewDataPath(listObjectPath)

	// Try to set dimensions and measures, null if not exist or not parsable (error)
	rawDimensions, _ := dimensions.Lookup(properties)
	if rawDimensions != nil {
		var dimensions []*enigma.NxDimension
		if err := jsonit.Unmarshal(rawDimensions, &dimensions); err != nil {
			return errors.WithStack(err)
		}
		obj.Dimensions = make([]AppStructureDimensionMeta, 0, len(dimensions))
		for _, dimension := range dimensions {
			if dimension == nil {
				continue
			}

			obj.Dimensions = append(obj.Dimensions, AppStructureDimensionMeta{
				LibraryId:       dimension.LibraryId,
				LabelExpression: dimension.Def.LabelExpression,
				Defs:            dimension.Def.FieldDefs,
				Labels:          dimension.Def.FieldLabels,
			})
		}
	}

	rawMeasures, _ := measures.Lookup(properties)
	if rawMeasures != nil {
		var measures []*enigma.NxMeasure
		if err := jsonit.Unmarshal(rawMeasures, &measures); err != nil {
			return errors.WithStack(err)
		}
		obj.Measures = make([]AppStructureMeasureMeta, 0, len(measures))
		for _, measure := range measures {
			if measure == nil {
				continue
			}
			obj.Measures = append(obj.Measures, AppStructureMeasureMeta{
				LibraryId: measure.LibraryId,
				Label:     measure.Def.Label,
				Def:       measure.Def.Def,
			})
		}
	}

	rawListObject, _ := listObjects.Lookup(properties)
	if rawListObject != nil {
		var listObject enigma.ListObjectDef
		if err := jsonit.Unmarshal(rawListObject, &listObject); err != nil {
			return errors.WithStack(err)
		}
		obj.Dimensions = []AppStructureDimensionMeta{
			{
				LibraryId:       listObject.LibraryId,
				LabelExpression: listObject.Def.LabelExpression,
				Defs:            listObject.Def.FieldDefs,
				Labels:          listObject.Def.FieldLabels,
			},
		}

		if obj.Measures == nil {
			obj.Measures = make([]AppStructureMeasureMeta, 0, len(listObject.Expressions))
		}

		for _, expression := range listObject.Expressions {
			obj.Measures = append(obj.Measures, AppStructureMeasureMeta{
				LibraryId: expression.LibraryId,
				Def:       expression.Expr,
			})
		}
	}

	if obj.Selectable && (len(obj.Dimensions)+len(obj.Measures)) < 1 {
		// object defined as selectable both doesn't have any data definitions found
		structure.warn(fmt.Sprintf("object<%s> visualization<%s> is expected to have data, but no measures or dimensions was found", obj.Id, vis))
	}

	// no dimension = not selectable
	if len(obj.Dimensions) < 1 {
		obj.Selectable = false
	}

	return nil
}

func (structure *AppStructure) handleMeasure(ctx context.Context, app *senseobjects.App, id string, obj *AppStructureObject) error {
	genMeasure, err := app.Doc.GetMeasure(ctx, id)
	if err != nil {
		fmt.Printf("Measure: %+v\n", genMeasure)
		return errors.WithStack(err)
	}
	obj.RawProperties, err = genMeasure.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	// Save measure information to structure
	var measure enigma.NxInlineMeasureDef
	measurePath := senseobjdef.NewDataPath("/qMeasure")
	rawMeasure, err := measurePath.Lookup(obj.RawProperties)
	if err != nil {
		structure.warn(fmt.Sprintf("measure<%s> definition not found", id))
	} else {
		if err := jsonit.Unmarshal(rawMeasure, &measure); err != nil {
			return errors.WithStack(err)
		}
	}

	// Save meta information to structure
	var meta MetaDef
	metaPath := senseobjdef.NewDataPath("/qMetaDef")
	rawMeta, err := metaPath.Lookup(obj.RawProperties)
	if err != nil {
		structure.warn(fmt.Sprintf("measure<%s> has not meta information", id))
	} else {
		if err := jsonit.Unmarshal(rawMeta, &meta); err != nil {
			return errors.WithStack(err)
		}
	}

	obj.Measures = []AppStructureMeasureMeta{
		{
			Meta:  &meta,
			Label: measure.Label,
			Def:   measure.Def,
		},
	}

	return nil
}

func (structure *AppStructure) handleDimension(ctx context.Context, app *senseobjects.App, id string, obj *AppStructureObject) error {
	genDim, err := app.Doc.GetDimension(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	obj.RawProperties, err = genDim.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	// Save dimension information to structure
	var dimension enigma.NxInlineDimensionDef
	dimensionPath := senseobjdef.NewDataPath("/qDim")
	rawDimension, err := dimensionPath.Lookup(obj.RawProperties)
	if err != nil {
		structure.warn(fmt.Sprintf("dimension<%s> defintion not found", id))
	} else {
		if err := jsonit.Unmarshal(rawDimension, &dimension); err != nil {
			return errors.WithStack(err)
		}
	}

	// Add dimension meta information to structure
	var meta MetaDef
	metaPath := senseobjdef.NewDataPath("/qMetaDef")
	rawMeta, err := metaPath.Lookup(obj.RawProperties)
	if err != nil {
		structure.warn(fmt.Sprintf("dimension<%s> has not meta information", id))
	} else {
		if err := jsonit.Unmarshal(rawMeta, &meta); err != nil {
			return errors.WithStack(err)
		}
	}

	obj.Dimensions = []AppStructureDimensionMeta{
		{
			Meta:            &meta,
			LabelExpression: dimension.LabelExpression,
			Defs:            dimension.FieldDefs,
			Labels:          dimension.FieldLabels,
		},
	}

	return nil
}

func (structure *AppStructure) handleBookmark(ctx context.Context, app *senseobjects.App, id string) error {
	bookmark, err := app.Doc.GetBookmark(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}

	properties, err := bookmark.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	var structureBookmark AppStructureBookmark
	if err := jsonit.Unmarshal(properties, &structureBookmark); err != nil {
		return errors.WithStack(err)
	}

	// Get bookmark meta information
	var meta MetaDef // meta shares title and description from this struct
	metaPath := senseobjdef.NewDataPath("/qMetaDef")
	rawMeta, err := metaPath.Lookup(properties)
	if err != nil {
		structure.warn(fmt.Sprintf("bookmark<%s> has no meta information", id))
	} else {
		if err = jsonit.Unmarshal(rawMeta, &meta); err != nil {
			structure.warn(fmt.Sprintf("bookmark<%s> failed to unmarshal meta information: %v", id, err))
		}
	}

	structureBookmark.Title = meta.Title
	structureBookmark.Description = meta.Description

	idPath := senseobjdef.NewDataPath("/qInfo/qId")
	rawId, err := idPath.Lookup(properties)
	if err != nil {
		return errors.Wrap(err, "failed to get ID of bookmark")
	}
	if err := jsonit.Unmarshal(rawId, &structureBookmark.ID); err != nil {
		return errors.Wrap(err, "failed to unmarshal ID of bookmark")
	}

	structure.AddBookmark(structureBookmark)

	return nil
}

// AddWarning to app structure report
func (report *AppStructureReport) AddWarning(warning string) {
	report.warningsLock.Lock()
	defer report.warningsLock.Unlock()

	if report.warnings == nil {
		report.warnings = make([]string, 0, 1)
	}

	report.warnings = append(report.warnings, warning)
}

// AddObject to app structure report
func (report *AppStructureReport) AddObject(obj AppObjectDef) {
	report.objectsLock.Lock()
	defer report.objectsLock.Unlock()

	if report.objects == nil {
		report.objects = make([]AppObjectDef, 0, 1)
	}

	report.objects = append(report.objects, obj)
}
