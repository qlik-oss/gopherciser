package scenario

import (
	"context"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/session"
)

// TODO support term from file
// TODO support session variables
// TODO supprt field search

type (
	ObjectSearchType int
	// ObjectSearchSettings ObjectSearch search listbox or field
	ObjectSearchSettings struct {
		ID           string           `json:"id" doc-key:"objectsearch.id"`
		SearchTerm   string           `json:"searchterm" doc-key:"objectsearch.searchterm"`
		SearchType   ObjectSearchType `json:"type" doc-key:"objectsearch.type"`
		ErrorOnEmpty bool             `json:"erroronempty" doc-key:"objectsearch.erroronempty"`
	}
)

const (
	ObjectSearchTypeListbox ObjectSearchType = iota
	ObjectSearchTypeField
)

var objectSearchTypeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
	"listbox": int(ObjectSearchTypeListbox),
	"field":   int(ObjectSearchTypeField),
})

// UnmarshalJSON unmarshal objectsearch type
func (value *ObjectSearchType) UnmarshalJSON(arg []byte) error {
	i, err := objectSearchTypeEnumMap.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal ObjectSearchType")
	}

	*value = ObjectSearchType(i)
	return nil
}

// MarshalJSON marshal objectsearch type
func (value ObjectSearchType) MarshalJSON() ([]byte, error) {
	str, err := objectSearchTypeEnumMap.String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown ObjectSearchType<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// String representation of ObjectSearchType
func (value ObjectSearchType) String() string {
	sType, err := objectSearchTypeEnumMap.String(int(value))
	if err != nil {
		return strconv.Itoa(int(value))
	}
	return sType
}

// GetEnumMap returns objectsearch type enum map to GUI
func (value ObjectSearchType) GetEnumMap() *enummap.EnumMap {
	return objectSearchTypeEnumMap
}

// Validate ObjectSearchSettings action (Implements ActionSettings interface)
func (settings ObjectSearchSettings) Validate() ([]string, error) {
	if settings.ID == "" {
		return nil, errors.Errorf("no id defined in  %s", ActionObjectSearch)
	}
	if settings.SearchTerm == "" {
		return nil, errors.Errorf("no search term defined in %s", ActionObjectSearch)
	}
	return nil, nil
}

// Execute ObjectSearchSettings action (Implements ActionSettings interface)
func (settings ObjectSearchSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	uplink := sessionState.Connection.Sense()

	var genObj *enigma.GenericObject
	var getDimInfo func() (*enigma.NxDimensionInfo, error)
	switch settings.SearchType {
	case ObjectSearchTypeField:
		app, err := sessionState.CurrentSenseApp()
		if err != nil {
			actionState.AddErrors(err)
			return
		}
		fieldList, err := app.GetFieldList(sessionState, actionState)
		if err != nil {
			actionState.AddErrors(err)
			return
		}
		if fieldList.Layout() == nil || fieldList.Layout().FieldList == nil {
			actionState.NewErrorf("no fieldlist layout")
			return
		}

		listbox, err := fieldList.GetOrCreateSessionListboxSync(sessionState, actionState, uplink.CurrentApp.Doc, settings.ID)
		if err != nil {
			// TODO if FieldNotFoundError, check dimensionlist instead
			actionState.AddErrors(err)
			return
		}
		genObj = listbox.EnigmaObject

		// Create 7 listboxes from fields and dimensions to simulate opening "selectors"?

		getDimInfo = func() (*enigma.NxDimensionInfo, error) {
			var dimInfo *enigma.NxDimensionInfo
			if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
				listObject, err := listbox.ListObject(ctx)
				if err != nil {
					return errors.WithStack(err)
				}
				dimInfo = listObject.DimensionInfo
				return nil
			}); err != nil {
				return nil, err
			}

			return dimInfo, nil
		}
	case ObjectSearchTypeListbox:
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
			genObj = gob.EnigmaObject.(*enigma.GenericObject)
		default:
			actionState.AddErrors(errors.Errorf("Unknown object type<%T>", t))
			return
		}

		getDimInfo = func() (*enigma.NxDimensionInfo, error) {
			listobject := gob.ListObject()
			if listobject == nil {
				return nil, errors.Errorf("listobject is nil")
			}
			return listobject.DimensionInfo, nil
		}
	}

	objInstance := sessionState.GetObjectHandlerInstance(genObj.GenericId, genObj.GenericType)
	selectPath, _, _, err := objInstance.GetObjectDefinition(genObj.GenericType)
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to get object definition"))
		return
	}

	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		return genObj.BeginSelections(ctx, []string{selectPath})
	}); err != nil {
		actionState.AddErrors(err)
	}

	searchChunks, err := newSearchTextChunks(sessionState.Randomizer(), settings.SearchTerm, false)
	if err != nil {
		actionState.AddErrors(err)
		return
	}
	termChan := searchChunks.simulate(sessionState.BaseContext(), func(errors ...error) {
		for _, err := range errors {
			actionState.AddErrors(err)
		}
	})

	for i := 0; i < cap(termChan)-1; i++ {
		doSearch(sessionState, actionState, genObj, selectPath, <-termChan, false)
	}
	doSearch(sessionState, actionState, genObj, selectPath, <-termChan, true)

	if sessionState.Wait(actionState) {
		return // an error occured
	}

	dimInfo, err := getDimInfo()
	if err != nil {
		actionState.AddErrors(err)
		return
	}
	if dimInfo == nil {
		actionState.AddErrors(errors.New("listobject dimension info is nil"))
		return
	}

	accept := true
	if dimInfo.Cardinal < 1 {
		accept = false
		if settings.ErrorOnEmpty {
			actionState.AddErrors(errors.New("no search results found"))
			return
		}
	} else {
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			return genObj.AcceptListObjectSearch(ctx, selectPath, true, false)
		})
		if err != nil {
			actionState.AddErrors(err)
			return
		}
	}
	// yes, abort is sent also after accept
	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		return genObj.AbortListObjectSearch(ctx, selectPath)
	}); err != nil {
		actionState.AddErrors(err)
		return
	}
	// end modal mode
	sessionState.QueueRequest(func(ctx context.Context) error {
		return genObj.EndSelections(ctx, accept)
	}, actionState, true, "end selections failed")

	sessionState.Wait(actionState) // Await all async requests, e.g. those triggered on changed objects
}

func doSearch(sessionState *session.State, actionState *action.State, genObj *enigma.GenericObject, path, term string, report bool) {
	sessionState.QueueRequest(func(ctx context.Context) error {
		success, err := genObj.SearchListObjectFor(ctx, path, term)
		if report {
			if err != nil {
				return err
			}
			if !success {
				return errors.Errorf("search was unsuccessful")
			}
		}
		return nil

	}, actionState, true, "object search failed")
}
