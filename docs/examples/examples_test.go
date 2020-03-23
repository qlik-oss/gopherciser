package examples

import (
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/config"
)

func TestExamples(t *testing.T) {
	files, err := ioutil.ReadDir("./")
	if err != nil {
		t.Errorf(err.Error())
	}

	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".json") ||
			strings.HasPrefix(f.Name(), "ignore") {

			continue
		}
		err := testFile(f.Name())
		if err != nil {
			t.Errorf(err.Error())
		}
	}
}

func testFile(filename string) error {
	jsonConfig, err := ioutil.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "Failed to open file <%s>", filename)
	}

	var cfg config.Config
	if err := json.Unmarshal(jsonConfig, &cfg); err != nil {
		return errors.Wrapf(err, "Failed to unmarshal file <%s>", filename)
	}

	if err := cfg.Validate(); err != nil {
		return errors.Wrapf(err, "Failed to validate file <%s>", filename)
	}
	return nil
}
