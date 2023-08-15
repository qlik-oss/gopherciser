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

// GetSheetList update sheet list for app
func (app *App) GetSheetList(sessionState SessionState, actionState *action.State) (*SheetList, error) {
	if app.sheetList != nil {
		return app.sheetList, nil
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

	if err := sessionState.SendRequest(actionState, app.sheetList.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	// update sheetList layout when sheetList has a change event
	onSheetListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(app.sheetList.UpdateLayout(ctx))
	}

	sessionState.RegisterEvent(app.sheetList.enigmaObject.Handle,
		onSheetListChanged, nil, true)

	return app.sheetList, nil
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
	if app.bookmarkList != nil {
		return app.bookmarkList, nil
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

	if err := sessionState.SendRequest(actionState, app.bookmarkList.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sessionState.SendRequest(actionState, app.bookmarkList.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}

	// update bookmark list layout when bookmark list has a change event
	onBookmarkListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(app.bookmarkList.UpdateLayout(ctx))
	}

	sessionState.RegisterEvent(app.bookmarkList.enigmaObject.Handle,
		onBookmarkListChanged, nil, true)

	return app.bookmarkList, nil
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
	if app.variablelist != nil {
		return app.variablelist, nil
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

	if err := sessionState.SendRequest(actionState, app.variablelist.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sessionState.SendRequest(actionState, app.variablelist.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}

	// update variable list layout when variable list has a change event
	onVariableListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(app.variablelist.UpdateLayout(ctx))
	}
	sessionState.RegisterEvent(app.variablelist.enigmaObject.Handle, onVariableListChanged, nil, true)

	return app.variablelist, nil
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
	if app.appPropsList != nil {
		return app.appPropsList, nil
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

	if err := sessionState.SendRequest(actionState, app.appPropsList.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		return app.appPropsList.UpdateLayout(ctx, app.Doc, sessionState, actionState)
	}); err != nil {
		return nil, errors.WithStack(err)
	}

	onAppPropsListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(app.appPropsList.UpdateLayout(ctx, app.Doc, sessionState, actionState))
	}
	sessionState.RegisterEvent(app.appPropsList.enigmaObject.Handle, onAppPropsListChanged, nil, true)

	return app.appPropsList, nil
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
	if app.storylist != nil {
		return app.storylist, nil
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

	if err := sessionState.SendRequest(actionState, app.storylist.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	// update story list layout when story list has a change event
	onStoryListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(app.storylist.UpdateLayout(ctx))
	}
	sessionState.RegisterEvent(app.storylist.enigmaObject.Handle, onStoryListChanged, nil, true)

	return app.storylist, nil
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
	if app.loadmodellist != nil {
		return app.loadmodellist, nil
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

	if err := sessionState.SendRequest(actionState, app.loadmodellist.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sessionState.SendRequest(actionState, app.loadmodellist.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}

	// update load model list layout when load model list has a change event
	onLoadModelListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(app.loadmodellist.UpdateLayout(ctx))
	}
	sessionState.RegisterEvent(app.loadmodellist.enigmaObject.Handle, onLoadModelListChanged, nil, true)

	return app.loadmodellist, nil
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
	fields := make([]string, 0, len(app.currentSelections.layout.SelectionObject.Selections))
	for _, selection := range app.currentSelections.layout.SelectionObject.Selections {
		fields = append(fields, selection.Field)
	}

	return strings.Join(fields, ",")
}

func (app *App) GetDimensionList(sessionState SessionState, actionState *action.State) (*DimensionList, error) {
	if app.dimensionList != nil {
		return app.dimensionList, nil
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

	// Get data
	if err := sessionState.SendRequest(actionState, app.dimensionList.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := sessionState.SendRequest(actionState, app.dimensionList.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}

	// setup automatic data update
	onDimensionListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(app.dimensionList.UpdateLayout(ctx))
	}
	sessionState.RegisterEvent(app.dimensionList.enigmaObject.Handle, onDimensionListChanged, nil, true)

	return app.dimensionList, nil

}

// GetFieldList session object containing list of fields
func (app *App) GetFieldList(sessionState SessionState, actionState *action.State) (*FieldList, error) {
	if app.fieldlist != nil {
		return app.fieldlist, nil
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

	// Get data
	if err := sessionState.SendRequest(actionState, app.fieldlist.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := sessionState.SendRequest(actionState, app.fieldlist.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}

	// setup automatic data update
	onFieldListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(app.fieldlist.UpdateLayout(ctx))
	}
	sessionState.RegisterEvent(app.fieldlist.enigmaObject.Handle, onFieldListChanged, nil, true)

	return app.fieldlist, nil
}

// GetCurrentSelections create current selection session object and add to list
func (app *App) GetCurrentSelections(sessionState SessionState, actionState *action.State) (*CurrentSelections, error) {
	if app.currentSelections != nil {
		return app.currentSelections, nil
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

	// Get layout
	if err := sessionState.SendRequest(actionState, app.currentSelections.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := sessionState.SendRequest(actionState, app.currentSelections.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	// update currentSelection layout when object is changed
	onCurrentSelectionChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(app.currentSelections.UpdateLayout(ctx))
	}
	sessionState.RegisterEvent(app.currentSelections.enigmaObject.Handle,
		onCurrentSelectionChanged, nil, true)

	return app.currentSelections, nil
}

// GetLocaleInfo send get locale info request
func (app *App) GetLocaleInfo(ctx context.Context) (*enigma.LocaleInfo, error) {
	if app.localeInfo != nil {
		return app.localeInfo, nil
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
