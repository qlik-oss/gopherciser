package scenario

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/structs"
)

type (
	ChangeStreamModeEnum int

	// ChangestreamSettings Changestream reads in apps from selected stream
	ChangestreamSettings struct {
		Mode   ChangeStreamModeEnum `json:"mode" displayname:"Mode" doc-key:"changestream.mode"`
		Stream string               `json:"stream" displayname:"Stream" doc-key:"changestream.stream"`
	}
)

const (
	ChangeStreamModeName ChangeStreamModeEnum = iota
	ChangestreamModeID
)

var (
	changeStreamModeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
		"name": int(ChangeStreamModeName),
		"id":   int(ChangestreamModeID),
	})
)

// GetEnumMap for change stream mode
func (mode ChangeStreamModeEnum) GetEnumMap() *enummap.EnumMap {
	return changeStreamModeEnumMap
}

// UnmarshalJSON unmarshal ChangeStreamModeEnum
func (mode *ChangeStreamModeEnum) UnmarshalJSON(arg []byte) error {
	i, err := changeStreamModeEnumMap.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal ChangeStreamModeEnum")
	}

	*mode = ChangeStreamModeEnum(i)
	return nil
}

// MarshalJSON marshal ChangeStreamModeEnum
func (mode ChangeStreamModeEnum) MarshalJSON() ([]byte, error) {
	str, err := changeStreamModeEnumMap.String(int(mode))
	if err != nil {
		return nil, errors.Errorf("unknown ChangeStreamModeEnum<%d>", mode)
	}

	return json.Marshal(str)
}

// String implements stringer interface
func (mode ChangeStreamModeEnum) String() string {
	return changeStreamModeEnumMap.StringDefault(int(mode), strconv.Itoa(int(mode)))
}

// Validate ChangestreamSettings action (Implements ActionSettings interface)
func (settings ChangestreamSettings) Validate() ([]string, error) {
	switch settings.Mode {
	case ChangeStreamModeName:
		if settings.Stream == "" {
			return nil, errors.Errorf("Changestream mode<%s> no stream name provided in stream field", settings.Mode)
		}
	case ChangestreamModeID:
		if settings.Stream == "" {
			return nil, errors.Errorf("Changestream mode<%s> no stream ID provided in stream field", settings.Mode)
		}
	default:
		return nil, errors.Errorf("Changestream mode<%s> not supported", settings.Mode)
	}
	return nil, nil
}

// Execute ChangestreamSettings action (Implements ActionSettings interface)
func (settings ChangestreamSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {

	host, err := connectionSettings.GetRestUrl()
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	val, ok := sessionState.GetCustomState(StreamsStateKey)
	if !ok {
		actionState.AddErrors(errors.New("no stream state found"))
		return
	}
	streamsState, ok := val.(*StreamsState)
	if !ok {
		actionState.AddErrors(errors.New("no stream state not of type<*StreamsState>"))
		return
	}

	sessionState.ArtifactMap.ClearArtifactMap()

	streamID := ""
	switch settings.Mode {
	case ChangeStreamModeName:
		for k, v := range *streamsState {
			if settings.Stream == k {
				streamID = v
				break
			}
		}
	case ChangestreamModeID:
		streamID = settings.Stream
	default:
		actionState.AddErrors(errors.Errorf("Changestream mode<%s> not supported", settings.Mode))
		return
	}

	xrfkey, err := sessionState.GetXrfKey(host)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	streamsUrl := fmt.Sprintf("%s/api/hub/v1/apps/stream/%s?xrfkey=%s", host, streamID, xrfkey)
	sessionState.Rest.GetAsyncWithCallback(streamsUrl, actionState, sessionState.LogEntry, nil, func(err error, req *session.RestRequest) {
		if err != nil {
			return
		}
		var stream structs.Stream
		if err = json.Unmarshal(req.ResponseBody, &stream); err != nil {
			actionState.AddErrors(err)
			return
		}

		if err := sessionState.ArtifactMap.FillAppsUsingStream(stream); err != nil {
			actionState.AddErrors(err)
			return
		}
	})

	// Same request again for api compliance
	sessionState.Rest.GetAsync(streamsUrl, actionState, sessionState.LogEntry, nil)

	sessionState.Wait(actionState)
}
