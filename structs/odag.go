package structs

import "time"

type OdagGetLinks []OdagGetLink

type OdagGetLink struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type IOdagPostRequest struct {
	SelectionApp       string                          `json:"selectionApp"`
	ActualRowEst       int                             `json:"actualRowEst"`
	BindSelectionState []OdagPostRequestSelectionState `json:"bindSelectionState"`
	SelectionState     []OdagPostRequestSelectionState `json:"selectionState"`
}

type OdagPostRequest struct {
	IOdagPostRequest
	Sheetname           string `json:"sheetname"`
	ClientContextHandle string `json:"clientContextHandle"`
}

type OdagPostRequestSelectionState struct {
	SelectionAppParamType string                          `json:"selectionAppParamType"`
	SelectionAppParamName string                          `json:"selectionAppParamName"`
	Values                []OdagPostRequestSelectionValue `json:"values"`
	SelectedSize          int                             `json:"selectedSize"`
}

type OdagPostRequestSelectionValue struct {
	SelStatus string `json:"selStatus"`
	StrValue  string `json:"strValue"`
	NumValue  string `json:"numValue"`
}

type OdagLinkBinding struct {
	TemplateAppVarName string   `json:"templateAppVarName"`
	SelectAppParamType string   `json:"selectAppParamType"`
	SelectAppParamName string   `json:"selectAppParamName"`
	SelectionStates    string   `json:"selectionStates"`
	NumericOnly        bool     `json:"numericOnly"`
	Formatting         struct{} `json:"formatting"`
	Range              struct{} `json:"range"`
}

type OdagGetLinkInfo struct {
	Bindings  []OdagLinkBinding `json:"bindings"`
	ObjectDef struct {
		ID                  string            `json:"id"`
		CreatedDate         time.Time         `json:"createdDate"`
		ModifiedDate        time.Time         `json:"modifiedDate"`
		ModifiedByUserName  string            `json:"modifiedByUserName"`
		Owner               struct{}          `json:"owner"`
		Name                string            `json:"name"`
		TemplateApp         struct{}          `json:"templateApp"`
		TemplateAppOrigName string            `json:"templateAppOrigName"`
		LoadScriptHash      int64             `json:"loadScriptHash"`
		RowEstExpr          string            `json:"rowEstExpr"`
		Bindings            []OdagLinkBinding `json:"bindings"`
		Properties          struct{}          `json:"properties"`
		ModelGroups         []interface{}     `json:"modelGroups"`
		Privileges          []string          `json:"privileges"`
		Status              string            `json:"status"`
		GenAppAccessible    bool              `json:"genAppAccessible"`
	} `json:"objectDef"`
	Feedback []interface{} `json:"feedback"`
}

type OdagGetRequests []OdagGetRequest

type OdagGetRequest struct {
	ID        string `json:"id"`
	LoadState struct {
		Status string `json:"status"`
	} `json:"loadState"`
}

type OdagPostRequestResponse struct {
	ID string `json:"id"`
}

type OdagRequestByLink struct {
	ID           string `json:"id"`
	GeneratedApp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"generatedApp"`
}

type OdagRequestsByLink []OdagRequestByLink
