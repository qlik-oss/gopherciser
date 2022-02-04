package scenario

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	SmartSearchSettings struct {
		SmartSearchSettingsCore
	}

	SmartSearchSettingsCore struct {
		SearchTextSource   SearchTextSource `json:"searchtextsource" displayname:"Search Text Source" doc-key:"smartsearch.searchtextsource"`
		SearchTextList     []string         `json:"searchtextlist" displayname:"Search Text List" doc-key:"smartsearch.searchtextlist"`
		SearchTextFilePath string           `json:"searchtextfile" displayname:"Search Text File" doc-key:"smartsearch.searchtextfile"`
		PasteSearchText    bool             `json:"pastesearchtext" displayname:"Simulate Pasting Search Text" doc-key:"smartsearch.pastesearchtext"`
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
			logEntry.LogDebugf(`search%d result group%d matching term%s %s in %s %s: %s`, id, sgIdx+1,
				ternary(len(searchTermsMatched) != 1, "s", ""), quoteList(searchTermsMatched),
				item.ItemType, quote(item.Identifier), quoteList(matches))
		}
	}
	if !hasResult {
		logEntry.LogDebugf("search%d result: NONE", id)
	}
}

var spaceRegex = regexp.MustCompile(`\s+`)

func parseSearchTerms(s string) []string {
	searchTerms := []string{}
	sb := &strings.Builder{}
	quoted := false
	escaped := false
	addSearchTerm := func() {
		defer sb.Reset()
		searchTerm := sb.String()
		searchTerm = strings.TrimSpace(searchTerm)
		searchTerm = spaceRegex.ReplaceAllString(searchTerm, " ")
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

func doSmartSearchRPCs(sessionState *session.State, actionState *action.State, appDoc *enigma.Doc, id int, searchText string) {
	searchTerms := parseSearchTerms(searchText)
	if sessionState.LogEntry.ShouldLogDebug() {
		sessionState.LogEntry.LogDebugf(`search%d text %s becomes search terms: %s`, id, quote(searchText), quoteList(searchTerms))
	}

	sessionState.QueueRequest(func(ctx context.Context) error {
		searchSuggestionResult, err := appDoc.SearchSuggest(
			ctx,
			&enigma.SearchCombinationOptions{},
			searchTerms,
		)
		logSearchSuggestionResult(sessionState.LogEntry, id, searchSuggestionResult)
		return err
	}, actionState, true, "SearchSuggest call failed")

	sessionState.QueueRequest(func(ctx context.Context) error {
		searchResult, err := appDoc.SearchResults(
			ctx,
			&enigma.SearchCombinationOptions{
				Context:      "CurrentSelections",
				CharEncoding: "Utf16",
			},
			searchTerms,
			&enigma.SearchPage{
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
			},
		)
		logSearchResult(sessionState.LogEntry, id, searchResult)
		return err
	}, actionState, true, "SearchResults call failed")
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

	const typeOneCharDuration = 300 * time.Millisecond
	const minPostTypingDuration = 700 * time.Millisecond
	const maxPostTypingDuration = 1300 * time.Millisecond

	chunks := searchTextChunks{}
	currentStart := 0
	for currentEnd := 1; currentEnd < len(searchText); currentEnd++ {
		randInt, err := randomizer.RandWeightedInt([]int{1, 9})
		if err != nil {
			return nil, errors.Wrap(err, "Failed to randomize search text chunk")

		}
		shallSplit := randInt == 0
		postTypingDuration, err := randomizer.RandDuration(minPostTypingDuration, maxPostTypingDuration)
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

func (chunks searchTextChunks) simulate(ctx context.Context, onErrors func(err ...error)) <-chan string {
	textChan := make(chan string, len(chunks))
	go func() {
		panicErr := helpers.RecoverWithErrorFunc(
			func() {
				defer close(textChan)
				onContextDone := func() {
					onErrors(errors.Wrap(ctx.Err(), "smart search typing simulation stopped"))
				}
				currentText := ""
				for _, chunk := range chunks {
					currentText = currentText + chunk.Text
					select {
					case <-time.After(chunk.TypingDuration):
					case <-ctx.Done():
						onContextDone()
						return
					}
					textChan <- currentText
					select {
					case <-time.After(chunk.PostTypingThinkDuration):
					case <-ctx.Done():
						onContextDone()
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

// Execute implements ActionSettings interface
func (settings SmartSearchSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	uplink := sessionState.Connection.Sense()
	if uplink.CurrentApp == nil {
		actionState.AddErrors(errors.New("not connected to app"))
		return
	}
	doc := uplink.CurrentApp.Doc
	rand := sessionState.Randomizer()
	searchText := settings.SearchTextList[rand.Rand(len(settings.SearchTextList))]

	searchTextChunks, err := newSearchTextChunks(rand, searchText, settings.PasteSearchText)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	sessionState.LogEntry.LogDebugf("search text chunks: %+v", searchTextChunks)

	reset()

	cnt := 0
	for searchText := range searchTextChunks.simulate(context.Background(), actionState.AddErrors) {
		cnt++
		doSmartSearchRPCs(sessionState, actionState, doc, cnt, searchText)
	}

	sessionState.LogEntry.LogDebugf("all %d search RPCs done", cnt)

	sessionState.Wait(actionState)
}
