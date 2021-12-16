package senseobjects

import (
	"context"
	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"sync"

	"github.com/qlik-oss/enigma-go/v3"
)

type (
	// DynamicAppViewList container with dav list in sense app
	DynamicAppViewList struct {
		enigmaObject *enigma.GenericObject
		layout       *DynamicAppViewListLayout
		mutex        sync.Mutex
	}

	DynamicAppViewListLayout struct {
		enigma.GenericObjectLayout
		AppObjectList *AppObjectList `json:"qAppObjectList"`
	}

	AppObjectList struct {
		Items []*AppObjectListItem `json:"qItems"`
	}

	AppObjectListItem struct {
		Info *AppObjectListItemInfo `json:"qInfo"`
		Meta *AppObjectListItemMeta `json:"qMeta"`
		Data *AppObjectListItemData `json:"qData"`
	}

	AppObjectListItemInfo struct {
		Id   string `json:"qId"`
		Type string `json:"qType"`
	}

	AppObjectListItemMeta struct {
		Name string `json:"qName"`
	}

	AppObjectListItemData struct {
		OdagLinkRef string `json:"odagLinkRef"`
	}
)

// UpdateLayout get and set a new layout for sheetlist
func (davlist *DynamicAppViewList) UpdateLayout(ctx context.Context) error {
	if davlist.enigmaObject == nil {
		return errors.Errorf("DynamicAppViewList enigma object is nil")
	}

	layoutRaw, err := davlist.enigmaObject.GetLayoutRaw(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get DynamicAppViewList layout")
	}

	var layout DynamicAppViewListLayout
	err = json.Unmarshal(layoutRaw, &layout)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal DynamicAppViewList layout")
	}

	davlist.setLayout(&layout)
	return nil
}

func (davlist *DynamicAppViewList) setLayout(layout *DynamicAppViewListLayout) {
	davlist.mutex.Lock()
	defer davlist.mutex.Unlock()
	davlist.layout = layout
}

// Layout for DynamicAppViewList
func (davlist *DynamicAppViewList) Layout() *DynamicAppViewListLayout {
	return davlist.layout //TODO DECISION: wait for write lock?
}

// CreateDynamicAppViewList create dav session object
func CreateDynamicAppViewList(ctx context.Context, doc *enigma.Doc) (*DynamicAppViewList, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "DynamicAppViewList",
		},
		AppObjectListDef: &enigma.AppObjectListDef{
			Type: "dynamicappview",
			Data: json.RawMessage("{\"odagLinkRef\":\"/qMetaDef/odagLinkRef\"}"),
		},
	}

	obj, err := doc.CreateSessionObjectRaw(ctx, properties)
	if err != nil {
		return nil, err
	}

	return &DynamicAppViewList{
		enigmaObject: obj,
	}, nil
}
