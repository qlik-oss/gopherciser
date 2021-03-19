package scenario

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/atomichandlers"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
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

var askHubAdvisorDefaultThinktimeSettings = ThinkTimeSettings{
	DistributionSettings: helpers.DistributionSettings{
		Type:      helpers.UniformDistribution,
		Delay:     0,
		Mean:      float64(8),
		Deviation: 4,
	},
}

// AskHubAdvisor-action settings
type (
	AskHubAdvisorSettingsCore struct {
		QuerySource       AdvisorQuerySourceEnum `json:"querysource" displayname:"Query source" doc-key:"askhubadvisor.querysource"`
		QueryList         []WeightedQuery        `json:"querylist" displayname:"List of queries" doc-key:"askhubadvisor.querylist"`
		Lang              language               `json:"lang" displayname:"Query language" doc-key:"askhubadvisor.lang"`
		FollowupDepth     uint                   `json:"maxfollowup" displayname:"Max depth of followup queries performed" doc-key:"askhubadvisor.maxfollowup"`
		FileName          string                 `json:"file" displayname:"File with one query per line" doc-key:"askhubadvisor.file"`
		App               string                 `json:"app" displayname:"App name (optional)" doc-key:"askhubadvisor.app"`
		SaveImages        bool                   `json:"saveimages" displayname:"Save images" doc-key:"askhubadvisor.saveimages"`
		SaveImageFile     session.SyncedTemplate `json:"saveimagefile" displayname:"File name (without suffix)" doc-key:"askhubadvisor.saveimagefile"`
		ThinkTimeSettings *ThinkTimeSettings     `json:"thinktime,omitempty" displayname:"Think time settings" doc-key:"askhubadvisor.thinktime"`
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
		saveImageFile session.SyncedTemplate
		query         *hubAdvisorQuery
		response      *hubAdvisorResponse
	}
)

// Messages used by hubAdvisorEndpont
type (
	hubAdvisorQuery struct {
		App                       *app                `json:"app,omitempty"`
		Text                      string              `json:"text"`
		Lang                      language            `json:"lang"`
		LimitedAccess             bool                `json:"limitedAccess"`
		GenerateNarrative         bool                `json:"generateNarrative"`
		EnableFollowups           bool                `json:"enableFollowups"`
		EnableConversationContext bool                `json:"enableConversationContext"`
		ConversationContext       conversationContext `json:"conversationContext"`
		SelectedRecommendation    *recommendation     `json:"selectedRecommendation"`
		ItemTokens                []interface{}       `json:"itemTokens"`
		ValueTokens               []interface{}       `json:"valueTokens"`
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
		App             *app            `json:"app"`
		Entity          json.RawMessage `json:"entity,omitempty"`
		ParserResults   json.RawMessage `json:"parserResults,omitempty"`
		Recommendations json.RawMessage `json:"recommendations,omitempty"`
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
		URL            string `json:"url,omitempty"`
	}
)

