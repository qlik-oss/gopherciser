package session

import "github.com/qlik-oss/gopherciser/enummap"

type (
	SessionVariableTypeEnum int

	SessionVariable struct {
		Type SessionVariableTypeEnum
	}
)

// SessionVariableTypeEnum enumerations
const (
	SessionVariableTypeUnknown SessionVariableTypeEnum = iota
	SessionVariableTypeString
	SessionVariableTypeInt
	SessionVariableTypeArray
)

var (
	SessionVariableTypeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
		"unknown": int(SessionVariableTypeUnknown),
		"string":  int(SessionVariableTypeString),
		"int":     int(SessionVariableTypeInt),
		"array":   int(SessionVariableTypeArray),
	})
)
