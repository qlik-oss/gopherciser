package session

import (
	"fmt"
	"strconv"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/synced"
	syncedcounter "github.com/qlik-oss/gopherciser/syncedCounter"
)

type (
	// AppSelectionModeEnum from where to get app to open, defaults to "current"
	AppSelectionModeEnum int

	// AppSelectionCore app selection settings shared between multiple actions
	AppSelectionCore struct {
		// AppMode app selection mode
		AppMode AppSelectionModeEnum `json:"appmode" displayname:"App selection mode" doc-key:"appselection.appmode"`
		// App name or GUID depending on AppMode
		App synced.Template `json:"app,omitempty" displayname:"App" doc-key:"appselection.app"`
		// AppList of app names or GUID's, depending on AppMode
		AppList []string `json:"list,omitempty" displayname:"App selection list" doc-key:"appselection.list"`
		// Filename of file containing app names, one app per line
		Filename helpers.RowFile `json:"filename,omitempty" displayname:"App selection filename" displayelement:"file" doc-key:"appselection.filename"`
	}

	// AppSelection contains the selected app for the current session
	AppSelection struct {
		AppSelectionCore

		//listCounter is used for round robin from list, should be shared by all users, but unique for each AppSelection instance
		listCounter *syncedcounter.Counter
	}
)

var (
	appModeEnum = enummap.NewEnumMapOrPanic(map[string]int{
		"current":            int(AppModeCurrent),
		"guid":               int(AppModeGUID),
		"name":               int(AppModeName),
		"random":             int(AppModeRandom),
		"randomnamefromlist": int(AppModeRandomNameFromList),
		"randomguidfromlist": int(AppModeRandomGUIDFromList),
		"randomnamefromfile": int(AppModeRandomNameFromFile),
		"randomguidfromfile": int(AppModeRandomGUIDFromFile),
		"round":              int(AppModeRound),
		"roundnamefromlist":  int(AppModeRoundNameFromList),
		"roundguidfromlist":  int(AppModeRoundGUIDFromList),
		"roundnamefromfile":  int(AppModeRoundNameFromFile),
		"roundguidfromfile":  int(AppModeRoundGUIDFromFile),
	})
)

// AppSelectionModeEnum from where to get app to open, defaults to "current"
const (
	AppModeCurrent AppSelectionModeEnum = iota
	AppModeGUID
	AppModeName
	AppModeRandom
	AppModeRandomNameFromList
	AppModeRandomGUIDFromList
	AppModeRandomNameFromFile
	AppModeRandomGUIDFromFile
	AppModeRound
	AppModeRoundNameFromList
	AppModeRoundGUIDFromList
	AppModeRoundNameFromFile
	AppModeRoundGUIDFromFile
)

// GetEnumMap for app selection mode
func (mode AppSelectionModeEnum) GetEnumMap() *enummap.EnumMap {
	return appModeEnum
}

// UnmarshalJSON unmarshal AppSelectionModeEnum
func (mode *AppSelectionModeEnum) UnmarshalJSON(arg []byte) error {
	i, err := appModeEnum.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal AppSelectionModeEnum")
	}

	*mode = AppSelectionModeEnum(i)
	return nil
}

