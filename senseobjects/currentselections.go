package senseobjects

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
)

type (
	// CurrentSelectionLayout layout object
	CurrentSelectionLayout struct {
		NxInfo          enigma.NxInfo          `json:"qInfo,omitempty"`
		NxSelectionInfo enigma.NxSelectionInfo `json:"qNxSelectionInfo,omitempty"`
		SelectionObject enigma.SelectionObject `json:"qSelectionObject,omitempty"`
	}

	// CurrentSelectionProperties properties object
	CurrentSelectionProperties struct {
		Info               enigma.NxInfo             `json:"qInfo"`
		MetaDef            enigma.NxMetaDef          `json:"qMetaDef"`
		SelectionObjectDef enigma.SelectionObjectDef `json:"qSelectionObjectDef"`
	}

	// CurrentSelections object containing current selection state
	CurrentSelections struct {
		enigmaObject *enigma.GenericObject
		layout       *CurrentSelectionLayout
		properties   *CurrentSelectionProperties
		mutex        sync.Mutex
	}
)

// UpdateLayout for current selections
func (cs *CurrentSelections) UpdateLayout(ctx context.Context) error {
	layoutRaw, err := cs.enigmaObject.GetLayoutRaw(ctx)
	if err != nil {
		return errors.Wrap(err, "Error getting layout for current selection object")
	}

	var layout CurrentSelectionLayout
	err = jsonit.Unmarshal(layoutRaw, &layout)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal CurrentSelection layout")
	}

	cs.setLayout(&layout)
	return nil
}

// UpdateProperties for current selections
func (cs *CurrentSelections) UpdateProperties(ctx context.Context) error {
	propertiesRaw, err := cs.enigmaObject.GetEffectivePropertiesRaw(ctx)
	if err != nil {
		return errors.Wrap(err, "Error getting properties for current selection object")
	}

	var properties CurrentSelectionProperties
	err = jsonit.Unmarshal(propertiesRaw, &properties)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal CurrentSelction properties")
	}

	cs.setProperties(&properties)

	return nil
}

func (cs *CurrentSelections) setLayout(layout *CurrentSelectionLayout) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.layout = layout
}

func (cs *CurrentSelections) setProperties(properties *CurrentSelectionProperties) {
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	cs.properties = properties
}

// Layout for current selections
func (cs *CurrentSelections) Layout() *CurrentSelectionLayout {
	return cs.layout //TODO DECISION: wait for write lock?
}

// Properties for current selections
func (cs *CurrentSelections) Properties() *CurrentSelectionProperties {
	return cs.properties //TODO DECISION: wait for write lock?
}

// CreateCurrentSelections create current selections session object
func CreateCurrentSelections(ctx context.Context, doc *enigma.Doc) (*CurrentSelections, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Id:   "CurrentSelection",
			Type: "CurrentSelection",
		},
		SelectionObjectDef: &enigma.SelectionObjectDef{},
	}

	// Create current selection object
	obj, err := doc.CreateSessionObjectRaw(ctx, properties)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create CurrentSelection session object in app<%s>", doc.GenericId)
	}

	cs := &CurrentSelections{
		enigmaObject: obj,
	}

	return cs, nil
}
