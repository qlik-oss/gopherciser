package session

import (
	"fmt"
	"sort"
	"sync"

	"github.com/pkg/errors"
	enigma "github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/elasticstructs"
	"github.com/qlik-oss/gopherciser/logger"
)

type (
	AppsResp struct {
		Title  string `json:"Title"` // QSCB
		Name   string `json:"name"`  // QCS
		ID     string `json:"resourceId"`
		ItemID string `json:"ID"`
	}

	// AppData struct to unmarshal list app response
	AppData struct {
		Data []AppsResp `json:"data"`
	}

	ArtifactListDict []*ArtifactEntry

	// ArtifactMap is the map between app names and GUIDs
	ArtifactMap struct {
		appTitleToID     *sync.Map
		appTitleToItemID *sync.Map
		streamTitleToID  *sync.Map
		spaceTitleToID   *sync.Map
		AppList          ArtifactListDict
		mu               sync.Mutex
	}

	// ArtifactEntry is a key value pair
	ArtifactEntry struct {
		Title  string
		GUID   string
		ItemID string
	}

	SpaceNameNotFoundError string
	SpaceIdNotFoundError   string
)

// Error implements error interface
func (err SpaceNameNotFoundError) Error() string {
	return fmt.Sprintf("space name <%s> not found", string(err))
}

func (err SpaceIdNotFoundError) Error() string {
	return fmt.Sprintf("space id <%s> not found", string(err))
}

func (d ArtifactListDict) Len() int {
	return len(d)
}
func (d ArtifactListDict) Less(i, j int) bool {
	return d[i].Title < d[j].Title
}
func (d ArtifactListDict) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// NewAppMap returns an empty ArtifactMap
func NewAppMap() *ArtifactMap {
	return &ArtifactMap{
		appTitleToID:     &sync.Map{},
		streamTitleToID:  &sync.Map{},
		appTitleToItemID: &sync.Map{},
		spaceTitleToID:   &sync.Map{},
	}
}

// IsEmpty returns true of entry is nil or has no GUID
func (entry *ArtifactEntry) IsEmpty() bool {
	if entry == nil {
		return true
	}

	// todo how to handle itemID?

	if entry.GUID == "" {
		return true
	}

	return false
}

// Copy of ArtifactEntry
func (entry *ArtifactEntry) Copy() *ArtifactEntry {
	if entry == nil {
		return nil
	}

	return &ArtifactEntry{
		Title:  entry.Title,
		GUID:   entry.GUID,
		ItemID: entry.ItemID,
	}
}

// fillAppMap puts the app name (key) and the app ID (value) in the ArtifactMap
func (am *ArtifactMap) fillAppMap(appData *AppData, lookFor string) error {
	if appData == nil || appData.Data == nil {
		return errors.New("empty AppData struct")
	}

	if len(appData.Data) == 0 {
		return nil
	}

	am.mu.Lock()
	defer am.mu.Unlock()
	if am.AppList == nil {
		am.AppList = make([]*ArtifactEntry, 0, len(appData.Data))
	}

	for _, i := range appData.Data {
		// This distinction is needed since QSCB and QCS list apps
		// slightly differently. The app Title is in a "Title" field
		// in QSCB, whereas it's in a "name" field in QCS.
		switch lookFor {
		case "Title":
			am.appTitleToID.Store(i.Title, i.ID)
			am.AppList = append(am.AppList, &ArtifactEntry{i.Title, i.ID, i.ItemID})
		case "name":
			am.appTitleToID.Store(i.Name, i.ID)
			am.appTitleToItemID.Store(i.Name, i.ItemID)
			am.AppList = append(am.AppList, &ArtifactEntry{i.Name, i.ID, i.ItemID})
		default:
			return fmt.Errorf("%s type not supported", lookFor)
		}
	}
	sort.Sort(am.AppList)
	return nil
}

