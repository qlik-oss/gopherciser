package senseobjects

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/logger"
)

type (
	// SessionObjects for the app
	SessionObjects struct {
		sheetList         lazySessionObject[*SheetList]
		bookmarkList      lazySessionObject[*BookmarkList]
		currentSelections lazySessionObject[*CurrentSelections]
		localeInfo        lazySessionObject[*enigma.LocaleInfo]
		variablelist      lazySessionObject[*VariableList]
		storylist         lazySessionObject[*StoryList]
		loadmodellist     lazySessionObject[*LoadModelList]
		fieldlist         lazySessionObject[*FieldList]
		dimensionList     lazySessionObject[*DimensionList]
		dimensions        map[string]*enigma.GenericDimension
		appPropsList      lazySessionObject[*AppPropsList]
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

// initSessionObject runs init for a newly created session object. If init fails the session object is destroyed in
// the engine, as it will not be cached and the next caller creates a new one. The init error is returned.
func (app *App) initSessionObject(sessionState SessionState, actionState *action.State, obj *enigma.GenericObject, init func() error) error {
	err := init()
	if err == nil || obj == nil {
		return err
	}

	if destroyErr := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		_, err := app.Doc.DestroySessionObject(ctx, obj.GenericId)
		return err
	}); destroyErr != nil {
		sessionState.LogDetail(logger.WarningLevel, fmt.Sprintf("failed to destroy session object<%s> after failed initialization", obj.GenericId), destroyErr.Error())
	}
	return err
}

