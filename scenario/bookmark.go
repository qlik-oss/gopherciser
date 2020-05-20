package scenario

import (
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/senseobjects"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	BookMarkSettings struct {
		Title session.SyncedTemplate `json:"title" displayname:"Bookmark title" doc-key:"bookmark.title"`
		ID    string                 `json:"id" displayname:"Bookmark ID" doc-key:"bookmark.id"`
	}
)

// getBookmark defined by BookMarkSettings, returns bookmark ID, sheet ID
func (settings BookMarkSettings) getBookmark(sessionState *session.State, actionState *action.State, uplink *enigmahandlers.SenseUplink) (string, string, error) {
	bl, err := uplink.CurrentApp.GetBookmarkList(sessionState, actionState)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to GetBookmarkList")
	}

	term, input, err := settings.getSearchTerm(sessionState)
	if err != nil {
		return "", "", err
	}

	id, sheetID, err := getBookmarkData(bl, input, term)
	if err != nil {
		return "", "", err
	}

	if id == "" { // No ID and named bookmark not found
		return id, sheetID, errors.New("could not find specified bookmark")
	}

	return id, sheetID, nil
}

func (settings BookMarkSettings) getSearchTerm(sessionState *session.State) (bmSearchTerm, string, error) {
	var input string
	var term bmSearchTerm
	var err error

	if settings.ID == "" {
		input, err = sessionState.ReplaceSessionVariables(&settings.Title)
		term = bmSearchTitle
	} else {
		input = sessionState.IDMap.Get(settings.ID)
		term = bmSearchId
	}
	return term, input, err
}

func getBookmarkData(bl *senseobjects.BookmarkList, input string, term bmSearchTerm) (string /*id*/, string /*sheetId*/, error) {
	for _, bookmark := range bl.GetBookmarks() {
		switch term {
		case bmSearchTitle:
			if input == bookmark.Data.Title {
				return bookmark.Info.Id, bookmark.Data.SheetId, nil
			}
		case bmSearchId:
			if input == bookmark.Info.Id {
				return bookmark.Info.Id, bookmark.Data.SheetId, nil
			}
		}
	}

	return "", "", errors.New("bookmark not found")
}

func (settings BookMarkSettings) getBookmarkObject(sessionState *session.State, actionState *action.State, uplink *enigmahandlers.SenseUplink) (*enigma.GenericBookmark, error) {
	id, _, err := settings.getBookmark(sessionState, actionState, uplink)
	if err != nil {
		return nil, err
	}

	return uplink.CurrentApp.GetBookmarkObject(sessionState, actionState, id)
}
