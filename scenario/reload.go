package scenario

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
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
func (settings ReloadSettings) Execute(sessionState *session.State,
	actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {

	if sessionState.Connection == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment (no connection)"))
		return
	}

	connection := sessionState.Connection.Sense()
	if connection == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense environment (no uplink)"))
		return
	}

	app := connection.CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("Not connected to a Sense app"))
		return
	}

	// Reserve a RequestID for use with the DoReload method
	ctxWithReservedRequestID, reservedRequestID := app.Doc.WithReservedRequestID(sessionState.BaseContext())

	progressMessage := ""
	reloadDone := make(chan struct{})
	defer close(reloadDone)
	go func() {
		for {
			select {
			case <-reloadDone:
				return
			case <-time.After(time.Duration(constant.ReloadPollInterval)):

				// Get the progress using the request id we reserved for the reload
				var progress *enigma.ProgressData
				getProgress := func(ctx context.Context) error {
					if sessionState.Connection == nil {
						return errors.New("Not connected to a Sense environment (no connection)")
					}
					uplink := sessionState.Connection.Sense()
					if uplink == nil {
						return errors.New("no sense connection")
					}
					var err error
					progress, err = uplink.Global.GetProgress(ctx, reservedRequestID)
					return err
				}
				if err := sessionState.SendRequest(actionState, getProgress); err != nil {
					actionState.AddErrors(errors.Wrap(err, "Error during reload"))
					if settings.SaveLog {
						sessionState.LogEntry.LogInfo("ReloadLog", progressMessage)
					}
					return
				}
				if settings.SaveLog {
					progressMessage = fmt.Sprintf("%s%s", progressMessage, progress.PersistentProgress)
				}
			}
		}
	}()

	var status bool
	doReload := func(ctx context.Context) error {
		var err error
		status, err = app.Doc.DoReload(ctxWithReservedRequestID, int(settings.ReloadMode), settings.Partial, false)
		return err
	}
	if err := sessionState.SendRequest(actionState, doReload); err != nil {
		actionState.AddErrors(errors.Wrap(err, "Error when reloading app"))
		sessionState.LogEntry.LogInfo("ReloadLog", progressMessage)
		return
	}

	if !status {
		actionState.AddErrors(errors.Errorf("Reload failed"))
		// don't return so ReloadLog will be save
	} else {
		// save the app after reload if it was successful
		if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			return connection.CurrentApp.Doc.DoSave(ctx, "")
		}); err != nil {
			actionState.AddErrors(errors.Wrap(err, "failed to save app"))
			// don't return so ReloadLog will be save
		}
	}

	if settings.SaveLog {
		sessionState.LogEntry.LogInfo("ReloadLog", progressMessage)
	}
	sessionState.Wait(actionState)
}
