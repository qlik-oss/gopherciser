package session

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/logger"
)

type (
	// IDMap should be used as a lookup table for Id defined in the script and id of object created.
	// We should never define the id to be sent to engine as this might cause collision errors, instead
	// this lookup table is used with the key id connecting object together for in-script purposes.
	IDMap struct {
		m *sync.Map
	}

	// DuplicateKeyError is returned in case of adding an already existing key id
	DuplicateKeyError string
)

// Error id already exist in map
func (err DuplicateKeyError) Error() string {
	return fmt.Sprintf("Duplicate id<%s>, id previously used.", string(err))
}

// allocate new sync.Map if not already allocated
func (idm *IDMap) newIfNil() {
	if idm.m == nil {
		idm.m = &sync.Map{}
	}
}

// IsDuplicateKey returns already in case of duplicate key
func (idm *IDMap) IsDuplicateKey(key string) error {
	if idm == nil {
		return errors.New("IDMap is nil")
	}
	idm.newIfNil()

	if _, exist := idm.m.Load(key); exist {
		return DuplicateKeyError(key)
	}
	return nil
}

// Add new key to id map
func (idm *IDMap) Add(key, id string, logEntry *logger.LogEntry) error {
	return idm.add(key, id, logEntry, false)
}

// Replace key in id map
func (idm *IDMap) Replace(key, id string, logEntry *logger.LogEntry) error {
	return idm.add(key, id, logEntry, true)
}

func (idm *IDMap) add(key, id string, logEntry *logger.LogEntry, overwrite bool) error {
	if idm == nil {
		return errors.New("IDMap is nil")
	}

	if id == "" {
		return errors.New("adding empty value string to IDMap")
	}

	idm.newIfNil()

	if !overwrite {
		// First check if key already exists
		if err := idm.IsDuplicateKey(key); err != nil {
			return errors.WithStack(err)
		}
	}

	idm.m.Store(key, id)
	logEntry.LogDebugf("Key pair added to IDMap %s:%s", key, id)

	return nil
}

// Get Id value for key, if key not found return key itself
func (idm *IDMap) Get(key string) string {
	if idm == nil || idm.m == nil || key == "" {
		return key
	}
	iid, ok := idm.m.Load(key)
	if !ok {
		return key
	}

	id, ok := iid.(string)
	if !ok {
		return key
	}

	return id
}
