package elasticstructs

import "time"

type OdagGetLinks []OdagGetLink

type OdagGetLink struct {
	ID                 string `json:"id"`
	CreatedDate        string `json:"createdDate"`
	ModifiedDate       string `json:"modifiedDate"`
	ModifiedByUserName string `json:"modifiedByUserName"`
	Owner              struct {
		ID            string      `json:"id"`
		UserID        string      `json:"userId"`
		UserDirectory string      `json:"userDirectory"`
		Name          string      `json:"name"`
		Privileges    interface{} `json:"privileges"`
	} `json:"owner"`
	Name        string `json:"name"`
	TemplateApp struct {
		ID          string `json:"id"`
		Name        string `json:"name"`
		AppID       string `json:"appId"`
		PublishTime string `json:"publishTime"`
		Published   bool   `json:"published"`
		Stream      struct {
			ID         string      `json:"id"`
			Name       string      `json:"name"`
			Privileges interface{} `json:"privileges"`
		} `json:"stream"`
		SavedInProductVersion string      `json:"savedInProductVersion"`
		MigrationHash         string      `json:"migrationHash"`
		AvailabilityStatus    int         `json:"availabilityStatus"`
		Privileges            interface{} `json:"privileges"`
	} `json:"templateApp"`
	TemplateAppOrigName string `json:"templateAppOrigName"`
	LoadScriptHash      int64  `json:"loadScriptHash"`
	RowEstExpr          string `json:"rowEstExpr"`
	Bindings            []struct {
		TemplateAppVarName string `json:"templateAppVarName"`
		SelectAppParamType string `json:"selectAppParamType"`
		SelectAppParamName string `json:"selectAppParamName"`
		SelectionStates    string `json:"selectionStates"`
		NumericOnly        bool   `json:"numericOnly"`
	} `json:"bindings"`
	Properties struct {
		RowEstRange []struct {
			Context   string `json:"context"`
			HighBound int    `json:"highBound"`
		} `json:"rowEstRange"`
		GenAppLimit []struct {
			Context string `json:"context"`
			Limit   int    `json:"limit"`
		} `json:"genAppLimit"`
		AppRetentionTime []struct {
			Context       string `json:"context"`
			RetentionTime string `json:"retentionTime"`
		} `json:"appRetentionTime"`
		MenuLabel []struct {
			Context string `json:"context"`
			Label   string `json:"label"`
		} `json:"menuLabel"`
		GenAppName []struct {
			UserContext  string   `json:"userContext"`
			FormatString string   `json:"formatString"`
			Params       []string `json:"params"`
		} `json:"genAppName"`
	} `json:"properties"`
	ModelGroups []interface{} `json:"modelGroups"`
	Privileges  []string      `json:"privileges"`
	Status      string        `json:"status"`
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
	ID                 string `json:"id"`
	CreatedDate        string `json:"createdDate"`
	ModifiedDate       string `json:"modifiedDate"`
	ModifiedByUserName string `json:"modifiedByUserName"`
	Owner              struct {
		ID            string      `json:"id"`
		UserID        string      `json:"userId"`
		UserDirectory string      `json:"userDirectory"`
		Name          string      `json:"name"`
		Privileges    interface{} `json:"privileges"`
	} `json:"owner"`
	CreatedByAnonymousUser string      `json:"createdByAnonymousUser"`
	Link                   interface{} `json:"link"`
	SelectionAppID         string      `json:"selectionAppId"`
	SelectionAppOrigName   string      `json:"selectionAppOrigName"`
	Sheetname              string      `json:"sheetname"`
	ClientContextHandle    string      `json:"clientContextHandle"`
	TargetSheet            string      `json:"targetSheet"`
	GeneratedAppOrigName   string      `json:"generatedAppOrigName"`
	GeneratedApp           struct {
		ID                    string      `json:"id"`
		Name                  string      `json:"name"`
		AppID                 string      `json:"appId"`
		PublishTime           string      `json:"publishTime"`
		Published             bool        `json:"published"`
		Stream                interface{} `json:"stream"`
		SavedInProductVersion string      `json:"savedInProductVersion"`
		MigrationHash         string      `json:"migrationHash"`
		AvailabilityStatus    int         `json:"availabilityStatus"`
		Privileges            interface{} `json:"privileges"`
	} `json:"generatedApp"`
	EngineGroup        interface{} `json:"engineGroup"`
	ParentRequestID    string      `json:"parentRequestId"`
	TimeToLive         int         `json:"timeToLive"`
	PurgeAfter         string      `json:"purgeAfter"`
	CurRowEstExpr      string      `json:"curRowEstExpr"`
	CurRowEstLowBound  int         `json:"curRowEstLowBound"`
	CurRowEstHighBound int         `json:"curRowEstHighBound"`
	ActualRowEst       int         `json:"actualRowEst"`
	BindingStateHash   int64       `json:"bindingStateHash"`
	SelectionState     []struct {
		SelectionAppParamType string `json:"selectionAppParamType"`
		SelectionAppParamName string `json:"selectionAppParamName"`
		Values                []struct {
			SelStatus string      `json:"selStatus"`
			StrValue  string      `json:"strValue"`
			NumValue  interface{} `json:"numValue"`
		} `json:"values"`
		SelectedSize int `json:"selectedSize"`
	} `json:"selectionState"`
	SelectionStateHash int64    `json:"selectionStateHash"`
	Privileges         []string `json:"privileges"`
	Kind               string   `json:"kind"`
	State              string   `json:"state"`
	ReloadCount        int      `json:"reloadCount"`
	LoadState          struct {
		Status     string `json:"status"`
		LoadHost   string `json:"loadHost"`
		StartedAt  string `json:"startedAt"`
		FinishedAt string `json:"finishedAt"`
		Messages   struct {
			Logfilepath  string   `json:"logfilepath"`
			Transient    []string `json:"transient"`
			ProgressData struct {
				PersistentProgressMessages []struct {
					QMessageCode       int      `json:"qMessageCode"`
					QMessageParameters []string `json:"qMessageParameters"`
				} `json:"persistentProgressMessages"`
				ErrorData []interface{} `json:"errorData"`
			} `json:"progressData"`
		} `json:"messages"`
	} `json:"loadState"`
	RetentionTime      int `json:"retentionTime"`
	BindSelectionState []struct {
		SelectionAppParamType string `json:"selectionAppParamType"`
		SelectionAppParamName string `json:"selectionAppParamName"`
		Values                []struct {
			SelStatus string      `json:"selStatus"`
			StrValue  string      `json:"strValue"`
			NumValue  interface{} `json:"numValue"`
		} `json:"values"`
	} `json:"bindSelectionState"`
}