// DeleteApp deletes an app from the ArtifactMap
func (am *ArtifactMap) DeleteApp(AppGUID string) {
	var desiredApp *ArtifactEntry
	index := 0
	for i, app := range am.AppList {
		if app.GUID == AppGUID {
			desiredApp = app
			index = i
			break
		}
	}
	if desiredApp != nil {
		am.AppList = append(am.AppList[:index], am.AppList[index+1:]...)
		am.appTitleToID.Delete(desiredApp.Title)
		am.appTitleToItemID.Delete(desiredApp.Title)
	}
}

// DeleteStream deletes a stream from the ArtifactMap
func (am *ArtifactMap) DeleteStream(streamName string) {
	am.streamTitleToID.Delete(streamName)
}

// FillAppsUsingTitle should be used to fillAppMap app map in QSCB
func (am *ArtifactMap) FillAppsUsingTitle(appData *AppData) error {
	return am.fillAppMap(appData, "Title")
}

// FillAppsUsingName should be used to fillAppMap app map in QCS
func (am *ArtifactMap) FillAppsUsingName(appData *AppData) error {
	return am.fillAppMap(appData, "name")
}

// EmptyApps Empty Apps from ArtifactMap
func (am *ArtifactMap) EmptyApps() {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.appTitleToID = &sync.Map{}
	am.appTitleToItemID = &sync.Map{}
	am.AppList = nil
}

// FillAppsUsingDocListEntries should be used to fillAppMap app map in QSEfW
func (am *ArtifactMap) FillAppsUsingDocListEntries(docListEntries []*enigma.DocListEntry) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for _, docListEntry := range docListEntries {
		id := docListEntry.DocId
		name := docListEntry.DocName
		am.appTitleToID.Store(name, id)
		am.AppList = append(am.AppList, &ArtifactEntry{name, id, "" /* QSEfW does not have item ID */})
	}
	sort.Sort(am.AppList)
	return nil
}

// FillStreams fills the stream map with the streams from the given list
func (am *ArtifactMap) FillStreams(streamList []elasticstructs.Collection) {
	for _, stream := range streamList {
		am.streamTitleToID.Store(stream.Name, stream.ID)
	}
}

// FillSpaces fills the spaces map with the spaces from the given list
func (am *ArtifactMap) FillSpaces(spaces []elasticstructs.Space) {
	for _, space := range spaces {
		am.spaceTitleToID.Store(space.Name, space)
	}
}

// AddSpace to artifact map
func (am *ArtifactMap) AddSpace(space elasticstructs.Space) {
	am.spaceTitleToID.Store(space.Name, space)
}

// GetAppID returns the app ID given the app Title. When multiple apps
// have the same Title, this will return the ID of the last app in the order
// of the struct passed to the Fill function.
func (am *ArtifactMap) GetAppID(appName string) (string, error) {
	value, found := am.appTitleToID.Load(appName)
	if !found {
		return "", errors.Errorf("key %s not found in app id map", appName)
	}
	return value.(string), nil
}

// GetItemId returns the item ID given the app Title. When multiple apps
// have the same Title, this will return the ID of the last app in the order
// of the struct passed to the Fill function.
func (am *ArtifactMap) GetItemId(appName string) (string, error) {
	value, found := am.appTitleToItemID.Load(appName)
	if !found {
		return "", errors.Errorf("key %s not found in app item map", appName)
	}
	return value.(string), nil
}

// GetStreamID returns the app ID given the stream
func (am *ArtifactMap) GetStreamID(streamName string) (string, error) {
	value, found := am.streamTitleToID.Load(streamName)
	if !found {
		return "", errors.Errorf("key %s not found in map", streamName)
	}
	return value.(string), nil
}

// GetSpaceByName returns the space ID given space name
func (am *ArtifactMap) GetSpaceByName(spaceName string) (*elasticstructs.Space, error) {
	value, found := am.spaceTitleToID.Load(spaceName)
	if !found {
		return nil, SpaceNameNotFoundError(spaceName)
	}

	retVal, ok := value.(elasticstructs.Space)
	if !ok {
		return nil, errors.Errorf("space<%s>", spaceName)
	}

	return &retVal, nil
}

