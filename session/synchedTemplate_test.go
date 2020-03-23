package session

import (
	"bytes"
	"testing"
)

var (
	data = struct{ Val1 string }{"val1"}
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
	}

	jsn := `{
	"param1" : "val1 is {{.Val1}}"
}`

	if err := jsonit.Unmarshal([]byte(jsn), &myStruct); err != nil {
		t.Fatal("failed to unmarshal struct:", err)
	}

	buf := bytes.NewBuffer(nil)
	if err := myStruct.Param1.Execute(buf, data); err != nil {
		t.Fatal("failed to execute template:", err)
	}

	expected := "val1 is val1"
	result := buf.String()
	if expected != result {
		t.Errorf("unexpected template result<%s> expected<%s>", result, expected)
	}
}
