package scenario

import (
	"context"
	"fmt"
	"github.com/qlik-oss/gopherciser/helpers"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/globals/constant"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjdef"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// SelectionType type of selection
	SelectionType int
	// SelectionSettings selection settings
	SelectionSettings struct {
		// ID object id
		ID string `json:"id" displayname:"Object ID" doc-key:"select.id" appstructure:"object"`
		// Type selection type
		Type SelectionType `json:"type" displayname:"Selection type" doc-key:"select.type"`
		// Accept true - confirm selection. false - abort selection
		Accept bool `json:"accept" displayname:"Accept selection" doc-key:"select.accept"`
		// WrapSelections
		WrapSelections bool `json:"wrap" displayname:"Wrap selections" doc-key:"select.wrap"`
		// Min minimum amount of values to select
		Min int `json:"min" displayname:"Minimum amount of values to select" doc-key:"select.min"`
		// Max maximum amount of values to select
		Max int `json:"max" displayname:"Maximum amount of values to select" doc-key:"select.max"`
		// Dimension in which dimension to select (defaults to 0)
		Dimension int `json:"dim" displayname:"Dimension to select in" doc-key:"select.dim"`
	}

	selectStates int

	uniqueInts map[int]struct{}
)

const (
	// RandomFromAll random from all values
	RandomFromAll SelectionType = iota
	// RandomFromEnabled random from white values
	RandomFromEnabled
	// RandomFromExcluded random from grey values
	RandomFromExcluded
	// RandomDeselect random deselect from selected values
	RandomDeselect
)

const (
	selectStateLocked selectStates = iota
	selectStateSelected
	selectStateOption
	selectStateDeselected
	selectStateAlternative
	selectStateExcluded
	selectStateExcludedSelected
	selectStateExcludedLocked
)

type Enum interface {
	GetEnumMap() *enummap.EnumMap
}

var selectionTypeEnumMap, _ = enummap.NewEnumMap(map[string]int{
	"randomfromall":      int(RandomFromAll),
	"randomfromenabled":  int(RandomFromEnabled),
	"randomfromexcluded": int(RandomFromExcluded),
	"randomdeselect":     int(RandomDeselect),
})

func (value SelectionType) GetEnumMap() *enummap.EnumMap {
	return selectionTypeEnumMap
}

var (
	selectStateHandler, _ = enummap.NewEnumMap(map[string]int{
		"l":  int(selectStateLocked),
		"s":  int(selectStateSelected),
		"o":  int(selectStateOption),
		"d":  int(selectStateDeselected),
		"a":  int(selectStateAlternative),
		"x":  int(selectStateExcluded),
		"xs": int(selectStateExcludedSelected),
		"xl": int(selectStateExcludedLocked),
	})
)

// UnmarshalJSON unmarshal selection type
func (value *SelectionType) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal SelectionType")
	}

	*value = SelectionType(i)
	return nil
}

// MarshalJSON marshal selection type
func (value SelectionType) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown selectiontype<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// String representation of StaticSelectionType
func (value SelectionType) String() string {
	sType, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return strconv.Itoa(int(value))
	}
	return sType
}

// IsExcludedOrDeselect true if type is RandomDeselect or RandomFromExcluded
func (value SelectionType) IsExcludedOrDeselect() bool {
	switch value {
	case RandomFromExcluded:
		return true
	case RandomDeselect:
		return true
	default:
		return false
	}
}

// Validate select action
func (settings SelectionSettings) Validate() error {
	if settings.ID == "" {
		return errors.Errorf("Empty object ID")
	}

	if settings.Dimension < 0 {
		return errors.Errorf("Illegal dimension<%d>", settings.Dimension)
	}

	if settings.Min < 1 {
		return errors.Errorf("min<%d> selections must be >1", settings.Min)
	}

	if settings.Max < 1 {
		return errors.Errorf("max<%d> selections must be >1", settings.Max)
	}

	if settings.Min > settings.Max {
		return errors.Errorf("min<%d> must be less than max<%d>", settings.Min, settings.Max)
	}

	return nil
}

func (state selectStates) isEnabled(binned bool) bool {
	switch state {
	case selectStateAlternative:
		return true
	case selectStateOption:
		return true
	case selectStateSelected:
		return true
	case selectStateLocked:
		return binned
	default:
		return false
	}
}