// GetSheetList update sheet list for app
func (app *App) GetSheetList(sessionState SessionState, actionState *action.State) (*SheetList, error) {
	return app.sheetList.get(func() (*SheetList, error) {
		var sheetList *SheetList
		createSheetList := func(ctx context.Context) error {
			var err error
			sheetList, err = CreateSheetListObject(ctx, app.Doc)
			return err
		}
		if err := sessionState.SendRequest(actionState, createSheetList); err != nil {
			return nil, errors.WithStack(err)
		}

		if err := app.initSessionObject(sessionState, actionState, sheetList.enigmaObject, func() error {
			return sessionState.SendRequest(actionState, sheetList.UpdateLayout)
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		// update sheetList layout when sheetList has a change event
		onSheetListChanged := func(ctx context.Context, actionState *action.State) error {
			return errors.WithStack(sheetList.UpdateLayout(ctx))
		}

		sessionState.RegisterEvent(sheetList.enigmaObject.Handle,
			onSheetListChanged, nil, true)

		return sheetList, nil
	})
}

// GetBookmarkList update bookmark list for app
func (app *App) GetBookmarkList(sessionState SessionState, actionState *action.State) (*BookmarkList, error) {
	return app.bookmarkList.get(func() (*BookmarkList, error) {
		var bookmarkList *BookmarkList
		createBookmarkList := func(ctx context.Context) error {
			var err error
			bookmarkList, err = CreateBookmarkListObject(ctx, app.Doc)
			return err
		}

		if err := sessionState.SendRequest(actionState, createBookmarkList); err != nil {
			return nil, errors.WithStack(err)
		}

		if err := app.initSessionObject(sessionState, actionState, bookmarkList.enigmaObject, func() error {
			if err := sessionState.SendRequest(actionState, bookmarkList.UpdateLayout); err != nil {
				return err
			}
			return sessionState.SendRequest(actionState, bookmarkList.UpdateProperties)
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		// update bookmark list layout when bookmark list has a change event
		onBookmarkListChanged := func(ctx context.Context, actionState *action.State) error {
			return errors.WithStack(bookmarkList.UpdateLayout(ctx))
		}

		sessionState.RegisterEvent(bookmarkList.enigmaObject.Handle,
			onBookmarkListChanged, nil, true)

		return bookmarkList, nil
	})
}

// GetVariableList create or return existing variable list session object
func (app *App) GetVariableList(sessionState SessionState, actionState *action.State) (*VariableList, error) {
	return app.variablelist.get(func() (*VariableList, error) {
		var variableList *VariableList
		createVariableList := func(ctx context.Context) error {
			var err error
			variableList, err = CreateVariableListObject(ctx, app.Doc)
			return err
		}

		if err := sessionState.SendRequest(actionState, createVariableList); err != nil {
			return nil, errors.WithStack(err)
		}

		if err := app.initSessionObject(sessionState, actionState, variableList.enigmaObject, func() error {
			return sessionState.SendRequest(actionState, variableList.UpdateLayout)
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		// update variable list layout when variable list has a change event
		onVariableListChanged := func(ctx context.Context, actionState *action.State) error {
			return errors.WithStack(variableList.UpdateLayout(ctx))
		}
		sessionState.RegisterEvent(variableList.enigmaObject.Handle, onVariableListChanged, nil, true)

		return variableList, nil
	})
}

// GetAppsPropsList create or return AppsPropsList
func (app *App) GetAppsPropsList(sessionState SessionState, actionState *action.State) (*AppPropsList, error) {
	return app.appPropsList.get(func() (*AppPropsList, error) {
		var appPropsList *AppPropsList
		createAppPropsList := func(ctx context.Context) error {
			var err error
			appPropsList, err = CreateAppPropsListObject(ctx, app.Doc)
			return err
		}

		if err := sessionState.SendRequest(actionState, createAppPropsList); err != nil {
			return nil, errors.WithStack(err)
		}

		if err := app.initSessionObject(sessionState, actionState, appPropsList.enigmaObject, func() error {
			return sessionState.SendRequest(actionState, func(ctx context.Context) error {
				return appPropsList.UpdateLayout(ctx, app.Doc, sessionState, actionState)
			})
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		onAppPropsListChanged := func(ctx context.Context, actionState *action.State) error {
			return errors.WithStack(appPropsList.UpdateLayout(ctx, app.Doc, sessionState, actionState))
		}
		sessionState.RegisterEvent(appPropsList.enigmaObject.Handle, onAppPropsListChanged, nil, true)

		return appPropsList, nil
	})
}

// GetStoryList create or return existing story list session object
func (app *App) GetStoryList(sessionState SessionState, actionState *action.State) (*StoryList, error) {
	return app.storylist.get(func() (*StoryList, error) {
		var storyList *StoryList
		createStoryList := func(ctx context.Context) error {
			var err error
			storyList, err = CreateStoryListObject(ctx, app.Doc)
			return err
		}

		if err := sessionState.SendRequest(actionState, createStoryList); err != nil {
			return nil, errors.WithStack(err)
		}

		if err := app.initSessionObject(sessionState, actionState, storyList.enigmaObject, func() error {
			return sessionState.SendRequest(actionState, storyList.UpdateLayout)
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		// update story list layout when story list has a change event
		onStoryListChanged := func(ctx context.Context, actionState *action.State) error {
			return errors.WithStack(storyList.UpdateLayout(ctx))
		}
		sessionState.RegisterEvent(storyList.enigmaObject.Handle, onStoryListChanged, nil, true)

		return storyList, nil
	})
}

// GetLoadModelList create or return existing load model list session object
func (app *App) GetLoadModelList(sessionState SessionState, actionState *action.State) (*LoadModelList, error) {
	return app.loadmodellist.get(func() (*LoadModelList, error) {
		var loadModelList *LoadModelList
		createLoadModelList := func(ctx context.Context) error {
			var err error
			loadModelList, err = CreateLoadModelListObject(ctx, app.Doc)
			return err
		}

		if err := sessionState.SendRequest(actionState, createLoadModelList); err != nil {
			return nil, errors.WithStack(err)
		}

		if err := app.initSessionObject(sessionState, actionState, loadModelList.enigmaObject, func() error {
			if err := sessionState.SendRequest(actionState, loadModelList.UpdateLayout); err != nil {
				return err
			}
			return sessionState.SendRequest(actionState, loadModelList.UpdateProperties)
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		// update load model list layout when load model list has a change event
		onLoadModelListChanged := func(ctx context.Context, actionState *action.State) error {
			return errors.WithStack(loadModelList.UpdateLayout(ctx))
		}
		sessionState.RegisterEvent(loadModelList.enigmaObject.Handle, onLoadModelListChanged, nil, true)

		return loadModelList, nil
	})
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
	currentSelections, ok := app.currentSelections.peek()
	if !ok {
		return ""
	}
	layout := currentSelections.Layout()
	if layout == nil {
		return ""
	}
	fields := make([]string, 0, len(layout.SelectionObject.Selections))
	for _, selection := range layout.SelectionObject.Selections {
		fields = append(fields, selection.Field)
	}

	return strings.Join(fields, ",")
}

func (app *App) GetDimensionList(sessionState SessionState, actionState *action.State) (*DimensionList, error) {
	return app.dimensionList.get(func() (*DimensionList, error) {
		var dimensionList *DimensionList
		createDimensionList := func(ctx context.Context) error {
			var err error
			dimensionList, err = CreateDimensionListObject(ctx, app.Doc)
			return err
		}
		if err := sessionState.SendRequest(actionState, createDimensionList); err != nil {
			return nil, errors.WithStack(err)
		}

		// Get data
		if err := app.initSessionObject(sessionState, actionState, dimensionList.enigmaObject, func() error {
			if err := sessionState.SendRequest(actionState, dimensionList.UpdateLayout); err != nil {
				return err
			}
			return sessionState.SendRequest(actionState, dimensionList.UpdateProperties)
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		// setup automatic data update
		onDimensionListChanged := func(ctx context.Context, actionState *action.State) error {
			return errors.WithStack(dimensionList.UpdateLayout(ctx))
		}
		sessionState.RegisterEvent(dimensionList.enigmaObject.Handle, onDimensionListChanged, nil, true)

		return dimensionList, nil
	})
}

// GetFieldList session object containing list of fields
func (app *App) GetFieldList(sessionState SessionState, actionState *action.State) (*FieldList, error) {
	return app.fieldlist.get(func() (*FieldList, error) {
		var fieldList *FieldList
		createFieldList := func(ctx context.Context) error {
			var err error
			fieldList, err = CreateFieldListObject(ctx, app.Doc)
			return err
		}
		if err := sessionState.SendRequest(actionState, createFieldList); err != nil {
			return nil, errors.WithStack(err)
		}

		// Get data
		if err := app.initSessionObject(sessionState, actionState, fieldList.enigmaObject, func() error {
			if err := sessionState.SendRequest(actionState, fieldList.UpdateLayout); err != nil {
				return err
			}
			return sessionState.SendRequest(actionState, fieldList.UpdateProperties)
		}); err != nil {
			return nil, errors.WithStack(err)
		}

		// setup automatic data update
		onFieldListChanged := func(ctx context.Context, actionState *action.State) error {
			return errors.WithStack(fieldList.UpdateLayout(ctx))
		}
		sessionState.RegisterEvent(fieldList.enigmaObject.Handle, onFieldListChanged, nil, true)

		return fieldList, nil
	})
}

// GetCurrentSelections create current selection session object and add to list
func (app *App) GetCurrentSelections(sessionState SessionState, actionState *action.State, requestData bool) (*CurrentSelections, error) {
	created := false
	currentSelections, err := app.currentSelections.get(func() (*CurrentSelections, error) {
		created = true

		var currentSelections *CurrentSelections
		createCurrentSelections := func(ctx context.Context) error {
			var err error
			currentSelections, err = CreateCurrentSelections(ctx, app.Doc)
			return err
		}
		if err := sessionState.SendRequest(actionState, createCurrentSelections); err != nil {
			return nil, errors.WithStack(err)
		}

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
			if err := app.initSessionObject(sessionState, actionState, currentSelections.enigmaObject, func() error {
				return onCurrentSelectionChanged(sessionState.BaseContext(), actionState)
			}); err != nil {
				return nil, errors.WithStack(err)
			}
		}

		sessionState.RegisterEvent(currentSelections.enigmaObject.Handle, onCurrentSelectionChanged, nil, true)

		return currentSelections, nil
	})
	if err != nil {
		return nil, err
	}

	// existing object, request data using registered event function
	if !created && requestData && currentSelections.enigmaObject != nil {
		if f := sessionState.GetEventFunc(currentSelections.enigmaObject.Handle); f != nil {
			if err := f(sessionState.BaseContext(), actionState); err != nil {
				return currentSelections, errors.WithStack(err)
			}
		}
	}
	return currentSelections, nil
}

// GetLocaleInfo send get locale info request
func (app *App) GetLocaleInfo(ctx context.Context) (*enigma.LocaleInfo, error) {
	return app.localeInfo.get(func() (*enigma.LocaleInfo, error) {
		localeInfo, err := app.Doc.GetLocaleInfo(ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return localeInfo, nil
	})
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
