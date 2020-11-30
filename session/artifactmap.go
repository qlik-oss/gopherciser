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
	// ArtifactEntry is the entry for most, but not all, item types in artifact map
	ArtifactEntry struct {
		Name         string `json:"name"`
		ID           string `json:"resourceId"`
		ItemID       string `json:"ID"`
		ResourceType string `json:"resourceType"`
	}

	// AppData struct to unmarshal list app response
	AppData struct {
		Data []ArtifactEntry `json:"data"`
	}

	// ArtifactList all items, spaces and collections
	ArtifactList []*ArtifactEntry

	// ArtifactMap is the map between app names and GUIDs
	ArtifactMap struct {
		nameResouceToItem *sync.Map
		artifactList      ArtifactList
		mu                sync.Mutex
	}

	// ArtifactKey used to find item using title
	ArtifactKey struct {
		Title string
		Type  string
	}

	// // ArtifactEntry is a key value pair
	// ArtifactEntry struct {
	// 	Title        string
	// 	GUID         string
	// 	ItemID       string
	// 	ResourceType string
	// }

	// SpaceNameNotFoundError space with name not found in artifact map error
	SpaceNameNotFoundError string
	// SpaceIDNotFoundError space with ID not found in artifact map error
	SpaceIDNotFoundError string
)

// Currently known resource types
const (
	ResourceTypeApp         = "app"
	ResourceTypeGenericLink = "genericlink"
	ResourceTypeDataset     = "dataset"
	ResourceTypeDataAsset   = "dataasset"

	// none Sense "resource types"
	ResourceTypeStream = "stream"
	ResourceTypeSpace  = "space"
)

// Error implements error interface
func (err SpaceNameNotFoundError) Error() string {
	return fmt.Sprintf("space name <%s> not found", string(err))
}

func (err SpaceIDNotFoundError) Error() string {
	return fmt.Sprintf("space id <%s> not found", string(err))
}

func (d ArtifactList) Len() int {
	return len(d)
}
func (d ArtifactList) Less(i, j int) bool {
	return d[i].Name < d[j].Name
}
func (d ArtifactList) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

// NewAppMap returns an empty ArtifactMap
func NewAppMap() *ArtifactMap {
	return &ArtifactMap{nameResouceToItem: &sync.Map{}}
}

// IsEmpty returns true of entry is nil or has no GUID
func (entry *ArtifactEntry) IsEmpty() bool {
	if entry == nil {
		return true
	}

	// todo how to handle itemID?

	if entry.ID == "" {
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
		Name:         entry.Name,
		ID:           entry.ID,
		ItemID:       entry.ItemID,
		ResourceType: entry.ResourceType,
	}
}

// fillAppMap puts the app name (key) and the app ID (value) in the ArtifactMap
func (am *ArtifactMap) fillAppMap(appData *AppData) error {
	if appData == nil || appData.Data == nil {
		return errors.New("empty AppData struct")
	}

	if len(appData.Data) == 0 {
		return nil
	}

	am.mu.Lock()
	defer am.mu.Unlock()
	if am.artifactList == nil {
		am.artifactList = make([]*ArtifactEntry, 0, len(appData.Data))
	}

	for _, i := range appData.Data {
		am.nameResouceToItem.Store(ArtifactKey{Title: i.Name, Type: i.ResourceType}, i)
		am.artifactList = append(am.artifactList, &ArtifactEntry{i.Name, i.ID, i.ItemID, i.ResourceType})
	}
	sort.Sort(am.artifactList)
	return nil
}

// DeleteApp deletes an app from the ArtifactMap
func (am *ArtifactMap) DeleteApp(AppGUID string) {
	var desiredApp *ArtifactEntry
	index := 0
	for i, app := range am.artifactList {
		if app.ID == AppGUID {
			desiredApp = app
			index = i
			break
		}
	}
	if desiredApp != nil {
		am.artifactList = append(am.artifactList[:index], am.artifactList[index+1:]...)
		am.nameResouceToItem.Delete(desiredApp.Name)
	}
}

// DeleteStream deletes a stream from the ArtifactMap
func (am *ArtifactMap) DeleteStream(streamName string) {
	am.nameResouceToItem.Delete(ArtifactKey{streamName, ResourceTypeStream})
}

// FillAppsUsingName should be used to fillAppMap app map in QCS
func (am *ArtifactMap) FillAppsUsingName(appData *AppData) error {
	return am.fillAppMap(appData)
}

// EmptyApps Empty Apps from ArtifactMap
// TODO empty all or only apps?
func (am *ArtifactMap) EmptyApps() {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.nameResouceToItem = &sync.Map{}
	am.artifactList = nil
}

// FillAppsUsingDocListEntries should be used to fillAppMap app map in QSEoW
func (am *ArtifactMap) FillAppsUsingDocListEntries(docListEntries []*enigma.DocListEntry) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for _, docListEntry := range docListEntries {
		id := docListEntry.DocId
		name := docListEntry.DocName
		am.nameResouceToItem.Store(ArtifactKey{name, ResourceTypeApp}, ArtifactEntry{Name: name, ID: id, ResourceType: ResourceTypeApp, ItemID: ""})
		am.artifactList = append(am.artifactList, &ArtifactEntry{name, id, "" /* QSEoW does not have item ID */, ResourceTypeApp})
	}
	sort.Sort(am.artifactList)
	return nil
}

// FillStreams fills the stream map with the streams from the given list
func (am *ArtifactMap) FillStreams(streamList []elasticstructs.Collection) {
	for _, stream := range streamList {
		am.nameResouceToItem.Store(ArtifactKey{stream.Name, ResourceTypeStream}, ArtifactEntry{Name: stream.Name, ID: stream.ID, ResourceType: ResourceTypeStream})
	}
}