func (state selectStates) isExcluded() bool {
	switch state {
	case selectStateDeselected:
		return true
	case selectStateExcluded:
		return true
	case selectStateExcludedSelected:
		return true
	case selectStateExcludedLocked:
		return true
	case selectStateLocked:
		return true
	default:
		return false
	}
}

func (state selectStates) isSelected() bool {
	switch state {
	case selectStateSelected:
		return true
	default:
		return false
	}
}

// Execute select action
func (settings SelectionSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	uplink := sessionState.Connection.Sense()
	objectID := sessionState.IDMap.Get(settings.ID)
	gob, err := uplink.Objects.GetObjectByID(objectID)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "Failed getting object<%s> from object list", objectID))
		return
	}

	linkedObjHandle := uplink.Objects.GetObjectLink(gob.Handle)
	if linkedObjHandle != 0 {
		var errLink error
		gob, errLink = uplink.Objects.GetObject(linkedObjHandle)
		if errLink != nil {
			actionState.AddErrors(errors.Wrapf(errLink, "Failed getting linked object<%d> object<%s>", linkedObjHandle, objectID))
			return
		}
	}

	switch t := gob.EnigmaObject.(type) {
	case *enigma.GenericObject:
		genObj := gob.EnigmaObject.(*enigma.GenericObject)

		// Get selection definitions for object
		def, defErr := senseobjdef.GetObjectDef(genObj.GenericType)
		if defErr != nil {
			actionState.AddErrors(errors.Wrapf(defErr, "Failed to get object<%s> selection definitions", genObj.GenericType))
			return
		}

		if validateErr := def.Validate(); validateErr != nil {
			actionState.AddErrors(errors.Wrapf(validateErr, "Error validating object<%s> selection definitions<%+v>", genObj.GenericType, def))
			return
		}

		if settings.WrapSelections {
			// Start selections
			sessionState.QueueRequest(func(ctx context.Context) error {
				return genObj.BeginSelections(ctx, []string{def.Select.Path})
			}, actionState, false, fmt.Sprintf("Begin selection error: %v", err))

			sessionState.Wait(actionState)
			if actionState.Errors() != nil {
				return
			}
		}

		cardinal, cardinalErr := getCardinal(gob, def, settings.Dimension)
		if cardinalErr != nil {
			actionState.AddErrors(errors.Wrapf(cardinalErr, "Failed to get cardinal for object<%s>", gob.ID))
			return
		}
		if cardinal < 0 {
			actionState.AddErrors(errors.Errorf("object<%s> has illegal cardinal<%d>", gob.ID, cardinal))
			return
		}

		rnd := sessionState.Randomizer()
		if rnd == nil {
			actionState.AddErrors(errors.New("No randomizer set on connection"))
			return
		}

		var selectPos []int
		var selectBins []string
		switch settings.Type {
		case RandomFromAll:
			var fillErr error
			selectPos, fillErr = fillSelectPosFromAll(settings.Min, settings.Max, cardinal, rnd)
			actionState.AddErrors(fillErr)
		case RandomFromExcluded:
			fallthrough // handled within getPossible
		case RandomDeselect:
			fallthrough
		case RandomFromEnabled:
			columns := false
			switch def.Select.Type {
			case senseobjdef.SelectTypeHypercubeColumnValues:
				columns = true
			}
			possible, bins, errPossible := getPossible(gob, def, settings.Dimension, settings.Type, columns)
			if errPossible != nil {
				actionState.AddErrors(errors.WithStack(errPossible))
				break
			}
			if bins != nil {
				var fillErr error
				selectBins, fillErr = fillSelectBinsFromBins(settings.Min, settings.Max, bins, rnd)
				actionState.AddErrors(fillErr)
			} else {
				var fillErr error
				selectPos, fillErr = fillSelectPosFromPossible(settings.Min, settings.Max, possible, rnd)
				actionState.AddErrors(fillErr)
			}
		default:
			actionState.AddErrors(errors.Errorf("Unknown select type<%s>", settings.Type.String()))
		}

		if actionState.Errors() != nil {
			return
		}

		if len(selectPos) < 1 && len(selectBins) < 1 {
			sessionState.LogEntry.Logf(logger.WarningLevel, "Nothing to select in object<%s>", gob.ID)
			return
		}

		if selectBins != nil {
			actionState.Details = fmt.Sprintf("%s;%v", gob.ID, selectBins)
		} else {
			actionState.Details = fmt.Sprintf("%s;%v", gob.ID, selectPos)
		}

		var selectFunc func(ctx context.Context) (bool, error)
		switch def.Select.Type {
		case senseobjdef.SelectTypeListObjectValues:
			selectFunc = func(ctx context.Context) (bool, error) {
				return genObj.SelectListObjectValues(ctx, def.Select.Path, selectPos, true, false)
			}
		case senseobjdef.SelectTypeHypercubeColumnValues:
			fallthrough
		case senseobjdef.SelectTypeHypercubeValues:
			if len(selectBins) > 0 {
				selectInfo, convertErr := convertBinsToSelectInfo(selectBins)
				if convertErr != nil {
					actionState.AddErrors(errors.WithStack(convertErr))
					return
				}
				selectFunc = func(ctx context.Context) (bool, error) {
					return genObj.MultiRangeSelectHyperCubeValues(ctx, def.Select.Path, selectInfo, false, false)
				}
			} else {
				selectFunc = func(ctx context.Context) (bool, error) {
					if len(selectPos) < 1 {
						return false, errors.Errorf("SelectHyperCubeValues SelectPos is nil")
					}
					return genObj.SelectHyperCubeValues(ctx, def.Select.Path, settings.Dimension, selectPos, true)
				}
			}
		default:
			actionState.AddErrors(errors.Errorf("Unknown select type<%v> for object<%v> type<%s>",
				def.Select.Type, gob.Type, genObj.GenericType))
			return
		}

		// Select
		sessionState.QueueRequest(func(ctx context.Context) error {
			sessionState.LogEntry.LogDebugf("Select in object<%s> h<%d> type<%s>", genObj.GenericId, genObj.Handle, genObj.GenericType)
			success, err := selectFunc(ctx)
			if err != nil {
				return errors.Wrapf(err, "Failed to select in object<%s>", genObj.GenericId)
			}
			if !success {
				return errors.Errorf("Select in object<%s> unsuccessful", genObj.GenericId)
			}
			sessionState.LogEntry.LogDebug(fmt.Sprint("Successful select in", genObj.GenericId))

			if settings.WrapSelections {
				//End Selections
				sessionState.QueueRequest(func(ctx context.Context) error {
					return genObj.EndSelections(ctx, settings.Accept)
				}, actionState, true, "End selections failed")
			}

			return nil
		}, actionState, true, fmt.Sprintf("Failed to select in %s", genObj.GenericId))
	default:
		actionState.AddErrors(errors.Errorf("Unknown object type<%T>", t))
		return
	}

	sessionState.Wait(actionState)
}

