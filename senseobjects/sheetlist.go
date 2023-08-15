package senseobjects

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/goccy/go-json"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
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
		Y      float64 `json:"y"`
		X      float64 `json:"x"`
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	}

	// SheetData data for a sheet
	SheetData struct {
		Cells []struct {
			Name    string      `json:"name"`
			Type    string      `json:"type"`
			Col     int         `json:"col"`
			Row     int         `json:"row"`
			Colspan int         `json:"colspan"`
			Rowspan int         `json:"rowspan"`
			Bounds  SheetBounds `json:"bounds,omitempty"`
		} `json:"cells,omitempty"`
		Columns               interface{}        `json:"columns"`
		Rows                  interface{}        `json:"rows"`
		Title                 string             `json:"title"`
		LabelExpression       string             `json:"labelExpression"`
		Description           string             `json:"description"`
		DescriptionExpression string             `json:"descriptionExpression"`
		Rank                  interface{}        `json:"rank"`
		ShowCondition         helpers.StringBool `json:"showCondition"`
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
	err = json.Unmarshal(layoutRaw, &layout)
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

// Layout for sheetlist
func (sheetList *SheetList) Layout() *SheetListLayout {
	return sheetList.layout //TODO DECISION: wait for write lock?
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
