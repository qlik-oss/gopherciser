package session

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjdef"
)

type (
	AutoChartInstance struct {
		DefaultHandlerInstance
		ObjectDef *senseobjdef.ObjectDef
	}

	AutoChartHandler struct {
		DefaultHandler
	}
)

const (
	GeneratedPropertiesPath = "/qUndoExclude/generated"
)

func (handler *AutoChartHandler) Instance(id string) ObjectHandlerInstance {
	return &AutoChartInstance{DefaultHandlerInstance{Id: id}, nil}
}

// GetObject implement ObjectHandler interface
func (instance *AutoChartInstance) SetObjectAndEvents(sessionState *State, actionState *action.State, obj *enigmahandlers.Object, genObj *enigma.GenericObject) {
	instance.handleAutoChart(sessionState, actionState, obj, genObj)
}

func (instance *AutoChartInstance) GetObjectDefinition(objectType string) (string, senseobjdef.SelectType, senseobjdef.DataDefType, error) {
	// Todo error checking
	return instance.ObjectDef.Select.Path, instance.ObjectDef.Select.Type, instance.ObjectDef.DataDef.Type, nil
}

func (instance *AutoChartInstance) handleAutoChart(sessionState *State, actionState *action.State, autochartObj *enigmahandlers.Object, autochartGen *enigma.GenericObject) {
	rawLayout, generatedProperties := getRawLayoutAndGeneratedProperties(sessionState, actionState, autochartObj, autochartGen)
	if actionState.Failed {
		return // at least one of the async request had an error
	}

	// Set child list and children from auto-chart main object
	if err := SetChildList(rawLayout, autochartObj); err != nil {
		actionState.AddErrors(err)
		return
	}
	if err := SetChildren(rawLayout, autochartObj); err != nil {
		actionState.AddErrors(err)
	}

	// get generated layout
	var generatedLayout json.RawMessage
	generatedLayoutPath := helpers.NewDataPath(GeneratedPropertiesPath) // Path same as for properties
	rawGeneratedLayout, errDataPath := generatedLayoutPath.Lookup(rawLayout)
	if errDataPath != nil {
		actionState.AddErrors(errors.Wrapf(errDataPath, "Failed to get generated layout for autochart<%s>", autochartGen.GenericId))
		return
	}
	if err := jsonit.Unmarshal(rawGeneratedLayout, &generatedLayout); err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to unmarshal auto-chart<%s> generated layout.", autochartObj.ID))
		return
	}

	if generatedLayout != nil {
		// Empty childList? Also check generated layout
		childList := autochartObj.ChildList()
		if childList == nil || len(childList.Items) < 1 {
			if err := SetChildList(generatedLayout, autochartObj); err != nil {
				actionState.AddErrors(err)
				return
			}
		}

		// No external references? Also check generated layout
		if len(autochartObj.ExternalReferenceApps()) < 1 {
			if err := SetChildren(generatedLayout, autochartObj); err != nil {
				actionState.AddErrors(err)
			}
		}
	}

	// Get object definitions for generated object
	objectDef, err := senseobjdef.GetObjectDef(generatedProperties.Info.Type)
	if err != nil {
		switch errors.Cause(err).(type) {
		case senseobjdef.NoDefError:
			sessionState.LogEntry.Logf(logger.WarningLevel, "Get Data for auto-chart generated object type<%s> not supported", generatedProperties.Info.Type)
			return
		default:
			actionState.AddErrors(err)
			return
		}
	}
	sessionState.LogEntry.LogDebugf("object<%s> objectdef<%+v>", autochartObj.ID, instance.ObjectDef)

	// clone and add paths to generated object def
	instance.ObjectDef = &senseobjdef.ObjectDef{
		DataDef: senseobjdef.DataDef{
			Type: objectDef.DataDef.Type,
			Path: helpers.DataPath(fmt.Sprintf("%s%s", GeneratedPropertiesPath, objectDef.DataDef.Path)),
		},
	}

	instance.SetObjectDefData(objectDef.Data)

	if objectDef.Select != nil {
		// de-reference pointer and update path
		selectDef := *objectDef.Select
		if selectDef.Path != "" {
			selectDef.Path = fmt.Sprintf("%s%s", GeneratedPropertiesPath, selectDef.Path)
		}
		instance.ObjectDef.Select = &selectDef
	}

	if err := SetObjectData(sessionState, actionState, rawLayout, instance.ObjectDef, autochartObj, autochartGen); err != nil {
		actionState.AddErrors(err)
		return
	}

	instance.setupEvents(sessionState, autochartObj, autochartGen)

	// Get any child objects
	children := autochartObj.ChildList()
	if children != nil && len(children.Items) > 0 {
		sessionState.LogEntry.LogDebugf("object<%s> type<%s> has children", autochartGen.GenericId, autochartGen.GenericType)
		for _, child := range children.Items {
			GetAndAddObjectAsync(sessionState, actionState, child.Info.Id)
		}
	}

	sessionState.LogEntry.LogDebugf("Getting layout for object<%s> handle<%d> type<%s> END", autochartObj.ID, autochartObj.Handle, autochartGen.GenericType)
}

