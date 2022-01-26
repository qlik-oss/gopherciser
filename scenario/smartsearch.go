package scenario

import (
	"context"

	"github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	SmartSearchSettings struct{}
)

// Validate implements ActionSettings interface
func (settings SmartSearchSettings) Validate() ([]string, error) {
	return nil, nil
}

// Execute implements ActionSettings interface
func (settings SmartSearchSettings) Execute(sessionState *session.State, actionState *action.State, connection *connection.ConnectionSettings, label string, reset func()) {
	uplink := sessionState.Connection.Sense()
	doc := uplink.CurrentApp.Doc
	doc.SearchSuggest(
		context.TODO(),
		&enigma.SearchCombinationOptions{},
		[]string{""},
	)
	doc.SearchResults(
		context.TODO(),
		&enigma.SearchCombinationOptions{},
		[]string{""},
		&enigma.SearchPage{},
	)
	return

}
