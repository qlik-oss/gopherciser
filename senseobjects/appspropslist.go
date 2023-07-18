package senseobjects

import (
	"context"
	"sync"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
)

type (
	AppPropsList struct {
		enigmaObject *enigma.GenericObject
		properties   *enigma.GenericObjectProperties
		layout       *enigma.GenericObjectLayout
		items        map[string]*enigma.GenericObject

		mu sync.Mutex
	}
)

// CreateAppPropsListObject session object
func CreateAppPropsListObject(ctx context.Context, doc *enigma.Doc) (*AppPropsList, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Id:   "AppPropsList",
			Type: "AppPropsList",
		},
		AppObjectListDef: &enigma.AppObjectListDef{
			Type: "appprops",
			Data: json.RawMessage(`{
				"sheetTitleBgColor": "/sheetTitleBgColor",
				"sheetTitleGradientColor": "/sheetTitleGradientColor",
				"sheetTitleColor": "/sheetTitleColor",
				"sheetLogoThumbnail": "/sheetLogoThumbnail",
				"sheetLogoPosition": "/sheetLogoPosition",
				"rtl": "/rtl",
				"theme": "/theme",
				"disableCellNavMenu": "/disableCellNavMenu"
			}`),
		},
	}
	obj, err := doc.CreateSessionObjectRaw(ctx, properties)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create AppPopsList session object")
	}

	return &AppPropsList{
		enigmaObject: obj,
		items:        make(map[string]*enigma.GenericObject, 1),
	}, nil
}

// UpdateProperties of AppPropsList
func (appPropsList *AppPropsList) UpdateProperties(ctx context.Context) error {
	if appPropsList.enigmaObject == nil {
		return errors.Errorf("AppPropsList enigma object is nil")
	}
	propertiesRaw, err := appPropsList.enigmaObject.GetEffectivePropertiesRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	var properties enigma.GenericObjectProperties
	err = json.Unmarshal(propertiesRaw, &properties)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal AppPopsList properties")
	}

	appPropsList.setProperties(&properties)
	return nil
}

func (appPropsList *AppPropsList) setProperties(properties *enigma.GenericObjectProperties) {
	appPropsList.mu.Lock()
	defer appPropsList.mu.Unlock()
	appPropsList.properties = properties
}

// Properties of AppPropsList
func (appPropsList *AppPropsList) Properties() *enigma.GenericObjectProperties {
	appPropsList.mu.Lock()
	defer appPropsList.mu.Unlock()
	return appPropsList.properties
}

// UpdateLayout of AppPropsList
func (appPropsList *AppPropsList) UpdateLayout(ctx context.Context, doc *enigma.Doc, sessionState SessionState, actionState *action.State) error {
	if appPropsList.enigmaObject == nil {
		return errors.Errorf("AppPropsList enigma object is nil")
	}

	layoutRaw, err := appPropsList.enigmaObject.GetLayoutRaw(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	var layout enigma.GenericObjectLayout
	err = json.Unmarshal(layoutRaw, &layout)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal AppPropsList layout")
	}

	appPropsList.setLayout(&layout)

	newItemsCmp := map[string]struct{}{}
	if layout.AppObjectList != nil {
		for _, item := range layout.AppObjectList.Items {
			if item == nil {
				continue
			}
			if item.Info == nil {
				continue
			}
			newItemsCmp[item.Info.Id] = struct{}{}
		}
	}

	removed := []*enigma.GenericObject{}
	oldItemsCmp := map[string]struct{}{}
	for _, item := range appPropsList.items {
		if _, exist := newItemsCmp[item.GenericId]; !exist {
			removed = append(removed, item)
		}
		oldItemsCmp[item.GenericId] = struct{}{}
	}

	added := []string{}
	for itemID := range newItemsCmp {
		if _, exist := oldItemsCmp[itemID]; !exist {
			added = append(added, itemID)
		}
	}

	for _, item := range removed {
		sessionState.DeRegisterEvent(item.Handle)
		delete(appPropsList.items, item.GenericId)
	}

	for _, itemID := range added {
		sessionState.QueueRequest(func(ctx context.Context) error {
			genObj, err := doc.GetObject(ctx, itemID)
			if err != nil {
				return errors.WithStack(err)
			}

			getLayout := func(ctx context.Context) error {
				_, err := genObj.GetLayoutRaw(ctx)
				return err
			}

			sessionState.QueueRequest(getLayout, actionState, true, "")
			sessionState.RegisterEvent(genObj.Handle, func(ctx context.Context, actionState *action.State) error {
				return getLayout(ctx)
			}, nil, true)

			appPropsList.addItem(genObj)

			return nil
		}, actionState, true, "")
	}

	return nil
}

func (appPropsList *AppPropsList) setLayout(layout *enigma.GenericObjectLayout) {
	appPropsList.mu.Lock()
	defer appPropsList.mu.Unlock()
	appPropsList.layout = layout
}

func (appPropsList *AppPropsList) addItem(item *enigma.GenericObject) {
	appPropsList.mu.Lock()
	defer appPropsList.mu.Unlock()
	appPropsList.items[item.GenericId] = item
}

// Layout of AppPropsList
func (appPropsList *AppPropsList) Layout() *enigma.GenericObjectLayout {
	appPropsList.mu.Lock()
	defer appPropsList.mu.Unlock()
	return appPropsList.layout
}

// Remove all items from list and de-register events
func (appPropsList *AppPropsList) RemoveAllItems(sessionState SessionState) {
	appPropsList.mu.Lock()
	defer appPropsList.mu.Unlock()

	for id, item := range appPropsList.items {
		sessionState.DeRegisterEvent(item.Handle)
		delete(appPropsList.items, id)
	}
}
