package session

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/goccy/go-json"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/globals/constant"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjdef"
)

type (
	ObjectHandler interface {
		Instance(id string) ObjectHandlerInstance
	}

	objectHandlerMap struct {
		m         map[string]ObjectHandler
		writeLock sync.Mutex
	}

	CalcEvalConditionFailedError string

	HyperCubeSection int

	NxValidationError struct {
		Err     enigma.NxValidationError
		Id      string
		Path    string
		Section HyperCubeSection
	}
)

const (
	maxNbrLines = 12
	maxNbrTicks = 300
)

const (
	HyperCubeSectionRoot HyperCubeSection = iota
	HyperCubeSectionMeasure
	HyperCubeSectionDimension
	HyperCubeSectionMeasureMinichart
)

var (
	GlobalObjectHandler objectHandlerMap
)

func init() {
	// Register default object handlers
	if err := GlobalObjectHandler.RegisterHandler("auto-chart", &AutoChartHandler{}, false); err != nil {
		panic(err)
	}
	if err := GlobalObjectHandler.RegisterHandler("container", &ContainerHandler{}, false); err != nil {
		panic(err)
	}
	if err := GlobalObjectHandler.RegisterHandler("sn-layout-container", &LayoutContainerHandler{}, false); err != nil {
		panic(err)
	}
	if err := GlobalObjectHandler.RegisterHandler("masterobject", &MasterObjectHandler{}, false); err != nil {
		panic(err)
	}
	if err := GlobalObjectHandler.RegisterHandler("sn-tabbed-container", &TabbedContainerHandler{}, false); err != nil {
		panic(err)
	}
	if err := GlobalObjectHandler.RegisterHandler("sn-nlg-chart", &NarrativesHandler{}, false); err != nil {
		panic(err)
	}
}

func (err NxValidationError) Error() string {
	return fmt.Sprintf("object<%s> has %s error<%s> ExtendMessage<%s>", err.Id, err.Path, EngineCodeToString(err.Err.ErrorCode), err.Err.ExtendedMessage)
}

// Implements engima.Error interface
func (err NxValidationError) Code() int {
	return err.Err.ErrorCode
}

// Implements engima.Error interface
func (err NxValidationError) Parameter() string {
	return ""
}

// Implements engima.Error interface
func (err NxValidationError) Message() string {
	return err.Err.ExtendedMessage
}

// Error implements error interface
func (err CalcEvalConditionFailedError) Error() string {
	return fmt.Sprintf("object has unsatisfied calculation condition in %s", string(err))
}

// RegisterHandler for object type, override existing handler with override flag
func (objects *objectHandlerMap) RegisterHandler(objectType string, handler ObjectHandler, override bool) error {
	objects.writeLock.Lock()
	defer objects.writeLock.Unlock()

	if objects.m == nil {
		objects.m = make(map[string]ObjectHandler)
	}

	// Does a handler exit?
	_, exists := objects.m[objectType]
	if exists && !override {
		return errors.New(fmt.Sprintf("object type<%s> already has a handler registered", objectType))
	}

	objects.m[objectType] = handler
	return nil
}

// GetObjectHandler for objectType
func (objects *objectHandlerMap) GetObjectHandler(objectType string) ObjectHandler {
	handler, ok := objects.m[objectType]
	if ok {
		return handler
	}
	return &DefaultHandler{}
}

// GetAndAddObjectAsync get and add object to object handling
func GetAndAddObjectAsync(sessionState *State, actionState *action.State, name string) {
	getAndAddObjectWithCallback(sessionState, actionState, name, func() {})
}

// GetAndAddObjectSync get and add object to object handling
func GetAndAddObjectSync(sessionState *State, actionState *action.State, name string) {
	var wg sync.WaitGroup
	wg.Add(1)
	getAndAddObjectWithCallback(sessionState, actionState, name, func() {
		wg.Done()
	})
	wg.Wait()
}

func getAndAddObjectWithCallback(sessionState *State, actionState *action.State, name string, callback func()) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		defer callback()
		sense := sessionState.Connection.Sense()

		var genObj *enigma.GenericObject
		getObject := func(ctx context.Context) error {
			var err error
			genObj, err = sense.CurrentApp.Doc.GetObject(ctx, name)
			return err
		}
		if err := sessionState.SendRequest(actionState, getObject); err != nil {
			return errors.Wrapf(err, "Failed go get object<%s>", name)
		}

		obj, err := sense.AddNewObject(genObj.Handle, enigmahandlers.ObjTypeGenericObject, name, genObj)
		if err != nil {
			return errors.Wrapf(err, "Failed to add object<%s> to object list", name)
		}

		objInstance := sessionState.GetObjectHandlerInstance(genObj.GenericId, genObj.GenericType)
		objInstance.SetObjectAndEvents(sessionState, actionState, obj, genObj)

		return nil
	}, actionState, true, fmt.Sprintf("Failed to get object<%s>", name))
}

