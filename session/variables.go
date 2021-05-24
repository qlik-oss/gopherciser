package session

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enummap"
)

type (
	SessionVariableTypeEnum int
)

// SessionVariableTypeEnum enumerations
const (
	SessionVariableTypeUnknown SessionVariableTypeEnum = iota
	SessionVariableTypeString
	SessionVariableTypeInt
	SessionVariableTypeArray
)

var (
	sessionVariableTypeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
		"undefined": int(SessionVariableTypeUnknown),
		"string":    int(SessionVariableTypeString),
		"int":       int(SessionVariableTypeInt),
		"array":     int(SessionVariableTypeArray),
	})
)

// GetEnumMap return sessionVariableTypeEnumMap to GUI
func (typ SessionVariableTypeEnum) GetEnumMap() *enummap.EnumMap {
	return sessionVariableTypeEnumMap
}

// String implements Stringer interface
func (typ SessionVariableTypeEnum) String() string {
	return sessionVariableTypeEnumMap.StringDefault(int(typ), "unknown")
}

// UnmarshalJSON SessionVariableTypeEnum type
func (typ *SessionVariableTypeEnum) UnmarshalJSON(arg []byte) error {
	i, err := sessionVariableTypeEnumMap.UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal SessionVariableType")
	}

	*typ = SessionVariableTypeEnum(i)

	return nil
}

// MarshalJSON SessionVariableTypeEnum type
func (typ SessionVariableTypeEnum) MarshalJSON() ([]byte, error) {
	str, err := sessionVariableTypeEnumMap.String(int(typ))
	if err != nil {
		return nil, errors.Errorf("unknown SheetDeletionModeEnum<%d>", typ)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// Validate SessionVariableTypeEnum type
func (typ *SessionVariableTypeEnum) Validate() error {
	if *typ == SessionVariableTypeUnknown {
		return errors.New("session variable type not defined")
	}

	return nil
}
