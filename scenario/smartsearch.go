package scenario

import (
	"context"

	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	SmartSearchSettings struct {
		SearchTerm string `json:"searchterm"`
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
	doc := uplink.CurrentApp.Doc
	searchTerms := []string{settings.SearchTerm}

	sessionState.SendRequest(actionState, func(ctx context.Context) error {
		searchSuggestionResult, err := doc.SearchSuggest(
			context.TODO(),
			&enigma.SearchCombinationOptions{},
			searchTerms,
		)
		return err
	})

	sessionState.SendRequest(actionState, func(ctx context.Context) error {
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
			},
		)
		return err
	})
}
