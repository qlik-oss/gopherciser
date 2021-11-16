package session

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/senseobjects"
)

// GetActiveDoc get active doc from engine
func (state *State) GetActiveDoc(actionState *action.State, upLink *enigmahandlers.SenseUplink) (*enigma.Doc, error) {
	var doc *enigma.Doc
	err := state.SendRequest(actionState, func(ctx context.Context) error {
		activeDoc, err := upLink.Global.GetActiveDoc(ctx)
		if err != nil {
			return errors.WithStack(NoActiveDocError{Err: err})
		} else if activeDoc == nil {
			return errors.WithStack(NoActiveDocError{Msg: "No Active doc found on reconnect."})
		}

		doc = activeDoc
		return nil
	})
	return doc, errors.WithStack(err)
}

// GetSheetList create and update sheetlist session object if not existing
func (state *State) GetSheetList(actionState *action.State, uplink *enigmahandlers.SenseUplink) {
	sheetList, err := uplink.CurrentApp.GetSheetList(state, actionState)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	slLayout := sheetList.Layout()
	if slLayout == nil {
		actionState.AddErrors(errors.New("sheetlist layout is nil"))
		return
	}

	if state.LogEntry.ShouldLogDebug() &&
		slLayout.AppObjectList != nil &&
		slLayout.AppObjectList.Items != nil {

		for _, v := range slLayout.AppObjectList.Items {
			state.LogEntry.LogDebugf("Sheet<%s> found", v.Info.Id)
		}
	}
}

// SetupChangeChan start listening to enigma change channel
func (state *State) SetupChangeChan() error {
	if state == nil {
		return errors.New("use of nil state")
	}

	if state.Connection == nil {
		return errors.New("connection is nil")
	}

	if state.Connection.Sense() == nil {
		return errors.New("sense uplink is nil")
	}

	if state.Connection.Sense().Global == nil {
		return errors.New("uplink global is nil")
	}

	changeChan := state.Connection.Sense().Global.ChangeListsChannel(true)
	go func() {
		for {
			select {
			case cl, ok := <-changeChan:
				if !ok {
					return
				}

				if len(cl.Changed) > 0 {
					state.LogEntry.LogInfo("Pushed ChangedList", fmt.Sprintf("%v", cl.Changed))
				}

				if len(cl.Closed) > 0 {
					state.LogEntry.LogInfo("Pushed ClosedList", fmt.Sprintf("%v", cl.Closed))
				}

				state.TriggerEvents(state.CurrentActionState, cl.Changed, cl.Closed)
			case <-state.BaseContext().Done():
				return
			}

		}
	}()

	return nil
}

// GetSheet get sheet and add to object list
func (state *State) GetSheet(actionState *action.State, upLink *enigmahandlers.SenseUplink, id string) (*enigmahandlers.Object, *senseobjects.Sheet, error) {
	app := upLink.CurrentApp
	if app == nil {
		err := errors.New("Not connected to a Sense app")
		return nil, nil, err
	}

	var sheet *senseobjects.Sheet
	getSheet := func(ctx context.Context) error {
		var err error
		id = state.IDMap.Get(id)
		sheet, err = senseobjects.GetSheet(ctx, app, id)
		return err
	}
	if err := state.SendRequest(actionState, getSheet); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	if sheet == nil {
		return nil, nil, errors.New("sheet is nil")
	}
	state.LogEntry.LogDebugf("Fetched sheet<%s> successfully", id)

	getProperties := func(ctx context.Context) error {
		_, err := sheet.GetProperties(ctx)
		return err
	}
	if err := state.SendRequest(actionState, getProperties); err != nil {
		return nil, nil, errors.WithStack(err)
	}

	sheetObject := enigmahandlers.NewObject(sheet.Handle, enigmahandlers.ObjTypeSheet, id, sheet)
	if err := upLink.Objects.AddObject(sheetObject); err != nil {
		return nil, nil, errors.Wrap(err, "failed to add object to object list")
	}

	return sheetObject, sheet, nil
}
