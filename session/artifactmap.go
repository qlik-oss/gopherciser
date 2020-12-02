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
		Name         string      `json:"name"`
		ID           string      `json:"resourceId"`
		ItemID       string      `json:"ID,omitempty"`
		ResourceType string      `json:"resourceType"`
		Data         interface{} `json:"-"`
	}

	// ItemData struct to unmarshal list app response
	ItemData struct {
		Data []ArtifactEntry `json:"data"`
	}

	// ArtifactList all items, spaces and collections
	ArtifactList struct {
		list   []*ArtifactEntry
		sorted bool
	}

	// ArtifactMap is the map between app names and GUIDs
	ArtifactMap struct {
		resourceMap map[string]*ArtifactList
		mu          sync.RWMutex
	}

	// ArtifactKey used to find item using title
	ArtifactKey struct {
		Title string
		Type  string
	}

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

func (d *ArtifactList) Len() int {
	if d == nil {
		return 0
	}
	return len(d.list)
}

func (d ArtifactList) Less(i, j int) bool {
	return d.list[i].Name < d.list[j].Name
}

func (d ArtifactList) Swap(i, j int) {
	d.list[i], d.list[j] = d.list[j], d.list[i]
}

// MarshalJSON marshal artifact list to JSON
func (d ArtifactList) MarshalJSON() ([]byte, error) {
	return jsonit.Marshal(d.list)
}

// NewArtifactMap returns an empty ArtifactMap
func NewArtifactMap() *ArtifactMap {
	return &ArtifactMap{resourceMap: make(map[string]*ArtifactList)}
}

// Copy of ArtifactEntry
func (entry *ArtifactEntry) Copy() *ArtifactEntry {
	if entry == nil {
		return nil
	}

	cpy := *entry
	return &cpy
}

// DataAsSpace return artifact entry Data as space
func (entry *ArtifactEntry) DataAsSpace() (*elasticstructs.Space, error) {
	space, ok := entry.Data.(*elasticstructs.Space)
	if !ok {
		if space == nil {
			return nil, errors.Errorf("no space data saved to artifact map for space name<%s> id<%s>", entry.Name, entry.ID)
		}
		return nil, errors.Errorf("unexpected space type<%T> saved to artifact map space name<%s> id<%s>", entry.Data, entry.Name, entry.ID)
	}
	return space, nil
}

// Append entry to artifact list
func (am *ArtifactMap) Append(resourceType string, entry *ArtifactEntry) {
	am.allocateResourceType(resourceType, 1)
	am.resourceMap[resourceType].list = append(am.resourceMap[resourceType].list, entry)
	am.resourceMap[resourceType].sorted = false
}

// Sort entries of resource type
func (am *ArtifactMap) Sort(resourceType string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.resourceMap[resourceType].Len() < 1 {
		return
	}

	if !am.resourceMap[resourceType].sorted {
		sort.Sort(am.resourceMap[resourceType])
	}
}

// FillArtifacts puts the app name (key) and the app ID (value) in the ArtifactMap
func (am *ArtifactMap) FillArtifacts(item *ItemData) error {
	if item == nil || item.Data == nil {
		return errors.New("empty ItemData struct")
	}

	if len(item.Data) == 0 {
		return nil
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	for _, data := range item.Data {
		am.Append(data.ResourceType, data.Copy())
	}

	return nil
}

// DeleteApp deletes an app from the ArtifactMap
func (am *ArtifactMap) DeleteApp(appGUID string) {
	am.DeleteItem(ResourceTypeApp, appGUID, true)
}

// DeleteItemUsingID with resource type and ID
func (am *ArtifactMap) DeleteItemUsingID(id, resourceType string) {
	am.DeleteItem(resourceType, id, true)
}

// DeleteItem from artifact map
func (am *ArtifactMap) DeleteItem(resourceType, lookfor string, id bool) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.resourceMap[resourceType].Len() < 1 {
		return
	}

	compare := getCompareFunc(id)
	for index, entry := range am.resourceMap[resourceType].list {
		if compare(entry, lookfor) {
			am.resourceMap[resourceType].list = append(am.resourceMap[resourceType].list[:index], am.resourceMap[resourceType].list[index+1:]...)
			return
		}
	}
}

