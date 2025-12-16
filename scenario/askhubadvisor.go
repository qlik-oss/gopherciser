package scenario

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/goccy/go-json"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/atomichandlers"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/synced"
)

const hubAdvisorEndpoint = "api/v1/sentences"

type (
	language               string
	AdvisorQuerySourceEnum int
)

const (
	english language = "en"
)

const (
	// QueryString queries from string or file
	QueryList AdvisorQuerySourceEnum = iota
	// QueryFromFile queries read from file
	QueryFromFile
)

var advisorQuerySourceEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
	"querylist": int(QueryList),
	"file":      int(QueryFromFile),
})

func (value AdvisorQuerySourceEnum) GetEnumMap() *enummap.EnumMap {
	return advisorQuerySourceEnumMap
}

var askHubAdvisorDefaultThinktimeSettings = func() ThinkTimeSettings {
	settings := ThinkTimeSettings{
		DistributionSettings: helpers.DistributionSettings{
			Type:      helpers.UniformDistribution,
			Delay:     0,
			Mean:      float64(8),
			Deviation: 4,
		},
	}
	warnings, err := settings.Validate()
	if err != nil {
		panic(fmt.Errorf("askHubAdvisorDefaultThinktimeSettings has validation error: %s", err.Error()))
	}
	if len(warnings) != 0 {
		panic(fmt.Errorf("askHubAdvisorDefaultThinktimeSettings has validation warnings: %#v", warnings))
	}
	return settings
}()

// AskHubAdvisor-action settings
type (
	AskHubAdvisorSettingsCore struct {
		QuerySource       AdvisorQuerySourceEnum `json:"querysource" displayname:"Query source" doc-key:"askhubadvisor.querysource"`
		QueryList         []WeightedQuery        `json:"querylist" displayname:"List of queries" doc-key:"askhubadvisor.querylist"`
		Lang              language               `json:"lang" displayname:"Query language" doc-key:"askhubadvisor.lang"`
		FollowupDepth     uint                   `json:"maxfollowup" displayname:"Max depth of followup queries performed" doc-key:"askhubadvisor.maxfollowup"`
		FileName          string                 `json:"file" displayname:"File with one query per line" doc-key:"askhubadvisor.file" displayelement:"file"`
		App               string                 `json:"app" displayname:"App name (optional)" doc-key:"askhubadvisor.app"`
		SaveImages        bool                   `json:"saveimages" displayname:"Save images" doc-key:"askhubadvisor.saveimages"`
		SaveImageFile     synced.Template        `json:"saveimagefile" displayname:"Saved image filename" doc-key:"askhubadvisor.saveimagefile" displayelement:"savefile"`
		ThinkTimeSettings ThinkTimeSettings      `json:"thinktime,omitempty" displayname:"Think time settings" doc-key:"askhubadvisor.thinktime"`
		FollowupTypes     []followupType         `json:"followuptypes,omitempty" displayname:"Followup query types" doc-key:"askhubadvisor.followuptypes"`
	}

	WeightedQueryCore struct {
		Weight int    `json:"weight" displayname:"Weight used for randomly picking this query" doc-key:"askhubadvisor.querylist.weight"`
		Query  string `json:"query" displayname:"A query sentence" doc-key:"askhubadvisor.querylist.query"`
	}

	WeightedQuery struct {
		WeightedQueryCore
	}

	AskHubAdvisorSettings struct {
		AskHubAdvisorSettingsCore
	}

	// hubAdvisorSettings is a sub action executed for the first query and any followup queries
	hubAdvisorRequestSettings struct {
		host          string
		saveImages    bool
		saveImageFile synced.Template
		query         *hubAdvisorQuery
		response      *hubAdvisorResponse
	}
	followupQuery struct {
		typ   followupType
		query *hubAdvisorQuery
	}
)

