package scenario

import (
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	UnsubscribeObjects struct{}
)

// Validate implements ActionSettings interface
func (settings UnsubscribeObjects) Validate() error {
	// Todo
	return nil
}

// Execute implements ActionSettings interface
func (settings UnsubscribeObjects) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	// Todo
	DebugPrintObjectSubscriptions(sessionState)
}

// AffectsAppObjectsAction implements AffectsAppObjectsAction interface, returns:
// * added *config.AppStructurePopulatedObjects - objects to be added to the selectable list by this action
// * removed []string - ids of objects that are removed (including any children) by this action
// * clearObjects bool - clears all objects except bookmarks and sheets
func (settings UnsubscribeObjects) AffectsAppObjectsAction(structure appstructure.AppStructure) (*appstructure.AppStructurePopulatedObjects, []string, bool) {
	// Todo
	return nil, nil, false
}
