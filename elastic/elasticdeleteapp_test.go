package elastic

import (
	"testing"

	"github.com/qlik-oss/gopherciser/scenario"
)

func TestElasticDeleteAppSettings(t *testing.T) {
	raw := []byte(`{
			"label": "clear collection",
			"action": "ElasticDeleteApp",
			"settings": {
				"mode" : "clearcollection",
				"collectionname" : "mycollectioname"
			}
}`)
	var item scenario.Action

	if err := jsonit.Unmarshal(raw, &item); err != nil {
		t.Fatal(err)
	}

	if err := item.Validate(); err != nil {
		t.Fatal(err)
	}

	settings, ok := item.Settings.(*ElasticDeleteAppSettings)
	if !ok {
		t.Fatalf("could not cast settings of type %T to ElasticDeleteAppSettings, %+v", item.Settings, item)
	}

	if settings.DeletionMode != ClearCollection {
		str, err := deletionModeEnumMap.String(int(settings.DeletionMode))
		if err != nil {
			t.Errorf("expected deletion mode ClearCollection got %s", str)
		} else {
			t.Errorf("expected deletion mode ClearCollection got %d", settings.DeletionMode)
		}
	}

	if settings.CollectionName != "mycollectioname" {
		t.Errorf("Expected collection name mycollectioname, got %s", settings.CollectionName)
	}

	j, err := jsonit.Marshal(settings)
	if err != nil {
		t.Fatal("error marshaling settings:", err)
	}

	expected := `{"appmode":"current","app":"","filename":"","mode":"clearcollection","collectionname":"mycollectioname"}`
	if string(j) != expected {
		t.Errorf("unexpected marshal:\n%s\nexpected:\n%s\n", j, expected)
	}
}

func TestElasticDeleteAppSettingsDeprecated(t *testing.T) {
	raw := []byte(`{
			"label": "clear collection",
			"action": "ElasticDeleteApp",
			"settings": {
				"appguid" : "myguid"
			}
}`)
	var item scenario.Action
	if err := jsonit.Unmarshal(raw, &item); err == nil {
		t.Error("Expected unmarshal error when using appguid setting")
	}
}
