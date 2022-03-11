package scenario

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/globals/constant"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

const (
	// typeOneCharDuration is based on:
	//
	// typing speed: 250 characters per minute
	//
	// 60 * 1000 / 250 ~= 240 ms
	//
	typeOneCharDuration   = 240 * time.Millisecond
	typingHaltProbability = 0.10
	// minTypingHaltDuration must be over 500ms since the delay until search is
	// done after a typing halt  is 500ms in the Sense client.
	minTypingHaltDuration = 700 * time.Millisecond
	maxTypingHaltDuration = 1300 * time.Millisecond
)

var (
	// searchResultsDefaultSearchPage is always sent in SearchResults websocket messages.
	searchResultsDefaultSearchPage = &enigma.SearchPage{
		Offset: 0,
		Count:  5,
		GroupOptions: []*enigma.SearchGroupOptions{
			{
				GroupType: "DatasetType",
				Offset:    0,
				Count:     intPtr(-1),
			},
		},
		GroupItemOptions: []*enigma.SearchGroupItemOptions{{
			GroupItemType: "Field",
			Offset:        0,
			Count:         intPtr(5),
		}},
	}

	// searchResultsDefaultSearchCombinationOptions is always sent in SearchResults websocket messages.
	searchResultsDefaultSearchCombinationOptions = &enigma.SearchCombinationOptions{
		Context:      "CurrentSelections",
		CharEncoding: "Utf16",
	}

	// searchSuggestDefaultSearchCombinationOptions is always sent in searchSuggest websocket messages.
	searchSuggestDefaultSearchCombinationOptions = &enigma.SearchCombinationOptions{}

	// selectAssociationsDefaultSearchCombinationOptions is always sent in selectAssociations websocket messages.
	selectAssociationsDefaultSearchCombinationOptions = &enigma.SearchCombinationOptions{
		Context: "CurrentSelections",
	}
)

type (
	SmartSearchSettings struct {
		SmartSearchSettingsCore
	}

	SmartSearchSettingsCore struct {
		SearchTextSource   SearchTextSource  `json:"searchtextsource" displayname:"Search Text Source" doc-key:"smartsearch.searchtextsource"`
		SearchTextList     []string          `json:"searchtextlist" displayname:"Search Text List" doc-key:"smartsearch.searchtextlist"`
		SearchTextFilePath string            `json:"searchtextfile" displayname:"Search Text File" doc-key:"smartsearch.searchtextfile"`
		PasteSearchText    bool              `json:"pastesearchtext" displayname:"Simulate Pasting Search Text" doc-key:"smartsearch.pastesearchtext"`
		MakeSelection      bool              `json:"makeselection" displayname:"Make selection from search result" doc-key:"smartsearch.makeselection"`
		SelectionThinkTime ThinkTimeSettings `json:"selectionthinktime,omitempty" displayname:"Think time before selection" doc-key:"smartsearch.selectionthinktime"`
	}

	SearchTextSource int

	searchTextChunk struct {
		Text                    string
		TypingDuration          time.Duration
		PostTypingThinkDuration time.Duration
	}

	searchTextChunks []searchTextChunk
)

const (
	SearchTextSourceList SearchTextSource = iota
	SearchTextSourceFile
)

var searchTextSourceEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
	"searchtextlist": int(SearchTextSourceList),
	"searchtextfile": int(SearchTextSourceFile),
})

var smartSearchDefaultThinktimeSettings = func() ThinkTimeSettings {
	settings := ThinkTimeSettings{
		DistributionSettings: helpers.DistributionSettings{
			Type:  helpers.StaticDistribution,
			Delay: 1,
		},
	}
	warnings, err := settings.Validate()
	if err != nil {
		panic(fmt.Errorf("smartSearchThinktimeSettings has validation error: %s", err.Error()))
	}
	if len(warnings) != 0 {
		panic(fmt.Errorf("smartSearchThinktimeSettings has validation warnings: %#v", warnings))
	}
	return settings
}()

func (SearchTextSource) GetEnumMap() *enummap.EnumMap {
	return searchTextSourceEnumMap
}

func (value *SearchTextSource) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal SearchTextSource")
	}

	*value = SearchTextSource(i)
	return nil
}

