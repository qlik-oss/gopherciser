package enigmahandlers

import (
	"sort"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	// ObjectsMap currently subscribed objects
	ObjectsMap struct {
		objects    map[int]*Object
		objectLink map[int]int
		mutex      sync.Mutex
	}
)

// GetObject get object from object list
func (o *ObjectsMap) GetObject(handle int) (*Object, error) {
	obj := o.Load(handle)
	if obj == nil {
		return nil, ObjectNotFound(handle)
	}

	return obj, nil
}

// Load object with handle
func (o *ObjectsMap) Load(handle int) *Object {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.objects == nil {
		return nil
	}
	return (o.objects)[handle]
}

// Store object with handle
func (o *ObjectsMap) Store(handle int, object *Object) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.objects == nil {
		o.objects = make(map[int]*Object)
	}
	o.objects[handle] = object
}

// GetObjectByID get object by id or ObjectIDNotFound error
func (o *ObjectsMap) GetObjectByID(id string) (*Object, error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	for _, v := range o.objects {
		if v != nil && v.ID == id {
			return v, nil
		}
	}
	return nil, ObjectIDNotFound(id)
}

// GetObjectsOfType get all objects of type
func (o *ObjectsMap) GetObjectsOfType(typ ObjectType) []*Object {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	objectList := make([]*Object, 0, len(o.objects))
	for _, v := range o.objects {
		if v.Type == typ {
			local := v
			objectList = append(objectList, local)
		}
	}

	return objectList
}

// AddObject to object list
func (o *ObjectsMap) AddObject(obj *Object) error {
	if o == nil {
		return errors.Errorf("Object map is nil")
	}
	if obj == nil {
		return errors.Errorf("Tried to store nil object in objectmap")
	}
	if obj.Handle < 0 {
		return errors.Errorf("Invalid object handle<%d>", obj.Handle)
	}

	o.Store(obj.Handle, obj)

	return nil
}

// RemoveObject from object list
func (o *ObjectsMap) RemoveObject(handle int) error {
	if o == nil {
		return errors.Errorf("Object map is nil")
	}
	if handle < 0 {
		return errors.Errorf("Invalid object handle<%d>", handle)
	}

	obj := o.Load(handle)

	o.mutex.Lock()
	defer o.mutex.Unlock()

	delete(o.objects, handle)

	if obj == nil {
		return nil
	}

	if err := obj.Close(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// ClearObjectsOfType remove all objects of type from list
func (o *ObjectsMap) ClearObjectsOfType(t ObjectType) ([]int, error) {
	handles := make([]int, 0, len(o.objects))

	for _, v := range o.objects {
		if v != nil && v.Type == t {
			handles = append(handles, v.Handle)
		}
	}

	return handles, errors.WithStack(o.ClearObjects(handles))
}

// ClearObjects remove all objects in list
func (o *ObjectsMap) ClearObjects(handles []int) error {
	var mErr *multierror.Error
	for _, handle := range handles {
		if err := o.ClearObject(handle); err != nil {
			mErr = multierror.Append(mErr, err)
		}
	}
	return errors.WithStack(helpers.FlattenMultiError(mErr))
}

// ClearObject remove object from list
func (o *ObjectsMap) ClearObject(handle int) error {
	err := o.RemoveObject(handle)
	o.RemoveObjectLink(handle)
	return errors.WithStack(err)
}

// AddObjectLink link object and sessionobject, to be used for e.g. auto-charts
func (o *ObjectsMap) AddObjectLink(baseHandle, linkedHandle int) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if o.objectLink == nil {
		o.objectLink = make(map[int]int)
	}

	o.objectLink[baseHandle] = linkedHandle
}

// GetObjectLink get linked object handle, defaults to 0 if not found
func (o *ObjectsMap) GetObjectLink(baseHandle int) int {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	var objectLink int
	if o.objectLink == nil {
		return objectLink
	}

	return o.objectLink[baseHandle]
}

// RemoveObjectLink for handle
func (o *ObjectsMap) RemoveObjectLink(baseHandle int) {
	o.mutex.Lock()
	defer o.mutex.Unlock()
	if o.objectLink == nil {
		return
	}
	if o.objectLink[baseHandle] != 0 {
		delete(o.objectLink, baseHandle)
	}
}

// ForEach Perform function on each object
func (o *ObjectsMap) ForEach(fe func(obj *Object) error) error {
	return errors.WithStack(o.forEach(fe, false))
}

// ForEachWithLock Perform function on each object and lock map during loop,
// don't combine lock with any other locking function
func (o *ObjectsMap) ForEachWithLock(fe func(obj *Object) error) error {
	return errors.WithStack(o.forEach(fe, true))
}

func (o *ObjectsMap) forEach(fe func(obj *Object) error, lock bool) error {
	if lock {
		o.mutex.Lock()
		defer o.mutex.Unlock()
	}
	for _, v := range o.objects {
		obj := v
		if err := fe(obj); err != nil {
			return errors.Wrapf(err, "ForEach failed for object<%+v>", obj)
		}
	}
	return nil
}

// GetAllObjectHandles returns a slice of the handles present in the ObjectsMap
func (o *ObjectsMap) GetAllObjectHandles(lock bool, t ObjectType) []int {
	if lock {
		o.mutex.Lock()
		defer o.mutex.Unlock()
	}
	keys := make([]int, 0, len(o.objects))
	for k, v := range o.objects {
		if v != nil && v.Type == t {
			keys = append(keys, k)
		}
	}
	// sort the keys so seeded randomization always gets the same order
	sort.Ints(keys)
	return keys
}
