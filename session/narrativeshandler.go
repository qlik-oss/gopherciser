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
	NarrativesHandler         struct{}
	NarrativesHandlerInstance struct {
		ID string
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

	NarrativesPayload struct {
		AlternateStateName string                        `json:"alternateStateName"`
		AnalysisTypes      []interface{}                 `json:"analysisTypes"`
		AppID              string                        `json:"appId"`
		Expressions        []NarrativesPayloadExpression `json:"expressions"`
		Fields             []interface{}                 `json:"fields"`
		Lang               string                        `json:"lang"`
		LibItems           []interface{}                 `json:"libItems"`
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
		return GetObjectProperties(sessionState, actionState, obj)
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

	// TODO check for library id's?
	expressions := make([]NarrativesPayloadExpression, len(obj.Properties().HyperCubeDef.Measures))
	for i, measure := range obj.Properties().HyperCubeDef.Measures {
		expressions[i] = NarrativesPayloadExpression{
			Expr: measure.Def.Def,
			Overrides: NarrativesPayloadExpressionOverrides{
				Classifications: []string{"measure"},
				Format:          measure.Def.NumFormat,
			},
			Label: "",
		}
	}

	for _, dimension := range obj.Properties().HyperCubeDef.Dimensions {
		expressions = append(expressions, NarrativesPayloadExpression{
			Expr: dimension.Def.FieldDefs[0],
			Overrides: NarrativesPayloadExpressionOverrides{
				Classifications: []string{"dimension"},
			},
			Label: "",
		})
	}

	payload := NarrativesPayload{
		AppID:         app.ID,
		Lang:          "en",
		Verbosity:     "full",
		Expressions:   expressions,
		AnalysisTypes: nil,
		Fields:        nil,
		LibItems:      nil,
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
