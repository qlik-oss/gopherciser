package senseobjects

import (
	"context"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
)

type (
	// SheetMetaDef sheet meta information
	SheetMetaDef struct {
		Title       string `json:"title,omitempty"`
		Description string `json:"description,omitempty"`
	}

	// SheetThumbnail thumbnail object of sheet properties
	SheetThumbnail struct {
		StaticContentURLef *enigma.StaticContentUrlDef `json:"qStaticContentUrlDef,omitempty"`
	}

	// SheetCells cells on sheet
	SheetCells struct {
		Name    string `json:"name"`
		Type    string `json:"type"`
		Col     int    `json:"col,omitempty"`
		Row     int    `json:"row,omitempty"`
		Colspan int    `json:"colspan,omitempty"`
		Rowspan int    `json:"rowspan,omitempty"`
	}

	// SheetProperties properties of sense sheet
	SheetProperties struct {
		Info      *enigma.NxInfo  `json:"qInfo"`
		MetaDef   *SheetMetaDef   `json:"qMetaDef"`
		Rank      float64         `json:"rank,omitempty"`
		Thumbnail *SheetThumbnail `json:"thumbnail,omitempty"`
		Columns   int             `json:"columns,omitempty"`
		Rows      int             `json:"rows,omitempty"`
		Cells     []*SheetCells   `json:"cells"`
	}

	// SheetLayout layout of sense sheet
	SheetLayout struct {
		Info *enigma.NxInfo `json:"qInfo"`
		Meta struct {
			Description  string      `json:"description"`
			Rank         int         `json:"rank"`
			Title        string      `json:"title"`
			Resourcetype string      `json:"_resourcetype"`
			Objecttype   string      `json:"_objecttype"`
			ID           string      `json:"id"`
			Approved     bool        `json:"approved"`
			Published    bool        `json:"published"`
			Owner        interface{} `json:"owner"` // string in QSEoK but struct in QSEoW
			OwnerID      string      `json:"ownerId"`
			CreatedDate  string      `json:"createdDate"`
			ModifiedDate string      `json:"modifiedDate"`
			Privileges   []string    `json:"privileges"`
		} `json:"qMeta"`
		SelectionInfo struct {
		} `json:"qSelectionInfo"`
		Cells []interface{} `json:"cells"`
	}

	// Sheet Sense sheet object
	Sheet struct {
		*enigma.GenericObject
		ID         string
		Properties *SheetProperties
		Layout     *SheetLayout
	}
)

// GetSheet in sense app
func GetSheet(ctx context.Context, app *App, id string) (*Sheet, error) {
	if app == nil {
		return nil, errors.New("app is nil")
	}

	enigmaSheet, errEnigmaSheet := app.Doc.GetObject(ctx, id)
	if errEnigmaSheet != nil {
		return nil, errors.Wrapf(errEnigmaSheet, "Failed to get sheet<%s>", id)
	}

	sheet := &Sheet{
		GenericObject: enigmaSheet,
		ID:            id,
	}

	return sheet, nil
}

// GetProperties for sheet
func (sheet *Sheet) GetProperties(ctx context.Context) (*SheetProperties, error) {
	if sheet == nil {
		return nil, errors.New("sheet is nil")
	}

	raw, err := sheet.GenericObject.GetEffectivePropertiesRaw(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get properties for sheet<%s>", sheet.ID)
	}

	var properties SheetProperties
	err = json.Unmarshal(raw, &properties)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal properties for sheet<%s>", sheet.ID)
	}

	sheet.Properties = &properties

	return sheet.Properties, nil
}

// GetLayout for sheet
func (sheet *Sheet) GetLayout(ctx context.Context) (*SheetLayout, error) {
	if sheet == nil {
		return nil, errors.New("sheet is nil")
	}

	raw, err := sheet.GenericObject.GetLayoutRaw(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get layout for sheet<%s>", sheet.ID)
	}

	var layout SheetLayout
	err = json.Unmarshal(raw, &layout)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal layout for sheet<%s>", sheet.ID)
	}

	sheet.Layout = &layout

	return sheet.Layout, err
}

// SetProperties send updates properties to Sense
func (sheet *Sheet) SetProperties(ctx context.Context) error {
	if sheet == nil {
		return errors.New("sheet is nil")
	}

	raw, err := json.Marshal(sheet.Properties)
	if err != nil {
		return errors.Wrapf(err, "failed marshaling properties for sheet<%s>", sheet.ID)
	}

	err = sheet.GenericObject.SetPropertiesRaw(ctx, raw)
	if err != nil {
		return errors.Wrapf(err, "failed to set properties for sheet<%s>", sheet.ID)
	}

	return nil
}