// DeleteStream deletes a stream from the ArtifactMap
func (am *ArtifactMap) DeleteStream(streamName string) {
	am.DeleteItem(ResourceTypeStream, streamName, false)
}

// ClearArtifactMap Empty Apps from ArtifactMap
func (am *ArtifactMap) ClearArtifactMap() {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.resourceMap = make(map[string]*ArtifactList)
}

// FillAppsUsingDocListEntries should be used to fillAppMap app map in QSEoW
func (am *ArtifactMap) FillAppsUsingDocListEntries(docListEntries []*enigma.DocListEntry) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	for _, docListEntry := range docListEntries {
		am.Append(ResourceTypeApp, &ArtifactEntry{docListEntry.DocName, docListEntry.DocId, "" /* QSEoW does not have item ID */, ResourceTypeApp, nil})
	}
	return nil
}

// allocateResourceType ArtifactMap should be locked before calling this function
func (am *ArtifactMap) allocateResourceType(resourceType string, len int) {
	if am.resourceMap == nil {
		am.resourceMap = make(map[string]*ArtifactList)
	}

	if am.resourceMap[resourceType] == nil {
		am.resourceMap[resourceType] = &ArtifactList{list: make([]*ArtifactEntry, 0, len)}
	}
}

// FillStreams fills the stream map with the streams from the given list
func (am *ArtifactMap) FillStreams(streamList []elasticstructs.Collection) {
	am.mu.Lock()
	defer am.mu.Unlock()

	for _, stream := range streamList {
		am.Append(ResourceTypeStream, &ArtifactEntry{Name: stream.Name, ID: stream.ID, ResourceType: ResourceTypeStream})
	}
}

// FillSpaces fills the spaces map with the spaces from the given list
func (am *ArtifactMap) FillSpaces(spaces []elasticstructs.Space) {
	am.mu.Lock()
	defer am.mu.Unlock()

	for _, space := range spaces {
		am.Append(ResourceTypeSpace, &ArtifactEntry{Name: space.Name, ID: space.ID, ResourceType: ResourceTypeSpace, Data: &space})
	}
}

// AddSpace to artifact map
func (am *ArtifactMap) AddSpace(space elasticstructs.Space) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.Append(ResourceTypeSpace, &ArtifactEntry{Name: space.Name, ID: space.ID, ResourceType: ResourceTypeSpace, Data: &space})
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

// GetAppItemID returns the item ID given the app Title. When multiple apps
// have the same Title, this will return the ID of the last app in the order
// of the struct passed to the Fill function.
func (am *ArtifactMap) GetAppItemID(appName string) (string, error) {
	entry, err := am.getAppEntry(appName)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return entry.ItemID, nil
}

func (am *ArtifactMap) getAppEntry(appName string) (*ArtifactEntry, error) {
	entry, err := am.lookup(ResourceTypeApp, appName, false)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return entry, nil
}

// GetStreamID returns the app ID given the stream
func (am *ArtifactMap) GetStreamID(streamName string) (string, error) {
	entry, err := am.lookup(ResourceTypeStream, streamName, false)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return entry.ID, nil
}

// GetSpaceID returns the space ID given space name
func (am *ArtifactMap) GetSpaceID(spaceName string) (string, error) {
	entry, err := am.lookup(ResourceTypeSpace, spaceName, false)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return entry.ID, nil
}

// GetSpaceByName return first found space with given name
func (am *ArtifactMap) GetSpaceByName(spaceName string) (*elasticstructs.Space, error) {
	entry, err := am.lookup(ResourceTypeSpace, spaceName, false)
	if err != nil {
		return nil, SpaceNameNotFoundError(spaceName)
	}

	space, err := entry.DataAsSpace()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return space, nil
}

