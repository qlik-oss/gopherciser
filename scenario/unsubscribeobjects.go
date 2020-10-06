package scenario

import (
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	UnsubscribeObjects struct {
		// IDs to unsubscribe to
		IDs []string `json:"ids" displayname:"ID's'" doc-key:"unsubscribeobjects.ids"` // todo add appstructure:"" when array is supported filling with subscribed objects
		// Clear unsubscribes to all objects
		Clear bool `json:"clear" displayname:"Clear" doc-key:"unsubscribeobjects.clear"`
	}
)

// Validate implements ActionSettings interface
func (settings UnsubscribeObjects) Validate() error {
	haveIds := len(settings.IDs) > 0
	if settings.Clear && haveIds {
		return errors.New("both clear and list of ID's given")
	}
	if !settings.Clear && !haveIds {
		return errors.New("no ID's given and set to not clear ")
	}
	return nil
}

// Execute implements ActionSettings interface
func (settings UnsubscribeObjects) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	if settings.Clear {
		ClearObjectSubscriptions(sessionState)
	} else {
		upLink := sessionState.Connection.Sense()
		for _, id := range settings.IDs {
			obj, err := upLink.Objects.GetObjectByID(id)
			if err != nil {
				actionState.AddErrors(errors.WithStack(err))
				return
			}
			if err := upLink.Objects.ClearObject(obj.Handle); err != nil {
				actionState.AddErrors(errors.WithStack(err))
				return
			}
			sessionState.DeRegisterEvent(obj.Handle)
		}
	}
	sessionState.Wait(actionState)
	DebugPrintObjectSubscriptions(sessionState)
}

// AffectsAppObjectsAction implements AffectsAppObjectsAction interface, returns:
// * added *config.AppStructurePopulatedObjects - objects to be added to the selectable list by this action
// * removed []string - ids of objects that are removed (including any children) by this action
// * clearObjects bool - clears all objects except bookmarks and sheets
func (settings UnsubscribeObjects) AffectsAppObjectsAction(structure appstructure.AppStructure) ([]*appstructure.AppStructurePopulatedObjects, []string, bool) {
	return nil, settings.IDs, settings.Clear
}