func (value SearchTextSource) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown SearchTextSource<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

func (settings *SmartSearchSettings) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, &settings.SmartSearchSettingsCore)
	if err != nil {
		return err
	}
	settings.SmartSearchSettingsCore.SelectionThinkTime = thinkTimeWithFallback(
		settings.SmartSearchSettingsCore.SelectionThinkTime,
		smartSearchDefaultThinktimeSettings,
	)
	switch settings.SearchTextSource {
	case SearchTextSourceList:
	case SearchTextSourceFile:
		rowFile, err := helpers.NewRowFile(settings.SearchTextFilePath)
		if err != nil {
			return err
		}
		settings.SearchTextList = rowFile.Rows()
	default:
		return errors.Errorf("Unknown SearchTextSource<%d>", settings.SearchTextSource)
	}
	return nil
}
func (settings SmartSearchSettings) MarshalJSON() ([]byte, error) {
	settings.SelectionThinkTime = thinkTimeWithFallback(
		settings.SelectionThinkTime,
		smartSearchDefaultThinktimeSettings,
	)
	return json.Marshal(settings)
}

func (settings *SmartSearchSettings) IsContainerAction() {}

// Validate implements ActionSettings interface
func (settings SmartSearchSettings) Validate() ([]string, error) {
	var warnings []string
	if len(settings.SearchTextList) == 0 {
		return warnings, errors.New("no searchtext found")
	}
	for idx, searchtext := range settings.SearchTextList {
		searchTerms := parseSearchTerms(searchtext)
		if len(searchTerms) == 0 {
			return warnings, errors.Errorf(`no search terms found in searchtext%d<%s> `, idx+1, settings.SearchTextList[idx])
		}
	}
	thinktimeWarnings, thinktimeErr := settings.SelectionThinkTime.Validate()
	warnings = append(warnings, thinktimeWarnings...)
	if thinktimeErr != nil {
		return warnings, thinktimeErr
	}

	return warnings, nil
}

func intPtr(i int) *int {
	return &i
}

func quote(s string) string {
	return fmt.Sprintf(`%#v`, s)
}

func quoteList(strList []string) string {
	if len(strList) == 0 {
		return "NONE"
	}
	quoted := []string{}
	for _, v := range strList {
		quoted = append(quoted, quote(v))
	}
	return strings.Join(quoted, ",")
}

func ternary(cond bool, ifCondTrue string, ifCondFalse string) string {
	if cond {
		return ifCondTrue
	}
	return ifCondFalse

}

func logSearchSuggestionResult(logEntry *logger.LogEntry, id int, searchSuggestionResult *enigma.SearchSuggestionResult) {
	if !logEntry.ShouldLogDebug() {
		return
	}
	suggestions := []string{}
	for _, suggestion := range searchSuggestionResult.Suggestions {
		suggestions = append(suggestions, suggestion.Value)
	}
	logEntry.LogDebugf("search%d suggestions: %s", id, quoteList(suggestions))
}

func logSearchResult(logEntry *logger.LogEntry, id int, searchResult *enigma.SearchResult) {
	if !logEntry.ShouldLogDebug() {
		return
	}
	hasResult := false
	for sgIdx, sga := range searchResult.SearchGroupArray {
		for _, item := range sga.Items {
			searchTermsMatched := []string{}
			for _, stIndex := range item.SearchTermsMatched {
				searchTermsMatched = append(searchTermsMatched, searchResult.SearchTerms[stIndex])
			}
			matches := []string{}
			for _, match := range item.ItemMatches {
				matches = append(matches, match.Text)
			}
			hasResult = true
			logEntry.LogDebugf(`search%d result group[%d] matching term%s %s in %s %s: %s`, id, sgIdx,
				ternary(len(searchTermsMatched) != 1, "s", ""), quoteList(searchTermsMatched),
				item.ItemType, quote(item.Identifier), quoteList(matches))
		}
	}
	if !hasResult {
		logEntry.LogDebugf("search%d result: NONE", id)
	}
}

