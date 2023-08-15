package enigmahandlers

import (
	"sync"

	"github.com/qlik-oss/enigma-go/v4"
)

type (
	// FieldCache
	FieldCache struct {
		fieldMap map[string]*enigma.Field
		mutex    sync.RWMutex
	}
)

func NewFieldCache() FieldCache {
	return FieldCache{fieldMap: make(map[string]*enigma.Field)}

}

func (fc *FieldCache) Lookup(name string) (field *enigma.Field, hit bool) {
	fc.mutex.RLock()
	defer fc.mutex.RUnlock()
	field, hit = fc.fieldMap[name]
	return
}

func (fc *FieldCache) Store(name string, field *enigma.Field) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()
	fc.fieldMap[name] = field
}
