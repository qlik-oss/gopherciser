package senseobjects

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
)

type (
	FieldList struct {
		enigmaObject *enigma.GenericObject
		layout       *enigma.GenericObjectLayout
		properties   *enigma.GenericObjectProperties
		listboxes    map[string]*ListBox
		mutex        sync.Mutex
	}

	FieldNotFoundError string
)

func (err FieldNotFoundError) Error() string {
	return fmt.Sprintf("field<%s> not found", string(err))
}

// CreateFieldListObject create fieldlist session object
func CreateFieldListObject(ctx context.Context, doc *enigma.Doc) (*FieldList, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "FieldList",
		},
		FieldListDef: &enigma.FieldListDef{
			ShowSystem:         false,
			ShowHidden:         false,
			ShowSemantic:       true,
			ShowSrcTables:      true,
			ShowDerivedFields:  true,
			ShowImplicit:       false,
			ShowDefinitionOnly: false,
		},
	}

	obj, err := doc.CreateSessionObject(ctx, properties)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create fieldlist session object in app<%s>", doc.GenericId)
	}

	return &FieldList{enigmaObject: obj, properties: properties}, nil
}

// UpdateLayout update object with new layout from engine
func (fieldlist *FieldList) UpdateLayout(ctx context.Context) error {
	layout, err := fieldlist.enigmaObject.GetLayout(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	fieldlist.setLayout(layout)
	return nil
}

func (fieldlist *FieldList) setLayout(layout *enigma.GenericObjectLayout) {
	fieldlist.mutex.Lock()
	defer fieldlist.mutex.Unlock()
	fieldlist.layout = layout
}

// Layout of fieldlist object
func (fieldlist *FieldList) Layout() *enigma.GenericObjectLayout {
	fieldlist.mutex.Lock()
	defer fieldlist.mutex.Unlock()
	return fieldlist.layout
}

// UpdateProperties update object with new properties from engine
func (fieldlist *FieldList) UpdateProperties(ctx context.Context) error {
	properties, err := fieldlist.enigmaObject.GetEffectiveProperties(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	fieldlist.setProperties(properties)
	return nil
}

func (fieldlist *FieldList) setProperties(properties *enigma.GenericObjectProperties) {
	fieldlist.mutex.Lock()
	defer fieldlist.mutex.Unlock()
	fieldlist.properties = properties
}

// Properties of fieldlist object
func (fieldlist *FieldList) Properties() *enigma.GenericObjectProperties {
	fieldlist.mutex.Lock()
	defer fieldlist.mutex.Unlock()
	return fieldlist.properties
}

// GetField searches field list and returns field with name if it exists
func (fieldlist *FieldList) GetField(fieldName string) (*enigma.NxFieldDescription, error) {
	fieldlist.mutex.Lock()
	defer fieldlist.mutex.Unlock()

	if fieldlist.layout == nil || fieldlist.layout.FieldList == nil {
		return nil, errors.Errorf("no fieldlist layout")
	}

	for _, field := range fieldlist.layout.FieldList.Items {
		if field == nil {
			continue
		}
		if field.Name == fieldName {
			return field, nil
		}
	}

	return nil, FieldNotFoundError(fieldName)
}

// GetOrCreateSessionListbox
func (fieldlist *FieldList) GetOrCreateSessionListboxSync(sessionState SessionState, actionState *action.State, doc *enigma.Doc, fieldName string) (*ListBox, error) {
	field, err := fieldlist.GetField(fieldName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if field == nil {
		return nil, FieldNotFoundError(fieldName)
	}

	if fieldlist.listboxes == nil || fieldlist.listboxes[fieldName] == nil {
		if err = fieldlist.createListbox(sessionState, actionState, doc, fieldName); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return fieldlist.listboxes[fieldName], nil
}

func (fieldlist *FieldList) createListbox(sessionState SessionState, actionState *action.State, doc *enigma.Doc, fieldName string) error {
	var listbox *ListBox
	createListBox := func(ctx context.Context) error {
		var err error
		listbox, err = CreateFieldListBoxObject(ctx, doc, fieldName)
		return errors.WithStack(err)
	}
	if err := sessionState.SendRequest(actionState, createListBox); err != nil {
		return errors.WithStack(err)
	}

	if err := sessionState.SendRequest(actionState, listbox.UpdateProperties); err != nil {
		return errors.WithStack(err)
	}
	if err := sessionState.SendRequest(actionState, listbox.UpdateLayout); err != nil {
		return errors.WithStack(err)
	}

	// mark dirty on update, don't update automatically since only updated when "selectors" are open
	onListboxChanged := func(ctx context.Context, actionState *action.State) error {
		listbox.Dirty = true
		return nil
	}
	sessionState.RegisterEvent(listbox.EnigmaObject.Handle, onListboxChanged, nil, true)

	fieldlist.mutex.Lock()
	defer fieldlist.mutex.Unlock()

	if fieldlist.listboxes == nil {
		fieldlist.listboxes = make(map[string]*ListBox, 1)
	}

	fieldlist.listboxes[fieldName] = listbox
	return nil
}
