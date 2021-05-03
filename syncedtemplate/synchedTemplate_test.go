package syncedtemplate

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
		Param4 SyncedTemplate `json:"param4"`
	}

	_ = os.Setenv("test-env", "val2")

	jsn := `{
	"param1" : "val1 is {{.Val1}} and val2 is {{env \"test-env\"}}",
	"param2" : "{{ add 1 \"3\" }}",
	"param3" : "{{ join .Val2 \",\" }},elem4",
	"param4" : "{{ join (slice .Val2 0 (add (len .Val2) -1)) \",\" }}"
}`

	if err := jsonit.Unmarshal([]byte(jsn), &myStruct); err != nil {
		t.Fatal("failed to unmarshal struct:", err)
	}

	testParam(t, &myStruct.Param1, "val1 is val1 and val2 is val2")
	testParam(t, &myStruct.Param2, "4")
	testParam(t, &myStruct.Param3, "elem1,elem2,elem3,elem4")
	testParam(t, &myStruct.Param4, "elem1,elem2")
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

type headerStruct struct {
	Params map[string]*SyncedTemplate
}

func TestMap(t *testing.T) {
	_ = os.Setenv("test-env", "val2")
	var myStruct headerStruct
	jsn := `{
		"Params" : {
			"param1" : "val1 is {{.Val1}} and val2 is {{env \"test-env\"}}",
			"param2" : "{{ add 1 \"3\" }}",
			"param3" : "{{ join .Val2 \",\" }},elem4",
			"param4" : "{{ join (slice .Val2 0 (add (len .Val2) -1)) \",\" }}"		
		}
	}`

	if err := jsonit.Unmarshal([]byte(jsn), &myStruct); err != nil {
		t.Fatal("failed to unmarshal struct:", err)
	}

	t.Logf("%+v", myStruct)

	testParam(t, myStruct.Params["param1"], "val1 is val1 and val2 is val2")
	testParam(t, myStruct.Params["param2"], "4")
	testParam(t, myStruct.Params["param3"], "elem1,elem2,elem3,elem4")
	testParam(t, myStruct.Params["param4"], "elem1,elem2")
}
