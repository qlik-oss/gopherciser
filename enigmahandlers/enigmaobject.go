package enigmahandlers

import (
	"fmt"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/qlik-oss/enigma-go"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	// ObjectType type of enigma object
	ObjectType int

	// HyperCube wrapping enigma hypercube with additional if binned or not
	HyperCube struct {
		*enigma.HyperCube
		Binned bool
	}

	objectData struct {
		properties    *enigma.GenericObjectProperties
		childlist     *enigma.ChildList
		children      *[]ObjChild
		listobject    *enigma.ListObject
		hypercube     *HyperCube
		treeDataPages []*enigma.NxTreeNode
	}

	// Object sense object handler
	Object struct {
		// Handle of sense object
		Handle int
		// ID of object
		ID string
		// Type of object
		Type ObjectType
		// EnigmaObject enigma object instance
		EnigmaObject interface{}
		closefuncs   []func() error
		data         *objectData
		lockData     sync.Mutex
	}

	ObjChild struct {
		ExternalReference struct {
			MasterID string `json:"masterId"`
			App      string `json:"app"`
			ViewID   string `json:"viewId"`
		} `json:"externalReference"`
		Label string `json:"label"`
		RefID string `json:"refId"`
		Type  string `json:"type"`
	}

	// ObjectNotFound error
	ObjectNotFound int
	// ObjectIDNotFound error
	ObjectIDNotFound string
)

const (
	// ObjTypeApp object is an app
	ObjTypeApp ObjectType = iota
	// ObjTypeSheet object is a sheet
	ObjTypeSheet
	// ObjTypeGenericObject object is sheet object
	ObjTypeGenericObject
)

// SetProperties set/update properties of object
func (obj *Object) SetProperties(properties *enigma.GenericObjectProperties) {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil {
		obj.data = &objectData{
			properties: properties,
		}
		return
	}
	obj.data.properties = properties
}

// Properties get object properties
func (obj *Object) Properties() *enigma.GenericObjectProperties {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil {
		return nil
	}
	return obj.data.properties
}

// SetChildList set object childlist
func (obj *Object) SetChildList(children *enigma.ChildList) {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil {
		obj.data = &objectData{
			childlist: children,
		}
		return
	}
	obj.data.childlist = children
}

// SetChildren set object children
func (obj *Object) SetChildren(children *[]ObjChild) {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil {
		obj.data = &objectData{
			children: children,
		}
		return
	}
	obj.data.children = children
}

// ChildList child objects of object
func (obj *Object) ChildList() *enigma.ChildList {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil {
		return nil
	}
	return obj.data.childlist
}

// SetListObject of object
func (obj *Object) SetListObject(listobject *enigma.ListObject) {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil {
		obj.data = &objectData{
			listobject: listobject,
		}
		return
	}
	obj.data.listobject = listobject
}

// ListObject of object
func (obj *Object) ListObject() *enigma.ListObject {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil {
		return nil
	}
	return obj.data.listobject
}

// SetHyperCube of object
func (obj *Object) SetHyperCube(hypercube *enigma.HyperCube) {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil {
		obj.data = &objectData{
			hypercube: &HyperCube{
				hypercube,
				false,
			},
		}
	}
	obj.data.hypercube = &HyperCube{hypercube, false}
}

// Get all apps from ExternalReference
func (obj *Object) ExternalReferenceApps() []string {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil || obj.data.children == nil {
		return nil
	}
	apps := make([]string, 0, len(*obj.data.children))
	for _, item := range *obj.data.children {
		apps = append(apps, item.ExternalReference.App)
	}
	return apps
}

// HyperCube of object
func (obj *Object) HyperCube() *HyperCube {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil {
		return nil
	}
	return obj.data.hypercube
}

