package senseobjects

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v2"
)

type (
	// StoryList used to keep track of story list session object
	StoryList struct {
		enigmaObject *enigma.GenericObject
		layout       *enigma.GenericObjectLayout
		properties   *enigma.GenericObjectProperties

		mutex sync.Mutex
	}
)

// CreateStoryListObject create StoryList session object
func CreateStoryListObject(ctx context.Context, doc *enigma.Doc) (*StoryList, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Id:   "StoryList",
			Type: "StoryList",
		},
		AppObjectListDef: &enigma.AppObjectListDef{
			Type: "story",
			Data: json.RawMessage(`{
				"title": "/qMetaDef/title",
				"description": "/qMetaDef/description",
				"thumbnail": "/thumbnail",
				"rank": "/rank"	
			}`),
		},
	}

	obj, err := doc.CreateSessionObjectRaw(ctx, properties)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create story list session object in app<%s>", doc.GenericId)
	}

	return &StoryList{
		enigmaObject: obj,
	}, nil
}

// UpdateLayout get and set a new layout for StoryList
func (storyList *StoryList) UpdateLayout(ctx context.Context) error {
	if storyList.enigmaObject == nil {
		return errors.Errorf("storyList enigma object is nil")
	}

	layoutRaw, err := storyList.enigmaObject.GetLayoutRaw(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get storyList layout")
	}

	var layout enigma.GenericObjectLayout
	err = jsonit.Unmarshal(layoutRaw, &layout)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal storyList layout")
	}

	storyList.setLayout(&layout)
	return nil
}

func (storyList *StoryList) setLayout(layout *enigma.GenericObjectLayout) {
	storyList.mutex.Lock()
	defer storyList.mutex.Unlock()
	storyList.layout = layout
}

// UpdateProperties get and set properties for StoryList
func (storyList *StoryList) UpdateProperties(ctx context.Context) error {
	if storyList.enigmaObject == nil {
		return errors.Errorf("storyList enigma object is nil")
	}

	propertiesRaw, err := storyList.enigmaObject.GetEffectivePropertiesRaw(ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal storyList properties")
	}

	var properties enigma.GenericObjectProperties
	err = jsonit.Unmarshal(propertiesRaw, &properties)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal storyList properties")
	}

	storyList.setProperties(&properties)

	return nil
}

func (storyList *StoryList) setProperties(properties *enigma.GenericObjectProperties) {
	storyList.mutex.Lock()
	defer storyList.mutex.Unlock()

	storyList.properties = properties
}
