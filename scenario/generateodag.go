package scenario

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//GenerateOdagSettings settings for GenerateOdag
	GenerateOdagSettings struct {
		Name session.SyncedTemplate `json:"linkname" displayname:"ODAG link name" doc-key:"generateodag.linkname"`
	}
)

// odagWindowsEndpoint - Windows, odagElasticEndpoint - elastic
const odagPollingInterval = 3 * time.Second
const odagStatusPending = "pending"
const odagStatusLoading = "loading"
const odagStatusSuccess = "success"

type OdagEndpointConfiguration struct {
	Main            string
	Requests        string
	EnabledEndpoint string
}

var WindowsOdagEndpointConfiguration = OdagEndpointConfiguration{
	Main:            "api/odag/v1/links",
	Requests:        "api/odag/v1/requests",
	EnabledEndpoint: "api/odag/v1/isodagavailable",
}

var ElasticOdagEndpointConfiguration = OdagEndpointConfiguration{
	Main:            "api/v1/odaglinks",
	Requests:        "api/v1/odagrequests",
	EnabledEndpoint: "api/v1/odagisavailable",
}

// Validate GenerateOdagSettings action (Implements ActionSettings interface)
func (settings GenerateOdagSettings) Validate() error {
	if settings.Name.String() == "" {
		return errors.New("no ODAG link name specified")
	}
	return nil
}

// Execute ElasticDeleteCollectionSettings action (Implements ActionSettings interface)
func (settings GenerateOdagSettings) Execute(sessionState *session.State, actionState *action.State,
	connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	odagEndpoint := WindowsOdagEndpointConfiguration
	err := generateOdag(sessionState, settings, actionState, connectionSettings, odagEndpoint)
	if err != nil {
		actionState.AddErrors(err)
	}
}

func generateOdag(sessionState *session.State, settings GenerateOdagSettings, actionState *action.State, connectionSettings *connection.ConnectionSettings, odagEndpoint OdagEndpointConfiguration) error {
	odagLinkName, err := sessionState.ReplaceSessionVariables(&settings.Name)
	if err != nil {
		return err
	}
	host, err := connectionSettings.GetRestUrl()
	if err != nil {
		return err
	}

	// first, find the ID of the ODAG link we want
	odagLink, err := getOdagLinkByName(odagLinkName, host, sessionState, actionState, odagEndpoint.Main)
	if err != nil {
		return errors.Wrap(err, "failed to find ODAG link")
	}

	// then, get all the details about this ODAG link
	odagLinkBindings, err := GetOdagSelectionBindings(host, odagLink.ID, sessionState, actionState, odagEndpoint.Main)
	if err != nil {
		return errors.Wrap(err, "failed to obtain ODAG bindings via GET links/{:id}")
	}

	// start constructing the request to POST a new ODAG request
	connection := sessionState.Connection.Sense()
	currentSheet, err := GetCurrentSheet(connection)
	if err != nil {
		return errors.Wrap(err, "failed to get current sheet")
	}
	postObject := elasticstructs.OdagPostRequest{
		IOdagPostRequest: elasticstructs.IOdagPostRequest{
			SelectionApp: connection.CurrentApp.GUID,
		},
		Sheetname: currentSheet.ID,
	}

	return MakeOdagRequest(sessionState, actionState, odagLinkBindings, host, odagEndpoint, odagLink.ID, postObject.IOdagPostRequest, connection)
}

