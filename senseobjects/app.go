package senseobjects

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
)

type (
	// SessionObjects for the app
	SessionObjects struct {
		sheetList         *SheetList
		bookmarkList      *BookmarkList
		currentSelections *CurrentSelections
		localeInfo        *enigma.LocaleInfo
		variablelist      *VariableList
		storylist         *StoryList
		loadmodellist     *LoadModelList
		fieldlist         *FieldList
		dimensionList     *DimensionList
		dimensions        map[string]*enigma.GenericDimension
		appPropsList      *AppPropsList
	}

	// App sense app object
	App struct {
		GUID      string
		Doc       *enigma.Doc
		Layout    *enigma.NxAppLayout
		bookmarks map[string]*enigma.GenericBookmark
		mutex     sync.Mutex
		SessionObjects
	}
)

// readLocked returns the value of an App field while holding the app mutex
func readLocked[T any](app *App, field *T) T {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	return *field
}

// GetSheetList update sheet list for app
func (app *App) GetSheetList(sessionState SessionState, actionState *action.State) (*SheetList, error) {
	if sheetList := readLocked(app, &app.sheetList); sheetList != nil {
		return sheetList, nil
	}

	// update sheetList to latest
	updateSheetList := func(ctx context.Context) error {
		sl, err := CreateSheetListObject(ctx, app.Doc)
		if err != nil {
			return err
		}
		app.setSheetList(sessionState, sl)
		return err
	}
	if err := sessionState.SendRequest(actionState, updateSheetList); err != nil {
		return nil, errors.WithStack(err)
	}

	sheetList := readLocked(app, &app.sheetList)

	if err := sessionState.SendRequest(actionState, sheetList.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	// update sheetList layout when sheetList has a change event
	onSheetListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(sheetList.UpdateLayout(ctx))
	}

	sessionState.RegisterEvent(sheetList.enigmaObject.Handle,
		onSheetListChanged, nil, true)

	return sheetList, nil
}

func (app *App) setSheetList(sessionState SessionState, sl *SheetList) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.sheetList != nil && app.sheetList.enigmaObject != nil && app.sheetList.enigmaObject.Handle > 0 && sl != app.sheetList {
		sessionState.LogDebugf("senseobjects.app.setSheetList: executing DeRegisterEvent on handle<%d>", app.GUID, app.sheetList.enigmaObject.Handle)
		sessionState.DeRegisterEvent(app.sheetList.enigmaObject.Handle)
	}
	app.sheetList = sl
}

// GetBookmarkList update bookmark list for app
func (app *App) GetBookmarkList(sessionState SessionState, actionState *action.State) (*BookmarkList, error) {
	if bookmarkList := readLocked(app, &app.bookmarkList); bookmarkList != nil {
		return bookmarkList, nil
	}

	// create bookmark list
	createBookmarkList := func(ctx context.Context) error {
		bl, err := CreateBookmarkListObject(ctx, app.Doc)
		if err != nil {
			return err
		}
		app.setBookmarkList(sessionState, bl)
		return err
	}

	if err := sessionState.SendRequest(actionState, createBookmarkList); err != nil {
		return nil, errors.WithStack(err)
	}

	bookmarkList := readLocked(app, &app.bookmarkList)

	if err := sessionState.SendRequest(actionState, bookmarkList.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sessionState.SendRequest(actionState, bookmarkList.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}

	// update bookmark list layout when bookmark list has a change event
	onBookmarkListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(bookmarkList.UpdateLayout(ctx))
	}

	sessionState.RegisterEvent(bookmarkList.enigmaObject.Handle,
		onBookmarkListChanged, nil, true)

	return bookmarkList, nil
}

func (app *App) setBookmarkList(sessionState SessionState, bl *BookmarkList) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.bookmarkList != nil && app.bookmarkList.enigmaObject != nil && app.bookmarkList.enigmaObject.Handle > 0 && bl != app.bookmarkList {
		sessionState.DeRegisterEvent(app.bookmarkList.enigmaObject.Handle)
	}
	app.bookmarkList = bl
}

// GetVariableList create or return existing variable list session object
func (app *App) GetVariableList(sessionState SessionState, actionState *action.State) (*VariableList, error) {
	if variableList := readLocked(app, &app.variablelist); variableList != nil {
		return variableList, nil
	}

	// create variable list
	createVariableList := func(ctx context.Context) error {
		vl, err := CreateVariableListObject(ctx, app.Doc)
		if err != nil {
			return err
		}
		app.setVariableList(sessionState, vl)
		return err
	}

	if err := sessionState.SendRequest(actionState, createVariableList); err != nil {
		return nil, errors.WithStack(err)
	}

	variableList := readLocked(app, &app.variablelist)

	if err := sessionState.SendRequest(actionState, variableList.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	// update variable list layout when variable list has a change event
	onVariableListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(variableList.UpdateLayout(ctx))
	}
	sessionState.RegisterEvent(variableList.enigmaObject.Handle, onVariableListChanged, nil, true)

	return variableList, nil
}

