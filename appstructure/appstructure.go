package appstructure

import (
	"encoding/json"

	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/enummap"
)

type (

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

	// AppStructureObjectChildren substructure adding children
	AppStructureObjectChildren struct {
		// Map of children to the sense object
		Map map[string]string `json:"children,omitempty"`
	}

	// AppStructureObject sense object structure
	AppStructureObject struct {
		AppObjectDef
		MetaDef
		AppStructureObjectChildren
		// RawBaseProperties of Sense object
		RawBaseProperties json.RawMessage `json:"rawBaseProperties,omitempty"`
		// RawExtendedProperties of extended Sense object
		RawExtendedProperties json.RawMessage `json:"rawExtendedProperties,omitempty"`
		// RawGeneratedProperties inner generated properties of auto-chart
		RawGeneratedProperties json.RawMessage `json:"rawGeneratedProperties,omitempty"`
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
		// RawProperties of Bookmark object
		RawProperties json.RawMessage `json:"rawProperties,omitempty"`
	}

	// AppStructureStoryObject list of objects used in stories
	AppStructureStoryObject struct {
		AppObjectDef
		AppStructureObjectChildren
		// RawProperties of Sense object
		RawProperties json.RawMessage `json:"rawProperties,omitempty"`
		// Visualization visualization of object, if exists
		Visualization string `json:"visualization,omitempty"`
		// SnapshotID of linked object snapshot object
		SnapshotID string `json:"snapshotid,omitempty"`
		// RawSnapShotProperties of extended snapshot object
		RawSnapShotProperties json.RawMessage `json:"rawSnapshotProperties,omitempty"`
	}

	// AppStructureField list of fields in the app
	AppStructureField struct {
		enigma.NxFieldDescription
	}

	// AppStructure of Sense app
	AppStructure struct {
		AppMeta AppStructureAppMeta `json:"meta"`
		// Objects in Sense app
		Objects map[string]AppStructureObject `json:"objects"`
		// Bookmark list of bookmarks in the app
		Bookmarks map[string]AppStructureBookmark `json:"bookmarks"`
		// Fields list of all fields in the app
		Fields map[string]AppStructureField `json:"fields"`
		// StoryObjects
		StoryObjects map[string]AppStructureStoryObject `json:"storyobjects"`
	}

	// AppStructurePopulatedObjects is the type returned by an action when prompted for selectable objects
	AppStructurePopulatedObjects struct {
		// Parent id of the parent object
		Parent string
		// Objects first level app objects returned by the current action
		Objects []AppStructureObject
		// Bookmark bookmarks returned by the current action
		Bookmarks []AppStructureBookmark
	}

	ObjectType                         int
	AppStructureObjectNotFoundError    string
	AppStructureNoScenarioActionsError struct{}
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

	// Objects connected to snapshots and stories
	ObjectSnapshotList
	ObjectSnapshot
	ObjectEmbeddedSnapshot
	ObjectStory
	ObjectSlide
	ObjectSlideItem
)

var (
	// ObjectTypeEnumMap enum of known object types which needs special handling
	ObjectTypeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
		"dimension":        int(ObjectTypeDimension),
		"measure":          int(ObjectTypeMeasure),
		"bookmark":         int(ObjectTypeBookmark),
		"masterobject":     int(ObjectTypeMasterObject),
		"auto-chart":       int(ObjectTypeAutoChart),
		"sheet":            int(ObjectSheet),
		"loadmodel":        int(ObjectLoadModel),
		"appprops":         int(ObjectAppprops),
		"snapshotlist":     int(ObjectSnapshotList),
		"snapshot":         int(ObjectSnapshot),
		"embeddedsnapshot": int(ObjectEmbeddedSnapshot),
		"story":            int(ObjectStory),
		"slide":            int(ObjectSlide),
		"slideitem":        int(ObjectSlideItem),
	})
)

// Error object was not found in app structure
func (err AppStructureObjectNotFoundError) Error() string {
	return string(err)
}

// Error no applicable actions found in scenario
func (err AppStructureNoScenarioActionsError) Error() string {
	return "no applicable actions in scenario"
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

	for id := range obj.Map {
		child, ok := structure.Objects[id]
		if !ok {
			continue
		}

		selectableChildren := structure.addSelectableChildren(child)
		selectables = append(selectables, selectableChildren...)
	}
	return selectables
}