// MarshalJSON marshal AppSelectionModeEnum
func (mode AppSelectionModeEnum) MarshalJSON() ([]byte, error) {
	str, err := appModeEnum.String(int(mode))
	if err != nil {
		return nil, errors.Errorf("unknown AppSelectionModeEnum<%d>", mode)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// String implements stringer interface
func (mode AppSelectionModeEnum) String() string {
	return appModeEnum.StringDefault(int(mode), strconv.Itoa(int(mode)))
}

// UnmarshalJSON unmarshal AppSelection
func (appSelection *AppSelection) UnmarshalJSON(arg []byte) error {
	var core AppSelectionCore
	if err := json.Unmarshal(arg, &core); err != nil {
		return errors.WithStack(err)
	}
	*appSelection = AppSelection{
		AppSelectionCore: core,
		listCounter:      &syncedcounter.Counter{},
	}
	return nil
}

// NewAppSelection creates new instance of AppSelection
func NewAppSelection(appMode AppSelectionModeEnum, app string, list []string) (*AppSelection, error) {
	tmpl, err := synced.New(app)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &AppSelection{
		AppSelectionCore: AppSelectionCore{
			AppMode: appMode,
			App:     *tmpl,
			AppList: list,
		},
		listCounter: &syncedcounter.Counter{},
	}, nil
}

// getRoundAppListEntry returns round robin entry from appList, based on local counter
func (appSelection *AppSelection) getRoundAppListEntry(sessionState *State, appList []string) string {
	appNumber := appSelection.listCounter.Inc() - 1
	return appList[appNumber%len(appList)]
}

// getRandomAppListEntry returns a random entry from list, chosen by a uniform distribution
func getRandomAppListEntry(sessionState *State, appList []string) (string, error) {
	n := len(appList)
	if n < 1 {
		return "", fmt.Errorf("specified app list is empty: Nothing to select from")
	}

	randomIndex := sessionState.Randomizer().Rand(n)
	selectedApp := appList[randomIndex]

	return selectedApp, nil
}

// Select new app
func (appSelection *AppSelection) Select(sessionState *State) (*ArtifactEntry, error) {
	var entry *ArtifactEntry
	switch appSelection.AppMode {
	case AppModeCurrent:
		if sessionState.CurrentApp == nil {
			return nil, errors.New("no current app defined, make sure to have preceeding app selection when using app selection mode<current>.")
		}
		return sessionState.CurrentApp, nil
	case AppModeGUID:
		app, err := sessionState.ReplaceSessionVariables(&appSelection.App)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse app guid")
		}
		if app == "" {
			return nil, errors.New("No app defined for app selection mode<guid>")
		}

		entry, err = sessionState.ArtifactMap.LookupAppGUID(app)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	case AppModeName:
		app, err := sessionState.ReplaceSessionVariables(&appSelection.App)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse app name")
		}
		if app == "" {
			return nil, errors.New("No app defined for app selection mode<name>")
		}

		entry, err = sessionState.ArtifactMap.LookupAppTitle(app)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	case AppModeRandom:
		var err error
		entry = &ArtifactEntry{}
		*entry, err = sessionState.ArtifactMap.GetRandomApp(sessionState)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	case AppModeRandomNameFromList:
		app, err := getRandomAppListEntry(sessionState, appSelection.AppList)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		entry, err = sessionState.ArtifactMap.LookupAppTitle(app)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	case AppModeRandomGUIDFromList:
		guid, err := getRandomAppListEntry(sessionState, appSelection.AppList)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		entry, err = sessionState.ArtifactMap.LookupAppGUID(guid)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	case AppModeRound:
		var err error
		entry = &ArtifactEntry{}
		*entry, err = sessionState.ArtifactMap.GetRoundRobin(sessionState)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	case AppModeRoundNameFromList:
		app := appSelection.getRoundAppListEntry(sessionState, appSelection.AppList)
		var err error
		entry, err = sessionState.ArtifactMap.LookupAppTitle(app)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to find app with title<%s>", app)
		}
	case AppModeRoundGUIDFromList:
		guid := appSelection.getRoundAppListEntry(sessionState, appSelection.AppList)
		entry = &ArtifactEntry{ // todo itemID, title?
			ID: guid,
		}
	case AppModeRandomNameFromFile:
		app, err := getRandomAppListEntry(sessionState, appSelection.Filename.Rows())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		entry, err = sessionState.ArtifactMap.LookupAppTitle(app)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	case AppModeRandomGUIDFromFile:
		guid, err := getRandomAppListEntry(sessionState, appSelection.Filename.Rows())
		if err != nil {
			return nil, errors.WithStack(err)
		}
		entry, err = sessionState.ArtifactMap.LookupAppGUID(guid)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	case AppModeRoundNameFromFile:
		app := appSelection.getRoundAppListEntry(sessionState, appSelection.Filename.Rows())
		var err error
		entry, err = sessionState.ArtifactMap.LookupAppTitle(app)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to find app with title<%s>", app)
		}
	case AppModeRoundGUIDFromFile:
		guid := appSelection.getRoundAppListEntry(sessionState, appSelection.Filename.Rows())
		entry = &ArtifactEntry{ // todo itemID, title?
			ID: guid,
		}
	default:
		return nil, errors.Errorf("app selection mode <%s> not supported", appSelection.AppMode)
	}

	if entry.ResourceType == "" {
		entry.ResourceType = ResourceTypeApp
	}
	sessionState.LogEntry.Session.AppName = entry.Name
	sessionState.LogEntry.Session.AppGUID = entry.ID
	sessionState.CurrentApp = entry

	return entry, nil
}