func parseSearchTerms(s string) []string {
	searchTerms := []string{}
	sb := &strings.Builder{}
	quoted := false
	escaped := false
	addSearchTerm := func() {
		defer sb.Reset()
		searchTerm := standardizeWhiteSpace(sb.String())
		if searchTerm != "" {
			searchTerms = append(searchTerms, searchTerm)
		}
	}
	for _, r := range s {
		if escaped {
			sb.WriteRune(r)
		} else if r == '\\' {
			escaped = true
			continue
		} else if r == '"' {
			quoted = !quoted
		} else if !quoted && unicode.IsSpace(r) {
			addSearchTerm()
		} else {
			sb.WriteRune(r)
		}
		escaped = false
	}
	addSearchTerm()
	return searchTerms
}

func standardizeWhiteSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func isAbortedError(err error) bool {
	if err == nil {
		return false
	}
	switch err := err.(type) {
	case enigma.Error:
		if err.Code() == constant.LocerrGenericAborted {
			return true
		}
	}
	return false
}

// doSmartSearchRPCs gets search suggestions and search result in parallel. Wait
// for queued requests with `sessionState.Wait(actionState)` before reading
// returned search result.
func doSmartSearchRPCs(sessionState *session.State, actionState *action.State, appDoc *enigma.Doc, id int, searchText string, doRetries, acceptAborted bool) **enigma.SearchResult {
	searchTerms := parseSearchTerms(searchText)
	if sessionState.LogEntry.ShouldLogDebug() {
		sessionState.LogEntry.LogDebugf(`search%d text %s becomes search terms: %s`, id, quote(searchText), quoteList(searchTerms))
	}

	contextDecorator := func(ctx context.Context) context.Context {
		if doRetries {
			return ctx
		}
		return enigmahandlers.ContextWithoutRetries(ctx)
	}

	sessionState.QueueRequest(func(ctx context.Context) error {
		searchSuggestionResult, err := appDoc.SearchSuggest(
			contextDecorator(ctx),
			searchSuggestDefaultSearchCombinationOptions,
			searchTerms,
		)
		if err != nil {
			if acceptAborted && isAbortedError(err) {
				return nil
			}
			return err
		}
		logSearchSuggestionResult(sessionState.LogEntry, id, searchSuggestionResult)
		return nil
	}, actionState, true, "SearchSuggest call failed")

	var searchResultRet *enigma.SearchResult
	sessionState.QueueRequest(func(ctx context.Context) error {
		searchResult, err := appDoc.SearchResults(
			contextDecorator(ctx),
			searchResultsDefaultSearchCombinationOptions,
			searchTerms,
			searchResultsDefaultSearchPage,
		)
		if err != nil {
			if acceptAborted && isAbortedError(err) {
				return nil
			}
			return err
		}
		searchResultRet = searchResult
		logSearchResult(sessionState.LogEntry, id, searchResult)
		return nil
	}, actionState, true, "SearchResults call failed")

	return &searchResultRet
}

func newSearchTextChunks(randomizer helpers.Randomizer, searchText string, pasteSearchText bool) (searchTextChunks, error) {
	if pasteSearchText {
		return searchTextChunks{
			searchTextChunk{
				Text:                    searchText,
				PostTypingThinkDuration: 0,
				TypingDuration:          0,
			},
		}, nil
	}

	chunks := searchTextChunks{}
	currentStart := 0
	for currentEnd := 1; currentEnd < len(searchText); currentEnd++ {
		randFloat := randomizer.Float64()
		shallSplit := randFloat < typingHaltProbability
		postTypingDuration, err := randomizer.RandDuration(minTypingHaltDuration, maxTypingHaltDuration)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to randomize search text chunk post typing duration")
		}
		if shallSplit {
			textChunk := searchText[currentStart:currentEnd]
			currentStart = currentEnd
			chunks = append(chunks, searchTextChunk{
				Text:                    textChunk,
				PostTypingThinkDuration: postTypingDuration,
				TypingDuration:          time.Duration(len(textChunk)) * typeOneCharDuration,
			})
		}
	}
	textChunk := searchText[currentStart:]
	chunks = append(chunks, searchTextChunk{
		Text:                    textChunk,
		PostTypingThinkDuration: 0,
		TypingDuration:          time.Duration(len(textChunk)) * typeOneCharDuration,
	})
	return chunks, nil

}

