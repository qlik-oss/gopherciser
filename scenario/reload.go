package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/globals/constant"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//ReloadModeEnum determines error handling mode
	ReloadModeEnum int

	// ReloadSettings loop through sheets in an app
	ReloadSettings struct {
		ReloadMode ReloadModeEnum `json:"mode" displayname:"Reload mode" doc-key:"reload.mode"`
		Partial    bool           `json:"partial" displayname:"Partial reload" doc-key:"reload.partial"`
		SaveLog    bool           `json:"log" displayname:"Save log" doc-key:"reload.log"`
		NoSave     bool           `json:"nosave" displayname:"Disable app save" doc-key:"reload.nosave"`
	}
)

const (
	// 0: for default mode.
	DefaultReloadMode ReloadModeEnum = iota
	// 1: for ABEND; the reload of the script ends if an error occurs.
	Abend
	// 2: for ignore; the reload of the script continues even if an error is detected in the script.
	Ignore
)

func (value ReloadModeEnum) GetEnumMap() *enummap.EnumMap {
	enumMap, _ := enummap.NewEnumMap(map[string]int{
		"default": int(DefaultReloadMode),
		"abend":   int(Abend),
		"ignore":  int(Ignore),
	})
	return enumMap
}

// UnmarshalJSON unmarshal ReloadModeEnum
func (value *ReloadModeEnum) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal ReloadModeEnum")
	}

	*value = ReloadModeEnum(i)
	return nil
}

// MarshalJSON marshal ReloadModeEnum type
func (value ReloadModeEnum) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("unknown ReloadModeEnum<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// Validate implements ActionSettings interface
func (settings ReloadSettings) Validate() ([]string, error) {
	return nil, nil
}

// Execute implements ActionSettings interface
func (settings ReloadSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {

	if sessionState.Connection == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment (no connection)"))
		return
	}
	uplink := sessionState.Connection.Sense()
	if uplink == nil {
		actionState.AddErrors(errors.New("no sense connection"))
		return
	}

	app := uplink.CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense app"))
		return
	}

	if err := settings.DoReload(sessionState, actionState, uplink, app.Doc); err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	sessionState.Wait(actionState)
}

func (settings *ReloadSettings) DoReload(sessionState *session.State, actionState *action.State, uplink *enigmahandlers.SenseUplink, doc *enigma.Doc) error {
	saveLog := func(progressMessage string) {
		if settings.SaveLog {
			sessionState.LogEntry.LogInfo("ReloadLog", progressMessage)
		}
	}

	progressMessage, err := settings.doReload(sessionState, actionState, uplink, doc)
	if err != nil {
		saveLog(progressMessage)
		return errors.WithStack(err)
	}

	saveLog(progressMessage)
	return nil
}

func (settings *ReloadSettings) doReload(sessionState *session.State, actionState *action.State, uplink *enigmahandlers.SenseUplink, doc *enigma.Doc) (string, error) {
	// Reserve a RequestID for use with the DoReload method
	ctxWithReservedRequestID, reservedRequestID := doc.WithReservedRequestID(sessionState.BaseContext())

	var progressMessage string
	reloadDone := make(chan struct{})
	progressPoller := func(ctx context.Context) error {
		var progress *enigma.ProgressData
		// Get the progress using the request id we reserved for the reload
		getProgress := func(ctx context.Context) error {
			var err error
			progress, err = uplink.Global.GetProgress(ctx, reservedRequestID)
			return err
		}
		for {
			select {
			case <-reloadDone:
				// Send one last progress message
				if err := sessionState.SendRequest(actionState, getProgress); err != nil {
					return errors.Wrap(err, "Error during reload")
				}
				if settings.SaveLog {
					progressMessage = fmt.Sprintf("%s%s", progressMessage, progress.PersistentProgress)
				}
				return nil
			case <-sessionState.BaseContext().Done():
				return nil
			case <-time.After(time.Duration(constant.ReloadPollInterval)):
				if err := sessionState.SendRequest(actionState, getProgress); err != nil {
					return errors.Wrap(err, "Error during reload")
				}
				if settings.SaveLog {
					progressMessage = fmt.Sprintf("%s%s", progressMessage, progress.PersistentProgress)
				}
			}
		}
	}

	var progressError error
	sessionState.QueueRequestWithCallback(progressPoller, actionState, false, "error during reload of app", func(err error) {
		progressError = err
	})

	var status bool
	doReload := func(ctx context.Context) error {
		defer close(reloadDone)
		var err error
		status, err = doc.DoReload(ctxWithReservedRequestID, int(settings.ReloadMode), settings.Partial, false)
		return err
	}

	if err := sessionState.SendRequest(actionState, doReload); err != nil {
		return progressMessage, errors.Wrap(err, "Error when reloading app")
	}

	if progressError != nil {
		return progressMessage, errors.Wrap(progressError, "Error when reloading app")
	}

	if !status {
		return progressMessage, errors.Errorf("Reload failed")
	} else if !settings.NoSave {
		// save the app after reload if it was successful
		if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			return doc.DoSave(ctx, "")
		}); err != nil {
			return progressMessage, errors.Wrap(err, "failed to save app")
		}
	}

	return progressMessage, nil
}
