package senseobjects

import (
	"context"
	"sync"

	"github.com/qlik-oss/gopherciser/senseobjdef"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
)

type (
	// ListBoxLayout listbox layout
	ListBoxLayout struct {
		enigma.GenericObjectLayout
		AppObjectList *enigma.ListObject `json:"qListBox,omitempty"`
	}

	// ListBoxProperties listbox properties
	ListBoxProperties struct {
		Info    enigma.NxInfo    `json:"qInfo,omitempty"`
		MetaDef enigma.NxMetaDef `json:"qMetaDef,omitempty"`
	}

	// ListBox container with listbox in sense app
	ListBox struct {
		enigmaObject *enigma.GenericObject
		layout       *ListBoxLayout
		properties   *ListBoxProperties
		mutex        sync.Mutex
	}
)

func (listBox *ListBox) setLayout(layout *ListBoxLayout) {
	listBox.mutex.Lock()
	defer listBox.mutex.Unlock()
	listBox.layout = layout
}

func (listBox *ListBox) setProperties(properties *ListBoxProperties) {
	listBox.mutex.Lock()
	defer listBox.mutex.Unlock()
	listBox.properties = properties
}

// UpdateLayout get and set a new layout for sheetlist
func (listBox *ListBox) UpdateLayout(ctx context.Context) error {
	if listBox.enigmaObject == nil {
		return errors.Errorf("listBox enigma object is nil")
	}

	layoutRaw, err := listBox.enigmaObject.GetLayoutRaw(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get listBox layout")
	}

	var layout ListBoxLayout
	err = jsonit.Unmarshal(layoutRaw, &layout)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal listBox layout")
	}

	listBox.setLayout(&layout)
	return nil
}

// UpdateProperties get and set properties for listBox
func (listBox *ListBox) UpdateProperties(ctx context.Context) error {
	if listBox.enigmaObject == nil {
		return errors.Errorf("listBox enigma object is nil")
	}

	propertiesRaw, err := listBox.enigmaObject.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal listBox properties")
	}

	var properties ListBoxProperties
	err = jsonit.Unmarshal(propertiesRaw, &properties)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal listBox properties")
	}

	listBox.setProperties(&properties)
	return nil
}

// GetListObjectData get datapages
func (listBox *ListBox) GetListObjectData(ctx context.Context) ([]*enigma.NxDataPage, error) {
	objDef, err := senseobjdef.GetObjectDef("listbox")
	if err != nil {
		return nil, err
	}
	return listBox.enigmaObject.GetListObjectData(ctx, string(objDef.Data[0].Requests[0].Path), []*enigma.NxPage{
		{
			Left:   0,
			Top:    0,
			Width:  1,
			Height: listBox.layout.GenericObjectLayout.ListObject.Size.Cy,
		},
	})
}

// Layout for listBox
func (listBox *ListBox) Layout() *ListBoxLayout {
	return listBox.layout //TODO DECISION: wait for write lock?
}

// CreateListBoxObject create listbox session object
func CreateListBoxObject(ctx context.Context, doc *enigma.Doc, name string) (*ListBox, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "ListBox",
		},
		ListObjectDef: &enigma.ListObjectDef{
			Def: &enigma.NxInlineDimensionDef{
				FieldDefs:   []string{name},
				FieldLabels: []string{name},
			},
			InitialDataFetch: []*enigma.NxPage{{Height: 20, Width: 1}},
		},
	}

	obj, err := doc.CreateSessionObjectRaw(ctx, properties)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create listbox session object in app<%s>", doc.GenericId)
	}

	return &ListBox{
		enigmaObject: obj,
	}, nil
}