// Messages used by hubAdvisorEndpont
type (
	hubAdvisorQuery struct {
		App                       *app            `json:"app,omitempty"`
		Text                      string          `json:"text"`
		Lang                      language        `json:"lang"`
		LimitedAccess             bool            `json:"limitedAccess"`
		GenerateNarrative         bool            `json:"generateNarrative"`
		EnableFollowups           bool            `json:"enableFollowups"`
		EnableConversationContext bool            `json:"enableConversationContext"`
		SelectedRecommendation    *recommendation `json:"selectedRecommendation"`
		ItemTokens                []interface{}   `json:"itemTokens"`
		ValueTokens               []interface{}   `json:"valueTokens"`
		EnableVisualizations      bool
		VisualizationTypes        []string
	}

	hubAdvisorResponse struct {
		ConversationalResponse conversationalResponse `json:"conversationalResponse"`
	}

	conversationalResponse struct {
		Apps                []app               `json:"apps"`
		Responses           []responseType      `json:"responses"`
		ConversationContext conversationContext `json:"conversationContext"`
	}

	responseType struct {
		Type             string `json:"type"`
		FollowupSentence string `json:"followupSentence"`
		ImageURL         string `json:"imageUrl"`
		TypedInfo
	}

	conversationContext struct {
		App *app `json:"app"`
	}

	recommendation struct {
		ID   string `json:"recId"`
		Name string `json:"name"`
	}

	TypedInfo struct {
		InfoType   string            `json:"infoType"`
		InfoValues []json.RawMessage `json:"infoValues"`
	}

	app struct {
		Name           string `json:"name"`
		ID             string `json:"id"`
		SpaceID        string `json:"space_id,omitempty"`
		SpaceName      string `json:"space_name,omitempty"`
		SpaceType      string `json:"space_type,omitempty"`
		LastReloadDate string `json:"last_reload_date,omitempty"`
		LimitedAccess  bool   `json:"limited_access"`
	}
)

const (
	infoTypeMeasures       = "measures"
	infoTypeDimensions     = "dimensions"
	infoTypeApps           = "apps"
	infoTypeRecomendations = "recommendations"
)

type followupType int

var followupTypeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
	"recommendation": int(followupRecommendation),
	"measure":        int(followupMeasure),
	"app":            int(followupApp),
	"dimension":      int(followupDimension),
	"sentence":       int(followupSentence),
})

const (
	followupApp followupType = iota
	followupRecommendation
	followupMeasure
	followupDimension
	followupSentence
)

// implements Enum interface compiler check
var _ Enum = followupType(0)
var _ Enum = (*followupType)(nil)

// GetEnumMap implements Enum interface
func (value followupType) GetEnumMap() *enummap.EnumMap {
	return followupTypeEnumMap
}

func (value *followupType) UnmarshalJSON(data []byte) error {
	ftInt, err := value.GetEnumMap().UnMarshal(data)

	if err != nil {
		return errors.Wrapf(err, "followup type has to be one of %v", value.GetEnumMap().Keys())
	}
	*value = followupType(ftInt)
	return nil
}

func (value followupType) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("unknown followup type<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

func (value followupType) String() string {
	return value.GetEnumMap().StringDefault(int(value), "UNDEFINED")
}

func (settings *AskHubAdvisorSettings) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, &settings.AskHubAdvisorSettingsCore)
	if err != nil {
		return err
	}
	settings.ThinkTimeSettings = thinkTimeWithFallback(
		settings.ThinkTimeSettings,
		askHubAdvisorDefaultThinktimeSettings,
	)
	switch settings.QuerySource {
	case QueryFromFile:
		if settings.FileName == "" {
			return errors.New("no file name")
		}
		if runtime.GOOS == "js" {
			return nil
		}
		file, err := os.Open(settings.FileName)
		if err != nil {
			return errors.Wrapf(err, "failed to open texts file<%s>", settings.FileName)
		}
		defer func() {
			if err := file.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "error closing file<%s>: %v\n", file.Name(), err)
			}
		}()
		wqs, err := ParseWeightedQueries(file)
		if err != nil {
			return errors.Wrapf(err, "failed parsing query-file<%s>", settings.FileName)
		}
		settings.QueryList = wqs
	case QueryList:
		return nil
	default:
		return errors.Errorf("Unknown AdvisorQuerySourceEnum<%d>", settings.QuerySource)

	}
	return nil
}