func GetObjectLayout(sessionState *State, actionState *action.State, obj *enigmahandlers.Object, def *senseobjdef.ObjectDef) error {
	enigmaObject, ok := obj.EnigmaObject.(*enigma.GenericObject)
	if !ok {
		return errors.Errorf("Failed to cast object<%s> to *enigma.GenericObject", obj.ID)
	}

	sessionState.LogEntry.LogDebugf("Getting layout for object<%s> handle<%d> type<%s> START", obj.ID, obj.Handle, enigmaObject.GenericType)

	rawLayout, layoutErr := sessionState.SendRequestRaw(actionState, enigmaObject.GetLayoutRaw)
	if layoutErr != nil {
		return errors.Wrapf(layoutErr, "object<%s>.GetLayout", enigmaObject.GenericId)
	}

	//TODO Investigate performance impact of datapath lookup and optimize!
	if err := SetChildList(rawLayout, obj); err != nil {
		return errors.Wrapf(err, "failed to get childlist for object<%s> type<%s>", obj.ID, enigmaObject.GenericType)
	}
	if err := SetChildren(rawLayout, obj); err != nil {
		return errors.Wrapf(err, "failed to get children for object<%s> type<%s>", obj.ID, enigmaObject.GenericType)
	}

	if def == nil {
		var err error
		def, err = senseobjdef.GetObjectDef(enigmaObject.GenericType)
		if err != nil {
			switch errors.Cause(err).(type) {
			case senseobjdef.NoDefError:
				sessionState.LogEntry.Logf(logger.WarningLevel, "Get Data for object<%s> type<%s> not supported", enigmaObject.GenericId, enigmaObject.GenericType)
				return nil
			default:
				return errors.WithStack(err)
			}
		}
	}

	sessionState.LogEntry.LogDebugf("object<%s> objectdef<%+v>", obj.ID, def)
	if def.DataDef.Type == senseobjdef.DataDefUnknown {
		sessionState.LogEntry.Logf(logger.WarningLevel,
			"object<%s> type<%s> has unknown data carrier, please add definition to config",
			obj.ID, enigmaObject.GenericType)
		return nil
	}

	if err := SetObjectData(sessionState, actionState, rawLayout, def, obj, enigmaObject); err != nil {
		return errors.WithStack(err)
	}

	sessionState.LogEntry.LogDebugf("Getting layout for object<%s> handle<%d> type<%s> END", obj.ID, obj.Handle, enigmaObject.GenericType)

	return nil
}

func DefaultSetObjectDataAndEvents(sessionState *State, actionState *action.State, obj *enigmahandlers.Object, genObj *enigma.GenericObject, def *senseobjdef.ObjectDef) {
	var wg sync.WaitGroup

	wg.Add(1)
	sessionState.QueueRequest(func(ctx context.Context) error {
		defer wg.Done()
		return GetObjectProperties(sessionState, actionState, obj)
	}, actionState, true, "")

	wg.Add(1)
	sessionState.QueueRequest(func(ctx context.Context) error {
		defer wg.Done()
		return GetObjectLayout(sessionState, actionState, obj, def)
	}, actionState, true, "")

	wg.Wait()

	properties := obj.Properties()
	if properties != nil && properties.ExtendsId != "" && properties.Info != nil && properties.Info.Type == "listbox" {
		// Special handling of objects wrapping listbox objects due to changes for listboxes belonging to filterpanes
		if err := sessionState.IDMap.Replace(properties.ExtendsId, obj.ID, sessionState.LogEntry); err != nil {
			sessionState.LogEntry.LogDetail(logger.WarningLevel, fmt.Sprintf("error adding id<%s> to IDMap", properties.ExtendsId), err.Error())
		}
	}

	event := func(ctx context.Context, as *action.State) error {
		return GetObjectLayout(sessionState, as, obj, def)
	}
	sessionState.RegisterEvent(genObj.Handle, event, nil, true)

	children := obj.ChildList()
	childListItems := make(map[string]interface{})
	if children != nil && children.Items != nil {
		sessionState.LogEntry.LogDebugf("object<%s> type<%s> has children", genObj.GenericId, genObj.GenericType)
		for _, child := range children.Items {
			sessionState.LogEntry.LogDebug(fmt.Sprintf("obj<%s> child<%s> found in ChildList", obj.ID, child.Info.Id))
			childListItems[child.Info.Id] = nil
			GetAndAddObjectAsync(sessionState, actionState, child.Info.Id)
		}
	}

	if genObj.GenericType == "sheet" {
		sessionState.QueueRequest(func(ctx context.Context) error {
			sheetList, err := sessionState.Connection.Sense().CurrentApp.GetSheetList(sessionState, actionState)
			if err != nil {
				return errors.WithStack(err)
			}
			if sheetList != nil {
				entry, err := sheetList.GetSheetEntry(genObj.GenericId)
				if err != nil {
					return errors.WithStack(err)
				}
				if entry != nil && entry.Data != nil {
					for _, cell := range entry.Data.Cells {
						if _, ok := childListItems[cell.Name]; !ok {
							// Todo should this be a warning?
							sessionState.LogEntry.LogDebug(fmt.Sprintf("cell<%s> missing from sheet<%s> childlist", cell.Name, genObj.GenericId))
							GetAndAddObjectAsync(sessionState, actionState, cell.Name)
						}
					}
				}
			}
			return nil
		}, actionState, true, "")
	}
}

