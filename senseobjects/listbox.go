package senseobjects

import (
	"context"
	"sync"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/senseobjdef"
)

type (
	// ListBoxProperties listbox properties
	ListBoxProperties struct {
		Info    enigma.NxInfo    `json:"qInfo,omitempty"`
		MetaDef enigma.NxMetaDef `json:"qMetaDef,omitempty"`
	}

	// ListBox container with listbox in sense app
	ListBox struct {
		Dirty bool // Set flag to have object update before accessing layout next time

		EnigmaObject *enigma.GenericObject
		layout       *enigma.GenericObjectLayout
		listobject   *enigma.ListObject
		properties   *ListBoxProperties
		mutex        sync.Mutex
	}
)

var (
	DefaultListboxInitialDataFetch = []*enigma.NxPage{{Height: 20, Width: 1}}
)

func (listBox *ListBox) setLayout(layout *enigma.GenericObjectLayout, lock bool) {
	if lock {
		listBox.mutex.Lock()
		defer listBox.mutex.Unlock()
	}
	listBox.layout = layout
	listBox.listobject = layout.ListObject
}

func (listBox *ListBox) setProperties(properties *ListBoxProperties) {
	listBox.mutex.Lock()
	defer listBox.mutex.Unlock()
	listBox.properties = properties
}

func (listBox *ListBox) setDataPages(datapages []*enigma.NxDataPage, lock bool) error {
	if lock {
		listBox.mutex.Lock()
		defer listBox.mutex.Unlock()
	}
	if listBox.listobject == nil {
		return errors.Errorf("listbox has no listobject")
	}
	listBox.listobject.DataPages = datapages
	return nil
}

// UpdateLayout get and set a new layout for listbox
func (listBox *ListBox) UpdateLayout(ctx context.Context) error {
	return listBox.updateLayout(ctx, true)
}

// UpdateLayout get and set a new layout for listbox
func (listBox *ListBox) updateLayout(ctx context.Context, lock bool) error {
	if listBox.EnigmaObject == nil {
		return errors.Errorf("listBox enigma object is nil")
	}

	layoutRaw, err := listBox.EnigmaObject.GetLayoutRaw(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get listBox layout")
	}

	var layout enigma.GenericObjectLayout
	err = json.Unmarshal(layoutRaw, &layout)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal listBox layout")
	}

	listBox.setLayout(&layout, lock)

	_, err = listBox.getListObjectData(ctx, lock)
	return errors.WithStack(err)
}

// UpdateProperties get and set properties for listBox
func (listBox *ListBox) UpdateProperties(ctx context.Context) error {
	if listBox.EnigmaObject == nil {
		return errors.Errorf("listBox enigma object is nil")
	}

	propertiesRaw, err := listBox.EnigmaObject.GetEffectivePropertiesRaw(ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal listBox properties")
	}

	var properties ListBoxProperties
	err = json.Unmarshal(propertiesRaw, &properties)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal listBox properties")
	}

	listBox.setProperties(&properties)
	return nil
}

// GetListObjectData get datapages
func (listBox *ListBox) getListObjectData(ctx context.Context, lock bool) ([]*enigma.NxDataPage, error) {
	objDef, err := senseobjdef.GetObjectDef("listbox")
	if err != nil {
		return nil, err
	}

	datapages, err := listBox.EnigmaObject.GetListObjectData(ctx, string(objDef.Data[0].Requests[0].Path), []*enigma.NxPage{
		{
			Left:   0,
			Top:    0,
			Width:  1,
			Height: listBox.layout.ListObject.Size.Cy,
		},
	})
	if err != nil {
		return nil, err
	}

	if err := listBox.setDataPages(datapages, lock); err != nil {
		return nil, errors.WithStack(err)
	}

	return datapages, nil
}

// Layout for listBox, if layout needs updating this will lock object and update synchronously
func (listBox *ListBox) Layout(ctx context.Context) (*enigma.GenericObjectLayout, error) {
	listBox.mutex.Lock()
	defer listBox.mutex.Unlock()

	if listBox.Dirty {
		if err := listBox.updateLayout(ctx, false); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return listBox.layout, nil
}

// Properties for listBox
func (listBox *ListBox) Properties() *ListBoxProperties {
	listBox.mutex.Lock()
	defer listBox.mutex.Unlock()
	return listBox.properties
}

// ListObject for listBox
func (listBox *ListBox) ListObject(ctx context.Context) (*enigma.ListObject, error) {
	listBox.mutex.Lock()
	defer listBox.mutex.Unlock()

	if listBox.Dirty {
		if err := listBox.updateLayout(ctx, false); err != nil {
			return nil, errors.WithStack(err)
		}
	}
	return listBox.listobject, nil
}

// CreateFieldListBoxObject create listbox session object
func CreateFieldListBoxObject(ctx context.Context, doc *enigma.Doc, name string) (*ListBox, error) {
	return createListBoxObject(ctx, doc, &enigma.ListObjectDef{
		Def: &enigma.NxInlineDimensionDef{
			FieldDefs:   []string{name},
			FieldLabels: []string{name},
		},
		InitialDataFetch: DefaultListboxInitialDataFetch,
	})
}

// CreateLibraryBoxObject create listbox session object from library ID
func CreateLibraryBoxObject(ctx context.Context, doc *enigma.Doc, id string) (*ListBox, error) {
	return createListBoxObject(ctx, doc, &enigma.ListObjectDef{
		LibraryId:        id,
		InitialDataFetch: DefaultListboxInitialDataFetch,
	})
}

func createListBoxObject(ctx context.Context, doc *enigma.Doc, objDef *enigma.ListObjectDef) (*ListBox, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "listbox",
		},
		ListObjectDef: objDef,
	}

	obj, err := doc.CreateSessionObjectRaw(ctx, properties)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create listbox session object in app<%s>", doc.GenericId)
	}

	return &ListBox{
		EnigmaObject: obj,
	}, nil
}
