package session

import (
	"github.com/qlik-oss/gopherciser/randomizer"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	emptyAppMap  = &ArtifactMap{}
	emptyAppData = &AppData{}

	someAppTitle     = "my-app-title"
	someAppID        = "my-app-id"
	someAppDataTitle = &AppData{
		Data: []AppsResp{
			{
				Title: someAppTitle,
				ID:    someAppID,
			},
			{
				Title: someAppTitle + "2",
				ID:    someAppID + "2",
			},
		},
	}

	someAppDataName = &AppData{
		Data: []AppsResp{
			{
				Name: someAppTitle,
				ID:   someAppID,
			},
		},
	}

	duplicateKeysApp = &AppData{
		Data: []AppsResp{
			{
				Title: someAppTitle,
				ID:    someAppID,
			},
			{
				Title: someAppTitle,
				ID:    someAppID + "2",
			},
		},
	}

	dummyState = &State{
		rand: &rand{
			mu: sync.Mutex{},
			rnd: &DefaultRandomizer{
				sync.Mutex{},
				randomizer.NewRandomizer(),
			},
		},
	}
)

func TestNewAppMap(t *testing.T) {
	am := NewAppMap()
	assert.NotNil(t, am)
	assert.IsType(t, emptyAppMap, am)
}

func TestAppMap_fill_name(t *testing.T) {
	am := NewAppMap()
	err := am.fillAppMap(someAppDataName, "name")
	assert.NoError(t, err)
	assert.NotNil(t, am.appTitleToID)
}

func TestAppMap_fill_title(t *testing.T) {
	am := NewAppMap()
	err := am.fillAppMap(someAppDataTitle, "Title")
	assert.NoError(t, err)
	assert.NotNil(t, am.appTitleToID)
}

func TestAppMap_fill_wrong(t *testing.T) {
	am := NewAppMap()
	err := am.fillAppMap(someAppDataName, "wrong")
	assert.Error(t, err)
}

func TestAppMap_FillWithName(t *testing.T) {
	am := NewAppMap()
	err := am.FillAppsUsingName(someAppDataName)
	assert.NoError(t, err)
	assert.NotNil(t, am.appTitleToID)
}

func TestAppMap_FillWithName_emptyAppData(t *testing.T) {
	am := NewAppMap()
	err := am.FillAppsUsingName(emptyAppData)
	assert.Error(t, err)
}

func TestAppMap_FillWithName_nil(t *testing.T) {
	am := NewAppMap()
	err := am.FillAppsUsingName(nil)
	assert.Error(t, err)
}

func TestAppMap_FillWithTitle(t *testing.T) {
	am := NewAppMap()
	err := am.FillAppsUsingTitle(someAppDataTitle)
	assert.NoError(t, err)
	assert.NotNil(t, am.appTitleToID)
}

func TestAppMap_FillWithTitle_emptyAppData(t *testing.T) {
	am := NewAppMap()
	err := am.FillAppsUsingTitle(emptyAppData)
	assert.Error(t, err)
}

func TestAppMap_FillWithTitle_nil(t *testing.T) {
	am := NewAppMap()
	err := am.FillAppsUsingTitle(nil)
	assert.Error(t, err)
}

func TestAppMap_GetAppID(t *testing.T) {
	am := NewAppMap()
	err := am.FillAppsUsingTitle(someAppDataTitle)
	assert.NoError(t, err)
	appID, err := am.GetAppID(someAppTitle)
	assert.NoError(t, err)
	assert.Equal(t, someAppID, appID)
}

func TestAppMap_GetAppID_notFound(t *testing.T) {
	am := NewAppMap()
	err := am.FillAppsUsingTitle(someAppDataTitle)
	assert.NoError(t, err)
	_, err = am.GetAppID("not-to-be-found")
	assert.Error(t, err)
}

func TestAppMap_GetAppID_duplicateKeys(t *testing.T) {
	// When 2 or more apps have the same Title, the
	// get should return the last of them
	am := NewAppMap()
	err := am.FillAppsUsingTitle(duplicateKeysApp)
	assert.NoError(t, err)
	appID1, err1 := am.GetAppID(someAppTitle)
	assert.NoError(t, err1)
	assert.Equal(t, someAppID+"2", appID1)
}

func TestAppMap_GetAppID_concurrent(t *testing.T) {
	am := NewAppMap()
	err := am.FillAppsUsingTitle(someAppDataTitle)
	assert.NoError(t, err)

	wg := sync.WaitGroup{}
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			appID, err := am.GetAppID(someAppTitle)
			assert.NoError(t, err)
			assert.Equal(t, someAppID, appID)
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestAppMap_GetRandomAppID(t *testing.T) {
	am := NewAppMap()
	err := am.FillAppsUsingTitle(someAppDataTitle)
	assert.NoError(t, err)
	_, err = am.GetRandomApp(dummyState)
	assert.NoError(t, err)
}

func TestAppMap_GetRandomAppID_noApps(t *testing.T) {
	am := NewAppMap()
	_, err := am.GetRandomApp(dummyState)
	assert.Error(t, err)
}
