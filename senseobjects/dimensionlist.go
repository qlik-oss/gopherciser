package senseobjects

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	DimensionList struct {
		enigmaObject *enigma.GenericObject
		layout       *enigma.GenericObjectLayout
		properties   *enigma.GenericObjectProperties
		listboxes    map[string]*ListBox
		mutex        sync.Mutex
	}

	DimensionNotFoundError string
)

func (err DimensionNotFoundError) Error() string {
	return fmt.Sprintf("dimension<%s> not found", string(err))
}

// CreateDimensionListObject create dimensionlist session object
func CreateDimensionListObject(ctx context.Context, doc *enigma.Doc) (*DimensionList, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "DimensionList",
			Id:   "DimensionList",
		},
		DimensionListDef: &enigma.DimensionListDef{
			Type: "dimension",
			Data: json.RawMessage(`{
				"title": "/qMetaDef/title",
				"tags": "/qMetaDef/tags",
				"grouping": "/qDim/qGrouping",
				"info": "/qDimInfos",
				"labelExpression": "/qDim/qLabelExpression"
			}`),
		},
	}

	obj, err := doc.CreateSessionObject(ctx, properties)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create dimensionlist session object in app<%s>", doc.GenericId)
	}

	return &DimensionList{enigmaObject: obj, properties: properties}, nil
}

// UpdateLayout update object with new layout from engine
func (dimensionlist *DimensionList) UpdateLayout(ctx context.Context) error {
	layout, err := dimensionlist.enigmaObject.GetLayout(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	dimensionlist.setLayout(layout)
	return nil
}

func (dimensionlist *DimensionList) setLayout(layout *enigma.GenericObjectLayout) {
	dimensionlist.mutex.Lock()
	defer dimensionlist.mutex.Unlock()
	dimensionlist.layout = layout
}

// Layout of dimensionlist object
func (dimensionlist *DimensionList) Layout() *enigma.GenericObjectLayout {
	dimensionlist.mutex.Lock()
	defer dimensionlist.mutex.Unlock()
	return dimensionlist.layout
}

// UpdateProperties update object with new properties from engine
func (dimensionlist *DimensionList) UpdateProperties(ctx context.Context) error {
	properties, err := dimensionlist.enigmaObject.GetEffectiveProperties(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	dimensionlist.setProperties(properties)
	return nil
}

func (dimensionlist *DimensionList) setProperties(properties *enigma.GenericObjectProperties) {
	dimensionlist.mutex.Lock()
	defer dimensionlist.mutex.Unlock()
	dimensionlist.properties = properties
}

// Properties of dimensionlist object
func (dimensionlist *DimensionList) Properties() *enigma.GenericObjectProperties {
	dimensionlist.mutex.Lock()
	defer dimensionlist.mutex.Unlock()
	return dimensionlist.properties
}

// GetDimension searches dimension list and returns container entry of dimension with title if it exists
func (dimensionlist *DimensionList) GetDimension(dimensionTitle string) (*enigma.NxContainerEntry, error) {
	dimensionlist.mutex.Lock()
	defer dimensionlist.mutex.Unlock()

	if dimensionlist.layout == nil || dimensionlist.layout.DimensionList == nil {
		return nil, errors.Errorf("no dimensionlist layout")
	}

	titlePath := helpers.NewDataPath("/title")
	for _, dimension := range dimensionlist.layout.DimensionList.Items {
		if dimension == nil {
			continue
		}

		rawTitle, err := titlePath.LookupNoQuotes(dimension.Data)
		if err != nil {
			return nil, errors.Wrap(err, "error getting dimension title")
		}
		if string(rawTitle) == dimensionTitle {
			return dimension, nil
		}
	}

	return nil, DimensionNotFoundError(dimensionTitle)
}

// GetOrCreateSessionListbox
func (dimensionlist *DimensionList) GetOrCreateSessionListboxSync(sessionState SessionState, actionState *action.State, doc *enigma.Doc, dimensionTitle string) (*ListBox, error) {
	dimension, err := dimensionlist.GetDimension(dimensionTitle)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if dimension == nil {
		return nil, DimensionNotFoundError(dimensionTitle)
	}

	if dimensionlist.listboxes == nil || dimensionlist.listboxes[dimensionTitle] == nil {
		listbox, err := dimensionlist.createListbox(sessionState, actionState, doc, dimension.Info.Id)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		dimensionlist.listboxes[dimensionTitle] = listbox
	}

	return dimensionlist.listboxes[dimensionTitle], nil
}

func (dimensionlist *DimensionList) createListbox(sessionState SessionState, actionState *action.State, doc *enigma.Doc, libraryID string) (*ListBox, error) {
	var listbox *ListBox
	createListBox := func(ctx context.Context) error {
		var err error
		listbox, err = CreateLibraryBoxObject(ctx, doc, libraryID)
		return errors.WithStack(err)
	}
	if err := sessionState.SendRequest(actionState, createListBox); err != nil {
		return nil, errors.WithStack(err)
	}

	if err := sessionState.SendRequest(actionState, listbox.UpdateProperties); err != nil {
		return nil, errors.WithStack(err)
	}
	if err := sessionState.SendRequest(actionState, listbox.UpdateLayout); err != nil {
		return nil, errors.WithStack(err)
	}

	// mark dirty on update, don't update automatically since only updated when "selection tool" is open
	onListboxChanged := func(ctx context.Context, actionState *action.State) error {
		listbox.Dirty = true
		return nil
	}
	sessionState.RegisterEvent(listbox.EnigmaObject.Handle, onListboxChanged, nil, true)

	dimensionlist.mutex.Lock()
	defer dimensionlist.mutex.Unlock()

	if dimensionlist.listboxes == nil {
		dimensionlist.listboxes = make(map[string]*ListBox, 1)
	}

	return listbox, nil
}
