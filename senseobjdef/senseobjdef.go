package senseobjdef

import (
	"fmt"
	"os"

	"github.com/goccy/go-json"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	// SelectType select function type
	SelectType int
	// DataType Get Data function type
	DataType int
	// DataDefType type of data definition, e.g. ListObject or HyperCube
	DataDefType int

	// DataDef type of data and path to data carrier
	DataDef struct {
		// Type of data
		Type DataDefType `json:"type"`
		// Path to data carrier
		Path helpers.DataPath `json:"path,omitempty"`
	}

	// GetDataRequests data requests to send
	GetDataRequests struct {
		// Type of data function
		Type DataType `json:"type"`
		// Path for get data function
		Path string `json:"path,omitempty"`
		// Height of data to get in GetData
		Height int `json:"height,omitempty"`
	}

	// Data Get data definitions
	DataCore struct {
		// Constraints constraint defining if to send requests
		Constraints []*Constraint `json:"constraints,omitempty"`
		// Requests List of data requests to send
		Requests []GetDataRequests `json:"requests,omitempty"`
	}

	Data struct {
		DataCore
	}

	// Select definitions for selecting in object
	Select struct {
		// Type of select function
		Type SelectType `json:"type"`
		// Path to use for selection
		Path string `json:"path,omitempty"`
	}

	// ObjectDef object definitions
	ObjectDef struct {
		// DataDef type of data and path to data carrier
		DataDef DataDef `json:"datadef,omitempty"`
		// Data Get data definitions
		Data []Data `json:"data,omitempty"`
		// Select definitions for selecting in object
		Select *Select `json:"select,omitempty"`
	}

	// ObjectDefs contains how to find and select data within sense objects
	ObjectDefs map[string]*ObjectDef

	//NoDefError No object definition found
	NoDefError string
)

//Non iota constants
const (
	DefaultDataHeight = 500
)

//When adding DataDefType, also:
// * add entry in dataDefTypeEnum
const (
	// DataDefUnknown unknown data definition type
	DataDefUnknown DataDefType = iota
	// DataDefListObject ListObject data carrier
	DataDefListObject
	// DataDefHyperCube HyperCube data carrier
	DataDefHyperCube
	// DataDefNoData object contains no data carrier
	DataDefNoData
)

//When adding SelectType, also:
// * add entry in selectTypeEnum
const (
	// SelectTypeUnknown unknown select func (default int)
	SelectTypeUnknown SelectType = iota
	// SelectTypeListObjectValues use SelectListObjectValues method
	SelectTypeListObjectValues
	// SelectTypeHypercubeValues use SelectHyperCubeValues method
	SelectTypeHypercubeValues
	// SelectTypeHypercubeColumnValues each dimension is a data page
	SelectTypeHypercubeColumnValues
)

//When adding DataType, also:
// * add entry in dataTypeEnum
const (
	// DataTypeLayout get data from layout
	DataTypeLayout DataType = iota
	// DataTypeListObject get data from listobject data
	DataTypeListObject
	// HyperCubeData get data from hypercube
	DataTypeHyperCubeData
	// DataTypeHyperCubeDataColumns
	DataTypeHyperCubeDataColumns
	// HyperCubeReducedData get hypercube reduced data
	DataTypeHyperCubeReducedData
	// DataTypeHyperCubeBinnedData get hypercube binned data
	DataTypeHyperCubeBinnedData
	// DataTypeHyperCubeStackData get hypercube stacked data
	DataTypeHyperCubeStackData
	// DataTypeHyperCubeContinuousData get hypercube continuous data
	DataTypeHyperCubeContinuousData
	// DataTypeHyperCubeTreeData get hypercube tree data
	DataTypeHyperCubeTreeData
)

