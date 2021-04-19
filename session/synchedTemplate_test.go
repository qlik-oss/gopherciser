package session

import (
	"bytes"
	"os"
	"testing"
)

var (
	data = struct {
		Val1 string
		Val2 interface{}
	}{
		Val1: "val1",
		Val2: []string{"elem1", "elem2", "elem3"},
	}
)

func TestEmptyTemplate(t *testing.T) {

	myStruct := struct {
		Param1 SyncedTemplate `json:"param1"`
	}{}

	jsn, err := jsonit.Marshal(myStruct)
	if err != nil {
		t.Fatal("error marshaling empty synced template:", err)
	}

	expected := `{"param1":""}`
	if string(jsn) != expected {
		t.Errorf("unexpected json result <%s> expected <%s>", string(jsn), expected)
	}

	if err = myStruct.Param1.parse(); err != nil {
		t.Error("error parsing empty synced template:", err)
	}

	buf := bytes.NewBuffer(nil)
	if err = myStruct.Param1.Execute(buf, data); err != nil {
		t.Error("failed executing empty synced template:", err)
	}

	// test without parse
	myStruct.Param1 = SyncedTemplate{}
	if err = myStruct.Param1.Execute(buf, data); err != nil {
		t.Error("failed executing empty synced template:", err)
	}
}

func TestTemplate(t *testing.T) {
	var myStruct struct {
		Param1 SyncedTemplate `json:"param1"`
		Param2 SyncedTemplate `json:"param2"`
		Param3 SyncedTemplate `json:"param3"`
	}

	_ = os.Setenv("test-env", "val2")

	jsn := `{
	"param1" : "val1 is {{.Val1}} and val2 is {{env \"test-env\"}}",
	"param2" : "{{ join .Val2 \",\" }},elem4",
	"param3" : "{{ add 1 \"3\" }}"
}`

	if err := jsonit.Unmarshal([]byte(jsn), &myStruct); err != nil {
		t.Fatal("failed to unmarshal struct:", err)
	}

	testParam(t, &myStruct.Param1, "val1 is val1 and val2 is val2")
	testParam(t, &myStruct.Param2, "elem1,elem2,elem3,elem4")
	testParam(t, &myStruct.Param3, "4")
}

func testParam(t *testing.T, tmpl *SyncedTemplate, expected string) {
	t.Helper()

	buf := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buf, data); err != nil {
		t.Fatal("failed to execute template:", err)
	}

	result := buf.String()
	if expected != result {
		t.Errorf("unexpected template result<%s> expected<%s>", result, expected)
	}
}