func (settings *AskHubAdvisorSettings) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, &settings.AskHubAdvisorSettingsCore)
	if err != nil {
		return err
	}
	switch settings.QuerySource {
	case QueryFromFile:

		if settings.FileName == "" {
			return errors.New("no file name")
		}
		if runtime.GOOS == "js" {
			return errors.Errorf("can not read file with GOOS=%s", runtime.GOOS)
		}
		file, err := os.Open(settings.FileName)
		if err != nil {
			return errors.Wrapf(err, "failed to open texts file<%s>", settings.FileName)
		}
		defer file.Close()
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
func (settings AskHubAdvisorSettings) Validate() error {
	switch settings.QuerySource {
	case QueryList:
		if len(settings.QueryList) == 0 {
			return errors.Errorf(`no items in querylist`)
		}
	case QueryFromFile:
		if settings.FileName == "" {
			return errors.Errorf(`no filename specified`)
		}
		if len(settings.QueryList) == 0 {
			return errors.Errorf(`no items in file<%s>`, settings.FileName)
		}
	default:
		return errors.Errorf(`unsupported querysource<%d>`, settings.QuerySource)
	}

	for _, q := range settings.QueryList {
		if q.Query == "" {
			return errors.New("empty query")
		}
		if q.Weight < 0 {
			return errors.New("negative weight")
		}
	}

	return nil
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

func ConversationContext(convContext conversationContext) HubAdvisorOption {
	return func(q *hubAdvisorQuery) {
		q.EnableConversationContext = true
		q.ConversationContext = convContext
		if q.ConversationContext.App != nil {
			q.App = q.ConversationContext.App
			if q.ConversationContext.App.ID != "" {
				q.ConversationContext.App.URL = fmt.Sprintf("/sense/app/%s/insightadvisor", q.ConversationContext.App.ID)
			}
		}
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

var defaultFileNameTemplate = func() *session.SyncedTemplate {
	aTemplate, err := session.NewSyncedTemplate(defaultFileNameTemplateString)
	if err != nil {
		panic(err)
	}
	return aTemplate
}()

func imageNameFromTemplate(sessionState *session.State, aTemplate *session.SyncedTemplate, templateName string, data *localData) (string, error) {
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
	aLocalData := localData{
		ImageCount:     savedImagesCount.Inc(),
		ServerFileName: imageNameFromHeader,
		Query:          settings.query.Text,
		AppName:        settings.query.App.Name,
		AppID:          settings.query.App.ID,
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
func (settings *hubAdvisorRequestSettings) Validate() error {
	return nil
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

func followupQuery(sessionState *session.State, actionState *action.State, res *responseType, convContext conversationContext, appToPick string, language language) *hubAdvisorQuery {
	if res.FollowupSentence == "" {
		return nil
	}
	currentApp := convContext.App

	// if current app context exist and followup sentence contains no variables which need to be substituted
	if currentApp != nil && currentApp.ID != "" && currentApp.Name != "" && !containVariable(res.FollowupSentence) {
		return HubAdvisorQuery(res.FollowupSentence, Language(language), ConversationContext(convContext))
	}

	if res.Type != "info" {
		return nil
	}

	switch res.InfoType {
	case "apps":
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
		return HubAdvisorQuery(res.FollowupSentence, Language(language), App(pickedApp))

	case "measures", "dimensions":
		if err := validateApp(currentApp); err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return nil
		}
		var pickedInfoValue string
		if err := unmarshalRandomInfoValue(sessionState.Randomizer(), res.InfoValues, &pickedInfoValue); err != nil {
			actionState.AddErrors(errors.WithStack(err))
			return nil
		}
		return HubAdvisorQuery(substituteVariable(res.FollowupSentence, pickedInfoValue), Language(language), ConversationContext(convContext))

	case "recommendations":
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
		return HubAdvisorQuery(
			substituteVariable(res.FollowupSentence, pickedRecommendation.Name),
			SelectedRecommendation(pickedRecommendation),
			Language(language),
			ConversationContext(convContext),
		)

	default:
		sessionState.LogEntry.LogDebugf(`unsupported hubadvisor infoType<%s>`, res.InfoType)
		return nil
	}
}

// followupQueries extract new queries from a hubAdvisorResponse
func followupQueries(sessionState *session.State, actionState *action.State, hubAdvisorResponse *hubAdvisorResponse, appToPick string, language language) []*hubAdvisorQuery {
	followupQueries := make([]*hubAdvisorQuery, 0, len(hubAdvisorResponse.ConversationalResponse.Responses))
	for _, res := range hubAdvisorResponse.ConversationalResponse.Responses {
		result := &res
		q := followupQuery(sessionState, actionState, result, hubAdvisorResponse.ConversationalResponse.ConversationContext, appToPick, language)
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
	// choose a random query from querysource
	randInt, err := sessionState.Randomizer().RandWeightedInt(Weights(settings.QueryList))
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "could not generate random int from weighted hubadvisor queries"))
		return
	}
	sentence := settings.QueryList[randInt].Query
	query := HubAdvisorQuery(sentence, Language(settings.Lang))

	settings.askHubAdvisorRec(sessionState, actionState, connection, query, label, 0)
}

// askHubAdvisorReq perform hubAdvisorQuery and asks followup queries in
// response recursively until no followup queries in respinse or until
// configured recursion depth is reached
func (settings AskHubAdvisorSettings) askHubAdvisorRec(sessionState *session.State, actionState *action.State,
	connection *connection.ConnectionSettings, query *hubAdvisorQuery, label string, depth uint) {
	if query == nil || depth == settings.FollowupDepth+1 {
		return
	}

	host, err := connection.GetRestUrl()
	if err != nil {
		actionState.AddErrors(errors.WithStack(err))
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
		thinktime(sessionState, actionState, connection, settings.ThinkTimeSettings, aHubadvisorQueryAction.Label)
	}

	if err := aHubadvisorQueryAction.Execute(sessionState, connection); err != nil {
		actionState.AddErrors(errors.WithStack(err))
		return
	}

	for _, query := range followupQueries(sessionState, actionState, request.response, settings.App, settings.Lang) {
		settings.askHubAdvisorRec(sessionState, actionState, connection, query, label, depth+1)
	}
}

func thinktime(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, thinkTimeSettings *ThinkTimeSettings, label string) {
	if thinkTimeSettings == nil {
		thinkTimeSettings = &askHubAdvisorDefaultThinktimeSettings
	}
	aThinkTimeAction := Action{
		ActionCore: ActionCore{
			Type:  ActionThinkTime,
			Label: fmt.Sprintf("thinktime before: %s", label),
		},
		Settings: thinkTimeSettings,
	}
	if err := aThinkTimeAction.Execute(sessionState, connection); err != nil {
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