func SetChildList(rawLayout json.RawMessage, obj *enigmahandlers.Object) error {
	childDataPath := helpers.NewDataPath("qChildList")

	rawChildren, err := childDataPath.Lookup(rawLayout)
	switch errors.Cause(err).(type) {
	case helpers.NoDataFound:
		return nil //object has no children
	case nil:
		//object has children
	default:
		return errors.Wrap(err, "error getting childlist")
	}

	var children enigma.ChildList
	if errUnMarshal := json.Unmarshal(rawChildren, &children); errUnMarshal != nil {
		return errors.Wrapf(errUnMarshal, "failed to unmarshal childlist")
	}

	obj.SetChildList(&children)
	return nil
}

func SetChildren(rawLayout json.RawMessage, obj *enigmahandlers.Object) error {
	childDataPath := helpers.NewDataPath("children")

	rawChildren, err := childDataPath.Lookup(rawLayout)
	switch errors.Cause(err).(type) {
	case helpers.NoDataFound:
		return nil //object has no children
	case nil:
		//object has children
	default:
		return errors.Wrap(err, "error getting children")
	}

	var children []enigmahandlers.ObjChild
	if errUnMarshal := json.Unmarshal(rawChildren, &children); errUnMarshal != nil {
		return errors.Wrapf(errUnMarshal, "failed to unmarshal children")
	}

	obj.SetChildren(&children)
	return nil
}

func SetListObject(rawLayout json.RawMessage, obj *enigmahandlers.Object, path helpers.DataPath) error {
	rawListObject, err := path.Lookup(rawLayout)
	if err != nil {
		return errors.Wrap(err, "error getting listObject")
	}

	var listObject *enigma.ListObject
	if err = json.Unmarshal(rawListObject, &listObject); err != nil {
		return errors.Wrap(err, "Failed to unmarshal listObject from layout subtree")
	}

	obj.SetListObject(listObject)
	return nil
}

func SetHyperCube(sessionState *State, actionState *action.State, rawLayout json.RawMessage, obj *enigmahandlers.Object, path helpers.DataPath) error {
	rawHyperCube, err := path.Lookup(rawLayout)
	if err != nil {
		return errors.Wrap(err, "error getting hypercube")
	}

	var hyperCube *enigma.HyperCube
	if err = json.Unmarshal(rawHyperCube, &hyperCube); err != nil {
		return errors.Wrap(err, "Failed to unmarshal hypercube from layout subtree")
	}

	obj.SetHyperCube(hyperCube)

	// Look for cyclic dimensions and add to app sessionobjects
	if len(hyperCube.DimensionInfo) > 0 {
		for i, dim := range hyperCube.DimensionInfo {
			if dim != nil && dim.Grouping == constant.NxDimensionInfoGroupingCollection {
				app, err := sessionState.CurrentSenseApp()
				if err != nil {
					return errors.WithStack(err)
				}
				if dim.LibraryId == "" {
					sessionState.LogEntry.Logf(logger.WarningLevel, "object<%s> dim<%d> has grouping<C>, but no library ID", obj.ID, i)
					continue
				}
				// GetDimension (adds it to sessionobjects list)
				if _, err = app.GetDimension(sessionState, actionState, dim.LibraryId); err != nil {
					actionState.AddErrors(err)
				}
			}
		}
	}

	return nil
}

func GetObjectProperties(sessionState *State, actionState *action.State, obj *enigmahandlers.Object) error {
	enigmaObject, ok := obj.EnigmaObject.(*enigma.GenericObject)
	if !ok {
		return errors.Errorf("Failed to cast object<%s> to *enigma.GenericObject", obj.ID)
	}

	//Get object properties
	getProperties := func(ctx context.Context) error {
		properties, err := enigmaObject.GetEffectiveProperties(ctx)
		if err != nil {
			return errors.Wrapf(err, "object<%s>.GetEffectiveProperties failed", obj.ID)
		}
		obj.SetProperties(properties)
		return nil
	}

	return sessionState.SendRequest(actionState, getProperties)
}

