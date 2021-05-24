package synced

import "github.com/pkg/errors"

type (
	TemplateMap map[string]*Template
)

//Execute templates and return new map[string]string
func (tm TemplateMap) Execute(data interface{}) (map[string]string, error) {
	m := make(map[string]string, len(tm))
	for key, templ := range tm {
		str, err := templ.ExecuteString(data)
		if err != nil {
			return nil, errors.Wrapf(err, "template key<%s>", key)
		}
		m[key] = str
	}
	return m, nil
}
