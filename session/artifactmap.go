package session

import (
	"sort"
	"sync"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	enigma "github.com/qlik-oss/enigma-go/v3"
	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/structs"
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
		NonEphemeralResourceTypes []string

		resourceMap map[string]*ArtifactList
		mu          sync.RWMutex
	}

	// ArtifactKey used to find item using title
	ArtifactKey struct {
		Title string
		Type  string
	}

	ArtifactEntryCompareType int
)

// Currently known resource types
const (
	ResourceTypeApp         = "app"
	ResourceTypeGenericLink = "genericlink"
	ResourceTypeDataset     = "dataset"
	ResourceTypeDataAsset   = "dataasset"
)

const (
	ArtifactEntryCompareTypeID ArtifactEntryCompareType = iota
	ArtifactEntryCompareTypeItemID
	ArtifactEntryCompareTypeName
)

func (d *ArtifactList) Len() int {
	if d == nil {
		return 0
	}
	return len(d.list)
}

func (d ArtifactList) Less(i, j int) bool {
	return d.list[i].ID < d.list[j].ID
}

func (d ArtifactList) Swap(i, j int) {
	d.list[i], d.list[j] = d.list[j], d.list[i]
}

// MarshalJSON marshal artifact list to JSON
func (d ArtifactList) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.list)
}

// NewArtifactMap returns an empty ArtifactMap
func NewArtifactMap() *ArtifactMap {
	return &ArtifactMap{
		NonEphemeralResourceTypes: []string{},
		resourceMap:               make(map[string]*ArtifactList),
	}
}

// Copy of ArtifactEntry
func (entry *ArtifactEntry) Copy() *ArtifactEntry {
	if entry == nil {
		return nil
	}

	cpy := *entry
	return &cpy
}

// Append locks ArtifactMap and appends entry to artifact list
func (am *ArtifactMap) Append(resourceType string, entry *ArtifactEntry) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.allocateResourceType(resourceType, 1)
	am.resourceMap[resourceType].list = append(am.resourceMap[resourceType].list, entry)
	am.resourceMap[resourceType].sorted = false
}

// Sort entries of resource type
func (am *ArtifactMap) Sort(resourceType string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.sortNoLock(resourceType)
}

func (am *ArtifactMap) sortNoLock(resourceType string) {
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

	for _, data := range item.Data {
		am.Append(data.ResourceType, data.Copy())
	}

	return nil
}

// DeleteApp deletes an app from the ArtifactMap
func (am *ArtifactMap) DeleteApp(appGUID string) {
	am.DeleteItem(ResourceTypeApp, appGUID, ArtifactEntryCompareTypeID)
}

// DeleteItemUsingID with resource type and ID
func (am *ArtifactMap) DeleteItemUsingID(id, resourceType string) {
	am.DeleteItem(resourceType, id, ArtifactEntryCompareTypeID)
}

// DeleteItemUsingItemID with resource type and item ID
func (am *ArtifactMap) DeleteItemUsingItemID(id, resourceType string) {
	am.DeleteItem(resourceType, id, ArtifactEntryCompareTypeItemID)
}

// DeleteItem from artifact map
func (am *ArtifactMap) DeleteItem(resourceType, lookfor string, compareType ArtifactEntryCompareType) {
	am.mu.Lock()
	defer am.mu.Unlock()

	if am.resourceMap[resourceType].Len() < 1 {
		return
	}

	compare := getCompareFunc(compareType)
	for index, entry := range am.resourceMap[resourceType].list {
		if compare(entry, lookfor) {
			am.resourceMap[resourceType].list = append(am.resourceMap[resourceType].list[:index], am.resourceMap[resourceType].list[index+1:]...)
			return
		}
	}
}

// ClearArtifactMap Empty Apps from ArtifactMap
func (am *ArtifactMap) ClearArtifactMap() {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Clear everything ephemeral from map
	for resource := range am.resourceMap {
		for _, typ := range am.NonEphemeralResourceTypes {
			if typ == resource {
				continue
			}
		}
		delete(am.resourceMap, resource)
	}
}

