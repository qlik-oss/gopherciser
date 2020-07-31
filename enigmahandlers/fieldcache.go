package enigmahandlers

import (
	"sync"

	"github.com/qlik-oss/enigma-go"
)

type (
	// FieldCache
	FieldCache struct {
		fieldMap map[string]*enigma.Field
		mutex    sync.RWMutex
	}
)

func (fc *FieldCache) Lookup(name string) (field *enigma.Field, hit bool) {
	fc.mutex.RLock()
	defer fc.mutex.RUnlock()
	field, hit = fc.fieldMap[name]
	return
}

func (fc *FieldCache) LookupWithFallback(fieldName string, fallback func(fieldName string) (*enigma.Field, error)) (*enigma.Field, error) {
	field, hit := fc.Lookup(fieldName)
	if hit {
		return field, nil
	}
	field, err := fallback(fieldName)
	if err != nil {
		return field, err
	}
	fc.Store(fieldName, field)
	return field, nil
}

func (fc *FieldCache) Store(name string, field *enigma.Field) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()
	fc.fieldMap[field.GenericId] = field
}
