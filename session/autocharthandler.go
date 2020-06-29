package session

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
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
	generatedLayoutPath := senseobjdef.NewDataPath(GeneratedPropertiesPath) // Path same as for properties
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
			Path: senseobjdef.DataPath(fmt.Sprintf("%s%s", GeneratedPropertiesPath, objectDef.DataDef.Path)),
		},
	}

		return nil
	}, actionState, true, "")

	wg.Add(1)
	var generatedProperties enigma.GenericObjectProperties
	sessionState.QueueRequest(func(ctx context.Context) error {
		defer wg.Done()
		rawAutoChartProperties, err := sessionState.SendRequestRaw(actionState, autochartGen.GetPropertiesRaw)
		if err != nil {
			return errors.Wrapf(err, "object<%s>.GetProperties", autochartGen.GenericId)
		}

		var autoChartProp enigma.GenericObjectProperties
		if err = jsonit.Unmarshal(rawAutoChartProperties, &autoChartProp); err != nil {
			return errors.Wrap(err, "Failed to unmarshal auto-chart properties to GenericObjectProperties")
		}

		// Get generated properties
		generatedPropPath := senseobjdef.NewDataPath(GeneratedPropertiesPath)
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