func (app *App) setVariableList(sessionState SessionState, vl *VariableList) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.variablelist != nil && app.variablelist.enigmaObject != nil && app.variablelist.enigmaObject.Handle > 0 && vl != app.variablelist {
		sessionState.DeRegisterEvent(app.variablelist.enigmaObject.Handle)
	}
	app.variablelist = vl
}

// GetAppsPropsList create or return AppsPropsList
func (app *App) GetAppsPropsList(sessionState SessionState, actionState *action.State) (*AppPropsList, error) {
	if appPropsList := readLocked(app, &app.appPropsList); appPropsList != nil {
		return appPropsList, nil
	}

	createAppPropsList := func(ctx context.Context) error {
		al, err := CreateAppPropsListObject(ctx, app.Doc)
		if err != nil {
			return err
		}
		app.setAppPropsList(sessionState, al)
		return nil
	}

	if err := sessionState.SendRequest(actionState, createAppPropsList); err != nil {
		return nil, errors.WithStack(err)
	}

	appPropsList := readLocked(app, &app.appPropsList)

	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		return appPropsList.UpdateLayout(ctx, app.Doc, sessionState, actionState)
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	onAppPropsListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(appPropsList.UpdateLayout(ctx, app.Doc, sessionState, actionState))
	}
	sessionState.RegisterEvent(appPropsList.enigmaObject.Handle, onAppPropsListChanged, nil, true)

	return appPropsList, nil
}

func (app *App) setAppPropsList(sessionState SessionState, al *AppPropsList) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.appPropsList != nil && app.appPropsList.enigmaObject != nil && app.appPropsList.enigmaObject.Handle > 0 && al != app.appPropsList {
		sessionState.DeRegisterEvent(app.appPropsList.enigmaObject.Handle)
		app.appPropsList.RemoveAllItems(sessionState)
	}
	app.appPropsList = al
}

// GetStoryList create or return existing story list session object
func (app *App) GetStoryList(sessionState SessionState, actionState *action.State) (*StoryList, error) {
	if storyList := readLocked(app, &app.storylist); storyList != nil {
		return storyList, nil
	}

	// create story list
	createStoryList := func(ctx context.Context) error {
		sl, err := CreateStoryListObject(ctx, app.Doc)
		if err != nil {
			return err
		}
		app.setStoryList(sessionState, sl)
		return err
	}

	if err := sessionState.SendRequest(actionState, createStoryList); err != nil {
		return nil, errors.WithStack(err)
	}

	storyList := readLocked(app, &app.storylist)

	if err := sessionState.SendRequest(actionState, storyList.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	// update story list layout when story list has a change event
	onStoryListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(storyList.UpdateLayout(ctx))
	}
	sessionState.RegisterEvent(storyList.enigmaObject.Handle, onStoryListChanged, nil, true)

	return storyList, nil
}

func (app *App) setStoryList(sessionState SessionState, sl *StoryList) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.storylist != nil && app.storylist.enigmaObject != nil && app.storylist.enigmaObject.Handle > 0 && sl != app.storylist {
		sessionState.DeRegisterEvent(app.storylist.enigmaObject.Handle)
	}
	app.storylist = sl
}

// GetLoadModelList create or return existing load model list session object
func (app *App) GetLoadModelList(sessionState SessionState, actionState *action.State) (*LoadModelList, error) {
	if loadModelList := readLocked(app, &app.loadmodellist); loadModelList != nil {
		return loadModelList, nil
	}

	// create load model list
	createLoadModelList := func(ctx context.Context) error {
		lml, err := CreateLoadModelListObject(ctx, app.Doc)
		if err != nil {
			return err
		}
		app.setLoadModelList(sessionState, lml)
		return err
	}

	if err := sessionState.SendRequest(actionState, createLoadModelList); err != nil {
		return nil, errors.WithStack(err)
	}

	loadModelList := readLocked(app, &app.loadmodellist)

	if err := sessionState.SendRequest(actionState, loadModelList.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sessionState.SendRequest(actionState, loadModelList.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}

	// update load model list layout when load model list has a change event
	onLoadModelListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(loadModelList.UpdateLayout(ctx))
	}
	sessionState.RegisterEvent(loadModelList.enigmaObject.Handle, onLoadModelListChanged, nil, true)

	return loadModelList, nil
}

