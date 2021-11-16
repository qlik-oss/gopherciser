package senseobjects

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
)

type (
	// BookmarkList container with bookmarks
	BookmarkList struct {
		enigmaObject *enigma.GenericObject
		layout       *BookmarkListLayout
		properties   *BookmarkListProperties
		mutex        sync.Mutex
	}

	// BookmarkListPropertiesData properties of bookmarklist
	BookmarkListPropertiesData struct {
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
		Cells       string `json:"cells,omitempty"`
	}

	// BookmarkListProperties bookmarklist properties
	BookmarkListProperties struct {
		Info    enigma.NxInfo            `json:"qInfo,omitempty"`
		MetaDef enigma.NxMetaDef         `json:"qMetaDef,omitempty"`
		Data    *SheetListPropertiesData `json:"qData,omitempty"`
	}

	BookmarkListLayout struct {
		enigma.GenericObjectLayout
		BookmarkList *BookmarkListAppObjectList `json:"qBookmarkList,omitempty"`
	}

	BookmarkListAppObjectList struct {
		enigma.AppObjectList
		Items []*BookmarkNxContainerEntry `json:"qItems,omitempty"`
	}

	// BookmarkNxContainerEntry container bookmark data
	BookmarkNxContainerEntry struct {
		enigma.NxContainerEntry
		Data *BookmarkData `json:"qData,omitempty"`
	}

	// BookmarkData data for a bookmark
	BookmarkData struct {
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
		SheetId     string `json:"sheetId,omitempty"`
	}
)

// CreateBookmarkListObject create bookmarklist session object
func CreateBookmarkListObject(ctx context.Context, doc *enigma.Doc) (*BookmarkList, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Id:   "BookmarkList",
			Type: "BookmarkList",
		},
		BookmarkListDef: &enigma.BookmarkListDef{ //TODO: How to find the SheetID?
			Type: "bookmark",
			Data: json.RawMessage(`{
				"title": "/qMetaDef/title",
				"description": "/qMetaDef/description",
                "selectionFields": "/selectionFields",
                "creationDate": "/creationDate",
                "sheetId": "/sheetId"
			}`),
		},
	}

	obj, err := doc.CreateSessionObjectRaw(ctx, properties)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create bookmarklist session object in app<%s>", doc.GenericId)
	}

	return &BookmarkList{
		enigmaObject: obj,
	}, nil
}

func (bookmarkList *BookmarkList) setLayout(layout *BookmarkListLayout) {
	bookmarkList.mutex.Lock()
	defer bookmarkList.mutex.Unlock()
	bookmarkList.layout = layout
}

// UpdateLayout get and set a new layout for bookmarklist
func (bookmarkList *BookmarkList) UpdateLayout(ctx context.Context) error {
	if bookmarkList.enigmaObject == nil {
		return errors.Errorf("bookmarklist enigma object is nil")
	}

	layoutRaw, err := bookmarkList.enigmaObject.GetLayoutRaw(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get bookmarklist layout")
	}

	var layout BookmarkListLayout
	err = jsonit.Unmarshal(layoutRaw, &layout)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal bookmarklist layout")
	}

	bookmarkList.setLayout(&layout)
	return nil
}

// UpdateProperties get and set properties for bookmarklist
func (bookmarkList *BookmarkList) UpdateProperties(ctx context.Context) error {
	if bookmarkList.enigmaObject == nil {
		return errors.Errorf("bookmarklist enigma object is nil")
	}

	propertiesRaw, err := bookmarkList.enigmaObject.GetEffectivePropertiesRaw(ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal bookmarklist properties")
	}

	var properties BookmarkListProperties
	err = jsonit.Unmarshal(propertiesRaw, &properties)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal bookmarklist properties")
	}

	bookmarkList.setProperties(&properties)

	return nil
}

func (bookmarkList *BookmarkList) setProperties(properties *BookmarkListProperties) {
	bookmarkList.mutex.Lock()
	defer bookmarkList.mutex.Unlock()
	bookmarkList.properties = properties
}

// GetBookmarks for bookmarkList
func (bookmarkList *BookmarkList) GetBookmarks() []*BookmarkNxContainerEntry {
	bookmarkList.mutex.Lock()
	defer bookmarkList.mutex.Unlock()
	var b []*BookmarkNxContainerEntry
	copy(b, bookmarkList.layout.BookmarkList.Items)
	return bookmarkList.layout.BookmarkList.Items
}