// UpdateObjectHyperCubeDataAsync send get straight hypercube data and update saved hypercube
func UpdateObjectHyperCubeDataAsync(sessionState *State, actionState *action.State, gob *enigma.GenericObject,
	obj *enigmahandlers.Object, requestDef senseobjdef.GetDataRequests, columns bool) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		sessionState.LogEntry.LogDebugf("Get hyper cube data for object<%s>", gob.GenericId)
		hypercube := obj.HyperCube()
		if hypercube == nil {
			return errors.Errorf("object<%s> has no hypercube", gob.GenericId)
		}

		if err := checkHyperCubeErrors(gob.GenericId, hypercube); err != nil {
			switch err.(type) {
			case CalcEvalConditionFailedError:
				sessionState.LogEntry.Logf(logger.WarningLevel, "object<%s>: %v", obj.ID, err)
				return errors.Wrap(obj.SetHyperCubeDataPages(make([]*enigma.NxDataPage, 0), false), "failed to set hypercube datapages")
			}
			return errors.WithStack(err)
		}

		if hypercube.Size == nil {
			return errors.Errorf("object<%s> has no hypercube size", gob.GenericId)
		}

		if hypercube.Size.Cx < 1 {
			return errors.Errorf("object<%s> has no hypercube width", gob.GenericId)
		}

		sessionState.LogEntry.LogDebugf("object<%s> type<%s> hypercube Cx<%d>", gob.GenericId, gob.GenericType, hypercube.Size.Cx)

		var pages []*enigma.NxPage
		if columns {
			for i := 0; i < hypercube.Size.Cx; i++ {
				pages = append(pages, &enigma.NxPage{
					Left:   i,
					Top:    0,
					Width:  1,
					Height: requestDef.MaxHeight(),
				})
			}
		} else {
			pages = append(pages, &enigma.NxPage{
				Left:   0,
				Top:    0,
				Width:  hypercube.Size.Cx, //TODO check if true for scatterplot
				Height: requestDef.MaxHeight(),
			})
		}

		//Will not be entirely correct for table, has multiple pages are sent.
		//Do like this for now to be compareable to current sdkexerciser logic
		datapages, err := gob.GetHyperCubeData(ctx, requestDef.Path, pages)
		err = checkEngineErr(err, sessionState, fmt.Sprintf("object<%s>.GetHyperCubeData", gob.GenericId))
		if err != nil {
			return errors.WithStack(err)
		}

		if err = obj.SetHyperCubeDataPages(datapages, false); err != nil {
			return errors.Wrap(err, "failed to set hypercube datapages")
		}

		return nil
	}, actionState, true, fmt.Sprintf("Failed to update object hypercube data for object<%s>", gob.GenericId))
}

// UpdateObjectHyperCubeReducedDataAsync send get hypercube reduced data request and update saved hypercube
func UpdateObjectHyperCubeReducedDataAsync(sessionState *State, actionState *action.State, gob *enigma.GenericObject,
	obj *enigmahandlers.Object, requestDef senseobjdef.GetDataRequests) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		sessionState.LogEntry.LogDebugf("Get hypercube reduced data for object<%s>", gob.GenericId)
		hypercube := obj.HyperCube()
		if hypercube == nil {
			return errors.Errorf("object<%s> has no hypercube", gob.GenericId)
		}

		if err := checkHyperCubeErrors(gob.GenericId, hypercube); err != nil {
			switch err.(type) {
			case CalcEvalConditionFailedError:
				sessionState.LogEntry.Logf(logger.WarningLevel, "object<%s>: %v", obj.ID, err)
				return errors.Wrap(obj.SetHyperCubeDataPages(make([]*enigma.NxDataPage, 0), false), "failed to set hypercube datapages")
			}
			return errors.WithStack(err)
		}

		if hypercube.Size == nil {
			return errors.Errorf("object<%s> has no hypercube size", gob.GenericId)
		}

		//Do we have stacked dimensions?
		var isStackedDims int
		if len(hypercube.DimensionInfo) > 1 &&
			(hypercube.Mode == constant.HyperCubeDataModePivotStack || hypercube.Mode == constant.HyperCubeDataModePivotStackL) {
			isStackedDims = 1
		} else {
			isStackedDims = 0
		}

		datapages, err := gob.GetHyperCubeReducedData(ctx, requestDef.Path, []*enigma.NxPage{
			{
				Left:   isStackedDims,
				Top:    0,
				Width:  len(hypercube.DimensionInfo) + len(hypercube.MeasureInfo) + isStackedDims,
				Height: int(math.Min(2000.0, 10000.0/float64(len(hypercube.MeasureInfo)+1))),
			},
		}, -1, constant.DataReductionModeOneDim)
		err = checkEngineErr(err, sessionState, fmt.Sprintf("object<%s>.GetHyperCubeReducedData", gob.GenericId))
		if err != nil {
			return errors.WithStack(err)
		}

		if err = obj.SetHyperCubeDataPages(datapages, false); err != nil {
			return errors.Wrap(err, "failed to set hypercube datapages")
		}

		return nil
	}, actionState, true, fmt.Sprintf("Failed to update object hypercube reduced data for object<%s>", gob.GenericId))
}

