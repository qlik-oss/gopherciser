package session

import (
	"fmt"
	"io"
	"os"
	"sync"
	"text/template"
	"time"

	uuid "github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	// Synced Template used for creating templates parsed once
	SyncedTemplate struct {
		// t text including parameters
		t string
		// s sync object
		s *sync.Once
		// template inner template
		template *template.Template
	}
)

func (syn SyncedTemplate) TreatAs() string {
	return "string"
}

var (
	funcMap = template.FuncMap{
		"now":       time.Now,
		"hostname":  os.Hostname,
		"timestamp": timestamp,
		"uuid":      uuid.New,
	}
	jsonit = jsoniter.ConfigCompatibleWithStandardLibrary
)

func timestamp() string {
	t := time.Now()
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

// NewSyncedTemplate parses string and creates new instance of SyncedTemplate
func NewSyncedTemplate(t string) (*SyncedTemplate, error) {
	syn := SyncedTemplate{t: t, s: &sync.Once{}}
	if err := syn.parse(); err != nil {
		return nil, errors.Wrapf(err, "failed to create sync template from string<%s>", t)
	}
	return &syn, nil
}

// UnmarshalJSON un-marshal from json string
func (syn *SyncedTemplate) UnmarshalJSON(arg []byte) error {

	var s string
	if err := jsonit.Unmarshal(arg, &s); err != nil {
		return errors.Wrap(err, "failed un-marshaling synced template to string")
	}

	*syn = SyncedTemplate{
		t: s,
		s: &sync.Once{},
	}

	return errors.WithStack(syn.parse())
}

// MarshalJSON marshal template to json string
func (syn SyncedTemplate) MarshalJSON() ([]byte, error) {
	return jsonit.Marshal(syn.t)
}

// Parse template
func (syn *SyncedTemplate) parse() error {
	if syn == nil {
		return errors.New("template is nil")
	}

	var parseErr error
	if syn.s == nil { // Not thread safe, but what can we do? Alternative is to throw error.
		syn.s = &sync.Once{}
	}
	syn.s.Do(func() {
		syn.template, parseErr = (&template.Template{}).Funcs(funcMap).Parse(syn.t)
	})
	return parseErr
}

// Execute template
func (syn *SyncedTemplate) Execute(writer io.Writer, data interface{}) error {
	if syn == nil {
		return errors.New("template is nil")
	}

	// make sure parsing has been done
	if err := syn.parse(); err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(syn.template.Execute(writer, data))
}

// String return text pattern
func (syn *SyncedTemplate) String() string {
	if syn == nil {
		return ""
	}
	return syn.t
}

// ReplaceWithoutSessionVariables execute template without session variables - only use if we do not have a session
func (input *SyncedTemplate) ReplaceWithoutSessionVariables(data interface{}) (string, error) {
	buf := helpers.GlobalBufferPool.Get()
	defer helpers.GlobalBufferPool.Put(buf)
	if err := input.Execute(buf, data); err != nil {
		return "", errors.Wrap(err, "failed to execute variables template")
	}
	return buf.String(), nil
}
