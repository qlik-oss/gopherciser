package config

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/qlik-oss/gopherciser/enummap"
	"os"
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
	getAppStructureSettings struct{}

	// LibraryMetaDef meta information for Library objects such as dimension and measure
	LibraryMetaDef struct {
		// Title of library item
		Title string `json:"title"`
		// Description of library item
		Description string `json:"description"`
		// Tags of  of library item
		Tags []string `json:"tags"`
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
		Meta *LibraryMetaDef `json:"meta,omitempty"`
		// LibraryId connects measure to separately defined measure
		LibraryId string `json:"libraryId,omitempty"`
		// Label of on measure
		Label string `json:"label,omitempty"`
		// Def the actual measure definition
		Def string `json:"def,omitempty"`
	}

	AppStructureDimensionMeta struct {
		// Meta information, only included for library items
		Meta *LibraryMetaDef `json:"meta,omitempty"`
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

	// AppStructure of Sense app
	AppStructure struct {
		AppMeta AppStructureAppMeta `json:"meta"`
		// Objects in Sense app
		Objects map[string]AppStructureObject `json:"objects"`
		// Bookmarks list of bookmarks in the app
		Bookmarks map[string]AppStructureBookmark `json:"bookmarks"`

		structureLock sync.Mutex
	}

	ObjectType int
)

const (
	ObjectTypeDefault ObjectType = iota
	ObjectTypeDimension
	ObjectTypeMeasure
	ObjectTypeBookmark
	ObjectTypeMasterObject
	ObjectTypeAutoChart
)

var (
	ObjectTypeEnumMap, _ = enummap.NewEnumMap(map[string]int{
		"dimension":    int(ObjectTypeDimension),
		"measure":      int(ObjectTypeMeasure),
		"bookmark":     int(ObjectTypeBookmark),
		"masterobject": int(ObjectTypeMasterObject),
		"auto-chart":   int(ObjectTypeAutoChart),
	})
)

func (cfg *Config) getAppStructureScenario() []scenario.Action {
	return evaluateActionList(cfg.Scenario)
}

func evaluateActionList(actions []scenario.Action) []scenario.Action {
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
					Settings: &getAppStructureSettings{},
				})

			}
			if len(subActions) > 0 {
				appStructureScenario = append(appStructureScenario, evaluateActionList(subActions)...)
			}
		}
	}
	return appStructureScenario
}

