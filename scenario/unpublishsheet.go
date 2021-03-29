package scenario

import (
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"

	"context"
)

// UnPublishSheetSettings specifies un-publish sheet settings
type UnPublishSheetSettings struct {
	Mode     PublishSheetMode `json:"mode" displayname:"Un-publish mode" doc-key:"unpublishsheet.mode"`
	SheetIDs []string         `json:"sheetIds" displayname:"Sheet IDs" doc-key:"unpublishsheet.sheetIds"`
}

// Execute performs the un-publish sheet action
func (unPublishSheetSettings UnPublishSheetSettings) Execute(sessionState *session.State,
	actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {

	unPublishAction := func(sheet *senseobjects.Sheet, ctx context.Context) error {
		return sheet.GenericObject.UnPublish(ctx)
	}

	executePubUnPubAction(unPublishSheetSettings.Mode, unPublishSheetSettings.SheetIDs,
		unPublishAction, "failed to un-publish sheet(s)",
		sessionState, actionState)
}

// Validate checks the settings of the un-publish sheet action
func (unPublishSheetSettings UnPublishSheetSettings) Validate() ([]string, error) {
	return validatePubUnPubSettings(unPublishSheetSettings.Mode, unPublishSheetSettings.SheetIDs)
}
