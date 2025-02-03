package session

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v4"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjdef"
)

type (
	NarrativesPropertiesNlgChartObject struct {
		ChartObjectId string                `json:"chartObjectId"`
		Label         string                `json:"label"`
		Type          string                `json:"type"`
		ExtendsId     string                `json:"qExtendsId"`
		Dimensions    []*enigma.NxDimension `json:"qDimensions"`
		Measures      []*enigma.NxMeasure   `json:"qMeasures"`
	}

	NarrativesProperties struct {
		*enigma.GenericObjectProperties
		NlgChartObject *NarrativesPropertiesNlgChartObject `json:"nlgChartObject"`
	}

	NarrativesHandler         struct{}
	NarrativesHandlerInstance struct {
		ID         string
		Properties *NarrativesProperties
	}

	NarrativesPayloadExpressionOverrides struct {
		Classifications []string                `json:"classifications"`
		Format          *enigma.FieldAttributes `json:"format,omitempty"`
	}

	NarrativesPayloadExpression struct {
		Expr      string                               `json:"expr"`
		Label     string                               `json:"label"`
		Overrides NarrativesPayloadExpressionOverrides `json:"overrides"`
	}

	NarrativesPayloadLibItem struct {
		LibId     string      `json:"libId"`
		Overrides interface{} `json:"overrides"`
	}

	NarrativesPayload struct {
		AlternateStateName string                        `json:"alternateStateName"`
		AnalysisTypes      []interface{}                 `json:"analysisTypes"`
		AppID              string                        `json:"appId"`
		Expressions        []NarrativesPayloadExpression `json:"expressions"`
		Fields             []interface{}                 `json:"fields"`
		Lang               string                        `json:"lang"`
		LibItems           []NarrativesPayloadLibItem    `json:"libItems"`
		Verbosity          string                        `json:"verbosity"`
	}
)

// Instance implements ObjectHandler  interface
func (handler *NarrativesHandler) Instance(id string) ObjectHandlerInstance {
	return &NarrativesHandlerInstance{ID: id}
}

// GetObjectDefinition implements ObjectHandlerInstance interface
func (handler *NarrativesHandlerInstance) GetObjectDefinition(objectType string) (string, senseobjdef.SelectType, senseobjdef.DataDefType, error) {
	if objectType != "sn-nlg-chart" {
		return "", senseobjdef.SelectTypeUnknown, senseobjdef.DataDefUnknown, errors.New("NarrativesHandlerInstance only handles objects of type sn-nlg-chart")
	}
	return (&DefaultHandlerInstance{}).GetObjectDefinition("sn-nlg-chart")
}

// SetObjectAndEvents implements ObjectHandlerInstance interface
func (handler *NarrativesHandlerInstance) SetObjectAndEvents(sessionState *State, actionState *action.State, obj *enigmahandlers.Object, genObj *enigma.GenericObject) {
	var wg sync.WaitGroup

	wg.Add(1)
	sessionState.QueueRequest(func(ctx context.Context) error {
		defer wg.Done()
		return handler.GetNarrativesProperties(sessionState, actionState, obj)
	}, actionState, true, "")

	wg.Add(1)
	sessionState.QueueRequest(func(ctx context.Context) error {
		defer wg.Done()
		return GetObjectLayout(sessionState, actionState, obj, nil)
	}, actionState, true, "")

	if sessionState.Rest == nil {
		sessionState.LogEntry.Log(logger.WarningLevel, "no resthandler defined, nl insights object will not generated correctly")
		return
	}

	wg.Wait()

	app := sessionState.CurrentApp
	if app == nil || app.ID == "" {
		actionState.AddErrors(errors.Errorf("no current app found"))
		return
	}

	payload := NarrativesPayload{
		AppID:         app.ID,
		Lang:          "en",
		Verbosity:     "full",
		Expressions:   []NarrativesPayloadExpression{},
		AnalysisTypes: []interface{}{},
		Fields:        []interface{}{},
		LibItems:      []NarrativesPayloadLibItem{},
	}

	measures := handler.Properties.HyperCubeDef.Measures
	if handler.Properties.NlgChartObject != nil {
		measures = handler.Properties.NlgChartObject.Measures
	}

	for _, measure := range measures {
		if measure.LibraryId != "" {
			payload.LibItems = append(payload.LibItems, NarrativesPayloadLibItem{
				LibId:     measure.LibraryId,
				Overrides: struct{}{},
			})
			continue
		}
		payload.Expressions = append(payload.Expressions, NarrativesPayloadExpression{
			Expr: measure.Def.Def,
			Overrides: NarrativesPayloadExpressionOverrides{
				Classifications: []string{"measure"},
				Format:          measure.Def.NumFormat,
			},
			Label: "",
		})
	}

	dimensions := handler.Properties.HyperCubeDef.Dimensions
	if handler.Properties.NlgChartObject != nil {
		dimensions = handler.Properties.NlgChartObject.Dimensions
	}

	for _, dimension := range dimensions {
		if dimension.LibraryId != "" {
			payload.LibItems = append(payload.LibItems, NarrativesPayloadLibItem{
				LibId:     dimension.LibraryId,
				Overrides: struct{}{},
			})
			continue
		}
		payload.Expressions = append(payload.Expressions, NarrativesPayloadExpression{
			Expr: dimension.Def.FieldDefs[0],
			Overrides: NarrativesPayloadExpressionOverrides{
				Classifications: []string{"dimension"},
			},
			Label: "",
		})
	}

	if obj.HyperCube().StateName != "" && obj.HyperCube().StateName != "$" {
		payload.AlternateStateName = obj.HyperCube().StateName
	}

	content, err := json.Marshal(payload)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to marshal narratives payload"))
		return
	}

	protocol := sessionState.Rest.Protocol()
	host := sessionState.Rest.Host()

	_, _ = sessionState.Rest.PostSync(fmt.Sprintf("%s%s/api/v1/narratives/actions/generate", protocol, host), actionState, sessionState.LogEntry, content, nil)

	event := func(ctx context.Context, as *action.State) error {
		if err := GetObjectLayout(sessionState, as, obj, nil); err != nil {
			return err
		}
		_, _ = sessionState.Rest.PostSync(fmt.Sprintf("%s%s/api/v1/narratives/actions/generate", protocol, host), actionState, sessionState.LogEntry, content, nil)
		return nil
	}

	sessionState.RegisterEvent(genObj.Handle, event, nil, true)
}

func (handler *NarrativesHandlerInstance) GetNarrativesProperties(sessionState *State, actionState *action.State, obj *enigmahandlers.Object) error {
	enigmaObject, ok := obj.EnigmaObject.(*enigma.GenericObject)
	if !ok {
		return errors.Errorf("Failed to cast object<%s> to *enigma.GenericObject", obj.ID)
	}

	//Get object properties
	getProperties := func(ctx context.Context) error {
		raw, err := enigmaObject.GetEffectivePropertiesRaw(ctx)
		if err != nil {
			return errors.Wrapf(err, "object<%s>.GetEffectiveProperties failed", obj.ID)
		}
		err = json.Unmarshal(raw, &handler.Properties)
		if err != nil {
			return errors.Wrapf(err, "object<%s>.GetEffectiveProperties unmarshal failed", obj.ID)
		}

		obj.SetProperties(handler.Properties.GenericObjectProperties)
		return nil
	}

	return sessionState.SendRequest(actionState, getProperties)
}