// FillSpaces fills the spaces map with the spaces from the given list
func (am *ArtifactMap) FillSpaces(spaces []elasticstructs.Space) {
	for _, space := range spaces {
		am.nameResouceToItem.Store(ArtifactKey{Title: space.Name, Type: ResourceTypeSpace}, space)
	}
}

// AddSpace to artifact map
func (am *ArtifactMap) AddSpace(space elasticstructs.Space) {
	am.nameResouceToItem.Store(ArtifactKey{Title: space.Name, Type: ResourceTypeSpace}, space)
}

// GetAppID returns the app ID given the app Title. When multiple apps
// have the same Title, this will return the ID of the last app in the order
// of the struct passed to the Fill function.
func (am *ArtifactMap) GetAppID(appName string) (string, error) {
	entry, err := am.getAppEntry(appName)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return entry.ID, nil
}

// GetItemID returns the item ID given the app Title. When multiple apps
// have the same Title, this will return the ID of the last app in the order
// of the struct passed to the Fill function.
func (am *ArtifactMap) GetItemID(appName string) (string, error) {
	entry, err := am.getAppEntry(appName)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return entry.ItemID, nil
}

func (am *ArtifactMap) getAppEntry(appName string) (*ArtifactEntry, error) {
	value, found := am.nameResouceToItem.Load(ArtifactKey{Title: appName, Type: ResourceTypeApp})
	if !found {
		return nil, errors.Errorf("key %s not found in app id map", appName)
	}
	entry, ok := value.(ArtifactEntry)
	if !ok {
		return nil, errors.Errorf("app name %s found in app id map, but type<%T> is not ArtifactEntry", appName, value)
	}
	return &entry, nil
}

// GetStreamID returns the app ID given the stream
func (am *ArtifactMap) GetStreamID(streamName string) (string, error) {
	value, found := am.nameResouceToItem.Load(ArtifactKey{Title: streamName, Type: ResourceTypeStream})
	if !found {
		return "", errors.Errorf("key %s not found in map", streamName)
	}
	return value.(string), nil
}

// GetSpaceByName returns the space ID given space name
func (am *ArtifactMap) GetSpaceByName(spaceName string) (*elasticstructs.Space, error) {
	value, found := am.nameResouceToItem.Load(ArtifactKey{Title: spaceName, Type: ResourceTypeSpace})
	if !found {
		return nil, SpaceNameNotFoundError(spaceName)
	}

	retVal, ok := value.(elasticstructs.Space)
	if !ok {
		return nil, errors.Errorf("space<%s> type<%T> not of type Space", spaceName, value)
	}

	return &retVal, nil
}

// GetSpaceByID returns the space ID given space name
func (am *ArtifactMap) GetSpaceByID(spaceID string) (*elasticstructs.Space, error) {
	var space *elasticstructs.Space
	am.nameResouceToItem.Range(func(key, value interface{}) bool {
		if sp, ok := value.(elasticstructs.Space); ok && sp.ID == spaceID {
			space = &sp
			return false
		}
		return true
	})

	if space != nil {
		return space, nil
	}

	return nil, SpaceIDNotFoundError(spaceID)
}

// GetRandomApp returns a random app for the map, chosen by a uniform distribution
func (am *ArtifactMap) GetRandomApp(sessionState *State) (ArtifactEntry, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	n := len(am.artifactList) // TODO this list should have all artifacts or only apps?
	if n < 1 {
		return ArtifactEntry{}, errors.New("cannot select random app: ArtifactMap is empty")
	}
	randomIndex := sessionState.Randomizer().Rand(n)
	selectedKVP := am.artifactList[randomIndex]
	return *selectedKVP, nil
}

// GetRoundRobin returns a app round robin for the map
func (am *ArtifactMap) GetRoundRobin(sessionState *State) (ArtifactEntry, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	appNumber := sessionState.Counters.AppCounter.Inc() - 1
	n := len(am.artifactList) // TODO list currently not only apps
	if n < 1 {
		return ArtifactEntry{}, errors.New("cannot select app round robin: ArtifactMap is empty")
	}
	selectedKVP := am.artifactList[appNumber%uint64(n)]
	return *selectedKVP, nil
}

// LookupAppTitle lookup app using title
func (am *ArtifactMap) LookupAppTitle(sessionState *State, title string) (*ArtifactEntry, error) {
	return am.lookup(sessionState, title, false)
}

// LookupAppGUID lookup app using GUID
func (am *ArtifactMap) LookupAppGUID(sessionState *State, guid string) (*ArtifactEntry, error) {
	entry, err := am.lookup(sessionState, guid, true)
	if err != nil {
		// GUID not found in map, create new entry with GUID (Supports using openapp with GUID and no preceeding OpenHub)
		entry = &ArtifactEntry{ID: guid} // todo how to handle itemID?
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
	for _, entry := range am.artifactList {
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
	return entry.Name == name
}

func compareGUID(entry *ArtifactEntry, guid string) bool {
	if entry == nil {
		return false
	}
	return entry.ID == guid
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

	js, err := am.JSON()
	if err != nil {
		return errors.Wrap(err, "failed to marshal artifacts map")
	}

	entry.LogDebugf("ArtifactMap:%s", js)
	return nil
}

// JSON locks artifact map and marshals json
func (am *ArtifactMap) JSON() ([]byte, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	return jsonit.Marshal(am.artifactList)
}

// Len length of artifact list
func (am *ArtifactMap) Len() int {
	am.mu.Lock()
	defer am.mu.Unlock()

	return len(am.artifactList)
}

// ForEach entry, execute function f, return true to continue loop
func (am *ArtifactMap) ForEach(f func(entry *ArtifactEntry) bool) {
	for _, entry := range am.artifactList {
		if !f(entry) {
			return
		}
	}
}
