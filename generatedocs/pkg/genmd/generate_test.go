package genmd

import (
	"os"
	"testing"

	generated "github.com/qlik-oss/gopherciser/generatedocs/pkg/genmd/testdata"
)

func TestGenerateMarkDown(t *testing.T) {
	t.Skip()
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
	expectedMDBytes, err := os.ReadFile("testdata/settingup.md")
	if err != nil {
		t.Error(err)
	}

	expectedMarkdown := string(expectedMDBytes)

	if expectedMarkdown != markdown {
		t.Error("unexpected result when generaterating markdown")
	}
}