func (settings AskHubAdvisorSettings) MarshalJSON() ([]byte, error) {
	settings.ThinkTimeSettings = thinkTimeWithFallback(
		settings.ThinkTimeSettings,
		askHubAdvisorDefaultThinktimeSettings,
	)
	return json.Marshal(settings.AskHubAdvisorSettingsCore)
}

func (value *AdvisorQuerySourceEnum) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal AdvisorQuerySourceEnum")
	}

	*value = AdvisorQuerySourceEnum(i)
	return nil
}

func (value AdvisorQuerySourceEnum) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown AdvisorQuerySourceEnum<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

func (settings *AskHubAdvisorSettings) IsContainerAction() {}

// Implements ActionSettings
func (settings AskHubAdvisorSettings) Validate() ([]string, error) {
	switch settings.QuerySource {
	case QueryList:
		if len(settings.QueryList) == 0 {
			return nil, errors.Errorf(`no items in querylist`)
		}
	case QueryFromFile:
		if settings.FileName == "" {
			return nil, errors.Errorf(`no filename specified`)
		}
		if len(settings.QueryList) == 0 {
			return nil, errors.Errorf(`no items in file<%s>`, settings.FileName)
		}
	default:
		return nil, errors.Errorf(`unsupported querysource<%d>`, settings.QuerySource)
	}

	for _, q := range settings.QueryList {
		if q.Query == "" {
			return nil, errors.New("empty query")
		}
		if q.Weight <= 0 {
			return nil, errors.New("weight must be a positive number")
		}
	}

	if settings.FollowupTypes != nil && len(settings.FollowupTypes) == 0 {
		return nil, errors.New(
			"empty followuptypes implies no followups, please set maxfollowup to 0 for this behaviuor")
	}

	warnings := []string{}
	thinktimeWarnings, thinktimeErr := settings.ThinkTimeSettings.Validate()
	warnings = append(warnings, thinktimeWarnings...)
	if thinktimeErr != nil {
		return warnings, thinktimeErr
	}

	return warnings, nil
}

type HubAdvisorOption func(*hubAdvisorQuery)

func Language(lang language) HubAdvisorOption {
	if lang == "" {
		lang = english
	}
	return func(q *hubAdvisorQuery) {
		q.Lang = lang
	}
}

func App(app *app) HubAdvisorOption {
	return func(q *hubAdvisorQuery) {
		q.App = app
	}
}

func SelectedRecommendation(rec *recommendation) HubAdvisorOption {
	return func(q *hubAdvisorQuery) {
		q.SelectedRecommendation = rec
	}
}

func HubAdvisorQuery(text string, options ...HubAdvisorOption) *hubAdvisorQuery {
	q := &hubAdvisorQuery{
		Text:                      text,
		Lang:                      english,
		LimitedAccess:             false,
		GenerateNarrative:         true,
		EnableConversationContext: true,
		EnableFollowups:           true,
		ItemTokens:                []interface{}{},
		ValueTokens:               []interface{}{},
		EnableVisualizations:      true,
		VisualizationTypes:        []string{"barchart", "linechart", "piechart", "mekkochart", "qlik-funnel-chart-ext", "qlik-sankey-chart-ext", "boxplot", "histogram", "distributionplot", "sn-grid-chart"},
	}

	for _, applyOption := range options {
		applyOption(q)
	}

	return q
}

var savedImagesCount atomichandlers.AtomicCounter

type localData struct {
	ServerFileName string
	ImageCount     uint64
	Query          string
	AppName        string
	AppID          string
	Language       string
}

const defaultFileNameTemplateString = "{{.Local.Query}}--app-{{.Local.AppName}}--user-{{.UserName}}--thread-{{.Thread}}--session-{{.Session}}"

