package scenario

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	SmartSearchSettings struct {
		SearchText string `json:"searchtext"`
	}
)

// Validate implements ActionSettings interface
func (settings SmartSearchSettings) Validate() ([]string, error) {
	return nil, nil
}

func intPtr(i int) *int {
	return &i
}

// Execute implements ActionSettings interface
func (settings SmartSearchSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	uplink := sessionState.Connection.Sense()
	if uplink.CurrentApp == nil {
		actionState.AddErrors(errors.New("not connected to app"))
		return
	}
	doc := uplink.CurrentApp.Doc
	searchTerms := strings.Fields(settings.SearchText)

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
	if sessionState.LogEntry.ShouldLogDebug() {
		suggestions := []string{}

		for _, suggestion := range searchSuggestionResult.Suggestions {
			suggestions = append(suggestions, suggestion.Value)
		}
		sessionState.LogEntry.LogDebugf("smart search suggestions: %s", strings.Join(suggestions, ", "))

		for _, sga := range searchResult.SearchGroupArray {
			for _, item := range sga.Items {
				for _, match := range item.ItemMatches {
					sessionState.LogEntry.LogDebugf("smart search result: (%s, %s, %s)", item.ItemType, item.Identifier, match.Text)
				}
			}
		}

	}

	sessionState.Wait(actionState)
}
