package genmd

import (
	"io/ioutil"
	"testing"

	"github.com/andreyvit/diff"
	generated "github.com/qlik-oss/gopherciser/generatedocs/pkg/genmd/testdata"
)

func TestGenerateMarkDown(t *testing.T) {
	unitTestMode = true
	compiledDocs := &CompiledDocs{
		Actions: generated.Actions,
		Params:  generated.Params,
		Config:  generated.Config,
		Groups:  generated.Groups,
		Extra:   generated.Extra,
	}
	mdBytes := generateFromCompiled(compiledDocs)

	markdown := string(mdBytes)
	expectedMDBytes, err := ioutil.ReadFile("testdata/settingup.md")
	if err != nil {
		t.Error(err)
	}

	expectedMarkdown := string(expectedMDBytes)

	if expectedMarkdown != markdown {
		t.Error("unexpected result when generaterating markdown")
		t.Log(diff.LineDiff(expectedMarkdown, markdown))
	}
}