// GetSpaceByID returns the space given space ID
func (am *ArtifactMap) GetSpaceByID(spaceID string) (*elasticstructs.Space, error) {
	entry, err := am.lookup(ResourceTypeSpace, spaceID, true)
	if err != nil {
		return nil, SpaceIDNotFoundError(spaceID)
	}

	space, err := entry.DataAsSpace()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return space, nil
}

// GetRandomApp returns a random app for the map, chosen by a uniform distribution
func (am *ArtifactMap) GetRandomApp(sessionState *State) (ArtifactEntry, error) {
	am.Sort(ResourceTypeApp)

	am.mu.RLock()
	defer am.mu.RUnlock()

	n := am.resourceMap[ResourceTypeApp].Len()
	if n < 1 {
		return ArtifactEntry{}, errors.New("cannot select random app: ArtifactMap is empty")
	}

	return *am.resourceMap[ResourceTypeApp].list[sessionState.Randomizer().Rand(n)], nil
}

// GetRoundRobin returns a app round robin for the map
func (am *ArtifactMap) GetRoundRobin(sessionState *State) (ArtifactEntry, error) {
	am.Sort(ResourceTypeApp)

	am.mu.RLock()
	defer am.mu.RUnlock()

	appNumber := sessionState.Counters.AppCounter.Inc() - 1
	n := am.resourceMap[ResourceTypeApp].Len()
	if n < 1 {
		return ArtifactEntry{}, errors.New("cannot select app round robin: ArtifactMap is empty")
	}

	return *am.resourceMap[ResourceTypeApp].list[appNumber%uint64(n)], nil
}

// LookupAppTitle lookup app using title
func (am *ArtifactMap) LookupAppTitle(title string) (*ArtifactEntry, error) {
	return am.lookup(ResourceTypeApp, title, false)
}

// LookupAppGUID lookup app using GUID
func (am *ArtifactMap) LookupAppGUID(guid string) (*ArtifactEntry, error) {
	entry, err := am.lookup(ResourceTypeApp, guid, true)
	if err != nil {
		// GUID not found in map, create new entry with GUID (Supports using openapp with GUID and no preceeding OpenHub)
		entry = &ArtifactEntry{ID: guid, ResourceType: ResourceTypeApp} // todo how to handle itemID?
	}
	return entry, nil
}

func (am *ArtifactMap) lookup(resourcetype, lookup string, id bool) (*ArtifactEntry, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	compare := getCompareFunc(id)

	if am.resourceMap[resourcetype] != nil {
		for _, entry := range am.resourceMap[resourcetype].list {
			if compare(entry, lookup) {
				return entry.Copy(), nil
			}
		}
	}

	return nil, errors.Errorf("item type<%s> id<%s> not found in artifact map", resourcetype, lookup)
}

func getCompareFunc(id bool) func(*ArtifactEntry, string) bool {
	if id {
		return compareID
	}
	return compareName
}

func compareName(entry *ArtifactEntry, name string) bool {
	if entry == nil {
		return false
	}
	return entry.Name == name
}

func compareID(entry *ArtifactEntry, id string) bool {
	if entry == nil {
		return false
	}
	return entry.ID == id
}

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
	am.mu.RLock()
	defer am.mu.RUnlock()

	return jsonit.Marshal(am.resourceMap)
}

// Count of specified resource type list
func (am *ArtifactMap) Count(resourceType string) int {
	am.mu.RLock()
	defer am.mu.RUnlock()

	return am.resourceMap[resourceType].Len()
}

// First entry of resource type
func (am *ArtifactMap) First(resourceType string) *ArtifactEntry {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if am.resourceMap[resourceType].Len() > 0 {
		return am.resourceMap[resourceType].list[0].Copy()
	}

	return nil
}
