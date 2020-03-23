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
		currentSelections *CurrentSelections
		localeInfo        *enigma.LocaleInfo
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

	// update sheetList to latest
	updateBookmarkList := func(ctx context.Context) error {
		bl, err := CreateBookmarkListObject(ctx, app.Doc)
		if err != nil {
			return err
		}
		app.setBookmarkList(sessionState, bl)
		return err
	}

	if err := sessionState.SendRequest(actionState, updateBookmarkList); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sessionState.SendRequest(actionState, app.bookmarkList.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sessionState.SendRequest(actionState, app.bookmarkList.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}

	// update sheetList layout when sheetList has a change event
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
