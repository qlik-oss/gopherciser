package synced

import (
	"os"
	"testing"

	"github.com/goccy/go-json"
)

func TestMap(t *testing.T) {
	_ = os.Setenv("test-env", "val2")
	jsn := `{
		"param1" : "val1 is {{.Val1}} and val2 is {{env \"test-env\"}}",
		"param2" : "{{ add 1 \"3\" }}",
		"param3" : "{{ join .Val2 \",\" }},elem4",
		"param4" : "{{ join (slice .Val2 0 (add (len .Val2) -1)) \",\" }}"		
	}`

	var tmplMap TemplateMap
	if err := json.Unmarshal([]byte(jsn), &tmplMap); err != nil {
		t.Fatal("failed to unmarshal struct:", err)
	}

	valueMap, err := tmplMap.Execute(data)
	if err != nil {
		t.Fatal(err)
	}

	cmpMap := map[string]string{
		"param1": "val1 is val1 and val2 is val2",
		"param2": "4",
		"param3": "elem1,elem2,elem3,elem4",
		"param4": "elem1,elem2",
	}

	for k, v := range cmpMap {
		if valueMap[k] != v {
			t.Errorf("key<%s> expected<%s> have<%s>", k, valueMap[k], v)
		}
	}
}
