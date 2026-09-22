package scenario

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	// SheetChangerSettings loop through sheets in an app
	SheetChangerSettings struct {
		// ThinkTime in between changing sheets
		InterThinkTimeSettings *ThinkTimeSettings `json:"thinktimesettings,omitempty" doc-key:"sheetchanger.thinktimesettings" displayname:"Think time inbetween actions"`
	}
)

// Validate implements ActionSettings interface
func (settings SheetChangerSettings) Validate() ([]string, error) {
	if settings.InterThinkTimeSettings != nil &&
		!(settings.InterThinkTimeSettings.Type == helpers.StaticDistribution && helpers.NearlyEqual(settings.InterThinkTimeSettings.Delay, 0.0)) { // don't return error on GUI creating empty section
		return settings.InterThinkTimeSettings.Validate()
	}

	return nil, nil
}

// Execute implements ActionSettings interface
func (settings SheetChangerSettings) Execute(sessionState *session.State,
	actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {

	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment"))
		return
	}
	uplink := sessionState.Connection.Sense()

	app := uplink.CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense app"))
		return
	}

	// Create list of existing sheets
	sheetList, err := app.GetSheetList(sessionState, actionState)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	items := sheetList.Layout().AppObjectList.Items
	sheetIDs := make([]string, 0, len(items))
	for _, item := range items {
		sheetIDs = append(sheetIDs, item.Info.Id)

	}

	// set a default label for sheet changer action if user set none
	if label == "" {
		label = "sheet changer"
	}

	var timerAction *Action
	if settings.InterThinkTimeSettings != nil &&
		!(settings.InterThinkTimeSettings.Type == helpers.StaticDistribution && helpers.NearlyEqual(settings.InterThinkTimeSettings.Delay, 0.0)) {
		timerAction = &Action{
			ActionCore{
				Type:  ActionChangeSheet,
				Label: fmt.Sprintf("%s - inter thinktime", label),
			},
			&ThinkTimeSettings{
				DistributionSettings: settings.InterThinkTimeSettings.DistributionSettings,
			},
		}
	}

	for i, sheetID := range sheetIDs {
		ac := Action{
			ActionCore{
				Type:  ActionChangeSheet,
				Label: fmt.Sprintf("%s (%s)", label, sheetID),
			},
			&ChangeSheetSettings{
				ID: sheetID,
			},
		}

		if isAborted, err := CheckActionError(ac.Execute(sessionState, connectionSettings)); isAborted {
			return // action is aborted, we should not continue
		} else if err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return
		}

		if i+1 < len(sheetIDs) && timerAction != nil {
			if isAborted, err := CheckActionError(timerAction.Execute(sessionState, connectionSettings)); isAborted {
				return // action is aborted, we should not continue
			} else if err != nil {
				actionState.AddErrors(errors.WithStack(err))
				return
			}
		}
	}

	sessionState.Wait(actionState)
}

// IsContainerAction implements ContainerAction interface
// and sets container action logging to original action entry
func (settings SheetChangerSettings) IsContainerAction() {}
