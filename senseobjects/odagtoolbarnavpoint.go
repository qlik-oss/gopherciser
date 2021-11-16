package senseobjects

import (
	"context"
	"sync"

	"github.com/qlik-oss/gopherciser/senseobjdef"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
)

type (
	// OdagToolbarNavpointLayout odag-toolbar-navpoint layout
	OdagToolbarNavpointLayout struct {
		enigma.GenericObjectLayout
	}

	// OdagToolbarNavpoint container with odag-toolbar-navpoint in sense app
	OdagToolbarNavpoint struct {
		enigmaObject *enigma.GenericObject
		layout       *OdagToolbarNavpointLayout
		mutex        sync.Mutex
	}
)

func (otn *OdagToolbarNavpoint) setLayout(layout *OdagToolbarNavpointLayout) {
	otn.mutex.Lock()
	defer otn.mutex.Unlock()
	otn.layout = layout
}

// UpdateLayout get and set a new layout for odag-toolbar-navpoint
func (otn *OdagToolbarNavpoint) UpdateLayout(ctx context.Context) error {
	if otn.enigmaObject == nil {
		return errors.Errorf("otn enigma object is nil")
	}

	layoutRaw, err := otn.enigmaObject.GetLayoutRaw(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to get otn layout")
	}

	var layout OdagToolbarNavpointLayout
	err = jsonit.Unmarshal(layoutRaw, &layout)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal otn layout")
	}

	otn.setLayout(&layout)
	return nil
}

// GetHypercubeData get datapages
func (otn *OdagToolbarNavpoint) GetHypercubeData(ctx context.Context) ([]*enigma.NxDataPage, error) {
	objDef, err := senseobjdef.GetObjectDef("odag-toolbar-navpoint")
	if err != nil {
		return nil, err
	}

	return otn.enigmaObject.GetHyperCubeData(ctx, string(objDef.DataDef.Path), []*enigma.NxPage{
		{
			Left:   0,
			Top:    0,
			Width:  1,
			Height: otn.layout.GenericObjectLayout.HyperCube.Size.Cy,
		},
	})
}

// Layout for odag-toolbar-navpoint
func (otn *OdagToolbarNavpoint) Layout() *OdagToolbarNavpointLayout {
	return otn.layout //TODO DECISION: wait for write lock?
}

// CreateOdagToolbarNavpoint create odag-toolbar-navpoint session object
func CreateOdagToolbarNavpoint(ctx context.Context, doc *enigma.Doc, name string) (*OdagToolbarNavpoint, error) {
	properties := &enigma.GenericObjectProperties{
		Info: &enigma.NxInfo{
			Type: "odag-toolbar-navpoint",
		},
		HyperCubeDef: &enigma.HyperCubeDef{
			Dimensions: []*enigma.NxDimension{
				{
					Def: &enigma.NxInlineDimensionDef{
						FieldDefs: []string{name},
					},
					NullSuppression: true,
				},
			},
			InitialDataFetch: []*enigma.NxPage{{Height: 20, Width: 1}},
		},
	}

	obj, err := doc.CreateSessionObjectRaw(ctx, properties)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create odag-toolbar-navpoint session object in app<%s>", doc.GenericId)
	}

	return &OdagToolbarNavpoint{
		enigmaObject: obj,
	}, nil
}