func (instance *AutoChartInstance) SetObjectDefData(objDefData []senseobjdef.Data) {
	if len(objDefData) < 1 {
		return
	}

	instance.ObjectDef.Data = make([]senseobjdef.Data, 0, len(objDefData))
	for _, defaultData := range objDefData {
		// make a copy of default data, de-reference pointers and update to new path
		data := senseobjdef.Data{}
		for i, constraint := range defaultData.Constraints {
			if i == 0 {
				data.Constraints = make([]*senseobjdef.Constraint, 0, len(defaultData.Constraints))
			}
			data.Constraints = append(data.Constraints, &senseobjdef.Constraint{
				Path:     helpers.DataPath(fmt.Sprint(GeneratedPropertiesPath, constraint.Path)),
				Value:    constraint.Value,
				Required: constraint.Required,
			})
		}

		for i, request := range defaultData.Requests {
			if i == 0 {
				data.Requests = make([]senseobjdef.GetDataRequests, 0, len(defaultData.Requests))
			}
			data.Requests = append(data.Requests, senseobjdef.GetDataRequests{
				Type:   request.Type,
				Path:   fmt.Sprint(GeneratedPropertiesPath, request.Path),
				Height: request.Height,
			})
		}

		instance.ObjectDef.Data = append(instance.ObjectDef.Data, data)
	}
}

func (instance *AutoChartInstance) setupEvents(sessionState *State, autochartObj *enigmahandlers.Object, autochartGen *enigma.GenericObject) {
	event := func(ctx context.Context, as *action.State) error {
		sessionState.LogEntry.LogDebugf("Getting layout for object<%s> handle<%d> type<%s> START", autochartObj.ID, autochartObj.Handle, autochartGen.GenericType)
		layout, err := sessionState.SendRequestRaw(as, autochartGen.GetLayoutRaw)
		if err != nil {
			return errors.Wrapf(err, "object<%s>.GetLayout", autochartGen.GenericId)
		}

		if err := SetObjectData(sessionState, as, layout, instance.ObjectDef, autochartObj, autochartGen); err != nil {
			return errors.WithStack(err)
		}
		sessionState.LogEntry.LogDebugf("Getting layout for object<%s> handle<%d> type<%s> END", autochartObj.ID, autochartObj.Handle, autochartGen.GenericType)
		return nil
	}
	sessionState.RegisterEvent(autochartObj.Handle, event, nil, true)
}

// returns auto-chart raw layout, and generated properties
func getRawLayoutAndGeneratedProperties(sessionState *State, actionState *action.State,
	autochartObj *enigmahandlers.Object, autochartGen *enigma.GenericObject) (json.RawMessage, enigma.GenericObjectProperties) {
	var wg sync.WaitGroup
	wg.Add(1)
	var rawLayout json.RawMessage
	sessionState.QueueRequest(func(ctx context.Context) error {
		defer wg.Done()
		sessionState.LogEntry.LogDebugf("Getting layout for object<%s> handle<%d> type<%s> START", autochartObj.ID, autochartObj.Handle, autochartGen.GenericType)
		var err error
		if rawLayout, err = autochartGen.GetLayoutRaw(ctx); err != nil {
			return errors.Wrapf(err, "object<%s>.GetLayout", autochartGen.GenericId)
		}
		return nil
	}, actionState, true, "")

	wg.Add(1)
	var generatedProperties enigma.GenericObjectProperties
	sessionState.QueueRequest(func(ctx context.Context) error {
		defer wg.Done()
		rawAutoChartProperties, err := sessionState.SendRequestRaw(actionState, autochartGen.GetEffectivePropertiesRaw)
		if err != nil {
			return errors.Wrapf(err, "object<%s>.GetEffectiveProperties", autochartGen.GenericId)
		}

		var autoChartProp enigma.GenericObjectProperties
		if err = jsonit.Unmarshal(rawAutoChartProperties, &autoChartProp); err != nil {
			return errors.Wrap(err, "Failed to unmarshal auto-chart properties to GenericObjectProperties")
		}

		// Get generated properties
		generatedPropPath := helpers.NewDataPath(GeneratedPropertiesPath)
		rawGeneratedProp, errDataPath := generatedPropPath.Lookup(rawAutoChartProperties)
		if errDataPath != nil {
			return errors.Wrapf(errDataPath, "Failed to get generated properties for autochart<%s>", autochartGen.GenericId)
		}
		if err := jsonit.Unmarshal(rawGeneratedProp, &generatedProperties); err != nil {
			return errors.Wrapf(err, "failed to unmarshal auto-chart<%s> generated properties.", autochartObj.ID)
		}
		autochartObj.SetProperties(&generatedProperties)

		return nil
	}, actionState, true, "")

	wg.Wait()
	return rawLayout, generatedProperties
}
