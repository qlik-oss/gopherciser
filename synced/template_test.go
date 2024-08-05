package synced

import (
	"bytes"
	"os"
	"testing"

	"github.com/goccy/go-json"
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
		Param1 Template `json:"param1"`
	}{}

	jsn, err := json.Marshal(myStruct)
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
	myStruct.Param1 = Template{}
	if err = myStruct.Param1.Execute(buf, data); err != nil {
		t.Error("failed executing empty synced template:", err)
	}
}

func TestTemplate(t *testing.T) {
	var myStruct struct {
		Param1 Template `json:"param1"`
		Param2 Template `json:"param2"`
		Param3 Template `json:"param3"`
		Param4 Template `json:"param4"`
		Param5 Template `json:"param5"`
	}

	_ = os.Setenv("test-env", "val2")

	jsn := `{
	"param1" : "val1 is {{.Val1}} and val2 is {{env \"test-env\"}}",
	"param2" : "{{ add 1 \"3\" }}",
	"param3" : "{{ join .Val2 \",\" }},elem4",
	"param4" : "{{ join (slice .Val2 0 (add (len .Val2) -1)) \",\" }}",
	"param5" : "{{ modulo 10 \"4\" }}"
}`

	if err := json.Unmarshal([]byte(jsn), &myStruct); err != nil {
		t.Fatal("failed to unmarshal struct:", err)
	}

	testParam(t, &myStruct.Param1, "val1 is val1 and val2 is val2")
	testParam(t, &myStruct.Param2, "4")
	testParam(t, &myStruct.Param3, "elem1,elem2,elem3,elem4")
	testParam(t, &myStruct.Param4, "elem1,elem2")
	testParam(t, &myStruct.Param5, "2")
}

func testParam(t *testing.T, tmpl *Template, expected string) {
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
