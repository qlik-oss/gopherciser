package scenario

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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
		searchTerms []string
		SmartSearchSettingsCore
	}

	SmartSearchSettingsCore struct {
		SearchText string `json:"searchtext" displayname:"Search Text" doc-key:"smartsearch.searchtext"`
	}
)

func (settings *SmartSearchSettings) UnmarshalJSON(bytes []byte) error {
	err := json.Unmarshal(bytes, &settings.SmartSearchSettingsCore)
	if err != nil {
		return err
	}
	settings.searchTerms = parseSearchTerms(settings.SearchText)
	return nil
}

// Validate implements ActionSettings interface
func (settings SmartSearchSettings) Validate() ([]string, error) {
	var warnings []string
	if len(settings.searchTerms) == 0 {
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

func logSearchSuggestionResult(logEntry *logger.LogEntry, searchSuggestionResult *enigma.SearchSuggestionResult) {
	suggestions := []string{}
	for _, suggestion := range searchSuggestionResult.Suggestions {
		suggestions = append(suggestions, quote(suggestion.Value))
	}
	if len(suggestions) > 0 {
		logEntry.LogDebugf("search suggestions: %s", strings.Join(suggestions, ","))
	}
}

func logSearchResult(logEntry *logger.LogEntry, searchResult *enigma.SearchResult) {
	if !logEntry.ShouldLogDebug() {
		return
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
	searchTerms := settings.searchTerms
	sessionState.LogEntry.LogDebugf(`search terms: "%s"`, strings.Join(searchTerms, `","`))

	sessionState.QueueRequest(func(ctx context.Context) error {
		searchSuggestionResult, err := doc.SearchSuggest(
			ctx,
			&enigma.SearchCombinationOptions{},
			searchTerms,
		)
		logSearchSuggestionResult(sessionState.LogEntry, searchSuggestionResult)
		return err
	}, actionState, true, "SearchSuggest call failed")

	sessionState.QueueRequest(func(ctx context.Context) error {
		searchResult, err := doc.SearchResults(
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
		logSearchResult(sessionState.LogEntry, searchResult)
		return err
	}, actionState, true, "SearchResults call failed")

	sessionState.Wait(actionState)
}