//AddValue to unique list
func (u *uniqueInts) AddValue(v int) {
	if u == nil || *u == nil {
		*u = make(map[int]struct{})
	}
	var emptyStruct struct{}
	(*u)[v] = emptyStruct
}

//Array of unique integers
func (u *uniqueInts) Array() []int {
	if u == nil || *u == nil {
		return []int{}
	}

	a := make([]int, len(*u))
	a = a[:0]

	for k := range *u {
		a = append(a, k)
	}
	// sort the keys so seeded randomization always gets the same order
	sort.Ints(a)
	return a
}

//HasValue test if collection includes value
func (u *uniqueInts) HasValue(v int) bool {
	if u == nil || *u == nil {
		return false
	}
	_, exist := (*u)[v]
	return exist
}

// TODO support stack hypercube and pivot hypercube + maps etc
func getCardinal(obj *enigmahandlers.Object, selectDef *senseobjdef.ObjectDef, dimension int) (int, error) {
	if selectDef == nil {
		return -1, errors.Errorf("selection def is nil")
	}

	switch selectDef.DataDef.Type {
	case senseobjdef.DataDefListObject:
		listobject := obj.ListObject()
		if listobject == nil {
			return -1, errors.Errorf("object<%s> has no listobject", obj.ID)
		}
		if listobject.DimensionInfo == nil {
			return -1, errors.Errorf("object<%s> does not have a dimension", obj.ID)
		}
		if err := verifyDimension(obj.ID, dimension, []*enigma.NxDimensionInfo{listobject.DimensionInfo}); err != nil {
			return -1, errors.WithStack(err)
		}
		return listobject.DimensionInfo.Cardinal, nil
	case senseobjdef.DataDefHyperCube:
		hypercube := obj.HyperCube()
		if hypercube == nil {
			return -1, errors.Errorf("object type<%v> has no hypercube", obj.Type)
		}
		return getHyperCubeCardinal(obj.ID, dimension, hypercube)
	default:
		return -1, errors.Errorf("object type<%v> doesn't have a supported data carrier definition", obj.Type)
	}
}

