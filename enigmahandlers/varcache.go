package enigmahandlers

import (
	"sync"

	"github.com/qlik-oss/enigma-go"
)

type (
	// varCache
	VarCache struct {
		varMap map[string]*enigma.GenericVariable
		mutex  sync.RWMutex
	}
)

func NewVarCache() VarCache {
	return VarCache{varMap: map[string]*enigma.GenericVariable{}}
}

func (vc *VarCache) Lookup(varName string) (varValue *enigma.GenericVariable, hit bool) {
	vc.mutex.RLock()
	defer vc.mutex.RUnlock()
	varValue, hit = vc.varMap[varName]
	return
}

func (vc *VarCache) Store(name string, variable *enigma.GenericVariable) {
	vc.mutex.Lock()
	defer vc.mutex.Unlock()
	vc.varMap[name] = variable
}