// GetSpaceByID returns the space ID given space name
func (am *ArtifactMap) GetSpaceByID(spaceID string) (*elasticstructs.Space, error) {
	var space *elasticstructs.Space
	am.spaceTitleToID.Range(func(key, value interface{}) bool {
		if sp, ok := value.(elasticstructs.Space); ok && sp.ID == spaceID {
			space = &sp
			return false
		}
		return true
	})

	if space != nil {
		return space, nil
	}

	return nil, SpaceIdNotFoundError(spaceID)
}

// GetRandomApp returns a random app for the map, chosen by a uniform distribution
func (am *ArtifactMap) GetRandomApp(sessionState *State) (ArtifactEntry, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	n := len(am.AppList)
	if n < 1 {
		return ArtifactEntry{}, errors.New("cannot select random app: ArtifactMap is empty")
	}
	randomIndex := sessionState.Randomizer().Rand(n)
	selectedKVP := am.AppList[randomIndex]
	return *selectedKVP, nil
}

// GetRoundRobin returns a app round robin for the map
func (am *ArtifactMap) GetRoundRobin(sessionState *State) (ArtifactEntry, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	appNumber := sessionState.Counters.AppCounter.Inc() - 1
	n := len(am.AppList)
	if n < 1 {
		return ArtifactEntry{}, errors.New("cannot select app round robin: ArtifactMap is empty")
	}
	selectedKVP := am.AppList[appNumber%uint64(n)]
	return *selectedKVP, nil
}

// LookupAppTitle lookup app using title
func (am *ArtifactMap) LookupAppTitle(sessionState *State, title string) (*ArtifactEntry, error) {
	return am.lookup(sessionState, title, false)
}

// SetCurrentAppTitle lookup app using GUID
func (am *ArtifactMap) LookupAppGUID(sessionState *State, guid string) (*ArtifactEntry, error) {
	entry, err := am.lookup(sessionState, guid, true)
	if err != nil {
		// GUID not found in map, create new entry with GUID (Supports using openapp with GUID and no preceeding OpenHub)
		entry = &ArtifactEntry{GUID: guid} // todo how to handle itemID?
	}
	return entry, nil
}

func (am *ArtifactMap) lookup(sessionState *State, lookup string, isGUID bool) (*ArtifactEntry, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	var compare func(*ArtifactEntry, string) bool

	if isGUID {
		compare = compareGUID
	} else {
		compare = compareName
	}

	// entry.Title == name
	for _, entry := range am.AppList {
		if compare(entry, lookup) {
			app := entry.Copy()
			return app, nil
		}
	}

	return nil, errors.Errorf("app<%s> not found in artifact map", lookup)
}

func compareName(entry *ArtifactEntry, name string) bool {
	if entry == nil {
		return false
	}
	return entry.Title == name
}

func compareGUID(entry *ArtifactEntry, guid string) bool {
	if entry == nil {
		return false
	}
	return entry.GUID == guid
}

// // GetRandomAppFromSubset returns a random app from the subset, chosen by a uniform distribution
// func (am *ArtifactMap) GetRandomAppFromSubset(sessionState *State, subset []string) (string, error) {
// 	if am == nil {
// 		return "", errors.New("Artifact map is nil")
// 	}
// 	am.mu.Lock()
// 	defer am.mu.Unlock()

// 	n := len(subset)
// 	if n < 1 {
// 		return "", fmt.Errorf("specified app subset is empty: Nothing to select from")
// 	}
// 	randomIndex := sessionState.Randomizer().Rand(n)
// 	selectedApp := subset[randomIndex]

// 	return selectedApp, nil
// }

// LogMap log entire map as debug logging
func (am *ArtifactMap) LogMap(entry *logger.LogEntry) error {
	// check if debug logging is turned on
	if entry == nil || !entry.ShouldLogDebug() {
		return nil
	}

	js, err := am.Json()
	if err != nil {
		return errors.Wrap(err, "failed to marshal artifacts map")
	}

	entry.LogDebugf("ArtifactMap:%s", js)
	return nil
}

// Json locks artifact map and marshals json
func (am *ArtifactMap) Json() ([]byte, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	return jsonit.Marshal(am.AppList)
}