func (app *App) setLoadModelList(sessionState SessionState, lml *LoadModelList) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.loadmodellist != nil && app.loadmodellist.enigmaObject != nil && app.loadmodellist.enigmaObject.Handle > 0 && lml != app.loadmodellist {
		sessionState.DeRegisterEvent(app.loadmodellist.enigmaObject.Handle)
	}
	app.loadmodellist = lml
}

// GetBookmarkObject with ID
func (app *App) GetBookmarkObject(sessionState SessionState, actionState *action.State, id string) (*enigma.GenericBookmark, error) {
	// Ge id from map of bookmarks
	bm := app.getBookmarkFromMap(id)
	if bm != nil {
		return bm, nil
	}

	// Bookmark object not in map, get it
	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		var err error
		bm, err = app.Doc.GetBookmark(ctx, id)
		if err != nil {
			return errors.Wrap(err, "failed to get bookmark object")
		}
		if _, err := bm.GetLayout(ctx); err != nil {
			return errors.Wrap(err, "failed to get bookmark layout")
		}

		app.addBookmarkToMap(bm)

		// Update data when object gets a changed event
		onBookmarkChanged := func(ctx context.Context, actionState *action.State) error {
			_, err := bm.GetLayout(ctx)
			return errors.WithStack(err)
		}

		// remove from list when event gets de-registered
		onEventDeregister := func() {
			if app == nil {
				return
			}
			app.mutex.Lock()
			defer app.mutex.Unlock()
			if app.bookmarks != nil {
				delete(app.bookmarks, id)
			}
		}
		sessionState.RegisterEvent(bm.Handle, onBookmarkChanged, onEventDeregister, true)
		return err
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	return bm, nil
}

func (app *App) getBookmarkFromMap(id string) *enigma.GenericBookmark {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.bookmarks == nil {
		app.bookmarks = make(map[string]*enigma.GenericBookmark)
	}
	return app.bookmarks[id]
}

func (app *App) addBookmarkToMap(bm *enigma.GenericBookmark) {
	if bm == nil {
		return
	}
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.bookmarks == nil {
		app.bookmarks = make(map[string]*enigma.GenericBookmark)
	}
	app.bookmarks[bm.GenericId] = bm
}

func (app *App) GetAggregatedSelectionFields() string {
	currentSelections := readLocked(app, &app.currentSelections)
	fields := make([]string, 0, len(currentSelections.layout.SelectionObject.Selections))
	for _, selection := range currentSelections.layout.SelectionObject.Selections {
		fields = append(fields, selection.Field)
	}

	return strings.Join(fields, ",")
}

func (app *App) GetDimensionList(sessionState SessionState, actionState *action.State) (*DimensionList, error) {
	if dimensionList := readLocked(app, &app.dimensionList); dimensionList != nil {
		return dimensionList, nil
	}
	updateDimensionList := func(ctx context.Context) error {
		dl, err := CreateDimensionListObject(ctx, app.Doc)
		if err != nil {
			return err
		}
		app.setDimensionList(sessionState, dl)
		return nil
	}
	if err := sessionState.SendRequest(actionState, updateDimensionList); err != nil {
		return nil, errors.WithStack(err)
	}

	dimensionList := readLocked(app, &app.dimensionList)

	// Get data
	if err := sessionState.SendRequest(actionState, dimensionList.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := sessionState.SendRequest(actionState, dimensionList.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}

	// setup automatic data update
	onDimensionListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(dimensionList.UpdateLayout(ctx))
	}
	sessionState.RegisterEvent(dimensionList.enigmaObject.Handle, onDimensionListChanged, nil, true)

	return dimensionList, nil

}

// GetFieldList session object containing list of fields
func (app *App) GetFieldList(sessionState SessionState, actionState *action.State) (*FieldList, error) {
	if fieldList := readLocked(app, &app.fieldlist); fieldList != nil {
		return fieldList, nil
	}

	// Create session object
	updateFieldList := func(ctx context.Context) error {
		fl, err := CreateFieldListObject(ctx, app.Doc)
		if err != nil {
			return err
		}
		app.setFieldList(sessionState, fl)
		return nil
	}
	if err := sessionState.SendRequest(actionState, updateFieldList); err != nil {
		return nil, errors.WithStack(err)
	}

	fieldList := readLocked(app, &app.fieldlist)

	// Get data
	if err := sessionState.SendRequest(actionState, fieldList.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := sessionState.SendRequest(actionState, fieldList.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}

	// setup automatic data update
	onFieldListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(fieldList.UpdateLayout(ctx))
	}
	sessionState.RegisterEvent(fieldList.enigmaObject.Handle, onFieldListChanged, nil, true)

	return fieldList, nil
}