// Validate  AppSelection settings
func (appSelection *AppSelection) Validate() error {
	switch appSelection.AppMode {
	case AppModeCurrent:
		if err := validateFuncs([]func() error{appSelection.validateNoApp, appSelection.validateNoAppList, appSelection.validateNoFileNAme}); err != nil {
			return err
		}
	case AppModeGUID:
		if err := validateFuncs([]func() error{appSelection.validateNoAppList, appSelection.validateNoFileNAme}); err != nil {
			return err
		}
	case AppModeName:
		if err := validateFuncs([]func() error{appSelection.validateNoAppList, appSelection.validateNoFileNAme}); err != nil {
			return err
		}
	case AppModeRandom:
		if err := validateFuncs([]func() error{appSelection.validateNoApp, appSelection.validateNoAppList, appSelection.validateNoFileNAme}); err != nil {
			return err
		}
	case AppModeRandomNameFromList:
		if err := validateFuncs([]func() error{appSelection.validateNoApp, appSelection.validateHasAppList, appSelection.validateNoFileNAme}); err != nil {
			return err
		}
	case AppModeRandomGUIDFromList:
		if err := validateFuncs([]func() error{appSelection.validateNoApp, appSelection.validateHasAppList, appSelection.validateNoFileNAme}); err != nil {
			return err
		}
	case AppModeRandomNameFromFile:
		if err := validateFuncs([]func() error{appSelection.validateNoApp, appSelection.validateNoAppList, appSelection.validateHasFileNAme}); err != nil {
			return err
		}
	case AppModeRandomGUIDFromFile:
		if err := validateFuncs([]func() error{appSelection.validateNoApp, appSelection.validateNoAppList, appSelection.validateHasFileNAme}); err != nil {
			return err
		}
	case AppModeRound:
		if err := validateFuncs([]func() error{appSelection.validateNoApp, appSelection.validateNoAppList, appSelection.validateNoFileNAme}); err != nil {
			return err
		}
	case AppModeRoundNameFromList:
		if err := validateFuncs([]func() error{appSelection.validateNoApp, appSelection.validateHasAppList, appSelection.validateNoFileNAme}); err != nil {
			return err
		}
	case AppModeRoundGUIDFromList:
		if err := validateFuncs([]func() error{appSelection.validateNoApp, appSelection.validateHasAppList, appSelection.validateNoFileNAme}); err != nil {
			return err
		}
	case AppModeRoundNameFromFile:
		if err := validateFuncs([]func() error{appSelection.validateNoApp, appSelection.validateNoAppList, appSelection.validateHasFileNAme}); err != nil {
			return err
		}
	case AppModeRoundGUIDFromFile:
		if err := validateFuncs([]func() error{appSelection.validateNoApp, appSelection.validateNoAppList, appSelection.validateHasFileNAme}); err != nil {
			return err
		}
	default:
		return errors.Errorf("app selection mode%s> not valid", appSelection.AppMode)
	}
	return nil
}

func validateFuncs(funcs []func() error) error {
	for _, f := range funcs {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

func (appSelection *AppSelection) validateNoFileNAme() error {
	if !appSelection.Filename.IsEmpty() {
		return errors.Errorf("filename<%s> not valid for mode<%s>", appSelection.Filename, appSelection.AppMode)
	}
	return nil
}

func (appSelection *AppSelection) validateHasFileNAme() error {
	if appSelection.Filename.IsEmpty() {
		return errors.Errorf("filename required for mode<%s>", appSelection.Filename)
	}
	if len(appSelection.Filename.Rows()) < 1 {
		return errors.Errorf("filename<%s> has no app entries", appSelection.Filename)
	}
	return nil
}

func (appSelection *AppSelection) validateNoAppList() error {
	if len(appSelection.AppList) > 0 {
		return errors.Errorf("app list<%v> not valid for mode<%s>", appSelection.AppList, appSelection.AppMode)
	}
	return nil
}

func (appSelection *AppSelection) validateHasAppList() error {
	if len(appSelection.AppList) < 1 {
		return errors.Errorf("app list required for mode<%s>", appSelection.AppMode)
	}
	return nil
}

func (appSelection *AppSelection) validateNoApp() error {
	if appSelection.App.String() != "" {
		return errors.Errorf("app<%s> not valid for mode<%s>", appSelection.App.String(), appSelection.AppMode)
	}
	return nil
}