func getHyperCubeCardinal(objId string, dimension int, hypercube *enigmahandlers.HyperCube) (int, error) {
	if hypercube == nil {
		return -1, errors.Errorf("object<%s> has no hypercube", objId)
	}

	if hypercube.DimensionInfo == nil {
		return -1, errors.Errorf("object<%s> does not have any dimensions", objId)
	}

	if err := verifyDimension(objId, dimension, hypercube.DimensionInfo); err != nil {
		return -1, errors.WithStack(err)
	}

	dimInfo := hypercube.DimensionInfo[dimension]
	if dimInfo == nil {
		return -1, errors.Errorf("object<%s> dimension<%d> is nil", objId, dimension)
	}

	return dimInfo.Cardinal, nil
}

func getSelectQty(min, max, possible int, rnd helpers.Randomizer) int {
	localMin := min
	localMax := max

	if rnd == nil {
		return -1
	}

	if localMax > possible {
		localMax = possible
	}

	if localMin > localMax {
		localMin = localMax
	}

	if localMin == localMax {
		return localMin
	}

	return rnd.Rand(localMax-localMin+1) + localMin
}

func fillSelectPosFromAll(min, max, cardinal int, rnd helpers.Randomizer) ([]int, error) {
	if rnd == nil {
		return nil, errors.Errorf("Randomizer not provided")
	}

	if cardinal < 1 {
		return nil, nil
	}

	if cardinal == 1 {
		return []int{0}, nil
	}

	selectQty := getSelectQty(min, max, cardinal, rnd)
	if selectQty < 1 {
		return nil, nil
	}

	if selectQty == cardinal {
		positions := make([]int, cardinal)
		for i := 0; i < cardinal; i++ {
			positions[i] = i
		}
		return positions, nil
	}

	selectPos := make(uniqueInts)

	failSafe := 0
	for len(selectPos) < selectQty {
		pos := rnd.Rand(cardinal)
		if selectPos.HasValue(pos) {
			failSafe++
			if failSafe > 10000 {
				return nil, errors.Errorf("Error randomizing positions, failsafe<10000> limit reached")
			}
			continue
		}
		selectPos.AddValue(pos)
	}

	return selectPos.Array(), nil
}

