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
		// Children to the sense object
		Children []AppObjectDef `json:"children,omitempty"`
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
		Objects []AppStructureObject `json:"objects"`
		// Bookmarks list of bookmarks in the app
		Bookmarks []AppStructureBookmark `json:"bookmarks"`

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
	// TODO handle "injectable auth actions" using interface returning action array
	// TODO new action type with parameter showing "app" connect or not
	actionsToKeep := map[string]interface{}{
		"openapp": nil,
	}

	appStructureScenario := make([]scenario.Action, 0, 1)
	for _, act := range cfg.Scenario {
		if _, ok := actionsToKeep[act.Type]; ok {
			appStructureScenario = append(appStructureScenario, act)
			// todo inject GetAppStructureAction if action is action set to handle app connect
			appStructureScenario = append(appStructureScenario, scenario.Action{
				ActionCore: scenario.ActionCore{
					Type:  "GetAppStructure",
					Label: "Get app structure",
				},
				Settings: &getAppStructureSettings{},
			})
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
	// Todo use outputs dir for structure result
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

	// Todo get and set object visualization

	sessionState.QueueRequest(func(ctx context.Context) error {
		obj := AppStructureObject{
			AppObjectDef: AppObjectDef{
				Id:   id,
				Type: typ,
			},
		}

		// Todo handle auto-chart subtype

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
		default:
			if err := handleObject(ctx, sessionState, app, id, typ, &obj); err != nil {
				return errors.WithStack(err)
			}
		}

		// Todo (Dev only) comment this line to turn on seeing raw properties in file
		obj.RawProperties = nil

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
		structure.Objects = make([]AppStructureObject, 0, 1)
	}
	structure.Objects = append(structure.Objects, obj)
}

// AddBookmark to structure
func (structure *AppStructure) AddBookmark(bookmark AppStructureBookmark) {
	structure.structureLock.Lock()
	defer structure.structureLock.Unlock()
	if structure.Bookmarks == nil {
		structure.Bookmarks = make([]AppStructureBookmark, 0, 1)
	}
	structure.Bookmarks = append(structure.Bookmarks, bookmark)
}

func handleObject(ctx context.Context, sessionState *session.State, app *senseobjects.App, id, typ string, obj *AppStructureObject) error {
	genObj, err := app.Doc.GetObject(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	obj.RawProperties, err = genObj.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	childInfos, err := genObj.GetChildInfos(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	var children []AppObjectDef
	if len(childInfos) > 0 {
		children = make([]AppObjectDef, 0, len(childInfos))
	}

	for _, child := range childInfos {
		if child == nil {
			continue
		}
		children = append(children, AppObjectDef{
			Id:   child.Id,
			Type: child.Type,
		})
	}
	obj.Children = children

	// Lookup and set ExtendsID
	extendsIdPath := senseobjdef.NewDataPath("/qExtendsId")
	rawExtendsID, _ := extendsIdPath.Lookup(obj.RawProperties)
	_ = jsonit.Unmarshal(rawExtendsID, &obj.ExtendsId)

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

	obj.Selectable = def.Select != nil // Todo also check if has dimensions
	if obj.Selectable {
		// Hyper cube dimensions and measures
		dimensions := senseobjdef.NewDataPath(def.Select.Path + "/qDimensions")
		measures := senseobjdef.NewDataPath(def.Select.Path + "/qMeasures")

		// Try to set dimensions and measures, null if not exist or not parsable (error)
		rawDimensions, _ := dimensions.Lookup(obj.RawProperties)
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

		rawMeasures, _ := measures.Lookup(obj.RawProperties)
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
		rawListObject, _ := listObjects.Lookup(obj.RawProperties)
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
	} else {
		// Todo handle non "selectable" objects
	}

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

	structure.AddBookmark(structureBookmark)

	return nil
}