var (
	od ObjectDefs

	dataDefTypeEnum = enummap.NewEnumMapOrPanic(map[string]int{
		"unknown":    int(DataDefUnknown),
		"listobject": int(DataDefListObject),
		"hypercube":  int(DataDefHyperCube),
		"nodata":     int(DataDefNoData),
	})

	selectTypeEnum = enummap.NewEnumMapOrPanic(map[string]int{
		"unknown":               int(SelectTypeUnknown),
		"listobjectvalues":      int(SelectTypeListObjectValues),
		"hypercubevalues":       int(SelectTypeHypercubeValues),
		"hypercubecolumnvalues": int(SelectTypeHypercubeColumnValues),
	})

	dataTypeEnum = enummap.NewEnumMapOrPanic(map[string]int{
		"layout":                  int(DataTypeLayout),
		"listobjectdata":          int(DataTypeListObject),
		"hypercubedata":           int(DataTypeHyperCubeData),
		"hypercubereduceddata":    int(DataTypeHyperCubeReducedData),
		"hypercubebinneddata":     int(DataTypeHyperCubeBinnedData),
		"hypercubestackdata":      int(DataTypeHyperCubeStackData),
		"hypercubedatacolumns":    int(DataTypeHyperCubeDataColumns),
		"hypercubecontinuousdata": int(DataTypeHyperCubeContinuousData),
		"hypercubetreedata":       int(DataTypeHyperCubeTreeData),
	})
)

func init() {
	od = DefaultObjectDefs
}

// Error No object definition found
func (err NoDefError) Error() string {
	return fmt.Sprintf("Definition for object<%s> not found", string(err))
}

// UnmarshalJSON unmarshal SelectType
func (t *SelectType) UnmarshalJSON(arg []byte) error {
	i, err := selectTypeEnum.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal SelectType")
	}

	*t = SelectType(i)

	return nil
}

// MarshalJSON marshal SelectType
func (t SelectType) MarshalJSON() ([]byte, error) {
	str, err := selectTypeEnum.String(int(t))
	if err != nil {
		return nil, errors.Errorf("Unknown SelectType<%d>", t)
	}

	if str == "" {
		return nil, errors.Errorf("Unknown SelectType<%d>", t)
	}

	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// String representation of select type
func (t SelectType) String() string {
	return selectTypeEnum.StringDefault(int(t), "unknown")
}

// UnmarshalJSON unmarshal DataType
func (typ *DataType) UnmarshalJSON(arg []byte) error {
	i, err := dataTypeEnum.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal DataType")
	}

	*typ = DataType(i)
	return nil
}

// MarshalJSON marshal DataType
func (typ DataType) MarshalJSON() ([]byte, error) {
	str, err := dataTypeEnum.String(int(typ))
	if err != nil {
		return nil, errors.Errorf("Unknown DataType<%d>", typ)
	}

	if str == "" {
		return nil, errors.Errorf("Unknown DataType<%d>", typ)
	}

	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// String representation of data type or "unknown"
func (typ DataType) String() string {
	return dataTypeEnum.StringDefault(int(typ), "unknown")
}

// UnmarshalJSON unmarshal DataDefType
func (d *DataDefType) UnmarshalJSON(arg []byte) error {
	i, err := dataDefTypeEnum.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal DataDefType")
	}

	*d = DataDefType(i)

	return nil
}