var defaultFileNameTemplate = func() *synced.Template {
	aTemplate, err := synced.New(defaultFileNameTemplateString)
	if err != nil {
		panic(err)
	}
	return aTemplate
}()

func imageNameFromTemplate(sessionState *session.State, aTemplate *synced.Template, templateName string, data *localData) (string, error) {
	imageName, err := sessionState.ReplaceSessionVariablesWithLocalData(aTemplate, &data)
	if err != nil {
		return "", errors.Wrapf(err, `failed to create filename image name using %s template<%s>`, templateName, aTemplate)
	}
	if imageName == "" {
		return "", errors.Errorf(`got an empty filename using %s template<%s>`, templateName, aTemplate)
	}
	return imageName, nil
}

func (settings *hubAdvisorRequestSettings) resolveImageName(sessionState *session.State, actionState *action.State, imageNameFromHeader string) string {
	app := app{}
	if settings.query.App != nil {
		app = *settings.query.App
	}
	aLocalData := localData{
		ImageCount:     savedImagesCount.Inc(),
		ServerFileName: imageNameFromHeader,
		Query:          settings.query.Text,
		AppName:        app.Name,
		AppID:          app.ID,
	}

	for _, f := range []func() string{
		func() string {
			if settings.saveImageFile.String() == "" {
				return ""
			}
			imageName, err := imageNameFromTemplate(sessionState, &settings.saveImageFile, "saveimagefile", &aLocalData)
			if err != nil {
				sessionState.LogEntry.Logf(logger.WarningLevel, "%v: image name is falling back on default template", err)
				return ""
			}
			return imageName
		},
		func() string {
			imageName, err := imageNameFromTemplate(sessionState, defaultFileNameTemplate, "default", &aLocalData)
			if err != nil {
				sessionState.LogEntry.Logf(logger.ErrorLevel, "%v: image name is falling back on default name", err)
				return ""
			}
			return imageName
		},
	} {
		if imageName := f(); imageName != "" {
			return imageName
		}
	}
	return fmt.Sprintf("hubadvisor-image-%d", aLocalData.ImageCount)
}

func (settings *hubAdvisorRequestSettings) resolveImagePath(sessionState *session.State, actionState *action.State, imageNameFromHeader string) string {
	imageName := settings.resolveImageName(sessionState, actionState, imageNameFromHeader)
	if path.Ext(imageName) == "" {
		imageName += ".png"
	}
	imageName = helpers.ToValidWindowsFileName(imageName)
	return path.Join(sessionState.OutputsDir, "hubadvisor-images", imageName)
}

func (settings *hubAdvisorRequestSettings) downloadAndSaveImage(sessionState *session.State, actionState *action.State, imageURL string) {
	if imageURL == "" {
		return
	}
	imageURLWithHost := fmt.Sprintf("%v/%v", settings.host, strings.TrimLeft(imageURL, "/"))
	imageNameFromHeader, image, err := downloadImage(sessionState, actionState, imageURLWithHost)
	if err != nil {
		sessionState.LogEntry.Logf(logger.WarningLevel, `failed to download image at "%s": %v`, imageURL, err)
		return
	}
	sessionState.LogEntry.LogDebugf(`successfully downloaded image<%s> at "%s"`, imageNameFromHeader, imageURLWithHost)
	if !settings.saveImages {
		return
	}
	imagePath := settings.resolveImagePath(sessionState, actionState, imageNameFromHeader)
	err = helpers.WriteToFile(imagePath, image)
	if err != nil {
		sessionState.LogEntry.Logf(logger.WarningLevel, "could not write chart image to file: %v", err)
		return
	}
	sessionState.LogEntry.LogDebugf(`succesfully wrote image to "%s"`, imagePath)
}

func (settings *hubAdvisorRequestSettings) downloadAndSaveImages(sessionState *session.State, actionState *action.State) {
	if settings.response == nil {
		return
	}
	for _, res := range settings.response.ConversationalResponse.Responses {
		settings.downloadAndSaveImage(sessionState, actionState, res.ImageURL)
	}

}

