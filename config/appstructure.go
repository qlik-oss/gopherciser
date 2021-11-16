package config

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/InVisionApp/tabular"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/scenario"
	"github.com/qlik-oss/gopherciser/scheduler"
	"github.com/qlik-oss/gopherciser/senseobjdef"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/synced"
)

type (
	// AppStructureReport reports warnings and fetched objects for app structure
	AppStructureReport struct {
		warnings     []string
		warningsLock sync.Mutex
	}

	// GeneratedAppStructure of Sense app
	GeneratedAppStructure struct {
		appstructure.AppStructure

		logEntry      *logger.LogEntry
		report        AppStructureReport
		structureLock sync.Mutex
	}
)

func (cfg *Config) getAppStructureScenario(includeRaw bool, summary SummaryType) []scenario.Action {
	return evaluateActionList(cfg.Scenario, includeRaw, summary)
}

func (structure *GeneratedAppStructure) printSummary(summary SummaryType, fileName string) {
	if structure == nil || summary == SummaryTypeNone {
		return
	}

	buf := helpers.NewBuffer()
	defer buf.WriteTo(ansiWriter)
	defer buf.WriteString(ansiReset)

	buf.WriteString(ansiBoldBlue)
	buf.WriteString(fileName)
	buf.WriteString(" created with ")

	// print object count
	objectCount := len(structure.Objects)
	buf.WriteString(ansiBoldWhite)
	buf.WriteString(strconv.Itoa(objectCount))
	buf.WriteString(" objects")
	buf.WriteString(ansiBoldBlue)
	buf.WriteString(" and ")

	warningCount := len(structure.report.warnings)

	warningColor := ansiBoldBlue
	if warningCount > 0 {
		warningColor = ansiBoldYellow
	}

	buf.WriteString(warningColor)
	buf.WriteString(strconv.Itoa(warningCount))
	buf.WriteString(" warning")
	if warningCount != 1 {
		buf.WriteString("s")
	}
	buf.WriteString(ansiBoldBlue)
	buf.WriteString(" found")

	if warningCount > 0 {
		buf.WriteString(":\n")
	} else {
		buf.WriteString("\n")
	}

	for _, warning := range structure.report.warnings {
		buf.WriteString(ansiBoldYellow)
		buf.WriteString(warning)
		buf.WriteString("\n")
	}

	if summary < SummaryTypeExtended {
		return
	}

	buf.WriteString(ansiBoldBlue)
	buf.WriteString("\n")

	// object table
	tabbedOutput := tabular.New()
	summaryHeaders := make(SummaryHeader)
	//objectTblData := make([]SummaryActionDataEntry, 0, objectCount)

	// Create headers and default column sizes
	summaryHeaders["id"] = &SummaryHeaderEntry{"ID", 2}
	summaryHeaders["vis"] = &SummaryHeaderEntry{"Visualization", 13}
	summaryHeaders["typ"] = &SummaryHeaderEntry{"Type", 4}

	// Update column widths
	for _, obj := range structure.Objects {
		summaryHeaders["id"].UpdateColSize(len(obj.Id))
		summaryHeaders["vis"].UpdateColSize(len(obj.Visualization))
		summaryHeaders["typ"].UpdateColSize(len(obj.Type))
	}

	// Set column widths
	for k := range summaryHeaders {
		summaryHeaders.Col(k, &tabbedOutput)
	}

	// Print table headers
	table := tabbedOutput.Parse("*")
	writeTableHeaders(buf, &table)

	// print all objects
	for _, obj := range structure.Objects {
		buf.WriteString(ansiBoldBlue)
		buf.WriteString(fmt.Sprintf(table.Format, obj.Id, obj.Visualization, obj.Type))
		buf.WriteString(ansiReset)
	}
}

