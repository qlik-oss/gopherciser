package scenario

import (
	"testing"
)

func TestStaticSelectionType(t *testing.T) {
	raw := "0"
	var st StaticSelectionType
	if err := jsonit.Unmarshal([]byte(raw), &st); err != nil {
		t.Fatalf("Failed to unmarsal json<%s> to SelectionType, err: %v", raw, err)
	}

	if value, err := jsonit.Marshal(&st); err != nil {
		t.Fatalf("Failed to marshal SelectionType<%v> to json, err: %v", st, err)
	} else if string(value) != `"hypercubecells"` {
		t.Fatalf(`Marshal SelectionType<%v> to json yielded unexpected value<%s> expected<"hypercubecells">`, st, string(value))
	}

	raw = `"HyperCubeCells"`
	if err := jsonit.Unmarshal([]byte(raw), &st); err != nil {
		t.Fatalf("Failed to unmarsal json<%s> to SelectionType, err: %v", raw, err)
	}

	raw = `{
		"type" : "HyperCubeCells",
		"path" : "/qHyperCubeDef",
		"rows" : [4,5],
		"cols" : [0],
		"accept" : true
	}`

	var settings StaticSelectSettings
	if err := jsonit.Unmarshal([]byte(raw), &settings); err != nil {
		t.Fatal("Unmarshal SelectionSettings failed, err:", err)
	}
	if settings.Type != HyperCubeCells {
		t.Errorf("Expected<HyperCubeCells> got<%v>", settings.Type)
	}
	if settings.Path != "/qHyperCubeDef" {
		t.Errorf("Expected</qHyperCubeDef> got<%s>", settings.Path)
	}
	if len(settings.Rows) != 2 {
		t.Errorf("Expected rows len<2> got <%d>", len(settings.Rows))
	} else {
		if settings.Rows[0] != 4 {
			t.Errorf("row 0 Expected<4> got<%d>", settings.Rows[0])
		}
		if settings.Rows[1] != 5 {
			t.Errorf("row 1 Expected<5> got<%d>", settings.Rows[1])
		}
	}
	if len(settings.Cols) != 1 {
		t.Errorf("Expected cols len<1> got<%d>", len(settings.Cols))
	} else {
		if settings.Cols[0] != 0 {
			t.Errorf("col 0 Expected<0> got<%d>", settings.Cols[0])
		}
	}
	if !settings.Accept {
		t.Error("Expected accepted<true> got accepted<false>")
	}

	j, err := jsonit.Marshal(settings)
	if err != nil {
		t.Errorf("Failed to marshal settings, err: %v", err)
	}
	if string(j) != `{"id":"","path":"/qHyperCubeDef","rows":[4,5],"cols":[0],"type":"hypercubecells","accept":true,"wrap":false}` {
		t.Errorf("unexpected result json:\n%s", string(j))
	}
}