func fillSelectPosFromPossible(min, max int, possible []int, rnd helpers.Randomizer) ([]int, error) {
	if rnd == nil {
		return nil, errors.Errorf("Randomizer not provided")
	}

	if possible == nil || len(possible) < 1 {
		return nil, nil
	}

	possibleLength := len(possible)

	selectQty := getSelectQty(min, max, possibleLength, rnd)
	if selectQty < 1 {
		return nil, nil
	}

	if selectQty > possibleLength {
		return nil, errors.Errorf("select quantity<%d> calculated to larger than possible<%d>",
			selectQty, possibleLength)
	}

	if possibleLength == selectQty {
		return possible, nil
	}

	selectPos := make(uniqueInts)
	failSafe := 0
	for len(selectPos) < selectQty {
		elValue, elPos, err := rnd.RandIntPos(possible)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if selectPos.HasValue(elValue) {
			failSafe++
			if failSafe > 10000 {
				return nil, errors.Errorf("Error randomizing positions, failsafe<10000> limit reached")
			}
		} else {
			selectPos.AddValue(elValue)
			if err = cutPosition(elPos, &possible); err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}

	return selectPos.Array(), nil
}

func fillSelectBinsFromBins(min, max int, bins []string, rnd helpers.Randomizer) ([]string, error) {
	if rnd == nil {
		return nil, errors.Errorf("Randomizer not provided")
	}

	binsLength := len(bins)
	if bins == nil || binsLength < 1 {
		return nil, nil
	}

	selectQty := getSelectQty(min, max, binsLength, rnd)
	if selectQty < 1 {
		return nil, nil
	}

	if selectQty > binsLength {
		return nil, errors.Errorf("select quantity<%d> calculated to larger than amount of bins<%d>",
			selectQty, binsLength)
	}

	if binsLength == selectQty {
		return bins, nil
	}

	selectPos := make(uniqueInts)
	failSafe := 0
	for len(selectPos) < selectQty {
		pos := rnd.Rand(binsLength)
		if selectPos.HasValue(pos) {
			failSafe++
			if failSafe > 10000 {
				return nil, errors.Errorf("Error randomizing positions, failsafe<10000> limit reached")
			}
			continue
		}
		selectPos.AddValue(pos)
	}

	positions := selectPos.Array()
	selectBins := make([]string, len(positions))
	for i, v := range positions {
		selectBins[i] = bins[v]
	}

	return selectBins, nil
}

func cutPosition(index int, slice *[]int) error {
	if slice == nil || *slice == nil || len(*slice) < 1 {
		return errors.Errorf("empty slice")
	}

	if index < 0 || index > (len(*slice)-1) {
		return errors.Errorf("index out of bounds")
	}

	*slice = append((*slice)[:index], (*slice)[(index+1):]...)
	return nil
}

//getPossible returns []possible, []bins, error
func getPossible(obj *enigmahandlers.Object, def *senseobjdef.ObjectDef, dim int,
	selectionType SelectionType, columns bool) ([]int, []string, error) {
	if selectStateHandler == nil {
		return nil, nil, errors.Errorf("No select state handler")
	}

	// TODO support pivot hypercube + maps etc
	var possible []int
	var bins []string
	var err error
	switch def.DataDef.Type {
	case senseobjdef.DataDefListObject:
		possible, err = getPossibleFromListObject(obj.ID, obj.ListObject(), dim, columns, selectionType)
	case senseobjdef.DataDefHyperCube:
		hypercube := obj.HyperCube()

		if hypercube == nil {
			return nil, nil, errors.Errorf("object<%s> has no hypercube", obj.ID)
		}

		switch hypercube.Mode {
		case constant.HyperCubeDataModePivot:
			fallthrough
		case constant.HyperCubeDataModePivotL:
			return nil, nil, errors.Errorf("Hypercube Pivot mode not supported")
		case constant.HyperCubeDataModePivotStack:
			fallthrough
		case constant.HyperCubeDataModePivotStackL:
			if selectionType.IsExcludedOrDeselect() {
				return nil, nil, errors.Errorf("selection type<%s> not supported for stacked/pivot hyper cube", selectionType.String())
			}
			possible, err = getPossibleFromStackedHyperCube(obj.ID, hypercube, dim)
		case constant.HyperCubeDataModeStraight:
			fallthrough
		case constant.HyperCubeDataModeStraightL:
			if hypercube.Binned {
				if selectionType.IsExcludedOrDeselect() {
					return nil, nil, errors.Errorf("selection type<%s> not supported for binned hyper cube", selectionType.String())
				}

				bins, err = getBinsFromStraightHyperCube(obj.ID, hypercube)
			} else {
				possible, err = getPossibleFromStraightHyperCube(obj.ID, hypercube, dim, columns, selectionType)
			}
		case constant.HyperCubeDataModeTree:
			fallthrough
		case constant.HyperCubeDataModeTreeL:
			return nil, nil, errors.Errorf("Hypercube tree mode not supported")
		default:
			return nil, nil, errors.Errorf("Hypercube mode<%d> not supported", def.DataDef.Type)
		}
	default:
		return nil, nil, errors.Errorf("DataDef type<%d> not implemented", def.DataDef.Type)
	}
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return possible, bins, nil
}

func getMatrix(datapages []*enigma.NxDataPage, dim int, columns bool) ([]enigma.NxCellRows, int, error) {
	var matrix []enigma.NxCellRows
	if columns {
		if dim >= len(datapages) {
			return nil, 0, errors.Errorf("No datapage for column<%d>", dim)
		}
		matrix = datapages[dim].Matrix
		dim = 0
	} else {
		matrix = datapages[0].Matrix
	}
	return matrix, dim, nil
}

func getHypercubeDataPages(id string, hypercube *enigmahandlers.HyperCube) ([]*enigma.NxDataPage, error) {
	if hypercube == nil {
		return nil, errors.Errorf("object<%s> has no hypercube", id)
	}
	dataPages := hypercube.DataPages
	if dataPages == nil || len(dataPages) < 1 {
		return nil, errors.Errorf("object<%s>  has no datapages", id)
	}
	return dataPages, nil
}

func getHypercubeStackedDataPages(id string, hypercube *enigmahandlers.HyperCube) ([]*enigma.NxStackPage, error) {
	if hypercube == nil {
		return nil, errors.Errorf("object<%s> has no hypercube", id)
	}
	dataPages := hypercube.StackedDataPages
	if dataPages == nil || len(dataPages) < 1 {
		return nil, errors.Errorf("object<%s>  has no datapages", id)
	}
	return dataPages, nil
}

func getListObjectDataPages(id string, listObject *enigma.ListObject) ([]*enigma.NxDataPage, error) {
	if listObject == nil {
		return nil, errors.Errorf("object<%s> has no listobject", id)
	}

	dataPages := listObject.DataPages
	if dataPages == nil || len(dataPages) < 1 {
		return nil, errors.Errorf("object<%s> has no datapages", id)
	}
	return dataPages, nil
}

func getPossibleFromMatrix(matrix []enigma.NxCellRows, id string, dim int, stype SelectionType, binned bool) ([]int, error) {
	if matrix == nil || len(matrix) < 1 {
		return nil, errors.Errorf("object<%s> matrix has no rows", id)
	}

	possibleMap := make(uniqueInts)

	for ri, row := range matrix {
		if len(row) < dim+1 {
			return nil, errors.Errorf("object<%s> matrix row<%d> doesn't have requested dim<%d>", id, ri, dim)
		}

		cell := row[dim]
		if cell != nil {
			state, err := selectStateHandler.Int(strings.ToLower(cell.State))
			if err != nil {
				return nil, errors.Wrapf(err, "object<%s> row<%d> dim<%d> has unknown state<%s>", id, ri, dim, cell.State)
			}

			switch stype {
			case RandomFromEnabled:
				if selectStates(state).isEnabled(binned) {
					possibleMap.AddValue(cell.ElemNumber)
				}
			case RandomDeselect:
				if selectStates(state).isSelected() {
					possibleMap.AddValue(cell.ElemNumber)
				}
			case RandomFromExcluded:
				if selectStates(state).isExcluded() {
					possibleMap.AddValue(cell.ElemNumber)
				}
			default:
				stypString, errStyp := stype.GetEnumMap().String(int(stype))
				if errStyp != nil {
					stypString = fmt.Sprintf("%d", stype)
				}
				return nil, errors.Errorf("Unsupported selection type<%s>", stypString)
			}
		}
	}

	return possibleMap.Array(), nil
}

func getPossibleFromStackedHyperCube(id string, hypercube *enigmahandlers.HyperCube, dim int) ([]int, error) {
	dataPages, err := getHypercubeStackedDataPages(id, hypercube)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if dataPages == nil || len(dataPages) < 1 {
		return nil, errors.Errorf("object<%s> Stacked hypercube contains no datapages", id)
	}

	var possibleMap uniqueInts
	for _, page := range dataPages {
		if err = getDataFromNxStackPage(id, page, dim, &possibleMap); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return possibleMap.Array(), nil
}

func getDataFromNxStackPage(id string, page *enigma.NxStackPage, dim int, possibleMap *uniqueInts) error {
	if page == nil || (*page).Data == nil { //No data in page
		return nil
	}

	rootLength := len((*page).Data)
	if rootLength == 0 { //No data in page
		return nil
	}

	if rootLength != 1 || (*page).Data[0] == nil || (*page).Data[0].Type != constant.NxDimCellRoot { //Malformed stack
		return errors.Errorf("object<%s> has malformed stacked hypercube", id)
	}

	recursiveDataFromStackedPivotCell((*page).Data[0], -1, dim, possibleMap)
	return nil
}

func recursiveDataFromStackedPivotCell(cell *enigma.NxStackedPivotCell, currentDim, getDim int, possibleMap *uniqueInts) {
	if cell == nil {
		return
	}

	if currentDim == getDim {
		if cell.Type != constant.NxDimCellNormal {
			return
		}
		possibleMap.AddValue(cell.ElemNo)
		return
	}

	if cell.SubNodes != nil && len(cell.SubNodes) > 1 {
		for _, subCell := range cell.SubNodes {
			recursiveDataFromStackedPivotCell(subCell, currentDim+1, getDim, possibleMap)
		}
	}
}

func getPossibleFromStraightHyperCube(id string, hypercube *enigmahandlers.HyperCube, dim int, columns bool, stype SelectionType) ([]int, error) {
	dataPages, err := getHypercubeDataPages(id, hypercube)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	matrix, useDim, err := getMatrix(dataPages, dim, columns)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	possible, err := getPossibleFromMatrix(matrix, id, useDim, stype, hypercube.Binned)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return possible, nil
}

func getPossibleFromListObject(id string, listobject *enigma.ListObject, dim int, columns bool, stype SelectionType) ([]int, error) {
	dataPages, err := getListObjectDataPages(id, listobject)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	matrix, useDim, err := getMatrix(dataPages, dim, columns)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	possible, err := getPossibleFromMatrix(matrix, id, useDim, stype, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return possible, nil
}

func getBinsFromStraightHyperCube(id string, hypercube *enigmahandlers.HyperCube) ([]string, error) {
	dataPages, err := getHypercubeDataPages(id, hypercube)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	matrix, _, err := getMatrix(dataPages, 0, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	bins := make([]string, len(matrix))
	bins = bins[:0]
	for _, row := range matrix {
		if row == nil || len(row) < 1 {
			continue
		}

		cell := row[0]

		//Simple bin checks
		if cell == nil || cell.Text == "" {
			continue
		}
		if []rune(cell.Text)[0] != '[' {
			continue
		}

		bins = append(bins, cell.Text)
	}

	return bins, nil
}

func verifyDimension(id string, dim int, dimensionList []*enigma.NxDimensionInfo) error {
	if dimensionList == nil || len(dimensionList) < 1 {
		return errors.Errorf("object<%s> does not contain a dimensionlist", id)
	}

	if len(dimensionList) < dim+1 {
		return errors.Errorf("object<%s> does not have requested dimension<%d>", id, dim)
	}

	dimension := dimensionList[dim]
	if dimension == nil {
		return errors.Errorf("requested object<%s> dimension<%d> is nil", id, dim)
	}

	// TODO check dimension.DimensionType?
	return nil
}

func convertBinsToSelectInfo(bins []string) ([]*enigma.NxMultiRangeSelectInfo, error) {
	selectInfoArray := make([]*enigma.NxMultiRangeSelectInfo, len(bins))
	selectInfoArray = selectInfoArray[:0]
	for _, bin := range bins {
		selectInfo, err := convertBinToSelectInfo(bin)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		selectInfoArray = append(selectInfoArray, selectInfo)
	}

	return selectInfoArray, nil
}

func convertBinToSelectInfo(bin string) (*enigma.NxMultiRangeSelectInfo, error) {
	//Example
	//"[2.828125,3.031250,3.031250,2.828125]"

	//Check if a valid bin
	runes := []rune(bin)
	if runes == nil || len(runes) < 9 || runes[0] != '[' || runes[len(runes)-1] != ']' {
		return nil, errors.Errorf("Invalid bin<%s>", bin)
	}
	bin = strings.Trim(bin, "[]")
	coords := strings.Split(bin, ",")
	if coords == nil || len(coords) != 4 {
		return nil, errors.Errorf("Bin<%s> does not contain 4 values", bin)
	}

	//Parse to float64
	rect := make([]float64, 4)
	for i, v := range coords {
		var err error
		rect[i], err = strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed parsing bin:%d from bin<%s> to float64", i, bin)
		}
	}

	selectInfo := enigma.NxMultiRangeSelectInfo{
		ColumnsToSelect: []int{},
		Ranges: []*enigma.NxRangeSelectInfo{{
			MeasureIx: 0,
			Range: &enigma.Range{
				Min:       enigma.Float64(math.Min(rect[0], rect[2])),
				Max:       enigma.Float64(math.Max(rect[0], rect[2])),
				MinInclEq: true,
				MaxInclEq: true,
			},
		}, {
			MeasureIx: 1,
			Range: &enigma.Range{
				Min:       enigma.Float64(math.Min(rect[1], rect[3])),
				Max:       enigma.Float64(math.Max(rect[1], rect[3])),
				MinInclEq: true,
				MaxInclEq: true,
			},
		}},
	}

	return &selectInfo, nil
}
