package enigmahandlers

import (
	"sync"

	"github.com/qlik-oss/enigma-go"
)

type (
	// varCache
	varCache struct {
		varMap map[string]*enigma.GenericVariable
		mutex  sync.RWMutex
	}
)

func NewVarCache() varCache {
	return varCache{varMap: map[string]*enigma.GenericVariable{}}
}

func (fc *varCache) Lookup(varName string) (varValue *enigma.GenericVariable, hit bool) {
	fc.mutex.RLock()
	defer fc.mutex.RUnlock()
	varValue, hit = fc.varMap[varName]
	return
}

func (fc *varCache) LookupWithFallback(varName string, fallback func(varName string) (*enigma.GenericVariable, error)) (*enigma.GenericVariable, error) {
	if variable, hit := fc.Lookup(varName); hit {

		return variable, nil
	}
	variable, err := fallback(varName)
	if err != nil {
		return variable, err
	}
	fc.Store(varName, variable)
	return variable, nil
}

func (fc *varCache) Store(name string, variable *enigma.GenericVariable) {
	fc.mutex.Lock()
	defer fc.mutex.Unlock()
	fc.varMap[name] = variable
}
