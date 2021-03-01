package session

import (
	"fmt"
	"net/url"
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
	space := string(err)
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
func (idMap *DataConnectionIDs) FillDataConnectionIDs(data ...elasticstructs.DataFilesRespData) {
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
func (state *State) FetchDataConnectionID(actionState *action.State, host, spaceID string) (string, error) {
	if dataConnectionID, found := state.DataConnectionIDs.GetDataConnectionID(spaceID); found && dataConnectionID != "" {
		return dataConnectionID, nil
	}
	requestURL, err := url.Parse(host)
	if err != nil {
		return "", errors.Wrapf(err, "faulty url<%s>", host)
	}
	requestURL.Path = "/api/v1/dc-dataconnections/DataFiles"
	query := requestURL.Query()
	if spaceID != "" {
		query.Set("space", spaceID)
	}
	query.Set("type", "connectionname")
	requestURL.RawQuery = query.Encode()
	requestURLString := requestURL.String()

	opts := DefaultReqOptions()
	opts.FailOnError = false
	req, err := state.Rest.GetSync(requestURLString, actionState, state.LogEntry, opts)
	if err != nil {
		return "", errors.WithStack(err)
	}

	var datafilesRespData elasticstructs.DataFilesRespData
	if err := jsonit.Unmarshal(req.ResponseBody, &datafilesRespData); err != nil {
		return "", errors.Wrap(err, "failed unmarshaling dataconnections data")
	}

	if datafilesRespData.ID == "" {
		return "", errors.Errorf(`data connection id from "%s" is empty`, requestURL.Path)
	}

	state.DataConnectionIDs.FillDataConnectionIDs(datafilesRespData)

	dataConnectionID, found := state.DataConnectionIDs.GetDataConnectionID(spaceID)
	if !found || dataConnectionID == "" {
		return "", DataConnectionIDNotFoundError(spaceID)
	}

	return dataConnectionID, nil
}

// FetchQixDataFiles for provided data connection ID
func (state *State) FetchQixDataFiles(actionState *action.State, host, connectionID string) ([]elasticstructs.QixDataFile, error) {
	req, err := state.Rest.GetSync(
		fmt.Sprintf("%s/api/v1/qix-datafiles?top=10000&connectionId=%s", host, connectionID), actionState,
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

// FetchQixDataFile with a specific name, for a provided data connection ID.
// Returns:
//   - nil and no error if file does not exist
//   - error if more than one datafile with the same name exist
//   - error if filename of returned data file is incorrect
func (state *State) FetchQixDataFile(actionState *action.State, host, connectionID string, fileName string) (*elasticstructs.QixDataFile, error) {
	requestURL := fmt.Sprintf("%s/api/v1/qix-datafiles?connectionId=%s&name=%s", host, connectionID, fileName)
	req, err := state.Rest.GetSync(requestURL, actionState, state.LogEntry, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	dataFiles := make([]*elasticstructs.QixDataFile, 0)
	if err := jsonit.Unmarshal(req.ResponseBody, &dataFiles); err != nil {
		return nil, errors.WithStack(err)
	}
	switch nbrDataFiles := len(dataFiles); nbrDataFiles {
	case 0:
		return nil, nil
	case 1:
		dataFile := dataFiles[0]
		if dataFile.Name != fileName {
			return nil, errors.Errorf("requested file with name<%s> but got file with name<%s>", fileName, dataFile.Name)
		}
		return dataFile, nil
	default:
		return nil, errors.Errorf("recieved %d datafiles when requesting<%s> 1 datafile", nbrDataFiles, requestURL)
	}
}
