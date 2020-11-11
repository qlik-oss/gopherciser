package senseobjects

import (
	"context"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
)

type (
	// App sense app object
	App struct {
		GUID              string
		Doc               *enigma.Doc
		Layout            *enigma.NxAppLayout
		sheetList         *SheetList
		bookmarkList      *BookmarkList
		bookmarks         map[string]*enigma.GenericBookmark
		currentSelections *CurrentSelections
		localeInfo        *enigma.LocaleInfo
		variablelist      *VariableList
		mutex             sync.Mutex
	}
)

var jsonit = jsoniter.ConfigCompatibleWithStandardLibrary

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

	if err := sessionState.SendRequest(actionState, app.sheetList.UpdateProperties); err != nil {
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

	// update sheetList layout when bookmark list has a change event
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

	// update sheetList layout when variable list has a change event
	onVariableListChanged := func(ctx context.Context, actionState *action.State) error {
		return errors.WithStack(app.variablelist.UpdateLayout(ctx))
	}

	sessionState.RegisterEvent(app.variablelist.enigmaObject.Handle,
		onVariableListChanged, nil, true)

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

// TODO StoryList
//{"delta":true,"handle":1,"method":"CreateSessionObject","params":[{"qInfo":{"qId":"StoryList","qType":"StoryList"},"qAppObjectListDef":{"qType":"story","qData":{"title":"/qMetaDef/title","description":"/qMetaDef/description","thumbnail":"/thumbnail","rank":"/rank"}}}],"id":13,"jsonrpc":"2.0"}

// TODO LoadModelList
//{"delta":true,"handle":1,"method":"CreateSessionObject","params":[{"qInfo":{"qId":"LoadModelList","qType":"LoadModelList"},"qAppObjectListDef":{"qType":"LoadModel"}}],"id":30,"jsonrpc":"2.0"}

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