// Execute executes one HubAdvisorRequest (implements ActionSettings)
func (settings *hubAdvisorRequestSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	settings.response = askHubAdvisor(sessionState, actionState, settings.host, settings.query)
	settings.downloadAndSaveImages(sessionState, actionState)
	sessionState.Wait(actionState)
}

// Validate implements ActionSettings
func (settings *hubAdvisorRequestSettings) Validate() ([]string, error) {
	return nil, nil
}

func downloadImage(sessionState *session.State, actionState *action.State, url string) (imageName string, image []byte, err error) {
	imgReqOpts := session.DefaultReqOptions()
	imgReqOpts.FailOnError = false
	imgReq, err := sessionState.Rest.GetSync(url, actionState, sessionState.LogEntry, imgReqOpts)
	if err != nil {
		return "", nil, errors.Errorf(`failed get request to "%s": %v`, url, err)
	}
	if imgReq.ResponseStatusCode != http.StatusOK {
		return "", nil, errors.Errorf(`unexpected response status code<%d><%s>`,
			imgReq.ResponseStatusCode, imgReq.ResponseStatus)
	}
	imgResContentType := imgReq.ResponseHeaders.Get("content-type")
	const expectedImgResContentType = "image/png"
	if imgResContentType != expectedImgResContentType {
		return "", nil, errors.Errorf(`expected content type "%s" but got "%s"`, expectedImgResContentType, imgResContentType)
	}
	image = imgReq.ResponseBody
	if len(image) == 0 {
		return "", nil, errors.Errorf(`no image in response`)
	}
	fileName, err := helpers.GetFileNameFromHTTPHeader(imgReq.ResponseHeaders)
	if err != nil {
		fileName = ""
	}
	return fileName, image, nil
}

// askhubadvisor performs a hubAdvisorQuery and fetches any images related to the response
func askHubAdvisor(sessionState *session.State, actionState *action.State, host string, query *hubAdvisorQuery) *hubAdvisorResponse {
	rawReqContent, err := json.Marshal(query)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return nil
	}

	req := sessionState.Rest.PostWithHeadersAsync(fmt.Sprintf("%v/%v", host, hubAdvisorEndpoint), actionState, sessionState.LogEntry, rawReqContent,
		map[string]string{
			"x-qlik-client-capability": "static",
		},
		&session.ReqOptions{
			ExpectedStatusCode: []int{http.StatusOK, http.StatusCreated},
			ContentType:        "application/json",
			FailOnError:        true,
		})

	if sessionState.Wait(actionState) {
		return nil
	}

	response := &hubAdvisorResponse{}
	if err := json.Unmarshal(req.ResponseBody, response); err != nil {
		actionState.AddErrors(errors.Wrapf(err, "could not unmarshal response<`%s`> from endpoint<%s>", req.ResponseBody, hubAdvisorEndpoint))
		return nil
	}
	// debug logging
	if sessionState.LogEntry.ShouldLogDebug() {
		var appName string
		if query.App != nil && query.App.Name != "" {
			appName = query.App.Name
		} else {
			appName = "ALL"
		}
		sessionState.LogEntry.LogDebugf(`query<%s> in app<%s>`, query.Text, appName)
	}

	return response
}

var valueVariableRegex = regexp.MustCompile(`\$value\b`)

// substituteVariable replaces "$value" in src with replacement
func substituteVariable(src, replacement string) string {
	return valueVariableRegex.ReplaceAllLiteralString(src, replacement)
}

// containVariable tells if sentence contains variables to be substituted
func containVariable(sentence string) bool {
	return strings.Contains(sentence, "$")
}

func validateApp(app *app) error {
	if app == nil || app.ID == "" || app.Name == "" {
		return errors.New(`insufficient info about current app to ask followup sentence`)
	}
	return nil

}

var errorNoInfoValues error = errors.New("no info values")