func evaluateActionList(actions []scenario.Action, includeRaw bool, summary SummaryType) []scenario.Action {
	appStructureScenario := make([]scenario.Action, 0, len(actions))
	for _, act := range actions {
		if act.Disabled {
			continue
		}
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
		return appstructure.AppStructureNoScenarioActionsError{}
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

	if _, err := cfg.Scheduler.Validate(); err != nil {
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
	stmpl, err := synced.New(appStructureLogPath)
	if err != nil {
		return errors.WithStack(err)
	}
	logSettings.FileName = *stmpl

	log, err := setupLogging(ctx, logSettings, nil, nil, &cfg.Counters)
	if err != nil {
		return errors.WithStack(err)
	}

	logEntry := log.NewLogEntry()
	logEntry.Log(logger.DebugLevel, fmt.Sprintf("outputs folder: %s", outputsDir))

	timeout := time.Duration(cfg.Settings.Timeout) * time.Second
	if err := cfg.Scheduler.Execute(ctx, log, timeout, appStructureScenario, outputsDir, cfg.LoginSettings, &cfg.ConnectionSettings, &cfg.Counters); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (structure *GeneratedAppStructure) getStructureForObjectAsync(sessionState *session.State, actionState *action.State, app *senseobjects.App, id, typ string, includeRaw bool) error {
	if structure == nil {
		return errors.New("appStructure object is nil")
	}

	sessionState.QueueRequest(func(ctx context.Context) error {
		obj := appstructure.AppStructureObject{
			AppObjectDef: appstructure.AppObjectDef{
				Id:   id,
				Type: typ,
			},
		}

		objectType := appstructure.ObjectTypeDefault
		if oType, err := appstructure.ObjectTypeEnumMap.Int(typ); err == nil {
			objectType = appstructure.ObjectType(oType)
		}

		sessionState.LogEntry.Log(logger.DebugLevel, fmt.Sprintf("get structure for object id<%s> type<%s>", id, typ))

		// handle some special types
		switch objectType {
		case appstructure.ObjectTypeDimension:
			if err := structure.handleDimension(ctx, app, id, &obj); err != nil {
				return errors.WithStack(err)
			}
		case appstructure.ObjectTypeMeasure:
			if err := structure.handleMeasure(ctx, app, id, &obj); err != nil {
				return errors.WithStack(err)
			}
		case appstructure.ObjectTypeBookmark:
			if err := structure.handleBookmark(ctx, app, id, includeRaw); err != nil {
				return errors.WithStack(err)
			}
			return nil
		case appstructure.ObjectTypeAutoChart:
			if err := structure.handleAutoChart(ctx, app, id, &obj); err != nil {
				return errors.WithStack(err)
			}
		case appstructure.ObjectEmbeddedSnapshot, appstructure.ObjectSnapshotList, appstructure.ObjectSnapshot:
			structure.handleSnapshots(id, typ)
			return nil
		case appstructure.ObjectStory, appstructure.ObjectSlide, appstructure.ObjectSlideItem:
			structure.handleStories(ctx, app, id, typ, includeRaw)
			return nil
		case appstructure.ObjectAlertBookmark, appstructure.ObjectHiddenBookmark:
			// ignore alert and hidden bookmarks
		default:
			if err := structure.handleDefaultObject(ctx, app, id, typ, &obj); err != nil {
				return errors.Wrapf(err, "id<%s> type<%s>", id, typ)
			}
		}

		if !includeRaw {
			// Remove raw properties from structure output
			obj.RawBaseProperties = nil
			obj.RawExtendedProperties = nil
			obj.RawGeneratedProperties = nil
		}

		structure.AddObject(obj)
		return nil
	}, actionState, true, "")

	return nil
}

// AddObject to structure
func (structure *GeneratedAppStructure) AddObject(obj appstructure.AppStructureObject) {
	structure.structureLock.Lock()
	defer structure.structureLock.Unlock()
	if structure.Objects == nil {
		structure.Objects = make(map[string]appstructure.AppStructureObject)
	}
	structure.Objects[obj.Id] = obj
}

// AddBookmark to structure
func (structure *GeneratedAppStructure) AddBookmark(bookmark appstructure.AppStructureBookmark) {
	structure.structureLock.Lock()
	defer structure.structureLock.Unlock()
	if structure.Bookmarks == nil {
		structure.Bookmarks = make(map[string]appstructure.AppStructureBookmark)
	}
	structure.Bookmarks[bookmark.ID] = bookmark
}

// AddStoryObject to structure
func (structure *GeneratedAppStructure) AddStoryObject(object appstructure.AppStructureStoryObject) {
	structure.structureLock.Lock()
	defer structure.structureLock.Unlock()

	if structure.StoryObjects == nil {
		structure.StoryObjects = make(map[string]appstructure.AppStructureStoryObject)
	}
	structure.StoryObjects[object.Id] = object
}

func (structure *GeneratedAppStructure) warn(warning string) {
	structure.report.AddWarning(warning)
	if structure.logEntry != nil {
		structure.logEntry.Log(logger.WarningLevel, warning)
	}
}

func (structure *GeneratedAppStructure) handleDefaultObject(ctx context.Context, app *senseobjects.App, id, typ string, obj *appstructure.AppStructureObject) error {
	genObj, err := structure.getObject(ctx, app, id, typ)
	if err != nil {
		return errors.WithStack(err)
	}

	obj.RawBaseProperties, err = genObj.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	// Lookup and set ExtendsID
	extendsIdPath := helpers.NewDataPath("/qExtendsId")
	rawExtendsID, _ := extendsIdPath.Lookup(obj.RawBaseProperties)
	_ = jsonit.Unmarshal(rawExtendsID, &obj.ExtendsId)

	if obj.ExtendsId != "" {
		extendedObject, err := structure.getObject(ctx, app, obj.ExtendsId, obj.Type)
		if err != nil {
			return errors.WithStack(err)
		}
		obj.RawExtendedProperties, err = extendedObject.GetPropertiesRaw(ctx)
		if err != nil {
			return errors.WithStack(err)
		}

		obj.RawGeneratedProperties = extractGeneratedProperties(obj.RawExtendedProperties)

		if err := handleChildren(ctx, extendedObject, &obj.AppStructureObjectChildren); err != nil {
			return errors.WithStack(err)
		}
	} else {
		if err := handleChildren(ctx, genObj, &obj.AppStructureObjectChildren); err != nil {
			return errors.WithStack(err)
		}
	}

	return errors.WithStack(structure.handleObject(typ, obj))
}

func (structure *GeneratedAppStructure) handleAutoChart(ctx context.Context, app *senseobjects.App, id string, obj *appstructure.AppStructureObject) error {
	genObj, err := structure.getObject(ctx, app, id, "auto-chart")
	if err != nil {
		return errors.WithStack(err)
	}

	obj.RawBaseProperties, err = genObj.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	obj.RawGeneratedProperties = extractGeneratedProperties(obj.RawBaseProperties)

	if err := handleChildren(ctx, genObj, &obj.AppStructureObjectChildren); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(structure.handleObject(appstructure.ObjectTypeEnumMap.StringDefault(int(appstructure.ObjectTypeAutoChart), "auto-chart"), obj))
}

func extractGeneratedProperties(properties json.RawMessage) json.RawMessage {
	generatedPropertiesPath := helpers.NewDataPath("/qUndoExclude/generated")
	properties, _ = generatedPropertiesPath.Lookup(properties)
	return properties
}

func handleChildren(ctx context.Context, genObj *enigma.GenericObject, children *appstructure.AppStructureObjectChildren) error {
	childInfos, err := genObj.GetChildInfos(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	for _, child := range childInfos {
		if child == nil {
			continue
		}
		if children.Map == nil {
			children.Map = make(map[string]string)
		}
		children.Map[child.Id] = child.Type
	}

	return nil
}

func (structure *GeneratedAppStructure) handleObject(typ string, obj *appstructure.AppStructureObject) error {
	// figure out which properties to use
	var properties json.RawMessage
	if obj.RawGeneratedProperties != nil {
		properties = obj.RawGeneratedProperties
	} else if obj.RawExtendedProperties != nil {
		properties = obj.RawExtendedProperties
	} else {
		properties = obj.RawBaseProperties
	}

	// Lookup and set Visualization
	visualizationPath := helpers.NewDataPath("/visualization")
	rawVisualization, _ := visualizationPath.Lookup(properties)
	_ = jsonit.Unmarshal(rawVisualization, &obj.Visualization)

	vis := obj.Visualization
	if vis == "" {
		vis = typ
	}

	metaDef := helpers.NewDataPath("/qMetaDef")
	rawMetaDef, _ := metaDef.Lookup(properties)
	_ = jsonit.Unmarshal(rawMetaDef, &obj.MetaDef)

	enumTyp, _ := appstructure.ObjectTypeEnumMap.Int(typ) // 0 will be default in case of "error" == ObjectTypeDefault

	// Should we look for measures and dimensions?
	switch appstructure.ObjectType(enumTyp) {
	case appstructure.ObjectSheet,
		appstructure.ObjectAppprops,
		appstructure.ObjectLoadModel:
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

	// Paths dimensions and measures in hypercube
	dimensions := helpers.NewDataPath(fmt.Sprintf("%sDef/qDimensions", def.DataDef.Path))
	measures := helpers.NewDataPath(fmt.Sprintf("%sDef/qMeasures", def.DataDef.Path))

	// Figure out list object path
	listObjectPath := fmt.Sprintf("%sDef", def.DataDef.Path)
	if !strings.HasSuffix(listObjectPath, "/qListObjectDef") {
		listObjectPath = fmt.Sprintf("%s/qListObjectDef", def.DataDef.Path)
	}
	listObjects := helpers.NewDataPath(listObjectPath)

	// Try to set dimensions and measures, null if not exist or not parsable (error)
	rawDimensions, _ := dimensions.Lookup(properties)
	if rawDimensions != nil {
		var dimensions []*enigma.NxDimension
		if err := jsonit.Unmarshal(rawDimensions, &dimensions); err != nil {
			return errors.WithStack(err)
		}
		obj.Dimensions = make([]appstructure.AppStructureDimensionMeta, 0, len(dimensions))
		for _, dimension := range dimensions {
			if dimension == nil {
				continue
			}

			obj.Dimensions = append(obj.Dimensions, appstructure.AppStructureDimensionMeta{
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
		obj.Measures = make([]appstructure.AppStructureMeasureMeta, 0, len(measures))
		for _, measure := range measures {
			if measure == nil {
				continue
			}
			obj.Measures = append(obj.Measures, appstructure.AppStructureMeasureMeta{
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
		obj.Dimensions = []appstructure.AppStructureDimensionMeta{
			{
				LibraryId:       listObject.LibraryId,
				LabelExpression: listObject.Def.LabelExpression,
				Defs:            listObject.Def.FieldDefs,
				Labels:          listObject.Def.FieldLabels,
			},
		}

		if obj.Measures == nil {
			obj.Measures = make([]appstructure.AppStructureMeasureMeta, 0, len(listObject.Expressions))
		}

		for _, expression := range listObject.Expressions {
			obj.Measures = append(obj.Measures, appstructure.AppStructureMeasureMeta{
				LibraryId: expression.LibraryId,
				Def:       expression.Expr,
			})
		}
	}

	if obj.Selectable && (len(obj.Dimensions)+len(obj.Measures)) < 1 {
		// object defined as selectable both doesn't have any data definitions found
		structure.warn(fmt.Sprintf("object<%s> visualization<%s> type<%s> is expected to have data, but no measures or dimensions were found", obj.Id, vis, typ))
	}

	// no dimension = not selectable
	if len(obj.Dimensions) < 1 {
		obj.Selectable = false
	}

	resolveTitle(obj, properties, []string{
		"/title",
		fmt.Sprintf("%sDef/qTitle", def.DataDef.Path),
	})

	return nil
}

func (structure *GeneratedAppStructure) handleMeasure(ctx context.Context, app *senseobjects.App, id string, obj *appstructure.AppStructureObject) error {
	genMeasure, err := app.Doc.GetMeasure(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	obj.RawBaseProperties, err = genMeasure.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	// Save measure information to structure
	var measure enigma.NxInlineMeasureDef
	measurePath := helpers.NewDataPath("/qMeasure")
	rawMeasure, err := measurePath.Lookup(obj.RawBaseProperties)
	if err != nil {
		structure.warn(fmt.Sprintf("measure<%s> definition not found", id))
	} else {
		if err := jsonit.Unmarshal(rawMeasure, &measure); err != nil {
			return errors.WithStack(err)
		}
	}

	// Save meta information to structure
	var meta appstructure.MetaDef
	metaPath := helpers.NewDataPath("/qMetaDef")
	rawMeta, err := metaPath.Lookup(obj.RawBaseProperties)
	if err != nil {
		structure.warn(fmt.Sprintf("measure<%s> has not meta information", id))
	} else {
		if err := jsonit.Unmarshal(rawMeta, &meta); err != nil {
			return errors.WithStack(err)
		}
	}

	obj.Measures = []appstructure.AppStructureMeasureMeta{
		{
			Meta:  &meta,
			Label: measure.Label,
			Def:   measure.Def,
		},
	}

	return nil
}

func (structure *GeneratedAppStructure) handleDimension(ctx context.Context, app *senseobjects.App, id string, obj *appstructure.AppStructureObject) error {
	genDim, err := app.Doc.GetDimension(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}
	obj.RawBaseProperties, err = genDim.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	// Save dimension information to structure
	var dimension enigma.NxInlineDimensionDef
	dimensionPath := helpers.NewDataPath("/qDim")
	rawDimension, err := dimensionPath.Lookup(obj.RawBaseProperties)
	if err != nil {
		structure.warn(fmt.Sprintf("dimension<%s> defintion not found", id))
	} else {
		if err := jsonit.Unmarshal(rawDimension, &dimension); err != nil {
			return errors.WithStack(err)
		}
	}

	// Add dimension meta information to structure
	var meta appstructure.MetaDef
	metaPath := helpers.NewDataPath("/qMetaDef")
	rawMeta, err := metaPath.Lookup(obj.RawBaseProperties)
	if err != nil {
		structure.warn(fmt.Sprintf("dimension<%s> has not meta information", id))
	} else {
		if err := jsonit.Unmarshal(rawMeta, &meta); err != nil {
			return errors.WithStack(err)
		}
	}

	obj.Dimensions = []appstructure.AppStructureDimensionMeta{
		{
			Meta:            &meta,
			LabelExpression: dimension.LabelExpression,
			Defs:            dimension.FieldDefs,
			Labels:          dimension.FieldLabels,
		},
	}

	return nil
}

func (structure *GeneratedAppStructure) handleSnapshots(id, typ string) {
	storyObject := appstructure.AppStructureStoryObject{
		AppObjectDef: appstructure.AppObjectDef{
			Id:   id,
			Type: typ,
		},
	}

	structure.AddStoryObject(storyObject)
}

func (structure *GeneratedAppStructure) handleStories(ctx context.Context, app *senseobjects.App, id, typ string, includeRaw bool) {
	storyObject := appstructure.AppStructureStoryObject{
		AppObjectDef: appstructure.AppObjectDef{
			Id:   id,
			Type: typ,
		},
	}

	// Add what we have on point of return, since we only warn for these types of objects
	defer func() {
		structure.AddStoryObject(storyObject)
	}()

	obj, err := structure.getObject(ctx, app, id, typ)
	if err != nil {
		structure.warn(err.Error())
		return
	}

	storyObject.RawProperties, err = obj.GetPropertiesRaw(ctx)
	if err != nil {
		structure.warn(fmt.Sprintf("id<%s> type<%s> failed to return properties error<%s>", id, typ, err))
		return
	}
	defer func() {
		if !includeRaw {
			storyObject.RawProperties = nil
			storyObject.RawSnapShotProperties = nil
		}
	}()

	// Lookup and set Visualization
	visualizationPath := helpers.NewDataPath("/visualization")
	rawVisualization, _ := visualizationPath.Lookup(storyObject.RawProperties)
	_ = jsonit.Unmarshal(rawVisualization, &storyObject.Visualization)

	if err := handleChildren(ctx, obj, &storyObject.AppStructureObjectChildren); err != nil {
		structure.warn(fmt.Sprintf("id<%s> type<%s> failed to get object children error<%s>", id, typ, err))
		return
	}

	if storyObject.Visualization == appstructure.ObjectTypeEnumMap.StringDefault(int(appstructure.ObjectSnapshot), "snapshot") {
		snapShotObj, err := obj.GetSnapshotObject(ctx)
		if err != nil {
			structure.warn(fmt.Sprintf("id<%s> type<%s> failed to get connected snapshot object", id, typ))
			return
		}
		storyObject.SnapshotID = snapShotObj.GenericId
		storyObject.RawSnapShotProperties, err = snapShotObj.GetPropertiesRaw(ctx)
		if err != nil {
			structure.warn(fmt.Sprintf("id<%s> type<%s> failed to get snapshot properties", snapShotObj.GenericId, snapShotObj.GenericType))
			return
		}

		// Lookup and set Visualization
		rawVisualization, _ := visualizationPath.Lookup(storyObject.RawSnapShotProperties)
		_ = jsonit.Unmarshal(rawVisualization, &storyObject.Visualization)
	}
}

func (structure *GeneratedAppStructure) handleBookmark(ctx context.Context, app *senseobjects.App, id string, includeRaw bool) error {
	bookmark, err := app.Doc.GetBookmark(ctx, id)
	if err != nil {
		return errors.WithStack(err)
	}

	properties, err := bookmark.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	var structureBookmark appstructure.AppStructureBookmark
	if err := jsonit.Unmarshal(properties, &structureBookmark); err != nil {
		return errors.WithStack(err)
	}

	// Get bookmark meta information
	var meta appstructure.MetaDef // meta shares title and description from this struct
	metaPath := helpers.NewDataPath("/qMetaDef")
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
	if includeRaw {
		structureBookmark.RawProperties = properties
	}

	idPath := helpers.NewDataPath("/qInfo/qId")
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

func (structure *GeneratedAppStructure) getFieldListAsync(sessionState *session.State, actionState *action.State, app *senseobjects.App) {
	// Create fieldlist object and handle fields
	sessionState.QueueRequest(func(ctx context.Context) error {
		fieldlist, err := senseobjects.CreateFieldListObject(ctx, app.Doc)
		if err != nil {
			return err
		}

		sessionState.QueueRequestWithCallback(fieldlist.UpdateLayout, actionState, true, "", func(err error) {
			properties := fieldlist.Layout()
			if properties == nil {
				actionState.AddErrors(errors.New("fieldlist layout is nil"))
				return
			}
			if properties.FieldList == nil {
				actionState.AddErrors(errors.New("FieldList missing from fieldlist layout"))
				return
			}
			for _, field := range properties.FieldList.Items {
				structure.addField(field)
			}
		})
		return nil
	}, actionState, true, "")
}

func (structure *GeneratedAppStructure) addField(field *enigma.NxFieldDescription) {
	if field == nil {
		return
	}
	structure.structureLock.Lock()
	defer structure.structureLock.Unlock()

	if structure.Fields == nil {
		structure.Fields = make(map[string]appstructure.AppStructureField)
	}
	structure.Fields[field.Name] = appstructure.AppStructureField{
		NxFieldDescription: *field,
	}
}

func (structure *GeneratedAppStructure) getObject(ctx context.Context, app *senseobjects.App, id, typ string) (*enigma.GenericObject, error) {
	obj, err := app.Doc.GetObject(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "id<%s> type<%s> failed to return object", id, typ)
	}

	if obj.Handle < 1 {
		return obj, errors.Wrapf(err, "GetObject id<%s> type<%s> returned object with handle<%d>", id, typ, obj.Handle)
	}
	return obj, nil
}

func (structure *GeneratedAppStructure) addSheetMeta(layout *senseobjects.SheetListLayout) error {
	objectSheet, err := appstructure.ObjectTypeEnumMap.String(int(appstructure.ObjectSheet))
	if err != nil {
		return err
	}
	for key, object := range structure.Objects {
		if object.Type == objectSheet {
			for _, item := range layout.AppObjectList.Items {
				if item.Info.Id == object.Id {
					object.SheetObjectMeta = &appstructure.SheetObjectMeta{
						Published: item.Meta.Published,
						Approved:  item.Meta.Approved,
					}
					structure.Objects[key] = object
				}
			}
			if object.SheetObjectMeta == nil {
				return errors.Errorf("sheet not in sheetlist: <%s>", object.Id)
			}
		}
	}
	return nil
}

func resolveTitle(obj *appstructure.AppStructureObject, properties json.RawMessage, paths []string) {
	if obj.MetaDef.Title != "" {
		return // We already have a title
	}

	for _, path := range paths {
		title := stringFromDataPath(path, properties)
		if title != "" {
			obj.MetaDef.Title = title
			return
		}
	}
}

func stringFromDataPath(path string, data json.RawMessage) string {
	dataPath := helpers.NewDataPath(path)
	rawData, _ := dataPath.Lookup(data)
	var str string
	_ = jsonit.Unmarshal(rawData, &str)
	return str
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
