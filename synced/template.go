package synced

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/goccy/go-json"
	uuid "github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	// Template used for creating synced templates parsed once
	Template struct {
		// t text including parameters
		t string
		// s sync object
		s *sync.Once
		// template inner template
		template *template.Template
	}
)

func (syn Template) TreatAs() string {
	return "string"
}

var (
	funcMap = template.FuncMap{
		"now":       time.Now,
		"hostname":  os.Hostname,
		"timestamp": timestamp,
		"uuid":      uuid.New,
		"env":       os.Getenv,
		"add":       add,
		"join":      strings.Join,
		"modulo":    modulo,
	}
)

func timestamp() string {
	t := time.Now()
	return fmt.Sprintf("%d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

// New parses string and creates new instance of Template
func New(t string) (*Template, error) {
	syn := Template{t: t, s: &sync.Once{}}
	if err := syn.parse(); err != nil {
		return nil, errors.Wrapf(err, "failed to create sync template from string<%s>", t)
	}
	return &syn, nil
}

// UnmarshalJSON un-marshal from json string
func (syn *Template) UnmarshalJSON(arg []byte) error {

	var s string
	if err := json.Unmarshal(arg, &s); err != nil {
		return errors.Wrap(err, "failed un-marshaling synced template to string")
	}

	*syn = Template{
		t: s,
		s: &sync.Once{},
	}

	return errors.WithStack(syn.parse())
}

// MarshalJSON marshal template to json string
func (syn Template) MarshalJSON() ([]byte, error) {
	return json.Marshal(syn.t)
}

// Parse template
func (syn *Template) parse() error {
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
func (syn *Template) Execute(writer io.Writer, data interface{}) error {
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
func (syn *Template) String() string {
	if syn == nil {
		return ""
	}
	return syn.t
}

// ExecuteString execute template with data
func (input *Template) ExecuteString(data interface{}) (string, error) {
	buf := helpers.GlobalBufferPool.Get()
	defer helpers.GlobalBufferPool.Put(buf)
	if err := input.Execute(buf, data); err != nil {
		return "", errors.Wrap(err, "failed to execute variables template")
	}
	return buf.String(), nil
}

func add(iVal1 interface{}, iVal2 interface{}) (int64, error) {
	val1, err := parseToInt64(iVal1)
	if err != nil {
		return 0, err
	}
	val2, err := parseToInt64(iVal2)
	if err != nil {
		return 0, err
	}

	return val1 + val2, nil
}

func modulo(iVal1 interface{}, iVal2 interface{}) (int64, error) {
	val1, err := parseToInt64(iVal1)
	if err != nil {
		return 0, err
	}
	val2, err := parseToInt64(iVal2)
	if err != nil {
		return 0, err
	}

	return val1 % val2, nil
}

func parseToInt64(val interface{}) (int64, error) {
	switch val := val.(type) {
	case string:
		val2, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, errors.Wrapf(err, "string<%s> not parseable to integer", val)
		}
		return val2, nil
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case uint:
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		return int64(val), nil
	default:
		return 0, errors.Errorf("type<%T> not parseable to int64", val)
	}
}
