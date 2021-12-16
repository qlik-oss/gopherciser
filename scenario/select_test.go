package scenario

import (
	"sort"
	"testing"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/randomizer"
)

type Rnd struct {
	*randomizer.Randomizer
}

func (rnd *Rnd) Reset(instance, session uint64, onlyinstanceSeed bool) {
	rnd.Randomizer = randomizer.NewSeededRandomizer(randomizer.GetPredictableSeed(int(instance), int(session)))
}

func TestSelectUnmarshal(t *testing.T) {
	t.Parallel()

	raw := `{
		"label" : "select from all",
		"action" : "select",
		"settings": {
			"id" : "objid1",
			"type" : "randomfromall",
			"accept" : true,
			"wrap" : false,
			"min" : 2,
			"max" : 5,
			"dim" : 2
		}
	}`
	var item Action
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		t.Fatal(err)
	}

	if _, err := item.Validate(); err != nil {
		t.Error(err)
	}

	validateString(t, "action", item.Type, "select")
	validateString(t, "label", item.Label, "select from all")

	settings, ok := item.Settings.(*SelectionSettings)
	if !ok {
		t.Fatalf("Failed to cast item settings<%T> to *SelectionSettings", item.Settings)
	}

	validateString(t, "ID", settings.ID, "objid1")
	typeString, err := settings.Type.GetEnumMap().String(int(settings.Type))
	if err != nil {
		t.Fatalf("Failed to cast selection type to string %T:%v", settings.Type, settings.Type)
	}
	validateString(t, "type", typeString, "randomfromall")

	validateBool(t, "accept", settings.Accept, true)
	validateBool(t, "wrap", settings.WrapSelections, false)

	validateInt(t, "min", settings.Min, 2)
	validateInt(t, "max", settings.Max, 5)
	validateInt(t, "dim", settings.Dimension, 2)

}

func TestSelectMarshal(t *testing.T) {
	t.Parallel()

	item := Action{
		ActionCore{
			Type:  ActionSelect,
			Label: "select from enabled",
		},
		&SelectionSettings{
			ID:             "",
			Type:           RandomFromEnabled,
			Accept:         false,
			WrapSelections: true,
			Min:            4,
			Max:            6,
			Dimension:      1,
		},
	}

	raw, err := json.Marshal(item)
	if err != nil {
		t.Fatal(err)
	}

	validateString(t, "json", string(raw), `{"action":"select","label":"select from enabled","disabled":false,"settings":{"id":"","type":"randomfromenabled","accept":false,"wrap":true,"min":4,"max":6,"dim":1,"values":null}}`)
}

func TestValidate(t *testing.T) {
	t.Parallel()

	settings := &SelectionSettings{}

	_, err := settings.Validate()
	validateError(t, err, "Empty object ID")
	settings.ID = "obj1"
	settings.Dimension = -1

	_, err = settings.Validate()
	validateError(t, err, "Illegal dimension<-1>")
	settings.Dimension = 0

	_, err = settings.Validate()
	validateError(t, err, "min<0> selections must be >1")
	settings.Min = 2

	_, err = settings.Validate()
	validateError(t, err, "max<0> selections must be >1")
	settings.Max = 1

	_, err = settings.Validate()
	validateError(t, err, "min<2> must be less than max<1>")
	settings.Max = 3

	_, err = settings.Validate()
	validateError(t, err, "")
}

func TestSelectQty(t *testing.T) {
	t.Parallel()

	rnd := &Rnd{}
	rnd.Reset(1, 2, false)

	v := getSelectQty(2, 5, 5, rnd)
	validateInt(t, "selectqty", v, 4)

	result := make(map[int]int)

	for i := 0; i < 1000000; i++ {
		v = getSelectQty(2, 5, 5, rnd)
		result[v]++
	}

	expectedKeys := make(map[int]int, 4)
	expectedKeys[2] = 250303
	expectedKeys[3] = 249072
	expectedKeys[4] = 250422
	expectedKeys[5] = 250203

	if len(result) > len(expectedKeys) {
		t.Fatalf("Unexpected length<%d> of result expected<%d> map<%+v>", len(result), len(expectedKeys), result)
	}

	for k, v := range result {
		if v != expectedKeys[k] {
			t.Fatalf("unexpected value<%d> for key<%d> expected<%d> map<%+v>", v, k, expectedKeys[k], result)
		}
	}
}

