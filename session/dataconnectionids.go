package session

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/elasticstructs"
)

type (
	// DataConnectionIDs datafile connection IDs keyed with
	DataConnectionIDs struct {
		m          map[string]string
		updateLock sync.RWMutex
	}

	// DataConnectionIDNotFoundError return if no data connection ID was found in map
	DataConnectionIDNotFoundError string
)

func (err DataConnectionIDNotFoundError) Error() string {
	space := err
	if space == "" {
		space = "personal"
	}
	return fmt.Sprintf("data connection ID for space<%s> not found", space)
}

// GetDataConnectionID for space, use empty string for persona
func (idMap *DataConnectionIDs) GetDataConnectionID(space string) (string, bool) {
	idMap.updateLock.RLock()
	defer idMap.updateLock.RUnlock()
	id, found := idMap.m[space]
	return id, found
}

// FillDataConnectionIDs fill DataConnectionIDs with space/data connection ID  mapping
func (idMap *DataConnectionIDs) FillDataConnectionIDs(data []elasticstructs.DataFilesRespData) {
	idMap.updateLock.Lock()
	defer idMap.updateLock.Unlock()

	if idMap.m == nil {
		idMap.m = make(map[string]string, len(data))
	}

	for _, entry := range data {
		if entry.QName == "DataFiles" {
			idMap.m[entry.Space] = entry.QID
		}
	}
}

// FetchDataConnectionID fetch connection id for space, use empty space for personal space
func (state *State) FetchDataConnectionID(actionState *action.State, host, space string) (string, error) {
	endpoint := fmt.Sprintf("%s/api/v1/dc-dataconnections?alldatafiles=true&allspaces=true&personal=true&owner=default&extended=true", host)
	opts := DefaultReqOptions()
	opts.FailOnError = false

	dataConnectionID, found := state.DataConnectionIDs.GetDataConnectionID(space)

	if found && dataConnectionID != "" {
		return dataConnectionID, nil
	}

	var requestError error
	_, _ = state.Rest.GetSyncWithCallback(endpoint, actionState, state.LogEntry, opts, func(err error, req *RestRequest) {
		if err != nil {
			requestError = err
			return
		}
		var datafilesResp elasticstructs.DataFilesResp

		if err := jsonit.Unmarshal(req.ResponseBody, &datafilesResp); err != nil {
			requestError = errors.Wrap(err, "failed unmarshaling dataconnections data")
			return
		}

		state.DataConnectionIDs.FillDataConnectionIDs(datafilesResp.Data)
	})

	if requestError != nil {
		return "", errors.WithStack(requestError)
	}

	dataConnectionID, found = state.DataConnectionIDs.GetDataConnectionID(space)
	if !found || dataConnectionID == "" {
		return "", DataConnectionIDNotFoundError(space)
	}

	return dataConnectionID, nil
}

// FetchQixDataFiles for provided data connection ID
func (state *State) FetchQixDataFiles(actionState *action.State, host, connectionID string) ([]elasticstructs.QixDataFile, error) {
	req, err := state.Rest.GetSync(
		fmt.Sprintf("%s/api/v1/qix-datafiles?top=656536&connectionId=%s", host, connectionID), actionState,
		state.LogEntry, nil,
	)
	dataFiles := make([]elasticstructs.QixDataFile, 0)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := jsonit.Unmarshal(req.ResponseBody, &dataFiles); err != nil {
		return nil, errors.WithStack(err)
	}
	return dataFiles, nil
}
