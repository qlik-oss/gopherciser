package scenario

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	SmartSearchSettings struct {
		SearchTerms []string
		SmartSearchSettingsCore
	}

	SmartSearchSettingsCore struct {
		SearchText string `json:"searchtext"`
	}
)

func (settings *SmartSearchSettings) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, &settings.SmartSearchSettingsCore)
	if err != nil {
		return err
	}
	settings.SearchTerms = parseSearchTerms(settings.SearchText)
	return nil
}

// Validate implements ActionSettings interface
func (settings SmartSearchSettings) Validate() ([]string, error) {
	var warnings []string
	if len(settings.SearchTerms) == 0 {
		return warnings, errors.New("no search terms")
	}
	return warnings, nil
}

func intPtr(i int) *int {
	return &i
}

func quote(s string) string {
	return fmt.Sprintf(`"%s"`, s)
}

func ternary(cond bool, ifCondTrue string, ifCondFalse string) string {
	if cond {
		return ifCondTrue
	}
	return ifCondFalse

}

func logSmartSearchResults(logEntry *logger.LogEntry, searchResult *enigma.SearchResult, searchSuggestionResult *enigma.SearchSuggestionResult) {
	if !logEntry.ShouldLogDebug() {
		return
	}
	suggestions := []string{}
	for _, suggestion := range searchSuggestionResult.Suggestions {
		suggestions = append(suggestions, quote(suggestion.Value))
	}
	if len(suggestions) > 0 {
		logEntry.LogDebugf("search suggestions: %s", strings.Join(suggestions, ","))
	}
	for _, sga := range searchResult.SearchGroupArray {
		searchTermsMatched := []string{}
		for _, stIndex := range sga.SearchTermsMatched {
			searchTermsMatched = append(searchTermsMatched, quote(searchResult.SearchTerms[stIndex]))
		}
		for _, item := range sga.Items {
			matches := []string{}
			for _, match := range item.ItemMatches {
				matches = append(matches, quote(match.Text))
			}
			logEntry.LogDebugf(`search result for term%s %s in %s %s: %s`,
				ternary(len(searchTermsMatched) != 1, "s", ""), strings.Join(searchTermsMatched, ","),
				item.ItemType, quote(item.Identifier), strings.Join(matches, ","))
		}
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

// Execute implements ActionSettings interface
func (settings SmartSearchSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	uplink := sessionState.Connection.Sense()
	if uplink.CurrentApp == nil {
		actionState.AddErrors(errors.New("not connected to app"))
		return
	}
	doc := uplink.CurrentApp.Doc
	searchTerms := settings.SearchTerms
	sessionState.LogEntry.LogDebugf(`search terms: "%s"`, strings.Join(searchTerms, `","`))

	var searchResult *enigma.SearchResult
	var searchSuggestionResult *enigma.SearchSuggestionResult
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			var err error
			searchSuggestionResult, err = doc.SearchSuggest(
				ctx,
				&enigma.SearchCombinationOptions{},
				searchTerms,
			)
			return err
		})
		if err != nil {
			actionState.AddErrors(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
			var err error
			searchResult, err = doc.SearchResults(
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
			return err
		})

		if err != nil {
			actionState.AddErrors(err)
		}
	}()

	wg.Wait()
	logSmartSearchResults(sessionState.LogEntry, searchResult, searchSuggestionResult)
	sessionState.Wait(actionState)
}