// UpdateObjectHyperCubeBinnedDataAsync send get hypercube binned data request and update saved hypercube
func UpdateObjectHyperCubeBinnedDataAsync(sessionState *State, actionState *action.State, gob *enigma.GenericObject,
	obj *enigmahandlers.Object, requestDef senseobjdef.GetDataRequests) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		sessionState.LogEntry.LogDebugf("Get hypercube binned data for object<%s>", gob.GenericId)
		hypercube := obj.HyperCube()
		if hypercube == nil {
			return errors.Errorf("object<%s> has no hypercube", gob.GenericId)
		}

		if err := checkHyperCubeErrors(gob.GenericId, hypercube); err != nil {
			switch err.(type) {
			case CalcEvalConditionFailedError:
				sessionState.LogEntry.Logf(logger.WarningLevel, "object<%s>: %v", obj.ID, err)
				return errors.Wrap(obj.SetHyperCubeDataPages(make([]*enigma.NxDataPage, 0), false), "failed to set hypercube datapages")
			}
			return errors.WithStack(err)
		}

		if hypercube.Size == nil {
			return errors.Errorf("object<%s> has no hypercube size", gob.GenericId)
		}

		if hypercube.Size.Cx < 1 {
			return errors.Errorf("object<%s> has no hypercube width", gob.GenericId)
		}

		if hypercube.MeasureInfo == nil {
			return errors.Errorf("object<%s> has no measureinfo<nil>", gob.GenericId)
		}

		if len(hypercube.MeasureInfo) < 2 {
			return errors.Errorf("object<%s> has less than two measures, GetHyperCubeBinnedData not possible",
				gob.GenericId)
		}

		measure0 := hypercube.MeasureInfo[0]
		measure1 := hypercube.MeasureInfo[1]

		if measure0 == nil || measure1 == nil {
			return errors.Errorf("object<%s> has nil measure, GetHyperCubeBinnedData not possible", gob.GenericId)
		}

		maxHeight := hypercube.Size.Cy
		if maxHeight > requestDef.MaxHeight() {
			maxHeight = requestDef.MaxHeight()
		}

		datapages, err := gob.GetHyperCubeBinnedData(ctx, requestDef.Path,
			[]*enigma.NxPage{
				{
					Left:   0,
					Top:    0,
					Width:  hypercube.Size.Cx,
					Height: maxHeight,
				},
			}, &enigma.NxViewPort{
				Height: 0,
				Width:  0,
			}, []*enigma.NxDataAreaPage{
				{
					Left:   measure0.Min,
					Top:    measure1.Max,
					Width:  measure0.Max - measure0.Min,
					Height: measure1.Max - measure1.Min,
				},
			},
			1000, //maxNbrCells, should this be definable?
			5,    //queryLevel, should this be definable?
			0,    //binningMethod, should this be definable?
		)
		err = checkEngineErr(err, sessionState, fmt.Sprintf("object<%s>.GetHyperCubeBinnedData", gob.GenericId))
		if err != nil {
			return errors.WithStack(err)
		}

		if err = obj.SetHyperCubeDataPages(datapages, true); err != nil {
			return errors.Wrap(err, "failed to set hypercube datapages")
		}

		return nil
	}, actionState, true, fmt.Sprintf("Failed to update object binned data for object<%s>", gob.GenericId))
}

// UpdateObjectHyperCubeStackDataAsync send get stacked hypercube data and update saved hypercube
func UpdateObjectHyperCubeStackDataAsync(sessionState *State, actionState *action.State, gob *enigma.GenericObject,
	obj *enigmahandlers.Object, requestDef senseobjdef.GetDataRequests) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		sessionState.LogEntry.LogDebugf("Get hypercube stack data for object<%s>", gob.GenericId)
		hypercube := obj.HyperCube()
		if hypercube == nil {
			return errors.Errorf("object<%s> has no hypercube", gob.GenericId)
		}

		if err := checkHyperCubeErrors(gob.GenericId, hypercube); err != nil {
			switch err.(type) {
			case CalcEvalConditionFailedError:
				sessionState.LogEntry.Logf(logger.WarningLevel, "object<%s>: %v", obj.ID, err)
				return errors.Wrap(obj.SetStackHyperCubePages(make([]*enigma.NxStackPage, 0)), "failed to set hypercube datapages")
			}
			return errors.WithStack(err)
		}

		if hypercube.Size == nil {
			return errors.Errorf("object<%s> has no hypercube size", gob.GenericId)
		}

		if hypercube.Size.Cx < 1 {
			return errors.Errorf("object<%s> has no hypercube width", gob.GenericId)
		}

		datapages, err := gob.GetHyperCubeStackData(ctx, requestDef.Path, []*enigma.NxPage{
			{
				Left:   0,
				Top:    0,
				Width:  hypercube.Size.Cx,
				Height: requestDef.MaxHeight(),
			},
		}, 0)
		err = checkEngineErr(err, sessionState, fmt.Sprintf("object<%s>.GetHyperCubeStackData", gob.GenericId))
		if err != nil {
			return errors.WithStack(err)
		}

		if err = obj.SetStackHyperCubePages(datapages); err != nil {
			return errors.Wrap(err, "failed to set hypercube datapages")
		}

		return nil
	}, actionState, true, fmt.Sprintf("Failed to update object stack data for object<%s>", gob.GenericId))
}