// FillAppsUsingDocListEntries should be used to fillAppMap app map in QSEoW
func (am *ArtifactMap) FillAppsUsingDocListEntries(docListEntries []*enigma.DocListEntry) error {
	for _, docListEntry := range docListEntries {
		am.Append(ResourceTypeApp, &ArtifactEntry{docListEntry.DocName, docListEntry.DocId, "" /* QSEoW does not have item ID */, ResourceTypeApp, nil})
	}
	return nil
}

// FillAppsUsingStream
func (am *ArtifactMap) FillAppsUsingStream(stream structs.Stream) error {
	for _, streamData := range stream.Data {
		if streamData.Type == structs.StreamTypeApp {
			am.Append(ResourceTypeApp, &ArtifactEntry{streamData.Attributes.Name, streamData.ID, "", ResourceTypeApp, nil})
		}
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
	entry, err := am.Lookup(ResourceTypeApp, appName, ArtifactEntryCompareTypeName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return entry, nil
}

// GetRandomApp returns a random app for the map, chosen by a uniform distribution
func (am *ArtifactMap) GetRandomApp(sessionState *State) (ArtifactEntry, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.sortNoLock(ResourceTypeApp)

	n := am.resourceMap[ResourceTypeApp].Len()
	if n < 1 {
		return ArtifactEntry{}, errors.New("cannot select random app: ArtifactMap is empty")
	}

	return *am.resourceMap[ResourceTypeApp].list[sessionState.Randomizer().Rand(n)], nil
}

// GetRoundRobin returns a app round robin for the map
func (am *ArtifactMap) GetRoundRobin(sessionState *State) (ArtifactEntry, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.sortNoLock(ResourceTypeApp)

	appNumber := sessionState.Counters.AppCounter.Inc() - 1
	n := am.resourceMap[ResourceTypeApp].Len()
	if n < 1 {
		return ArtifactEntry{}, errors.New("cannot select app round robin: ArtifactMap is empty")
	}

	return *am.resourceMap[ResourceTypeApp].list[appNumber%uint64(n)], nil
}

// LookupAppTitle lookup app using title
func (am *ArtifactMap) LookupAppTitle(title string) (*ArtifactEntry, error) {
	return am.Lookup(ResourceTypeApp, title, ArtifactEntryCompareTypeName)
}

// LookupItemID lookup resource using Item ID
func (am *ArtifactMap) LookupItemID(resourcetype, itemID string) (*ArtifactEntry, error) {
	return am.Lookup(resourcetype, itemID, ArtifactEntryCompareTypeItemID)
}

// LookupAppGUID lookup app using GUID
func (am *ArtifactMap) LookupAppGUID(guid string) (*ArtifactEntry, error) {
	entry, err := am.Lookup(ResourceTypeApp, guid, ArtifactEntryCompareTypeID)
	if err != nil {
		// GUID not found in map, create new entry with GUID (Supports using openapp with GUID and no preceeding OpenHub)
		entry = &ArtifactEntry{ID: guid, ResourceType: ResourceTypeApp} // todo how to handle itemID?
	}
	return entry, nil
}

// Lookup resource type with using lookup (name or ID, defined by id flag)
func (am *ArtifactMap) Lookup(resourcetype, lookup string, compareType ArtifactEntryCompareType) (*ArtifactEntry, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	compare := getCompareFunc(compareType)

	if am.resourceMap[resourcetype] != nil {
		for _, entry := range am.resourceMap[resourcetype].list {
			if compare(entry, lookup) {
				return entry.Copy(), nil
			}
		}
	}

	return nil, errors.Errorf("item type<%s> id<%s> not found in artifact map", resourcetype, lookup)
}

func getCompareFunc(compareType ArtifactEntryCompareType) func(*ArtifactEntry, string) bool {
	switch compareType {
	case ArtifactEntryCompareTypeID:
		return compareID
	case ArtifactEntryCompareTypeName:
		return compareName
	case ArtifactEntryCompareTypeItemID:
		return compareItemID
	}
	return nil
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

func compareItemID(entry *ArtifactEntry, id string) bool {
	if entry == nil {
		return false
	}
	return entry.ItemID == id
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

	return json.Marshal(am.resourceMap)
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