// unmarshalRandomInfoValue from infoValues into v (initialized pointer), using randomizer
func unmarshalRandomInfoValue(randomizer helpers.Randomizer, infoValues []json.RawMessage, v interface{}) error {
	if len(infoValues) == 0 {
		return errorNoInfoValues
	}
	pickedInfoValueRaw := infoValues[randomizer.Rand(len(infoValues))]
	if err := json.Unmarshal(pickedInfoValueRaw, v); err != nil {
		return err
	}
	return nil
}

func createFollowupQuery(sessionState *session.State, actionState *action.State, res *responseType, convContext conversationContext, appToPick string, language language) *followupQuery {
	if res.FollowupSentence == "" {
		return nil
	}
	currentApp := convContext.App

	// if current app context exist and followup sentence contains no variables which need to be substituted
	if currentApp != nil && currentApp.ID != "" && currentApp.Name != "" && !containVariable(res.FollowupSentence) {
		return &followupQuery{
			typ:   followupSentence,
			query: HubAdvisorQuery(res.FollowupSentence, Language(language)),
		}
	}

	if res.Type != "info" {
		return nil
	}

	switch infoType := res.InfoType; infoType {
	case infoTypeApps:
		// if one or no app, return error an answer to query should already be computed
		if len(res.InfoValues) < 2 {
			actionState.AddErrors(errors.Errorf(
				`hub advisor reponse contain followup sentence and app info "%s", but %d apps`, res.FollowupSentence, len(res.InfoValues)))
			return nil
		}
		var pickedApp *app
		if appToPick == "" {
			// pick random app
			pickedApp = &app{}
			if err := unmarshalRandomInfoValue(sessionState.Randomizer(), res.InfoValues, pickedApp); err != nil {
				actionState.AddErrors(errors.WithStack(err))
				return nil
			}
		} else {
			// pick app from settings
			for _, infoValue := range res.InfoValues {
				app := &app{}
				if err := json.Unmarshal(infoValue, app); err != nil {
					actionState.AddErrors(errors.Wrapf(err, "could not unmarshal app in response (%s)", hubAdvisorEndpoint))
					return nil
				}
				if app.Name == appToPick {
					pickedApp = app
					break
				}
			}
			if pickedApp == nil {
				actionState.AddErrors(errors.Errorf("no app was picked: configured app<%s> was not present in response", appToPick))
				return nil
			}

		}

		return &followupQuery{
			typ:   followupApp,
			query: HubAdvisorQuery(res.FollowupSentence, Language(language), App(pickedApp)),
		}

	case infoTypeMeasures, infoTypeDimensions:
		if err := validateApp(currentApp); err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return nil
		}
		fq := &followupQuery{}
		if infoType == infoTypeMeasures {
			fq.typ = followupMeasure
		} else {
			fq.typ = followupDimension
		}
		var pickedInfoValue string
		if err := unmarshalRandomInfoValue(sessionState.Randomizer(), res.InfoValues, &pickedInfoValue); err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return nil
		}
		fq.query = HubAdvisorQuery(
			substituteVariable(res.FollowupSentence, pickedInfoValue),
			Language(language),
		)
		return fq

	case infoTypeRecomendations:
		if err := validateApp(currentApp); err != nil {
			actionState.AddErrors(err)
			return nil
		}
		pickedRecommendation := &recommendation{}
		if err := unmarshalRandomInfoValue(sessionState.Randomizer(), res.InfoValues, pickedRecommendation); err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return nil
		}
		if pickedRecommendation.ID == "" {
			actionState.AddErrors(errors.New("hub advisor followup recommendation has no id"))
			return nil
		}
		if pickedRecommendation.Name == "" {
			actionState.AddErrors(errors.New("hub advisor followup recommendation has no name"))
			return nil
		}
		return &followupQuery{
			typ: followupRecommendation,
			query: HubAdvisorQuery(
				substituteVariable(res.FollowupSentence, pickedRecommendation.Name),
				SelectedRecommendation(pickedRecommendation),
				Language(language),
				App(currentApp),
			),
		}

	default:
		sessionState.LogEntry.LogDebugf(`unsupported hubadvisor infoType<%s>`, res.InfoType)
		return nil
	}
}

