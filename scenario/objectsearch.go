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
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	ObjectSearchType   int
	ObjectSearchSource int

	// ObjectSearchSettings ObjectSearch search listbox, field or master dimension
	ObjectSearchSettings struct {
		ID           string             `json:"id" doc-key:"objectsearch.id"`
		SearchTerms  []string           `json:"searchterms" doc-key:"objectsearch.searchterms"`
		SearchType   ObjectSearchType   `json:"type" doc-key:"objectsearch.type"`
		SearchSource ObjectSearchSource `json:"source" doc-key:"objectsearch.source"`
		ErrorOnEmpty bool               `json:"erroronempty" doc-key:"objectsearch.erroronempty"`
		Filename     helpers.RowFile    `json:"searchtermsfile" doc-key:"objectsearch.searchtermsfile"`
	}
)

// ObjectSearchType
const (
	ObjectSearchTypeListbox ObjectSearchType = iota
	ObjectSearchTypeField
	ObjectSearchTypeDimension
)

var objectSearchTypeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
	"listbox":   int(ObjectSearchTypeListbox),
	"field":     int(ObjectSearchTypeField),
	"dimension": int(ObjectSearchTypeDimension),
})

// ObjectSearchSource
const (
	ObjectSearchSourceFromList ObjectSearchSource = iota
	ObjectSearchSourceFromFile
)

var objectSearchSourceEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
	"fromlist": int(ObjectSearchSourceFromList),
	"fromfile": int(ObjectSearchSourceFromFile),
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

// UnmarshalJSON unmarshal objectsearch source
func (value *ObjectSearchSource) UnmarshalJSON(arg []byte) error {
	i, err := objectSearchSourceEnumMap.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal ObjectSearchSource")
	}

	*value = ObjectSearchSource(i)
	return nil
}

// String implements Stringer interface
func (value ObjectSearchSource) String() string {
	return objectSearchSourceEnumMap.StringDefault(int(value), strconv.Itoa(int(value)))
}

// MarshalJSON marshal objectsearch source
func (value ObjectSearchSource) MarshalJSON() ([]byte, error) {
	str, err := objectSearchSourceEnumMap.String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown ObjectSearchSource<%d>", value)
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
		return nil, errors.Errorf("%s no id defined", ActionObjectSearch)
	}
	switch settings.SearchSource {
	case ObjectSearchSourceFromList:
		if len(settings.SearchTerms) < 1 {
			return nil, errors.Errorf("%s no search terms defined", ActionObjectSearch)
		}
	case ObjectSearchSourceFromFile:
		if settings.Filename.IsEmpty() {
			return nil, errors.Errorf("%s search source<%s> defined, but no searchtermsfile<%s> found or no filename set",
				ActionObjectSearch, settings.SearchSource, settings.Filename)
		}
		if len(settings.Filename.Rows()) < 1 {
			return nil, errors.Errorf("%s search source<%s> defined, but searchtermsfile<%s> contains no search terms",
				ActionObjectSearch, settings.SearchSource, settings.Filename)
		}
	default:
		return nil, errors.Errorf("%s source<%s> not supported", ActionObjectSearch, settings.SearchSource)
	}
	return nil, nil
}

// Execute ObjectSearchSettings action (Implements ActionSettings interface)
func (settings ObjectSearchSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	uplink := sessionState.Connection.Sense()

	app, err := sessionState.CurrentSenseApp()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	var genObj *enigma.GenericObject
	var getDataSize func() (int, error)
	switch settings.SearchType {
	case ObjectSearchTypeField:
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
			actionState.AddErrors(err)
			return
		}
		genObj = listbox.EnigmaObject

		getDataSize = func() (int, error) {
			size := 0
			err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
				listObject, err := listbox.ListObject(ctx)
				if err != nil {
					return errors.WithStack(err)
				}

				size = getListObjectDataSize(listObject)
				return nil
			})
			return size, err
		}
	case ObjectSearchTypeDimension:
		dimensionList, err := app.GetDimensionList(sessionState, actionState)
		if err != nil {
			actionState.AddErrors(err)
			return
		}
		if dimensionList.Layout() == nil || dimensionList.Layout().DimensionList == nil {
			actionState.NewErrorf("no dimensionList layout")
			return
		}
		listbox, err := dimensionList.GetOrCreateSessionListboxSync(sessionState, actionState, uplink.CurrentApp.Doc, settings.ID)
		if err != nil {
			actionState.AddErrors(err)
			return
		}
		genObj = listbox.EnigmaObject

		getDataSize = func() (int, error) {
			size := 0
			err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
				listObject, err := listbox.ListObject(ctx)
				if err != nil {
					return errors.WithStack(err)
				}
				size = getListObjectDataSize(listObject)
				return nil
			})
			return size, err
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

		getDataSize = func() (int, error) {
			listObject := gob.ListObject()
			if listObject == nil {
				return 0, errors.Errorf("listobject is nil")
			}

			return getListObjectDataSize(listObject), nil
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

	searchTerm := ""
	switch settings.SearchSource {
	case ObjectSearchSourceFromList:
		searchTerm, err = getRandomSearchTerm(sessionState, settings.SearchTerms)
	case ObjectSearchSourceFromFile:
		searchTerm, err = getRandomSearchTerm(sessionState, settings.Filename.Rows())
	default:
		err = errors.Errorf("source<%s> not supported", settings.SearchSource)
	}
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	searchChunks, err := newSearchTextChunks(sessionState.Randomizer(), searchTerm, false)
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

	dataPageSize, err := getDataSize()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	accept := true
	if dataPageSize < 1 {
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

func getListObjectDataSize(listObject *enigma.ListObject) int {
	if listObject == nil {
		return 0
	}
	size := 0
	for _, dataPage := range listObject.DataPages {
		size += len(dataPage.Matrix)
	}
	return size
}

// getRandomSearchTerm returns a random entry from list, chosen by a uniform distribution
func getRandomSearchTerm(sessionState *session.State, terms []string) (string, error) {
	n := len(terms)
	if n < 1 {
		return "", fmt.Errorf("specified terms list is empty: Nothing to select from")
	}

	randomIndex := sessionState.Randomizer().Rand(n)
	selectedTerm := terms[randomIndex]

	return selectedTerm, nil
}
