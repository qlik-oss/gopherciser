package scenario

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	SubscribeObjectsSettings struct {
		ClearCurrent bool     `json:"clear" displayname:"Clear" doc-key:"subscribeobjects.clear"`
		IDs          []string `json:"ids" displayname:"ID's'" doc-key:"subscribeobjects.ids"` // todo add appstructure:"" when array is supported filling with all objects
	}
)

// Validate implements ActionSettings interface
func (settings SubscribeObjectsSettings) Validate() error {
	if len(settings.IDs) < 1 {
		return errors.New("no ID defined to subscribe to")
	}

	for i, id := range settings.IDs {
		if id == "" {
			return errors.Errorf("id in array position %d is empty", i)
		}
	}
	return nil
}

// Execute implements ActionSettings interface
func (settings SubscribeObjectsSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	actionState.Details = strings.Join(settings.IDs, ",")

	if settings.ClearCurrent {
		sessionState.ClearObjectSubscriptions()
	}

	for _, id := range settings.IDs {
		session.GetAndAddObjectAsync(sessionState, actionState, sessionState.IDMap.Get(id))
	}

	sessionState.Wait(actionState)
	DebugPrintObjectSubscriptions(sessionState)
}

// AffectsAppObjectsAction implements AffectsAppObjectsAction interface, returns:
// * added *config.AppStructurePopulatedObjects - objects to be added to the selectable list by this action
// * removed []string - ids of objects that are removed (including any children) by this action
// * clearObjects bool - clears all objects except bookmarks and sheets
func (settings SubscribeObjectsSettings) AffectsAppObjectsAction(structure appstructure.AppStructure) ([]*appstructure.AppStructurePopulatedObjects, []string, bool) {
	if len(settings.IDs) < 1 {
		return nil, nil, settings.ClearCurrent
	}

	parents := make([]*appstructure.AppStructurePopulatedObjects, 0, len(settings.IDs))
	for _, id := range settings.IDs {
		selectables, err := structure.GetSelectables(id)
		if err != nil {
			return nil, nil, settings.ClearCurrent
		}
		newObjs := appstructure.AppStructurePopulatedObjects{
			Parent:    id,
			Objects:   selectables,
			Bookmarks: nil,
		}
		parents = append(parents, &newObjs)
	}

	return parents, nil, settings.ClearCurrent
}
