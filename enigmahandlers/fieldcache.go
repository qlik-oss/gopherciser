package enigmahandlers

import (
	"sync"

	"github.com/qlik-oss/enigma-go"
)

type (
	// fieldCache
	fieldCache struct {
		fieldMap map[string]*enigma.Field
		mutex    sync.RWMutex
	}
)

func NewFieldCache() fieldCache {
	return fieldCache{fieldMap: make(map[string]*enigma.Field)}

}

func (fc *fieldCache) Lookup(name string) (field *enigma.Field, hit bool) {
	fc.mutex.RLock()
	defer fc.mutex.RUnlock()
	field, hit = fc.fieldMap[name]
	return
}

func (fc *fieldCache) LookupWithFallback(fieldName string, fallback func(fieldName string) (*enigma.Field, error)) (*enigma.Field, error) {
	if field, hit := fc.Lookup(fieldName); hit {
		return field, nil
	}
	field, err := fallback(fieldName)
	if err != nil {
		return field, err
	}
	fc.Store(fieldName, field)
	return field, nil
}

func (fc *fieldCache) Store(name string, field *enigma.Field) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()
	fc.fieldMap[name] = field
}
