package senseobjects

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v2"
)

type (
	FieldList struct {
		enigmaObject *enigma.GenericObject
		layout       *enigma.GenericObjectLayout
		properties   *enigma.GenericObjectProperties
		mutex        sync.Mutex
	}
)

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
	properties, err := fieldlist.enigmaObject.GetProperties(ctx)
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