// SetListObjectDataPages on object listobject
func (obj *Object) SetListObjectDataPages(datapages []*enigma.NxDataPage) error {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil || obj.data.listobject == nil {
		return errors.Errorf("object<%s> has no listobject", obj.ID)
	}

	obj.data.listobject.DataPages = datapages

	return nil
}

// ListObjectDataPages from object listobject
func (obj *Object) ListObjectDataPages() []*enigma.NxDataPage {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil || obj.data.listobject == nil {
		return nil
	}
	return obj.data.listobject.DataPages
}

// SetHyperCubeDataPages on object hypercube
func (obj *Object) SetHyperCubeDataPages(datapages []*enigma.NxDataPage, binned bool) error {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil || obj.data.hypercube == nil {
		return errors.Errorf("object<%s> has no hypercube", obj.ID)
	}
	obj.data.hypercube.DataPages = datapages
	obj.data.hypercube.Binned = binned
	return nil
}

// SetStackHyperCubePages on object hypercube
func (obj *Object) SetStackHyperCubePages(datapages []*enigma.NxStackPage) error {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil || obj.data.hypercube == nil {
		return errors.Errorf("object<%s> has no hypercube", obj.ID)
	}
	obj.data.hypercube.StackedDataPages = datapages
	return nil
}

// SetPivotHyperCubePages on object hypercube
func (obj *Object) SetPivotHyperCubePages(datapages []*enigma.NxPivotPage) error {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil || obj.data.hypercube == nil {
		return errors.Errorf("object<%s> has no hypercube", obj.ID)
	}
	obj.data.hypercube.PivotDataPages = datapages
	return nil
}

// SetTreeDataPages on object
func (obj *Object) SetTreeDataPages(treeDataPages []*enigma.NxTreeNode) error {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()

	if obj.data == nil {
		return errors.Errorf("object<%s> has no data structure", obj.ID)
	}

	obj.data.treeDataPages = treeDataPages
	return nil
}

// HyperCubeDataPages from object hypercube
func (obj *Object) HyperCubeDataPages() []*enigma.NxDataPage {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil || obj.data.hypercube == nil {
		return nil
	}
	return obj.data.hypercube.DataPages
}

// HyperCubeStackPages from object hypercube
func (obj *Object) HyperCubeStackPages() []*enigma.NxStackPage {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil || obj.data.hypercube == nil {
		return nil
	}
	return obj.data.hypercube.StackedDataPages
}

// HyperPivotPages from object hypercube
func (obj *Object) HyperPivotPages() []*enigma.NxPivotPage {
	obj.lockData.Lock()
	defer obj.lockData.Unlock()
	if obj.data == nil || obj.data.hypercube == nil {
		return nil
	}
	return obj.data.hypercube.PivotDataPages
}

// Close execute registered close functions
func (obj *Object) Close() error {
	if obj == nil {
		return nil
	}

	if obj.closefuncs == nil {
		return nil
	}

	var mErr *multierror.Error
	for _, f := range obj.closefuncs {
		err := f()
		if err != nil {
			mErr = multierror.Append(mErr, err)
		}
	}
	return errors.WithStack(helpers.FlattenMultiError(mErr))
}

// RegisterCloseFunc register function to be executed on object close
func (obj *Object) RegisterCloseFunc(f func() error) {
	if obj.closefuncs == nil {
		obj.closefuncs = make([]func() error, 1)
		obj.closefuncs[0] = f
		return
	}
	obj.closefuncs = append(obj.closefuncs, f)
}

// NewObject container for enigma object
func NewObject(handle int, t ObjectType, id string, enigmaobject interface{}) *Object {
	obj := &Object{
		Handle:       handle,
		Type:         t,
		ID:           id,
		EnigmaObject: enigmaobject,
	}
	return obj
}

func (onf ObjectNotFound) Error() string {
	return fmt.Sprintf("handle<%d> not found in object list", int(onf))
}

func (onf ObjectIDNotFound) Error() string {
	return fmt.Sprintf("Object<%s> not found in object list", string(onf))
}
