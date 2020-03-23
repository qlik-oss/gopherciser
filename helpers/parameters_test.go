package helpers_test

import (
	"testing"

	"github.com/qlik-oss/gopherciser/helpers"
)

func TestParameters(t *testing.T) {
	params := make(helpers.Params)
	params["sort"] = "-createdAt"
	expected := "?sort=-createdAt"
	result := params.String()
	if result != expected {
		t.Errorf("expected<%s> got %s", expected, result)
	}
	params["limit"] = "30"
	expected = "?sort=-createdAt&limit=30"
	expected2 := "?limit=30&sort=-createdAt"
	result = params.String()
	if result != expected && result != expected2 { // order not guaranteed
		t.Errorf("expected<%s> got %s", expected, result)
	}
}
