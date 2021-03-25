package scenario

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/action"
	"github.com/qlik-oss/gopherciser/appstructure"
	"github.com/qlik-oss/gopherciser/connection"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
)

type (
	//BookmarkDeletionModeEnum defines what bookmarks to remove
	BookmarkDeletionModeEnum int

	//DeleteBookmarkSettings create bookmark settings
	DeleteBookmarkSettings struct {
		BookMarkSettings
		DeletionMode BookmarkDeletionModeEnum `json:"mode" displayname:"Deletion mode" doc-key:"deletebookmark.mode"`
	}
)

const (
	// SingleBookmark delete specified bookmark
	SingleBookmark BookmarkDeletionModeEnum = iota
	// MatchingBookmarks delete all bookmarks in app
	MatchingBookmarks
	// AllBookmarks delete all bookmarks in app
	AllBookmarks
)

func (value BookmarkDeletionModeEnum) GetEnumMap() *enummap.EnumMap {
	enumMap, _ := enummap.NewEnumMap(map[string]int{
		"single":   int(SingleBookmark),
		"matching": int(MatchingBookmarks),
		"all":      int(AllBookmarks),
	})
	return enumMap
}

// UnmarshalJSON unmarshal BookmarkDeletionModeEnum
func (value *BookmarkDeletionModeEnum) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal BookmarkDeletionModeEnum")
	}

	*value = BookmarkDeletionModeEnum(i)
	return nil
}

// MarshalJSON marshal BookmarkDeletionModeEnum type
func (value BookmarkDeletionModeEnum) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("unknown BookmarkDeletionModeEnum<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// Validate DeleteBookmarkSettings action (Implements ActionSettings interface)
func (settings DeleteBookmarkSettings) Validate() ([]string, error) {
	if settings.DeletionMode == SingleBookmark {
		if settings.Title.String() == "" && settings.ID == "" {
			return nil, errors.New("either specify bookmark title or bookmark id")
		}
	}
	if settings.DeletionMode == MatchingBookmarks {
		if settings.Title.String() == "" {
			return nil, errors.New("please specify bookmark title")
		}
	}
	if settings.DeletionMode == AllBookmarks {
		if settings.Title.String() != "" || settings.ID != "" {
			return nil, errors.New("neither bookmark title nor id cannot be specified when deleting all bookmarks")
		}
	}
	if settings.Title.String() != "" && settings.ID != "" {
		return nil, errors.New("specify only one of the following - bookmark title and bookmark id")
	}
	return nil, nil
}

// Execute DeleteBookmarkSettings action (Implements ActionSettings interface)
func (settings DeleteBookmarkSettings) Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string, reset func()) {
	if sessionState.Connection == nil || sessionState.Connection.Sense() == nil {
		actionState.AddErrors(errors.New("not connected to a Sense environment"))
		return
	}

	app := sessionState.Connection.Sense().CurrentApp
	if app == nil {
		actionState.AddErrors(errors.New("not connected to a Sense app"))
		return
	}

	uplink := sessionState.Connection.Sense()

	title, err := sessionState.ReplaceSessionVariables(&settings.Title)
	if err != nil {
		actionState.AddErrors(err)
		return
	}

	bl, err := uplink.CurrentApp.GetBookmarkList(sessionState, actionState)
	if err != nil {
		actionState.AddErrors(errors.Wrap(err, "failed to GetBookmarkList"))
		return
	}

	n := 0
	switch settings.DeletionMode {
	case SingleBookmark:
		id := sessionState.IDMap.Get(settings.ID)
		if id == "" { // Search by title if ID not specified
			for _, bookmark := range bl.GetBookmarks() {
				name := bookmark.Data.Title
				if name == title {
					id = bookmark.Info.Id
					break
				}
			}
		}

		if id == "" { // No ID and named bookmark not found
			actionState.AddErrors(errors.New("could not find specified bookmark"))
			return
		}

		err = settings.destroyBookmarkById(sessionState, actionState, uplink, id)
		if err != nil {
			actionState.AddErrors(errors.Wrap(err, fmt.Sprintf("failed to delete bookmark <%s>", id)))
			return
		}
		n = 1
	case MatchingBookmarks:
		for _, bookmark := range bl.GetBookmarks() {
			name := bookmark.Data.Title
			if name != title {
				continue
			}
			// Bookmark matches name, destroy it
			err = settings.destroyBookmarkById(sessionState, actionState, uplink, bookmark.Info.Id)
			if err != nil {
				actionState.AddErrors(errors.Wrap(err, fmt.Sprintf("failed to delete bookmark <%s>", bookmark.Info.Id)))
				sessionState.LogEntry.LogInfo("NumDeletedBookmarks", fmt.Sprintf("%d", n))
				sessionState.Wait(actionState)
				return
			}
			n++
		}
	case AllBookmarks:
		for _, bookmark := range bl.GetBookmarks() {
			err = settings.destroyBookmarkById(sessionState, actionState, uplink, bookmark.Info.Id)
			if err != nil {
				actionState.AddErrors(errors.Wrap(err, fmt.Sprintf("failed to delete bookmark <%s>", bookmark.Info.Id)))
				sessionState.LogEntry.LogInfo("NumDeletedBookmarks", fmt.Sprintf("%d", n))
				sessionState.Wait(actionState)
				return
			}
			n++
		}
	}

	sessionState.LogEntry.LogInfo("NumDeletedBookmarks", fmt.Sprintf("%d", n))

	if n == 0 {
		sessionState.LogEntry.Logf(logger.WarningLevel, "no bookmarks deleted")
	}

	sessionState.Wait(actionState)
}

func (settings DeleteBookmarkSettings) destroyBookmarkById(sessionState *session.State, actionState *action.State, uplink *enigmahandlers.SenseUplink, id string) error {
	err := sessionState.SendRequest(actionState, func(ctx context.Context) error {
		success, err := uplink.CurrentApp.Doc.DestroyBookmark(ctx, id)
		if err != nil {
			return err
		}
		if !success {
			return errors.Errorf("destroy bookmark unsuccessful")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "failed to destroy bookmark")
	}
	return nil
}

// AffectsAppObjectsAction implements AffectsAppObjectsAction interface
func (settings DeleteBookmarkSettings) AffectsAppObjectsAction(structure appstructure.AppStructure) ([]*appstructure.AppStructurePopulatedObjects, []string, bool) {
	switch settings.DeletionMode {
	case SingleBookmark:
		for _, obj := range structure.Bookmarks {
			if settings.ID != "" && obj.ID == settings.ID {
				return nil, []string{settings.ID}, false
			}
			if obj.Title == settings.Title.String() {
				return nil, []string{obj.ID}, false
			}
		}
		return nil, nil, false // Not found
	case MatchingBookmarks:
		list := make([]string, 0)
		for _, obj := range structure.Bookmarks {
			if (settings.ID != "" && obj.ID == settings.ID) || obj.Title == settings.Title.String() {
				list = append(list, obj.ID)
			}
		}
		return nil, list, false
	case AllBookmarks:
		list := make([]string, 0)
		for _, obj := range structure.Bookmarks {
			list = append(list, obj.ID)
		}
		return nil, list, false
	default:
		return nil, nil, false
	}
}