// UpdateListObjectDataAsync send get listobject data and update saved list object
func UpdateListObjectDataAsync(sessionState *State, actionState *action.State, gob *enigma.GenericObject,
	obj *enigmahandlers.Object, requestDef senseobjdef.GetDataRequests) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		sessionState.LogEntry.LogDebugf("Get listobject data for object<%s>", gob.GenericId)
		dataPages, err := gob.GetListObjectData(ctx, requestDef.Path, []*enigma.NxPage{
			{
				Left:   0,
				Top:    0,
				Width:  1,
				Height: requestDef.MaxHeight(),
			},
		})
		err = checkEngineErr(err, sessionState, fmt.Sprintf("object<%s>.GetHypercubeData", gob.GenericId))
		if err != nil {
			return errors.WithStack(err)
		}

		if err = obj.SetListObjectDataPages(dataPages); err != nil {
			return errors.WithStack(err)
		}

		return nil
	}, actionState, true, fmt.Sprintf("failed to get listobject data for object<%s>", gob.GenericId))
}

// UpdateObjectHyperCubeContinuousDataAsync send get continous data request
func UpdateObjectHyperCubeContinuousDataAsync(sessionState *State, actionState *action.State, gob *enigma.GenericObject,
	obj *enigmahandlers.Object, requestDef senseobjdef.GetDataRequests) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		sessionState.LogEntry.LogDebugf("Get continuous data for object<%s>", gob.GenericId)
		hypercube := obj.HyperCube()
		if err := checkHyperCubeErrors(gob.GenericId, hypercube); err != nil {
			switch err.(type) {
			case CalcEvalConditionFailedError:
				sessionState.LogEntry.Logf(logger.WarningLevel, "object<%s>: %v", obj.ID, err)
				return nil
			}
			return errors.WithStack(err)
		}
		maxLines := maxNbrLines
		start, end, err := GetFullContinuousRange(hypercube)
		sessionState.LogEntry.LogDebugf("Get continuous data for object with start <%v> and end <%v>", start, end)
		if err != nil {
			return errors.WithStack(err)
		}
		_, _, err = gob.GetHyperCubeContinuousData(ctx, requestDef.Path, &enigma.NxContinuousDataOptions{
			Start:          start,
			End:            end,
			NbrPoints:      GetApproriateNrOfBins(hypercube),
			MaxNbrTicks:    maxNbrTicks,
			MaxNumberLines: &maxLines,
		}, false)
		if err != nil {
			return errors.WithStack(err)
		}
		return nil
	}, actionState, true, fmt.Sprintf("failed to get continous data for object<%s>", gob.GenericId))
}

// UpdateObjectHyperCubeTreeDataAsync send get hypercube tree data request and update saved data
func UpdateObjectHyperCubeTreeDataAsync(sessionState *State, actionState *action.State, gob *enigma.GenericObject,
	obj *enigmahandlers.Object, requestDef senseobjdef.GetDataRequests) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		sessionState.LogEntry.LogDebugf("Get tree data for object<%s> type<%s>", gob.GenericId, gob.GenericType)

		hypercube := obj.HyperCube()
		if hypercube == nil {
			return errors.Errorf("no hybercube found for object<%s> type<%s>", gob.GenericId, gob.GenericType)
		}

		dimInfo := hypercube.DimensionInfo
		if len(dimInfo) < 1 {
			return errors.Errorf("no dimensions found for object<%s> type<%s>", gob.GenericId, gob.GenericType)
		}

		nodes := make([]*enigma.NxPageTreeNode, 0, len(dimInfo))

		for i, dim := range dimInfo {
			height := dim.Cardinal
			if height > requestDef.Height {
				height = requestDef.Height
			}

			node := &enigma.NxPageTreeNode{
				Area: &enigma.Rect{
					Left:   i,
					Top:    0,
					Width:  1,
					Height: height,
				},
				AllValues: i != 0,
			}
			nodes = append(nodes, node)
		}

		treeNodes, err := gob.GetHyperCubeTreeData(ctx, requestDef.Path, &enigma.NxTreeDataOption{
			TreeNodes: nodes,
		})
		err = checkEngineErr(err, sessionState, fmt.Sprintf("object<%s>.GetHyperCubeTreeData", gob.GenericId))
		if err != nil {
			return errors.WithStack(err)
		}

		if err := obj.SetTreeDataPages(treeNodes); err != nil {
			return errors.WithStack(err)
		}

		return nil
	}, actionState, true, fmt.Sprintf("failed to get tree data for object<%s>", gob.GenericId))
}