func MakeOdagRequest(sessionState *session.State, actionState *action.State, odagLinkBindings []elasticstructs.OdagLinkBinding, host string, odagEndpoint OdagEndpointConfiguration, odagLinkId string, postObject elasticstructs.IOdagPostRequest, connection *enigmahandlers.SenseUplink) error {
	var currentSelections *senseobjects.CurrentSelections
	var err error
	if currentSelections, err = connection.CurrentApp.GetCurrentSelections(sessionState, actionState); err != nil {
		return errors.WithStack(err)
	}

	// iterate through selections and gather data on what values all the fields with bindings currently have
	selections := currentSelections.Layout().SelectionObject.Selections
	postObject.BindSelectionState = []elasticstructs.OdagPostRequestSelectionState{}
	postObject.SelectionState = []elasticstructs.OdagPostRequestSelectionState{}
	for _, selection := range selections {
		// for each selection, create an object specifying that selection
		selectionState := elasticstructs.OdagPostRequestSelectionState{
			SelectionAppParamName: selection.Field,
			SelectionAppParamType: "Field",
			Values:                []elasticstructs.OdagPostRequestSelectionValue{},
		}
		selectionValue := elasticstructs.OdagPostRequestSelectionValue{
			SelStatus: "S", // S - value explicitly selected
		}
		if selection.IsNum {
			for _, selectedNumValue := range selection.SelectedFieldSelectionInfo {
				// selection is a number
				_, err := strconv.Atoi(selectedNumValue.Name)
				if err != nil {
					return errors.Wrap(err, "IsNum is true but failed to parse as int")
				}
				selectionValue.StrValue = selectedNumValue.Name
				selectionValue.NumValue = selectedNumValue.Name
				selectionState.Values = append(selectionState.Values, selectionValue)
			}
		} else {
			// selection is a string
			for _, selectedStrValue := range selection.SelectedFieldSelectionInfo {
				selectionValue.NumValue = "NaN"
				selectionValue.StrValue = selectedStrValue.Name
				selectionState.Values = append(selectionState.Values, selectionValue)
			}
		}

		selectionState.SelectedSize = len(selectionState.Values)
		postObject.SelectionState = append(postObject.SelectionState, selectionState)
	}

	// for all bindings, figure out the current state and add to the ODAG request
	for _, binding := range odagLinkBindings {
		bindSelectionState, err := getSelectionStateFromBinding(binding, sessionState, connection, actionState)
		if err != nil {
			return err
		}
		postObject.BindSelectionState = append(postObject.BindSelectionState, *bindSelectionState)
	}

	// time to send the final request
	postObjJson, err := jsonit.Marshal(postObject)
	if err != nil {
		actionState.AddErrors(err)
		//return
	}
	postRequest := session.RestRequest{
		Method:      session.POST,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%s/%s/%s/requests", host, odagEndpoint.Main, odagLinkId),
		Content:     postObjJson,
	}
	sessionState.Rest.QueueRequest(actionState, true, &postRequest, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return errors.New("failed during wait for POST ODAG request")
	}
	if postRequest.ResponseStatusCode != http.StatusCreated && postRequest.ResponseStatusCode != http.StatusOK {
		return errors.Errorf("failed to POST ODAG request: unexpected response code %d <%s>", postRequest.ResponseStatusCode, postRequest.ResponseBody)
	}
	var odagPostResponse elasticstructs.OdagPostRequestResponse
	if err := jsonit.Unmarshal(postRequest.ResponseBody, &odagPostResponse); err != nil {
		actionState.AddErrors(errors.Wrap(err, fmt.Sprintf("failed unmarshaling ODAG request POST response: %s", postRequest.ResponseBody)))
	}

	// and now we wait
	status := odagStatusPending
	for status == odagStatusPending || status == odagStatusLoading {
		helpers.WaitFor(sessionState.BaseContext(), time.Duration(odagPollingInterval))
		if sessionState.IsAbortTriggered() {
			return nil
		}

		odagRequests := session.RestRequest{
			Method:      session.GET,
			ContentType: "application/json",
			Destination: fmt.Sprintf("%s/%s/%s/requests?pending=true", host, odagEndpoint.Main, odagLinkId),
		}
		sessionState.Rest.QueueRequest(actionState, true, &odagRequests, sessionState.LogEntry)
		if sessionState.Wait(actionState) {
			return errors.New("Failed to execute REST request")
		}
		if odagRequests.ResponseStatusCode != http.StatusOK {
			return errors.Errorf("failed to get ODAG requests: %s", odagRequests.ResponseBody)
		}
		var odagGetRequests elasticstructs.OdagGetRequests
		if err := jsonit.Unmarshal(odagRequests.ResponseBody, &odagGetRequests); err != nil {
			return errors.Wrapf(err, "failed unmarshaling ODAG requests GET reponse: %s", odagRequests.ResponseBody)
		}

		var myRequestStatus elasticstructs.OdagGetRequest
		for _, odagRequest := range odagGetRequests {
			if odagRequest.ID == odagPostResponse.ID {
				myRequestStatus = odagRequest
			}
		}
		if myRequestStatus.ID == "" {
			return errors.Errorf("ODAG request with ID <%s> not in status list", odagPostResponse.ID)
		}
		status = myRequestStatus.LoadState.Status
	}
	if status != odagStatusSuccess {
		return errors.Errorf("ODAG generation finished with unexpected status <%s>", status)
	}
	return nil
}

