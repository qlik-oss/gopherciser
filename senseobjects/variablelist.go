package senseobjects

import (
	"context"
	"sync"

	"github.com/goccy/go-json"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
)

type (
	// VariableList used to keep track of variable list session object
	VariableList struct {
		enigmaObject *enigma.GenericObject
		layout       *enigma.GenericObjectLayout

		mutex sync.Mutex
	}
)

// CreateVariableListObject create VariableList session object
func CreateVariableListObject(ctx context.Context, doc *enigma.Doc) (*VariableList, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Id:   "VariableList",
			Type: "VariableList",
		},
		VariableListDef: &enigma.VariableListDef{
			Type:         "variable",
			ShowReserved: true,
			ShowConfig:   true,
			Data: json.RawMessage(`{
				"tags": "/tags"
			}`),
		},
	}

	obj, err := doc.CreateSessionObjectRaw(ctx, properties)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create variablelist session object in app<%s>", doc.GenericId)
	}

	return &VariableList{
		enigmaObject: obj,
	}, nil
}

// UpdateLayout get and set a new layout for VariableList
func (variableList *VariableList) UpdateLayout(ctx context.Context) error {
	if variableList.enigmaObject == nil {
		return errors.Errorf("variableList enigma object is nil")
	}

	layoutRaw, err := variableList.enigmaObject.GetLayoutRaw(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get variableList layout")
	}

	var layout enigma.GenericObjectLayout
	err = json.Unmarshal(layoutRaw, &layout)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal variableList layout")
	}

	variableList.setLayout(&layout)
	return nil
}

func (variableList *VariableList) setLayout(layout *enigma.GenericObjectLayout) {
	variableList.mutex.Lock()
	defer variableList.mutex.Unlock()

	variableList.layout = layout
}