// MarshalJSON marshal DataFuncType
func (d DataDefType) MarshalJSON() ([]byte, error) {
	str, err := dataDefTypeEnum.String(int(d))
	if err != nil {
		return nil, errors.Errorf("Unknown DataDefType<%d>", d)
	}

	if str == "" {
		return nil, errors.Errorf("Unknown DataDefType<%d>", d)
	}

	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// String representation of DataDefType
func (d DataDefType) String() string {
	return dataDefTypeEnum.StringDefault(int(d), "unknown")
}

// UnmarshalJSON unmarshal Data
func (d *Data) UnmarshalJSON(arg []byte) error {
	if err := helpers.HasDeprecatedFields(arg, []string{
		"/constraint",
	}); err != nil {
		return errors.New("Deprecated field 'constraint' - please replace with 'constraints' array'")
	}
	dc := DataCore{}
	err := json.Unmarshal(arg, &dc)
	if err != nil {
		return err
	}
	*d = Data{dc}
	return nil
}

// OverrideFromFile read config from file (using default config)
func OverrideFromFile(cfgFile string) (ObjectDefs, error) {
	err := od.OverrideFromFile(cfgFile)
	return od, errors.WithStack(err)
}

// FromFile read config from file
func (defs ObjectDefs) OverrideFromFile(cfgFile string) error {
	if defs == nil {
		return errors.Errorf("defs is nil")
	}

	if cfgFile == "" {
		return errors.Errorf("No config file defined")
	}

	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		return errors.Wrapf(err, "file not found<%s>", cfgFile)
	}

	jsonOverrides, err := os.ReadFile(cfgFile)
	if err != nil {
		return errors.Wrapf(err, "Error reading config from file<%s>", cfgFile)
	}

	var overrides ObjectDefs
	err = json.Unmarshal(jsonOverrides, &overrides)
	if err != nil {
		return errors.Wrapf(err, "Error unmarshaling file<%s>", cfgFile)
	}

	for k, v := range overrides {
		value := v // de-reference pointer
		defs[k] = value
	}

	return nil
}

// GetObjectDef get definitions for object type (using default config)
func GetObjectDef(object string) (*ObjectDef, error) {
	return od.GetObjectDef(object)
}

// GetObjectDef get definitions for object type
func (defs ObjectDefs) GetObjectDef(object string) (*ObjectDef, error) {
	if defs == nil {
		return nil, errors.Errorf("defs is nil")
	}

	def, ok := defs[object]
	if !ok {
		return nil, errors.WithStack(NoDefError(object))
	}

	return def, nil
}

// Validate object definition
func (def *ObjectDef) Validate() error {
	if def == nil {
		return errors.Errorf("object definition is nil")
	}

	switch def.DataDef.Type {
	case DataDefUnknown:
	case DataDefNoData:
	default:
		dataDefPath := string(def.DataDef.Path)
		if dataDefPath == "" || dataDefPath[0] != '/' {
			str, _ := dataDefTypeEnum.String(int(def.DataDef.Type))
			return errors.Errorf("data def type<%s> requires a path", str)
		}
	}

	//Validate get data requests def
	for _, d := range def.Data {
		for _, r := range d.Requests {
			switch r.Type {
			case DataTypeLayout:
			default:
				if r.Path == "" || r.Path[0] != '/' {
					str, _ := dataTypeEnum.String(int(r.Type))
					return errors.Errorf("data type<%s> requires a path", str)
				}
			}
		}
	}

	if def.Select != nil {
		switch def.Select.Type {
		case SelectTypeUnknown:
		default:
			if def.Select.Path == "" || def.Select.Path[0] != '/' {
				str, _ := selectTypeEnum.String(int(def.Select.Type))
				return errors.Errorf("select type<%s> requires a path", str)
			}
		}
	}

	return nil
}

// Evaluate which constraint section applies
func (def *ObjectDef) Evaluate(data json.RawMessage) ([]GetDataRequests, error) {
	for _, v := range def.Data {
		meetsConstraints := true
		for _, c := range v.Constraints {

			result, err := c.Evaluate(data)
			if err != nil {
				return nil, errors.Wrap(err, "Failed to evaluate get data function")
			}
			if !result {
				meetsConstraints = false
				break
			}
		}

		if meetsConstraints {
			return v.Requests, nil
		}
	}

	return nil, errors.Errorf("No constraint section applies")
}

//MaxHeight max data to get
func (data GetDataRequests) MaxHeight() int {
	if data.Height < 1 {
		return DefaultDataHeight
	}
	return data.Height
}