// getSelectionStateFromBinding this function gets the selection state from the specified binding field
func getSelectionStateFromBinding(binding elasticstructs.OdagLinkBinding, sessionState *session.State,
	uplink *enigmahandlers.SenseUplink, actionState *action.State) (*elasticstructs.OdagPostRequestSelectionState, error) {
	bindSelectionState := elasticstructs.OdagPostRequestSelectionState{
		SelectionAppParamName: binding.SelectAppParamName,
		SelectionAppParamType: binding.SelectAppParamType,
		Values:                []elasticstructs.OdagPostRequestSelectionValue{},
	}

	// create a listbox session object for the binding field
	obj, err := createFieldListboxAsync(sessionState, actionState, uplink.CurrentApp.Doc, binding.SelectAppParamName)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating session listbox object for <%s>", binding.SelectAppParamName)
	}

	// get the data we need and move into our struct
	var dataPages []*enigma.NxDataPage
	err = sessionState.SendRequest(actionState, func(ctx context.Context) error {
		var err error
		dataPages, err = obj.GetListObjectData(ctx)
		return err
	})
	if err != nil {
		actionState.AddErrors(errors.Wrapf(err, "failed to get listobject data for field <%s>", binding.SelectAppParamName))
	}
	for _, dataPage := range dataPages {
		for _, listObjectValue := range dataPage.Matrix {
			if len(listObjectValue) < 1 {
				return nil, errors.Errorf("len(listObjectValue) < 1")
			}

			// Skip values that are not selected (green) or optional (white) depending on the binding configuration
			state := listObjectValue[0].State
			switch binding.SelectionStates {
			case "S":
				if state != "S" {
					continue
				}
			case "O":
				if state != "O" {
					continue
				}
			case "SO":
				if state != "S" && state != "O" {
					continue
				}
			default:
				return nil, errors.Errorf("unknown SelectionStates: <%s>", binding.SelectionStates)
			}

			value := elasticstructs.OdagPostRequestSelectionValue{}
			value.StrValue = listObjectValue[0].Text
			if math.IsNaN(float64(listObjectValue[0].Num)) {
				value.NumValue = "NaN"
			} else {
				value.NumValue = fmt.Sprintf("%f", listObjectValue[0].Num)
			}
			value.SelStatus = listObjectValue[0].State
			bindSelectionState.Values = append(bindSelectionState.Values, value)
		}
	}

	// delete session objects after they're no longer needed
	// we need to wait some time before doing so, or the deletion will be unsuccessful
	idToDelete := obj.Layout().Info.Id
	destroySessionObj := func() {
		sessionState.QueueRequest(func(ctx context.Context) error {
			success, err := uplink.CurrentApp.Doc.DestroyObject(ctx, idToDelete)
			if !success {
				sessionState.LogEntry.Logf(logger.WarningLevel, "unsuccessful destruction of <%s> (session object for field <%s>)", idToDelete, binding.SelectAppParamName)
			}
			return err
		}, actionState, false, fmt.Sprintf("failed to destroy session object for field <%s>", binding.SelectAppParamName))
	}
	go func() {
		select {
		case <-sessionState.BaseContext().Done():
			return
		case <-time.After(10 * time.Second):
			destroySessionObj()
		}
	}()

	return &bindSelectionState, nil
}

