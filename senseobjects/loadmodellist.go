package senseobjects

import (
	"context"
	"sync"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
)

type (
	// LoadModelList used to keep track of story list session object
	LoadModelList struct {
		enigmaObject *enigma.GenericObject
		layout       *enigma.GenericObjectLayout
		properties   *enigma.GenericObjectProperties

		mutex sync.Mutex
	}
)

// CreateLoadModelListObject create LoadModelList session object
func CreateLoadModelListObject(ctx context.Context, doc *enigma.Doc) (*LoadModelList, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Id:   "LoadModelList",
			Type: "LoadModelList",
		},
		AppObjectListDef: &enigma.AppObjectListDef{
			Type: "LoadModel",
		},
	}

	obj, err := doc.CreateSessionObjectRaw(ctx, properties)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create load model list session object in app<%s>", doc.GenericId)
	}

	return &LoadModelList{
		enigmaObject: obj,
	}, nil
}

// UpdateLayout get and set a new layout for LoadModelList
func (loadModelList *LoadModelList) UpdateLayout(ctx context.Context) error {
	if loadModelList.enigmaObject == nil {
		return errors.Errorf("loadModelList enigma object is nil")
	}

	layoutRaw, err := loadModelList.enigmaObject.GetLayoutRaw(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get loadModelList layout")
	}

	var layout enigma.GenericObjectLayout
	err = json.Unmarshal(layoutRaw, &layout)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal loadModelList layout")
	}

	loadModelList.setLayout(&layout)
	return nil
}

func (loadModelList *LoadModelList) setLayout(layout *enigma.GenericObjectLayout) {
	loadModelList.mutex.Lock()
	defer loadModelList.mutex.Unlock()
	loadModelList.layout = layout
}

// UpdateProperties get and set properties for LoadModelList
func (loadModelList *LoadModelList) UpdateProperties(ctx context.Context) error {
	if loadModelList.enigmaObject == nil {
		return errors.Errorf("loadModelList enigma object is nil")
	}

	propertiesRaw, err := loadModelList.enigmaObject.GetEffectivePropertiesRaw(ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal loadModelList properties")
	}

	var properties enigma.GenericObjectProperties
	err = json.Unmarshal(propertiesRaw, &properties)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal loadModelList properties")
	}

	loadModelList.setProperties(&properties)

	return nil
}

func (loadModelList *LoadModelList) setProperties(properties *enigma.GenericObjectProperties) {
	loadModelList.mutex.Lock()
	defer loadModelList.mutex.Unlock()

	loadModelList.properties = properties
}