func TestCutSlice(t *testing.T) {
	t.Parallel()

	var possible []int

	validateError(t, cutPosition(1, &possible), "empty slice")

	possible = []int{}
	validateError(t, cutPosition(1, &possible), "empty slice")

	possible = []int{0, 1, 2, 3, 4}
	validateError(t, cutPosition(5, &possible), "index out of bounds")

	if err := cutPosition(1, &possible); err != nil {
		t.Fatal(err)
	}
	validateIntArray(t, "cutPossible", possible, []int{0, 2, 3, 4})

	if err := cutPosition(0, &possible); err != nil {
		t.Fatal(err)
	}
	validateIntArray(t, "cutPossible", possible, []int{2, 3, 4})

	if err := cutPosition(2, &possible); err != nil {
		t.Fatal(err)
	}
	validateIntArray(t, "cutPossible", possible, []int{2, 3})
}

func TestFillPos(t *testing.T) {
	t.Parallel()

	rnd := &Rnd{}
	rnd.Reset(1, 3245345, false)

	selectPos, err := fillSelectPosFromAll(3, 6, 5, rnd)
	if err != nil {
		t.Fatal(err)
	}
	validateIntArraySortable(t, "selectpos1", selectPos, []int{0, 1, 3, 4}, true)
	selectPos, err = fillSelectPosFromAll(3, 6, 50, rnd)
	if err != nil {
		t.Fatal(err)
	}
	validateIntArraySortable(t, "selectpos2", selectPos, []int{39, 4, 3, 27, 15, 26}, true)
	selectPos, err = fillSelectPosFromPossible(2, 3, []int{3}, rnd)
	if err != nil {
		t.Fatal(err)
	}
	validateIntArray(t, "selectpos3", selectPos, []int{3})

	selectPos, err = fillSelectPosFromPossible(2, 3, []int{0, 1, 2, 3, 4, 6, 7, 9, 10, 12, 56}, rnd)
	if err != nil {
		t.Fatal(err)
	}
	validateIntArraySortable(t, "selectpos4", selectPos, []int{4, 9, 56}, true)
}

func validateString(t *testing.T, key, value, expected string) {
	t.Helper()
	if value != expected {
		t.Errorf("Unexpected %s<%s> expected<%s>", key, value, expected)
	}
}

func validateInt(t *testing.T, key string, value, expected int) {
	t.Helper()
	if value != expected {
		t.Errorf("Unexpected %s<%d> expected<%d>", key, value, expected)
	}
}

func validateBool(t *testing.T, key string, value, expected bool) {
	t.Helper()
	if value != expected {
		t.Errorf("Unexpected %s<%v> expected<%v>", key, value, expected)
	}
}

func validateError(t *testing.T, err error, expected string) {
	t.Helper()
	cause := errors.Cause(err)

	if err == nil {
		if expected == "" {
			return
		}
		t.Fatalf("Expected error<%s> got<nil>", expected)
	}

	errString := cause.Error()
	if errString != expected {
		t.Fatalf("Expected error:\n%s\ngot:\n%s\n", expected, errString)
	}
}

func validateIntArray(t *testing.T, key string, value, expected []int) {
	t.Helper()

	validateIntArraySortable(t, key, value, expected, false)
}

func validateIntArraySortable(t *testing.T, key string, value, expected []int, doSort bool) {
	t.Helper()

	if len(value) != len(expected) {
		t.Errorf("%s: Unexpected array length<%d> expected<%d> array<%+v>", key, len(value), len(expected), value)
		return
	}

	if doSort { //Order not important
		sort.Ints(value)
		sort.Ints(expected)
	}

	for i, v := range value {
		if v != expected[i] {
			t.Errorf("%s: Unexpected value<%d> pos<%d> expected<%d> array<%+v>", key, v, i, expected[i], value)
			return
		}
	}
}
