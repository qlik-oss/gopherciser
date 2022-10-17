package enigmahandlers

import (
	"context"
	"strconv"
	"testing"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/requestmetrics"
)

func TestConnection(t *testing.T) {
	connection := NewSenseUplink(context.Background(), nil, &requestmetrics.RequestMetrics{}, nil)
	connection.MockMode = true

	if err := connection.Connect(context.Background(), "wss://localhost", nil, nil, false, 0, false); err != nil {
		t.Error(err)
	}
}

func TestObjectList(t *testing.T) {
	connection := &SenseUplink{}

	if _, err := connection.Objects.GetObject(1); err != ObjectNotFound(1) {
		t.Errorf("Unexpected error when getting object from object list. Expected<%s> got<%s>",
			ObjectNotFound(1).Error(), err.Error())
	}

	obj := &Object{
		Handle: 0,
	}
	if err := connection.Objects.AddObject(obj); err != nil {
		t.Fatalf("Failed to add obj<0> to object list: %v", err)
	}

	if _, err := connection.Objects.GetObject(1); err != ObjectNotFound(1) {
		t.Errorf("Unexpected error when getting object from object list. Expected<%s> got<%s>",
			ObjectNotFound(1).Error(), err.Error())
	}

	obj = &Object{
		Handle: 1,
		ID:     "obj1",
	}
	if err := connection.Objects.AddObject(obj); err != nil {
		t.Fatalf("Failed to add obj<1> to object list: %v", err)
	}

	var err error
	if obj, err = connection.Objects.GetObject(1); err != nil {
		t.Errorf("Unexpected error when getting object from object list: %v", err)
	} else {
		if obj.Handle != 1 {
			t.Errorf("Unexpected object handle, expected<1> got<%d>", obj.Handle)
		}
	}

	if obj, err = connection.Objects.GetObjectByID("obj1"); err != nil {
		t.Errorf("Unexpected error when getting object from object list: %v", err)
	} else {
		if obj.Handle != 1 {
			t.Errorf("GetObjectByID: Unexpected object, expected<1> got<%d>", obj.Handle)
		}
	}

	if err := connection.Objects.AddObject(obj); err != nil {
		t.Fatalf("Failed to re-add obj<1> to object list: %v", err)
	}

	expectedError := "Object<not empty> not found in object list"
	if _, err := connection.Objects.GetObjectByID("not empty"); err == nil {
		t.Errorf("Expected error<%s>, got ok", expectedError)
	} else if expectedError != errors.Cause(err).Error() {
		t.Errorf("Unexpected error. Expected<%s> got<%s>", expectedError, errors.Cause(err).Error())
	}

	err = connection.Objects.RemoveObject(0)
	if err != nil {
		t.Errorf("Got unexpected error when removing object: %v", err)
	}
}

func TestObjectListMulti(t *testing.T) {
	connection := &SenseUplink{}

	objectCount := 10

	// Add
	for i := 0; i < objectCount; i++ {
		err := connection.Objects.AddObject(&Object{
			Handle: i,
			ID:     strconv.Itoa(i),
		})
		if err != nil {
			t.Fatalf("Failed adding object<%d>, err: %v", i, err)
		}
	}

	// Get
	for i := 0; i < objectCount; i++ {
		if _, err := connection.Objects.GetObject(i); err != nil {
			t.Fatalf("Failed getting handle<%d>, err: %v", i, err)
		}
	}

	// Get again
	for i := 0; i < objectCount; i++ {
		if _, err := connection.Objects.GetObject(i); err != nil {
			t.Fatalf("Failed getting handle<%d>, err: %v", i, err)
		}
	}

	for _, v := range connection.Objects.objects {
		t.Logf("object: %+v", v)
	}

	// GetByID
	for i := 0; i < objectCount; i++ {
		if _, err := connection.Objects.GetObjectByID(strconv.Itoa(i)); err != nil {
			t.Fatalf("Failed getting byID object<%d>, err: %v", i, err)
		}
	}

	//Remove
	for i := 0; i < objectCount; i++ {
		if err := connection.Objects.RemoveObject(i); err != nil {
			t.Errorf("Got unexpected error when removing object: %v", err)
		}

	}
}

//TODO test App
//TODO test Object
//TODO test ResponseTime

func BenchmarkObjectListAdd(b *testing.B) {
	connection := &SenseUplink{}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = connection.Objects.AddObject(&Object{
			Handle: n,
			ID:     strconv.Itoa(n),
		})
	}
}

func BenchmarkObjectListDelete(b *testing.B) {
	connection := &SenseUplink{}
	for n := 0; n < b.N; n++ {
		_ = connection.Objects.AddObject(&Object{
			Handle: n,
			ID:     strconv.Itoa(n),
		})
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = connection.Objects.RemoveObject(n)
	}
}

func BenchmarkObjectListGet(b *testing.B) {
	connection := &SenseConnection{}
	for n := 0; n < b.N; n++ {
		_ = connection.Objects.AddObject(&Object{
			Handle: n,
			ID:     strconv.Itoa(n),
		})
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = connection.Objects.GetObject(n)
	}
}

func BenchmarkObjectListGetByID(b *testing.B) {
	connection := &SenseConnection{}
	for n := 0; n < b.N; n++ {
		_ = connection.Objects.AddObject(&Object{
			Handle: n,
			ID:     strconv.Itoa(n),
		})
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = connection.Objects.GetObjectByID(strconv.Itoa(n))
	}
}