// GetCurrentSelections create current selection session object and add to list
func (app *App) GetCurrentSelections(sessionState SessionState, actionState *action.State, requestData bool) (*CurrentSelections, error) {
	if currentSelections := readLocked(app, &app.currentSelections); currentSelections != nil {
		if requestData && currentSelections.enigmaObject != nil {
			if f := sessionState.GetEventFunc(currentSelections.enigmaObject.Handle); f != nil {
				if err := f(sessionState.BaseContext(), actionState); err != nil {
					return currentSelections, errors.WithStack(err)
				}
			}
		}
		return currentSelections, nil
	}

	// Create session object
	updateCurrentSelections := func(ctx context.Context) error {
		cs, err := CreateCurrentSelections(ctx, app.Doc)
		if err != nil {
			return err
		}
		app.setCurrentSelections(sessionState, cs)
		return nil
	}
	if err := sessionState.SendRequest(actionState, updateCurrentSelections); err != nil {
		return nil, errors.WithStack(err)
	}

	currentSelections := readLocked(app, &app.currentSelections)

	var once sync.Once
	// update currentSelection layout when object is changed
	onCurrentSelectionChanged := func(ctx context.Context, actionState *action.State) error {
		var err error
		once.Do(func() {
			err = sessionState.SendRequest(actionState, currentSelections.UpdateProperties)
		})
		if err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(currentSelections.UpdateLayout(ctx))
	}

	if requestData {
		if err := onCurrentSelectionChanged(sessionState.BaseContext(), actionState); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	sessionState.RegisterEvent(currentSelections.enigmaObject.Handle, onCurrentSelectionChanged, nil, true)

	return currentSelections, nil
}

// GetLocaleInfo send get locale info request
func (app *App) GetLocaleInfo(ctx context.Context) (*enigma.LocaleInfo, error) {
	if localeInfo := readLocked(app, &app.localeInfo); localeInfo != nil {
		return localeInfo, nil
	}

	localeInfo, err := app.Doc.GetLocaleInfo(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	app.mutex.Lock()
	defer app.mutex.Unlock()
	app.localeInfo = localeInfo

	return localeInfo, nil
}

func (app *App) setCurrentSelections(sessionState SessionState, cs *CurrentSelections) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.currentSelections != nil && app.currentSelections.enigmaObject != nil && app.currentSelections.enigmaObject.Handle > 0 && cs != app.currentSelections {
		sessionState.DeRegisterEvent(app.currentSelections.enigmaObject.Handle)
	}
	app.currentSelections = cs
}

func (app *App) setFieldList(SessionState SessionState, fl *FieldList) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.fieldlist != nil && app.fieldlist.enigmaObject != nil && app.fieldlist.enigmaObject.Handle > 0 && fl != app.fieldlist {
		SessionState.DeRegisterEvent(app.fieldlist.enigmaObject.Handle)
	}
	app.fieldlist = fl
}

func (app *App) setDimensionList(SessionState SessionState, dl *DimensionList) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	if app.dimensionList != nil && app.dimensionList.enigmaObject != nil && app.dimensionList.enigmaObject.Handle > 0 && dl != app.dimensionList {
		SessionState.DeRegisterEvent(app.dimensionList.enigmaObject.Handle)
	}
	app.dimensionList = dl
}

func (app *App) GetDimension(sessionState SessionState, actionState *action.State, id string) (*enigma.GenericDimension, error) {
	dim := app.getDimensionFromMap(id)
	if dim != nil {
		return dim, nil
	}

	err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		var err error
		dim, err = app.Doc.GetDimension(ctx, id)
		return err
	})
	if err != nil {
		return nil, err
	}

	app.addDimension(dim)

	return dim, nil
}

func (app *App) getDimensionFromMap(id string) *enigma.GenericDimension {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	return app.dimensions[id]
}

func (app *App) addDimension(dim *enigma.GenericDimension) {
	app.mutex.Lock()
	defer app.mutex.Unlock()

	if app.dimensions == nil {
		app.dimensions = map[string]*enigma.GenericDimension{dim.GenericId: dim}
		return
	}

	app.dimensions[dim.GenericId] = dim
}
