package senseobjects

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	// SheetListLayout  sheetlist layout
	SheetListLayout struct {
		enigma.GenericObjectLayout
		AppObjectList *SheetListAppObjectList `json:"qAppObjectList,omitempty"`
	}

	// SheetListAppObjectList sheetlist app object list
	SheetListAppObjectList struct {
		enigma.AppObjectList
		Items []*SheetNxContainerEntry `json:"qItems,omitempty"`
	}

	// SheetNxContainerEntry container sheet data
	SheetNxContainerEntry struct {
		enigma.NxContainerEntry
		Data *SheetData `json:"qData,omitempty"`
		Meta SheetMeta  `json:"qMeta,omitempty"`
	}

	SheetMeta struct {
		enigma.NxMeta
		Published bool `json:"published,omitempty"`
		Approved  bool `json:"approved,omitempty"`
	}

	SheetBounds struct {
		Y      int `json:"y,omitempty"`
		X      int `json:"x,omitempty"`
		Width  int `json:"width,omitempty"`
		Height int `json:"height,omitempty"`
	}

	// SheetData data for a sheet
	SheetData struct {
		Cells []struct {
			Name string `json:"name,omitempty"`
			Type string `json:"type,omitempty"`
		} `json:"cells,omitempty"`
		Columns               int                `json:"columns,omitempty"`
		Rows                  int                `json:"rows,omitempty"`
		Colspan               int                `json:"colspan,omitempty"`
		Rowspan               int                `json:"rowspan,omitempty"`
		Bounds                SheetBounds        `json:"bounds,omitempty"`
		Title                 string             `json:"title,omitempty"`
		LabelExpression       string             `json:"labelExpression,omitempty"`
		Description           string             `json:"description,omitempty"`
		DescriptionExpression string             `json:"descriptionExpression,omitempty"`
		Rank                  interface{}        `json:"rank,omitempty"`
		ShowCondition         helpers.StringBool `json:"showCondition,omitempty"`
	}

	// SheetListPropertiesData properties of sheetlist
	SheetListPropertiesData struct {
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
		Cells       string `json:"cells,omitempty"`
	}

	// SheetListProperties SheetList properties
	SheetListProperties struct {
		Info    enigma.NxInfo            `json:"qInfo,omitempty"`
		MetaDef enigma.NxMetaDef         `json:"qMetaDef,omitempty"`
		Data    *SheetListPropertiesData `json:"qData,omitempty"`
	}

	// SheetList container with sheet in sense app
	SheetList struct {
		enigmaObject *enigma.GenericObject
		layout       *SheetListLayout
		properties   *SheetListProperties
		mutex        sync.Mutex
	}

	// SheetEntryNotFoundError error returned when sheet entry was not found in sheet list
	SheetEntryNotFoundError string
)

// Error returned when sheet entry was not found in sheet list
func (err SheetEntryNotFoundError) Error() string {
	return fmt.Sprintf("no sheet entry found for id<%s>", string(err))
}

func (sheetList *SheetList) setLayout(layout *SheetListLayout) {
	sheetList.mutex.Lock()
	defer sheetList.mutex.Unlock()
	sheetList.layout = layout
}

// UpdateLayout get and set a new layout for sheetlist
func (sheetList *SheetList) UpdateLayout(ctx context.Context) error {
	if sheetList.enigmaObject == nil {
		return errors.Errorf("sheetlist enigma object is nil")
	}

	layoutRaw, err := sheetList.enigmaObject.GetLayoutRaw(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get sheetlist layout")
	}

	var layout SheetListLayout
	err = jsonit.Unmarshal(layoutRaw, &layout)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal sheetlist layout")
	}

	// sort sheets after rank (same order as shown in app)
	if layout.AppObjectList != nil || layout.AppObjectList.Items != nil {
		sheetItems := layout.AppObjectList.Items
		sort.Slice(sheetItems, func(i, j int) bool {
			item1, item2 := sheetItems[i], sheetItems[j]
			if item1 == nil || item2 == nil || item1.Data == nil || item2.Data == nil {
				return false
			}
			iRank, ok := sheetItems[i].Data.Rank.(float64)
			if !ok {
				iRank = 0
			}
			jRank, ok := sheetItems[j].Data.Rank.(float64)
			if !ok {
				jRank = 0
			}
			return iRank < jRank
		})
	}

	sheetList.setLayout(&layout)
	return nil
}

// UpdateProperties get and set properties for sheetlist
func (sheetList *SheetList) UpdateProperties(ctx context.Context) error {
	if sheetList.enigmaObject == nil {
		return errors.Errorf("sheetlist enigma object is nil")
	}

	propertiesRaw, err := sheetList.enigmaObject.GetPropertiesRaw(ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to unmarshal sheetlist properties")
	}

	var properties SheetListProperties
	err = jsonit.Unmarshal(propertiesRaw, &properties)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal sheetlist properties")
	}

	sheetList.setProperties(&properties)

	return nil
}

func (sheetList *SheetList) setProperties(properties *SheetListProperties) {
	sheetList.mutex.Lock()
	defer sheetList.mutex.Unlock()
	sheetList.properties = properties
}

// Layout for sheetlist
func (sheetList *SheetList) Layout() *SheetListLayout {
	return sheetList.layout //TODO DECISION: wait for write lock?
}

// Properties for sheetlist
func (sheetList *SheetList) Properties() *SheetListProperties {
	return sheetList.properties
}

// GetSheetEntry Get sheet entry from sheet list
func (sheetList *SheetList) GetSheetEntry(sheetid string) (*SheetNxContainerEntry, error) {
	for _, v := range sheetList.layout.AppObjectList.Items {
		if v.Info.Id == sheetid {
			return v, nil
		}
	}
	return nil, SheetEntryNotFoundError(sheetid)
}

// CreateSheetListObject create sheetlist session object
func CreateSheetListObject(ctx context.Context, doc *enigma.Doc) (*SheetList, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Id:   "SheetList",
			Type: "SheetList",
		},
		AppObjectListDef: &enigma.AppObjectListDef{
			Type: "sheet",
			Data: json.RawMessage(`{
				"title": "/qMetaDef/title",
				"labelExpression": "/labelExpression",
				"showCondition": "/showCondition",
				"description": "/qMetaDef/description",
				"descriptionExpression": "/descriptionExpression",
				"thumbnail": "/thumbnail",
				"cells": "/cells",
				"rank": "/rank",
				"columns": "/columns",
				"rows": "/rows"
}`),
		},
	}

	obj, err := doc.CreateSessionObjectRaw(ctx, properties)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create sheetlist session object in app<%s>", doc.GenericId)
	}

	return &SheetList{
		enigmaObject: obj,
	}, nil
}