type OdagPostRequestResponse struct {
	ID                 string `json:"id"`
	CreatedDate        string `json:"createdDate"`
	ModifiedDate       string `json:"modifiedDate"`
	ModifiedByUserName string `json:"modifiedByUserName"`
	Owner              struct {
		ID            string      `json:"id"`
		UserID        string      `json:"userId"`
		UserDirectory string      `json:"userDirectory"`
		Name          string      `json:"name"`
		Privileges    interface{} `json:"privileges"`
	} `json:"owner"`
	CreatedByAnonymousUser string      `json:"createdByAnonymousUser"`
	Link                   interface{} `json:"link"`
	SelectionAppID         string      `json:"selectionAppId"`
	SelectionAppOrigName   string      `json:"selectionAppOrigName"`
	Sheetname              string      `json:"sheetname"`
	ClientContextHandle    string      `json:"clientContextHandle"`
	TargetSheet            string      `json:"targetSheet"`
	GeneratedAppOrigName   string      `json:"generatedAppOrigName"`
	GeneratedApp           interface{} `json:"generatedApp"`
	EngineGroup            interface{} `json:"engineGroup"`
	ParentRequestID        string      `json:"parentRequestId"`
	Engine                 string      `json:"engine"`
	TimeToLive             int         `json:"timeToLive"`
	PurgeAfter             string      `json:"purgeAfter"`
	CurRowEstExpr          string      `json:"curRowEstExpr"`
	CurRowEstLowBound      int         `json:"curRowEstLowBound"`
	CurRowEstHighBound     int         `json:"curRowEstHighBound"`
	ActualRowEst           int         `json:"actualRowEst"`
	BindingStateHash       int64       `json:"bindingStateHash"`
	SelectionState         []struct {
		SelectionAppParamType string `json:"selectionAppParamType"`
		SelectionAppParamName string `json:"selectionAppParamName"`
		Values                []struct {
			SelStatus string      `json:"selStatus"`
			StrValue  string      `json:"strValue"`
			NumValue  interface{} `json:"numValue"`
		} `json:"values"`
		SelectedSize int `json:"selectedSize"`
	} `json:"selectionState"`
	SelectionStateHash int64    `json:"selectionStateHash"`
	Privileges         []string `json:"privileges"`
	Kind               string   `json:"kind"`
	State              string   `json:"state"`
	ReloadCount        int      `json:"reloadCount"`
	LoadState          struct {
		Status string `json:"status"`
	} `json:"loadState"`
	RetentionTime      int `json:"retentionTime"`
	BindSelectionState []struct {
		SelectionAppParamType string `json:"selectionAppParamType"`
		SelectionAppParamName string `json:"selectionAppParamName"`
		Values                []struct {
			SelStatus string      `json:"selStatus"`
			StrValue  string      `json:"strValue"`
			NumValue  interface{} `json:"numValue"`
		} `json:"values"`
	} `json:"bindSelectionState"`
}

type OdagRequestByLink struct {
	ID           string `json:"id"`
	CreatedDate  string `json:"createdDate"`
	ModifiedDate string `json:"modifiedDate"`
	Owner        struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Subject  string `json:"subject"`
		Tenantid string `json:"tenantid"`
	} `json:"owner"`
	Link             string `json:"link"`
	SelectionApp     string `json:"selectionApp"`
	SelectionAppName string `json:"selectionAppName"`
	Sheetname        string `json:"sheetname"`
	GeneratedApp     struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"generatedApp"`
	GeneratedAppName string `json:"generatedAppName"`
	ParentRequestID  string `json:"parentRequestId"`
	Kind             string `json:"kind"`
	State            string `json:"state"`
	TemplateApp      string `json:"templateApp"`
	TemplateAppName  string `json:"templateAppName"`
	LoadState        struct {
		LoadHost   string `json:"loadHost"`
		StartedAt  string `json:"startedAt"`
		FinishedAt string `json:"finishedAt"`
		Status     string `json:"status"`
	} `json:"loadState"`
	TimeToLive         int      `json:"timeToLive"`
	RetentionTime      int      `json:"retentionTime"`
	PurgeAfter         string   `json:"purgeAfter"`
	CurRowEstExpr      string   `json:"curRowEstExpr"`
	CurRowEstLowBound  int      `json:"curRowEstLowBound"`
	CurRowEstHighBound int      `json:"curRowEstHighBound"`
	ActualRowEst       int      `json:"actualRowEst"`
	TargetSheet        string   `json:"targetSheet"`
	BindingStateHash   int      `json:"bindingStateHash"`
	SelectionStateHash int      `json:"selectionStateHash"`
	Validation         []string `json:"validation"`
	ErrorMessage       string   `json:"errorMessage"`
}

type OdagRequestsByLink []OdagRequestByLink