// createFollowupQueries extract new queries from a hubAdvisorResponse
func createFollowupQueries(sessionState *session.State, actionState *action.State, hubAdvisorResponse *hubAdvisorResponse, appToPick string, language language) []*followupQuery {
	followupQueries := make([]*followupQuery, 0, len(hubAdvisorResponse.ConversationalResponse.Responses))
	for _, res := range hubAdvisorResponse.ConversationalResponse.Responses {
		result := &res
		q := createFollowupQuery(sessionState, actionState, result, hubAdvisorResponse.ConversationalResponse.ConversationContext, appToPick, language)
		if q != nil {
			followupQueries = append(followupQueries, q)
		}
	}
	return followupQueries
}

// AskHubAdvisorSettings implements ActionSettings
func (settings AskHubAdvisorSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	if label == "" {
		label = "hubadvisorquery"
	}

	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	reqHeaders := map[string]string{"x-qlik-client-capability": "static"}
	reqOptions := session.DefaultReqOptions()
	reqOptions.ExpectedStatusCode = []int{http.StatusOK, http.StatusCreated}

	var wg sync.WaitGroup

	query := HubAdvisorQuery("clear", Language(settings.Lang))
	payload, err := json.Marshal(query)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}
	wg.Add(1)
	sessionState.Rest.PostAsyncWithCallback(fmt.Sprintf("%v/%v", host, hubAdvisorEndpoint), actionState, sessionState.LogEntry, payload, reqHeaders, reqOptions, func(err error, req *session.RestRequest) {
		wg.Done()
	})

	query = HubAdvisorQuery("Start", Language(settings.Lang))
	payload, err = json.Marshal(query)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}
	wg.Add(1)
	sessionState.Rest.PostAsyncWithCallback(fmt.Sprintf("%v/%v", host, hubAdvisorEndpoint), actionState, sessionState.LogEntry, payload, reqHeaders, reqOptions, func(err error, req *session.RestRequest) {
		wg.Done()
	})

	query = HubAdvisorQuery("Questions", Language(settings.Lang))
	payload, err = json.Marshal(query)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}
	wg.Add(1)
	sessionState.Rest.PostAsyncWithCallback(fmt.Sprintf("%v/%v", host, hubAdvisorEndpoint), actionState, sessionState.LogEntry, payload, reqHeaders, reqOptions, func(err error, req *session.RestRequest) {
		wg.Done()
	})

	wg.Wait()

	// choose a random query from querysource
	randInt, err := sessionState.Randomizer().RandWeightedInt(Weights(settings.QueryList))
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "could not generate random int from weighted hubadvisor queries"))
		return
	}
	sentence := settings.QueryList[randInt].Query
	query = HubAdvisorQuery(sentence, Language(settings.Lang))

	settings.askHubAdvisorRec(sessionState, actionState, connection, query, host, label, 0)
}

