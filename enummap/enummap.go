package enummap

import (
	"fmt"
	"sort"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

type (
	// EnumMap struct with internal maps containing string and int representations
	EnumMap struct {
		asString map[int]string
		asInt    map[string]int
	}

	// IntKeyNotFoundError Key(int) not found in enum map
	IntKeyNotFoundError int
	// StringKeyNotFoundError Key(string) not found in enum map
	StringKeyNotFoundError string
	// DuplicateValueError Duplicate value
	DuplicateValueError int
	// DuplicateKeyError Duplicate key
	DuplicateKeyError string
	// NonLowercaseKeyError Key contains non-lowercase character
	NonLowercaseKeyError string
)

var jsonit = jsoniter.ConfigCompatibleWithStandardLibrary

func (ik IntKeyNotFoundError) Error() string {
	return fmt.Sprintf("Key<%d> not found", int(ik))
}

func (sk StringKeyNotFoundError) Error() string {
	return fmt.Sprintf("Key<%s> not found", string(sk))
}

func (dv DuplicateValueError) Error() string {
	return fmt.Sprintf("Value<%d> is duplicated", int(dv))
}

func (dk DuplicateKeyError) Error() string {
	return fmt.Sprintf("Key<%s> is duplicated", string(dk))
}

func (nlk NonLowercaseKeyError) Error() string {
	return fmt.Sprintf("Key<%s> must use lowercase characters only", string(nlk))
}

// New empty enum map
func New() *EnumMap {
	em := &EnumMap{
		asInt:    make(map[string]int),
		asString: make(map[int]string),
	}
	return em
}

// NewEnumMap new enum map from map
func NewEnumMap(m map[string]int) (*EnumMap, error) {
	em := &EnumMap{
		asInt: m,
	}

	em.asString = make(map[int]string, len(m))

	for k, v := range m {
		//Validate duplicate
		if _, ok := em.asString[v]; ok {
			return nil, DuplicateValueError(v)
		}
		if strings.ToLower(k) != k {
			return nil, NonLowercaseKeyError(k)
		}

		em.asString[v] = k
	}

	return em, nil
}

func (em *EnumMap) AsInt() map[string]int {
	return em.asInt
}

// Int Get integer representation of enum
func (em *EnumMap) Int(s string) (int, error) {
	i, ok := em.asInt[strings.ToLower(s)]
	if !ok {
		return 0, StringKeyNotFoundError(s)
	}
	return i, nil
}

// String Get string representation of enum
func (em *EnumMap) String(i int) (string, error) {
	s, ok := em.asString[i]
	if !ok {
		return "", IntKeyNotFoundError(i)
	}
	return s, nil
}

// StringDefault Get string representation of enum or default value
func (em *EnumMap) StringDefault(i int, dflt string) string {
	s, ok := em.asString[i]
	if !ok {
		return dflt
	}
	return s
}

// Add entry to EnumMap
func (em *EnumMap) Add(k string, v int) error {
	if k == "" {
		return fmt.Errorf("key is empty")
	}

	k = strings.ToLower(k)

	if _, ok := em.asString[v]; ok {
		return DuplicateValueError(v)
	}

	if _, ok := em.asInt[k]; ok {
		return DuplicateKeyError(k)
	}

	em.asInt[k] = v
	em.asString[v] = k

	return nil
}

// UnMarshal Get int enum representation from byte array
func (em *EnumMap) UnMarshal(arg []byte) (int, error) {
	var i int
	// Is integer
	if err := jsonit.Unmarshal(arg, &i); err == nil {
		return i, nil
	}

	// Is string
	var s string
	if err := jsonit.Unmarshal(arg, &s); err != nil {
		return 0, errors.Wrapf(err, "Failed to unmarshal byte array<%v>", arg)
	}

	i, err := em.Int(strings.ToLower(s))
	if err != nil {
		return 0, errors.Wrapf(err, "Unknown enum<%s>", s)
	}
	return i, nil
}

// ForEach execute function for each enum entry
func (em *EnumMap) ForEach(f func(k int, v string)) {
	if em == nil || len(em.asString) == 0 {
		return
	}
	for k, v := range em.asString {
		f(k, v)
	}
}

// ForEachSorted execute function for each enum with strict ordering
// this is somewhat heavier than using regular ForEach
func (em *EnumMap) ForEachSorted(f func(k int, v string)) {
	if em == nil || len(em.asString) == 0 {
		return
	}

	for _, k := range em.sortedIntKeys() {
		f(k, em.asString[k])
	}
}

func (em *EnumMap) sortedIntKeys() []int {
	keys := make([]int, 0, len(em.asString))
	for k := range em.asString {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}

func (em *EnumMap) Keys() []string {
	keys := make([]string, 0, len(em.asInt))
	for k := range em.asInt {
		keys = append(keys, k)
	}
	return keys
}