// getOdagLinkByName returns the ODAG link by the specified name
func getOdagLinkByName(name string, host string, sessionState *session.State,
	actionState *action.State, odagEndpoint string) (*elasticstructs.OdagGetLink, error) {
	odagLinks := session.RestRequest{
		Method:      session.GET,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%s/%s", host, odagEndpoint),
	}
	sessionState.Rest.QueueRequest(actionState, true, &odagLinks, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return nil, errors.New("Failed to execute REST request")
	}
	if odagLinks.ResponseStatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("failed to get ODAG links: %s", odagLinks.ResponseBody))
	}
	var odagGetLinksResponse elasticstructs.OdagGetLinks
	if err := jsonit.Unmarshal(odagLinks.ResponseBody, &odagGetLinksResponse); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed unmarshaling ODAG links GET reponse: %s", odagLinks.ResponseBody))
	}
	odagLink := elasticstructs.OdagGetLink{}
	for _, maybeOdagLink := range odagGetLinksResponse {
		if maybeOdagLink.Name == name {
			odagLink = maybeOdagLink
			break
		}
	}
	if odagLink.ID == "" {
		return nil, errors.Errorf("found no such ODAG link <%s>", name)
	}
	return &odagLink, nil
}

// GetOdagSelectionBindings gets information about the ODAG link, including bindings
func GetOdagSelectionBindings(host string, odagLinkId string, sessionState *session.State,
	actionState *action.State, odagEndpoint string) ([]elasticstructs.OdagLinkBinding, error) {
	odagLinkInfo := session.RestRequest{
		Method:      session.GET,
		ContentType: "application/json",
		Destination: fmt.Sprintf("%s/%s/%s", host, odagEndpoint, odagLinkId),
	}
	sessionState.Rest.QueueRequest(actionState, true, &odagLinkInfo, sessionState.LogEntry)
	if sessionState.Wait(actionState) {
		return nil, errors.New("Failed to execute REST request")
	}
	if odagLinkInfo.ResponseStatusCode != http.StatusOK {
		return nil, errors.New(fmt.Sprintf("failed to get ODAG link info: %s", odagLinkInfo.ResponseBody))
	}
	var odagLinkInfoStruct elasticstructs.OdagGetLinkInfo
	if err := jsonit.Unmarshal(odagLinkInfo.ResponseBody, &odagLinkInfoStruct); err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed unmarshaling ODAG links info GET reponse: %s", odagLinkInfo.ResponseBody))
	}
	if odagLinkInfoStruct.Bindings != nil {
		return odagLinkInfoStruct.Bindings, nil // this is expected for elastic
	} else if odagLinkInfoStruct.ObjectDef.Bindings != nil {
		return odagLinkInfoStruct.ObjectDef.Bindings, nil // this is expected for Windows
	}
	return nil, errors.New("failed to find any bindings for the ODAG link") // no bueno
}

// createFieldListboxAsync creates a listbox session object for specified field
func createFieldListboxAsync(sessionState *session.State, actionState *action.State, doc *enigma.Doc, field string) (*senseobjects.ListBox, error) {
	var obj *senseobjects.ListBox
	err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		var err error
		obj, err = senseobjects.CreateListBoxObject(ctx, doc, field)
		return err
	})
	if err != nil {
		return nil, err
	}
	err = sessionState.SendRequest(actionState, func(ctx context.Context) error {
		return obj.UpdateLayout(ctx)
	})
	if err != nil {
		return nil, err
	}
	err = sessionState.SendRequest(actionState, func(ctx context.Context) error {
		return obj.UpdateProperties(ctx)
	})
	if err != nil {
		return nil, err
	}
	return obj, nil
}