func checkHyperCubeErrors(id string, hypercube *enigmahandlers.HyperCube) error {
	if hypercube == nil {
		return errors.Errorf("object<%s> has no hypercube", id)
	}

	if err, _ := checkHyperCubeErrorInner(id, "hypercube", HyperCubeSectionRoot, hypercube.Error); err != nil {
		return err
	}

	var firstWarnError error
	if hypercube.DimensionInfo != nil {
		for i, dimInfo := range hypercube.DimensionInfo {
			if err, warning := checkHyperCubeErrorInner(id, fmt.Sprintf("hypercube.DimensionInfo[%d]", i), HyperCubeSectionDimension, dimInfo.Error); err != nil {
				if !warning {
					return errors.Wrapf(err, "object<%s> has hypercube error<%s> in DimensionInfo[%d]", id, EngineCodeToString(dimInfo.Error.ErrorCode), i)
				}
				if firstWarnError == nil {
					firstWarnError = err
				}
			}
		}
	}

	if hypercube.MeasureInfo != nil {
		for i, measureInfo := range hypercube.MeasureInfo {
			if err, warning := checkHyperCubeErrorInner(id, fmt.Sprintf("hypercube.MeasureInfo[%d]", i), HyperCubeSectionMeasure, measureInfo.Error); err != nil {
				if !warning {
					return errors.Wrapf(err, "object<%s> has hypercube error<%s> in MeasureInfo[%d]", id, EngineCodeToString(measureInfo.Error.ErrorCode), i)
				}
				if firstWarnError == nil {
					firstWarnError = err
				}
			}
			if measureInfo.MiniChart != nil {
				if err, warning := checkHyperCubeErrorInner(id, fmt.Sprintf("hypercube.MeasureInfo[%d].MiniChart", i), HyperCubeSectionMeasureMinichart, measureInfo.MiniChart.Error); err != nil {
					if !warning {
						return errors.Wrapf(err, "object<%s> has hypercube error<%s> in MeasureInfo[%d].MiniChart", id, EngineCodeToString(measureInfo.MiniChart.Error.ErrorCode), i)
					}
					if firstWarnError == nil {
						firstWarnError = err
					}
				}
			}
		}
	}

	return firstWarnError
}

func checkHyperCubeErrorInner(id, path string, section HyperCubeSection, nve *enigma.NxValidationError) (error, bool) {
	if nve == nil {
		return nil, false
	}
	switch nve.ErrorCode {
	case constant.LocerrCalcEvalConditionFailed:
		return CalcEvalConditionFailedError(path), true
	default:
		return NxValidationError{Err: *nve, Id: id, Section: section, Path: path}, false
	}
}

func checkEngineErr(err error, sessionState *State, req string) error {
	switch e := err.(type) {
	case enigma.Error:
		switch e.Code() {
		case constant.LocerrGenericAborted:
			sessionState.LogEntry.LogDebugf("Request<%s> was aborted", req)
			return nil
		case constant.LocerrCalcEvalConditionFailed:
			sessionState.LogEntry.Logf(logger.WarningLevel, "Request<%s> has unsatisfied calculation condition", req)
			return nil
		default:
			return err
		}
	default:
		return err
	}
}

// Logic as written in client.js as of sense april 2018:
//
//	getFullContinuousRange: function (t) {
//		var e = t.qHyperCube.qDimensionInfo[0].qMin,
//		n = t.qHyperCube.qDimensionInfo[0].qMax;
//		return n < e || "NaN" === n ? e = n = "NaN" : e === n && (e -= .5, n += .5), {
//			min: e,
//			max: n
//		}
//	},
func GetFullContinuousRange(hypercube *enigmahandlers.HyperCube) (enigma.Float64, enigma.Float64, error) {
	if hypercube == nil || hypercube.DimensionInfo == nil || len(hypercube.DimensionInfo) < 1 {
		return -1, -1, errors.Errorf("hypercube has no dimension")
	}
	e := hypercube.DimensionInfo[0].Min
	n := hypercube.DimensionInfo[0].Max
	if n < e || math.IsNaN(float64(n)) {
		//Should be set to zero then according to the client code
		e = enigma.Float64(0.0)
		n = e
	} else if e == n {
		e -= .5
		n += .5
	}
	return e, n, nil
}