// simulate writes the full search string to the returned channel at the point
// in time the search string shall be sent to the server. The returned channel
// has a capacity equal to the number of search strings scheduled to be written
// to the same channel.
func (chunks searchTextChunks) simulate(ctx context.Context, onErrors func(err ...error)) <-chan string {
	textChan := make(chan string, len(chunks))
	go func() {
		panicErr := helpers.RecoverWithErrorFunc(
			func() {
				defer close(textChan)
				currentText := ""
				for _, chunk := range chunks {
					currentText = currentText + chunk.Text
					select {
					case <-time.After(chunk.TypingDuration):
					case <-ctx.Done():
						return
					}
					textChan <- currentText
					select {
					case <-time.After(chunk.PostTypingThinkDuration):
					case <-ctx.Done():
						return
					}
				}
			},
		)
		if panicErr != nil {
			onErrors(panicErr)
		}
	}()
	return textChan
}

func selectFromSearchResult(sessionState *session.State, actionState *action.State, searchResult *enigma.SearchResult) {
	app := sessionState.Connection.Sense().CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("not connected to app"))
		return
	}
	if len(searchResult.SearchGroupArray) == 0 {
		actionState.AddErrors(errors.New("can not select from empty search results"))
		return
	}
	terms := searchResult.SearchTerms
	searchGroupIdx := sessionState.Randomizer().Rand(len(searchResult.SearchGroupArray))
	err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		sessionState.LogEntry.LogDebugf("selecting search result group[%d]", searchGroupIdx)
		return app.Doc.SelectAssociations(ctx, selectAssociationsDefaultSearchCombinationOptions, terms, searchGroupIdx, false)
	})
	if err != nil {
		actionState.AddErrors(err)
		return
	}
	sessionState.Wait(actionState)
}

func smartSearch(sessionState *session.State, actionState *action.State, reset func(), searchTextList []string, pasteSearchText bool) *enigma.SearchResult {
	uplink := sessionState.Connection.Sense()
	if uplink.CurrentApp == nil {
		actionState.AddErrors(errors.New("not connected to app"))
		return nil
	}
	doc := uplink.CurrentApp.Doc
	rand := sessionState.Randomizer()
	searchText := searchTextList[rand.Rand(len(searchTextList))]

	searchTextChunks, err := newSearchTextChunks(rand, searchText, pasteSearchText)
	if err != nil {
		actionState.AddErrors(err)
		return nil
	}
	reset()
	defer sessionState.Wait(actionState)
	cnt := 0
	var searchResult **enigma.SearchResult
	searchTexts := searchTextChunks.simulate(sessionState.BaseContext(), actionState.AddErrors)
	totalCnt := cap(searchTexts)
	for searchText := range searchTexts {
		cnt++
		isLast := cnt == totalCnt
		searchResult = doSmartSearchRPCs(sessionState, actionState, doc, cnt, searchText, isLast, !isLast)
	}
	if !sessionState.Wait(actionState) {
		if searchResult == nil {
			actionState.AddErrors(errors.New("searchResult is unexpectedly nil"))
			return nil
		} else if *searchResult == nil {
			actionState.AddErrors(errors.New("*searchResult is unexpectedly nil: the last search was aborted"))
		}
	}
	return *searchResult
}

// Execute implements ActionSettings interface
func (settings SmartSearchSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {

	executeSubAction := executeSubActionFuncFactory(sessionState, connectionSettings, ActionSmartSearch, label)

	var searchResult *enigma.SearchResult

	searchSettings := executeFunc(func(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
		searchResult = smartSearch(sessionState, actionState, reset, settings.SearchTextList, settings.PasteSearchText)
	})

	selectSettings := executeFunc(func(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
		selectFromSearchResult(sessionState, actionState, searchResult)
	})

	err := executeSubAction("search", searchSettings)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to execute smart search subaction search"))
		return
	}

	if settings.MakeSelection {
		err := executeSubAction(ActionThinkTime, settings.SelectionThinkTime)
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "failed to execute smart search subaction pre selection thinktime"))
		}

		err = executeSubAction(ActionSelect, selectSettings)
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, "failed to execute smart search subaction select"))
			return
		}
	}

	sessionState.Wait(actionState)
}
