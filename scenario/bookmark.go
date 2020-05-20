package scenario

import (
	"context"
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
		ID    string                 `json:"id" displayname:"Bookmark ID" doc-key:"bookmark.id" appstructure:"bookmark"`
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

	var bm *enigma.GenericBookmark
	if err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		var err error
		bm, err = uplink.CurrentApp.Doc.GetBookmark(ctx, id)
		if err != nil {
			return errors.Wrap(err, "Failed to get bookmark object")
		}
		if _, err := bm.GetLayout(ctx); err != nil {
			return errors.Wrap(err, "Failed to get bookmark layout")
		}
		onBookmarkChanged := func(ctx context.Context, actionState *action.State) error {
			_, err := bm.GetLayout(ctx)
			return errors.WithStack(err)
		}
		sessionState.RegisterEvent(bm.Handle, onBookmarkChanged, nil, true)
		return err
	}); err != nil {
		return nil, err
	}

	return bm, nil
}