// GetAppStructures for all apps in scenario
func (cfg *Config) GetAppStructures(ctx context.Context) error {
	// find all auth and actions
	appStructureScenario := cfg.getAppStructureScenario()
	if len(appStructureScenario) < 1 {
		return errors.New("no applicable actions in scenario")
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

	// Todo where to log? Override for now during development
	stmpl, err := session.NewSyncedTemplate("./logs/appstructure.tsv")
	if err != nil {
		return errors.WithStack(err)
	}
	logSettings := LogSettings{
		Traffic:        cfg.Settings.LogSettings.Traffic,
		Debug:          cfg.Settings.LogSettings.Debug,
		TrafficMetrics: false,
		FileName:       *stmpl,
		Format:         0,
		Summary:        0,
	}

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
	}

	for _, info := range allInfos {
		if info == nil {
			continue
		}
		if err := getStructureForObjectAsync(sessionState, actionState, app, info.Id, info.Type, appStructure); err != nil {
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
}

func getStructureForObjectAsync(sessionState *session.State, actionState *action.State, app *senseobjects.App, id, typ string, appStructure *AppStructure) error {
	if appStructure == nil {
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
			if err := handleDimension(ctx, app, id, &obj); err != nil {
				return errors.WithStack(err)
			}
		case ObjectTypeMeasure:
			if err := handleMeasure(ctx, app, id, &obj); err != nil {
				return errors.WithStack(err)
			}
		case ObjectTypeBookmark:
			if err := handleBookmark(ctx, app, id, appStructure); err != nil {
				return errors.WithStack(err)
			}
		case ObjectTypeAutoChart:
			if err := handleAutoChart(ctx, sessionState, app, id, &obj); err != nil {
				return errors.WithStack(err)
			}
		default:
			if err := handleDefaultObject(ctx, sessionState, app, id, typ, &obj); err != nil {
				return errors.WithStack(err)
			}
		}

		// Todo (Dev only) comment this line to turn on seeing raw properties in file
		obj.RawProperties = nil
		obj.RawExtendedProperties = nil

		appStructure.AddObject(obj)
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

func (structure *AppStructure) GetSelectables(rooObject string) ([]AppStructureObject, error) {
	rootObj, ok := structure.Objects[rooObject]
	if !ok {
		return nil, errors.New("not found") // todo make defined error
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

func handleDefaultObject(ctx context.Context, sessionState *session.State, app *senseobjects.App, id, typ string, obj *AppStructureObject) error {
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

	return errors.WithStack(handleObject(sessionState, typ, obj))
}

func handleAutoChart(ctx context.Context, sessionState *session.State, app *senseobjects.App, id string, obj *AppStructureObject) error {
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

	return errors.WithStack(handleObject(sessionState, ObjectTypeEnumMap.StringDefault(int(ObjectTypeAutoChart), "auto-chart"), obj))
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

func handleObject(sessionState *session.State, typ string, obj *AppStructureObject) error {
	// Lookup and set Visualization
	visualizationPath := senseobjdef.NewDataPath("/visualization")
	rawVisualization, _ := visualizationPath.Lookup(obj.RawProperties)
	_ = jsonit.Unmarshal(rawVisualization, &obj.Visualization)

	vis := obj.Visualization
	if vis == "" {
		vis = typ
	}

	def, err := senseobjdef.GetObjectDef(vis)
	if err != nil {
		switch errors.Cause(err).(type) {
		case senseobjdef.NoDefError:
			sessionState.LogEntry.Logf(logger.WarningLevel, "Object type<%s> not supported", vis)
			return nil
		default:
			return errors.WithStack(err)
		}
	}

	properties := obj.RawProperties
	if obj.ExtendsId != "" && obj.RawExtendedProperties != nil {
		properties = obj.RawExtendedProperties
	}

	obj.Selectable = def.Select != nil
	if obj.Selectable {
		// Hyper cube dimensions and measures
		dimensions := senseobjdef.NewDataPath(def.Select.Path + "/qDimensions")
		measures := senseobjdef.NewDataPath(def.Select.Path + "/qMeasures")

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
				// Todo error or warning here?
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

		// Handle list object
		path := def.Select.Path
		if !strings.HasSuffix(path, "/qListObjectDef") {
			path = def.Select.Path + "/qListObjectDef"
		}
		listObjects := senseobjdef.NewDataPath(path)
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

		// no dimension = not selectable
		if len(obj.Dimensions) < 1 {
			obj.Selectable = false
		}
	}
	// Todo handle non "selectable" objects
	//else {
	//
	//}

	return nil
}

func handleMeasure(ctx context.Context, app *senseobjects.App, id string, obj *AppStructureObject) error {
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
	measurePath := senseobjdef.NewDataPath("/qMeasure")
	rawMeasure, err := measurePath.Lookup(obj.RawProperties)
	if err != nil {
		return errors.WithStack(err)
	}
	var measure enigma.NxInlineMeasureDef
	if err := jsonit.Unmarshal(rawMeasure, &measure); err != nil {
		return errors.WithStack(err)
	}

	metaPath := senseobjdef.NewDataPath("/qMetaDef")
	rawMeta, err := metaPath.Lookup(obj.RawProperties)
	if err != nil {
		return errors.WithStack(err)
	}

	var meta LibraryMetaDef
	if err := jsonit.Unmarshal(rawMeta, &meta); err != nil {
		return errors.WithStack(err)
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

func handleDimension(ctx context.Context, app *senseobjects.App, id string, obj *AppStructureObject) error {
	genDim, err := app.Doc.GetDimension(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	obj.RawProperties, err = genDim.GetPropertiesRaw(ctx)
	if err != nil {
		fmt.Printf("Dim: %+v\n", genDim)
		return errors.WithStack(err)
	}

	// Save dimension information to structure
	dimensionPath := senseobjdef.NewDataPath("/qDim")
	rawDimension, err := dimensionPath.Lookup(obj.RawProperties)
	if err != nil {
		return errors.WithStack(err)
	}

	var dimension enigma.NxInlineDimensionDef
	if err := jsonit.Unmarshal(rawDimension, &dimension); err != nil {
		return errors.WithStack(err)
	}

	metaPath := senseobjdef.NewDataPath("/qMetaDef")
	rawMeta, err := metaPath.Lookup(obj.RawProperties)
	if err != nil {
		return errors.WithStack(err)
	}

	var meta LibraryMetaDef
	if err := jsonit.Unmarshal(rawMeta, &meta); err != nil {
		return errors.WithStack(err)
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

func handleBookmark(ctx context.Context, app *senseobjects.App, id string, structure *AppStructure) error {
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

	metaPath := senseobjdef.NewDataPath("/qMetaDef")
	rawMeta, _ := metaPath.Lookup(properties)
	var meta LibraryMetaDef // meta shares title and description from this struct
	_ = jsonit.Unmarshal(rawMeta, &meta)

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