// askHubAdvisorRec performs a hubAdvisorQuery and asks followup queries created
// using the response. This is done recursively until there is no followup
// queries in response or until configured recursion depth is reached.
func (settings AskHubAdvisorSettings) askHubAdvisorRec(sessionState *session.State, actionState *action.State,
	connection *connection.ConnectionSettings, query *hubAdvisorQuery, host string, label string, depth uint) {
	if query == nil || depth == settings.FollowupDepth+1 {
		return
	}

	var subLabel string
	if depth == 0 {
		subLabel = label
	} else {
		subLabel = fmt.Sprintf("%s-followup-%d", label, depth)
	}

	request := &hubAdvisorRequestSettings{
		host:          host,
		saveImageFile: settings.SaveImageFile,
		saveImages:    settings.SaveImages,
		query:         query,
		response:      nil,
	}

	aHubadvisorQueryAction := Action{
		ActionCore: ActionCore{
			Type:  "hubadvisorquery",
			Label: subLabel,
		},
		Settings: request,
	}

	if depth > 0 {
		preFollowupThinktime(sessionState, actionState, connection, &settings.ThinkTimeSettings, aHubadvisorQueryAction.Label)
	}

	if err := aHubadvisorQueryAction.Execute(sessionState, connection); err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	followupQueries := createFollowupQueries(sessionState, actionState, request.response, settings.App, settings.Lang)
	followupQueries = followupsOfType(settings.FollowupTypes, followupQueries)
	for _, fq := range followupQueries {
		sessionState.LogEntry.LogDebugf("has followup<%s> of type<%s>", fq.query.Text, fq.typ)
		settings.askHubAdvisorRec(sessionState, actionState, connection, fq.query, host, label, depth+1)
	}
}

func followupsOfType(types []followupType, followups []*followupQuery) []*followupQuery {
	if types == nil {
		return append(make([]*followupQuery, 0, len(followups)), followups...)
	}
	filtered := []*followupQuery{}
	for _, typ := range types {
		for _, f := range followups {
			if f.typ == typ {
				filtered = append(filtered, f)
			}
		}
	}
	return filtered
}

func preFollowupThinktime(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, thinkTimeSettings *ThinkTimeSettings, label string) {
	if thinkTimeSettings == nil {
		thinkTimeSettings = &askHubAdvisorDefaultThinktimeSettings
	}
	err := executeThinkTimeSubAction(sessionState, connection, fmt.Sprintf("thinktime before: %s", label), thinkTimeSettings)
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}
}

// UnmarshalJSON accepts plain string and WeightedQueryCore json object.
// Plain strings are assigned weight 1.
func (wq *WeightedQuery) UnmarshalJSON(bytes []byte) error {
	wq.Weight = 1
	if strErr := json.Unmarshal(bytes, &wq.Query); strErr == nil {
		return strErr
	}

	if objErr := json.Unmarshal(bytes, &wq.WeightedQueryCore); objErr != nil {
		return errors.New("failed to unmarshal hub-advisor-query as string or json object")
	}
	return nil
}

func Weights(wqs []WeightedQuery) []int {
	var weights []int
	for _, wq := range wqs {
		weights = append(weights, wq.Weight)

	}
	return weights
}

// WeightedQueryFromString creates a new WeightedQuery from string on format
// [WEIGHT;]QUERY, where weight is optional.
func WeightedQueryFromString(str string) (WeightedQuery, error) {
	weightQuerySeparator := ";"
	weightQuery := strings.SplitN(str, weightQuerySeparator, 2)
	wq := WeightedQuery{WeightedQueryCore{
		Weight: 1,
	}}
	if len(weightQuery) == 1 {
		wq.Query = strings.TrimSpace(weightQuery[0])
		return wq, nil
	}

	weight, err := strconv.Atoi(strings.TrimSpace(weightQuery[0]))
	if err != nil {
		return wq, errors.Wrapf(err, "WEIGHT must be an integer in hubadvisor query file syntax [WEIGHT;]QUERY")
	}
	wq.Weight = weight
	wq.Query = strings.TrimSpace(weightQuery[1])
	return wq, nil
}

func ParseWeightedQueries(reader io.Reader) ([]WeightedQuery, error) {
	wqs := []WeightedQuery{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		wq, err := WeightedQueryFromString(line)
		if err != nil {
			return wqs, err
		}
		wqs = append(wqs, wq)
	}
	if err := scanner.Err(); err != nil {
		return wqs, errors.Wrap(err, "scan failed")
	}
	return wqs, nil
}

func (AskHubAdvisorSettings) DefaultValuesForGUI() ActionSettings {
	newSettings := &AskHubAdvisorSettings{}
	newSettings.ThinkTimeSettings = askHubAdvisorDefaultThinktimeSettings
	return newSettings
}