// Logic as written in client.js as of sense april 2018:
//
//	getApproriateNrOfBins: function (t) {
//		var e = t.qHyperCube.qMeasureInfo.length || 1,
//		n = 4 + 2 * (e - 1);
//		return t.qHyperCube.qDimensionInfo.length > 1 && (e = Math.max(1, Math.min(h.maxNumberOfLines, t.qHyperCube.qDimensionInfo[1].qStateCounts.qLocked + t.qHyperCube.qDimensionInfo[1].qStateCounts.qOption + t.qHyperCube.qDimensionInfo[1].qStateCounts.qSelected)), n = 4),
//		Math.ceil(2e3 / (e * n))
//	},
func GetApproriateNrOfBins(hypercube *enigmahandlers.HyperCube) int {
	e := 1
	if hypercube != nil && hypercube.MeasureInfo != nil {
		e = len(hypercube.MeasureInfo)
	}
	n := 4 + 2*(e-1)

	if hypercube != nil && hypercube.DimensionInfo != nil && len(hypercube.DimensionInfo) > 1 {
		n = 4
		stateCounts := hypercube.DimensionInfo[1].StateCounts
		states := stateCounts.Locked + stateCounts.Option + stateCounts.Selected
		e = int(math.Max(1.0, math.Min(float64(maxNbrLines), float64(states))))
	}
	return int(math.Ceil(2000.0 / float64(e*n)))
}

// SetObjectData sets data to obj from layout and data requests according to objectDef
func SetObjectData(sessionState *State, actionState *action.State, rawLayout json.RawMessage, objectDef *senseobjdef.ObjectDef,
	obj *enigmahandlers.Object, enigmaObject *enigma.GenericObject) error {
	switch objectDef.DataDef.Type {
	case senseobjdef.DataDefNoData:
		return nil
	case senseobjdef.DataDefListObject:
		if string(objectDef.DataDef.Path) == "" {
			return errors.Errorf(
				"object<%s> is defined as listobject carrier, but has not listobject path definition", enigmaObject.GenericType)
		}

		if err := SetListObject(rawLayout, obj, objectDef.DataDef.Path); err != nil {
			return errors.Wrapf(err, "object<%s> type<%s>", obj.ID, enigmaObject.GenericType)
		}
	case senseobjdef.DataDefHyperCube:
		if objectDef.DataDef.Path == "" {
			return errors.Errorf(
				"object<%s> is defined as hypercube carrier, but has not hypercube path definition", enigmaObject.GenericType)
		}
		if err := SetHyperCube(sessionState, actionState, rawLayout, obj, objectDef.DataDef.Path); err != nil {
			return errors.Wrapf(err, "object<%s> type<%s>", obj.ID, enigmaObject.GenericType)
		}
	default:
		sessionState.LogEntry.Logf(logger.WarningLevel, "Get Data for object<%s> type<%s> not supported", enigmaObject.GenericId, enigmaObject.GenericType)
		return nil
	}

	// Evaluate data requests
	dataRequests, err := objectDef.Evaluate(rawLayout)
	if err != nil {
		return errors.Wrapf(err, "object<%s> type<%s>", obj.ID, enigmaObject.GenericType)
	}
	sessionState.LogEntry.LogDebugf("object<%s> type<%s> request evaluation result<%+v>", obj.ID, enigmaObject.GenericType, dataRequests)
	if obj.HyperCube() != nil {
		sessionState.LogEntry.LogDebugf("object<%s> type<%s> hypercube mode<%s>", obj.ID, enigmaObject.GenericType, obj.HyperCube().Mode)
	}
	if len(dataRequests) < 1 {
		return nil
	}

	for _, r := range dataRequests {
		columns := false
		switch r.Type {
		case senseobjdef.DataTypeLayout:
		case senseobjdef.DataTypeListObject:
			UpdateListObjectDataAsync(sessionState, actionState, enigmaObject, obj, r)
		case senseobjdef.DataTypeHyperCubeDataColumns:
			columns = true
			fallthrough
		case senseobjdef.DataTypeHyperCubeData:
			UpdateObjectHyperCubeDataAsync(sessionState, actionState, enigmaObject, obj, r, columns)
		case senseobjdef.DataTypeHyperCubeReducedData:
			UpdateObjectHyperCubeReducedDataAsync(sessionState, actionState, enigmaObject, obj, r)
		case senseobjdef.DataTypeHyperCubeBinnedData:
			UpdateObjectHyperCubeBinnedDataAsync(sessionState, actionState, enigmaObject, obj, r)
		case senseobjdef.DataTypeHyperCubeStackData:
			UpdateObjectHyperCubeStackDataAsync(sessionState, actionState, enigmaObject, obj, r)
		case senseobjdef.DataTypeHyperCubeContinuousData:
			UpdateObjectHyperCubeContinuousDataAsync(sessionState, actionState, enigmaObject, obj, r)
		case senseobjdef.DataTypeHyperCubeTreeData:
			UpdateObjectHyperCubeTreeDataAsync(sessionState, actionState, enigmaObject, obj, r)
		default:
			sessionState.LogEntry.Logf(logger.WarningLevel,
				"Get Data for object<%s> type<%s> not supported", enigmaObject.GenericId, enigmaObject.GenericType)
		}
	}
	return nil
}

func EngineCodeToString(errorCode int) string {
	return fmt.Sprintf("ErrorCode:%d (%s)", errorCode, enigma.ErrorCodeLookup(errorCode))
}
