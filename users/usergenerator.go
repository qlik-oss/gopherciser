package users

import (
	"encoding/json"
	"fmt"
	"runtime"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/statistics"
)

type (
	// Password
	Password string

	// Type type of scheduler
	Type int
	// Settings interface to be implemented by user generator
	Settings interface {
		Iterate(iteration uint64) *User
		Validate() error
	}

	GeneratorCore struct {
		GeneratorType Type `json:"type" displayname:"User generator type" doc-key:"config.loginSettings.type"`
	}

	generatorTmp struct {
		GeneratorCore
		Settings json.RawMessage `json:"settings"`
	}

	// UserGenerator of users
	UserGenerator struct {
		GeneratorCore
		Settings Settings `json:"settings" doc-key:"config.loginSettings.settings"`
	}
)

const (
	// UserGeneratorUnknown unknown user generator
	UserGeneratorUnknown Type = iota
	// UserGeneratorCircular users according to userlist
	UserGeneratorCircular
	// UserGeneratorPrefix users with a prefix and an enumeration
	UserGeneratorPrefix
	// UserGeneratorNone no user creation
	UserGeneratorNone
	// UserGeneratorCircularFile userlist read from file
	UserGeneratorCircularFile
)

var (
	userGeneratorTypeEnumMap = enummap.NewEnumMapOrPanic(map[string]int{
		"userlist": int(UserGeneratorCircular),
		"prefix":   int(UserGeneratorPrefix),
		"none":     int(UserGeneratorNone),
		"fromfile": int(UserGeneratorCircularFile),
	})
	jsonit = jsoniter.ConfigCompatibleWithStandardLibrary
)

func (value Type) GetEnumMap() *enummap.EnumMap {
	return userGeneratorTypeEnumMap
}

// String implements stringer interface
func (value Type) String() string {
	return userGeneratorTypeEnumMap.StringDefault(int(value), fmt.Sprintf("%d", value))
}

func UserGenHandler(generator Type) interface{} {
	switch generator {
	case UserGeneratorUnknown:
		return nil
	case UserGeneratorCircular:
		return NewCircularUsers()
	case UserGeneratorPrefix:
		return &PrefixUsers{}
	case UserGeneratorCircularFile:
		return NewCircularUsersFromFile()
	case UserGeneratorNone:
		return &NoneUsers{}
	default:
		return nil
	}
}

// NewUserGeneratorCircular create new circular user generator
func NewUserGeneratorCircular(users []*User) UserGenerator {
	circularUsers := NewCircularUsers()
	circularUsers.UserList = users

	uGen := UserGenerator{
		GeneratorCore{
			GeneratorType: UserGeneratorCircular,
		},
		circularUsers,
	}

	return uGen
}

// NewUserGeneratorPrefix create new prefix user generator
func NewUserGeneratorPrefix(prefix string) UserGenerator {
	return UserGenerator{
		GeneratorCore{
			GeneratorType: UserGeneratorPrefix,
		},
		&PrefixUsers{
			Prefix: prefix,
		},
	}
}

// NewUserGeneratorNone create new "none" user generator (to be used with e.g. Qlik Core)
func NewUserGeneratorNone() UserGenerator {
	return UserGenerator{
		GeneratorCore{
			GeneratorType: UserGeneratorNone,
		},
		&NoneUsers{},
	}
}

// UnmarshalJSON unmarshal password from json
func (passwd *Password) UnmarshalJSON(arg []byte) error {
	var s string
	if err := jsonit.Unmarshal(arg, &s); err != nil {
		return errors.Wrap(err, "failed to unmarshal password")
	}
	*passwd = Password(s)

	return nil
}

//  MarshalJSON marshal password to json: replace with ***
func (passwd *Password) MarshalJSON() ([]byte, error) {
	if runtime.GOOS != "js" {
		return []byte(`"***"`), nil
	} else {
		return []byte(fmt.Sprintf("\"%s\"", string(*passwd))), nil
	}
}

// UnmarshalJSON unmarshal user generator type from json
func (value *Type) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal Type")
	}

	*value = Type(i)

	return nil
}

// MarshalJSON marshal scheduler type to json
func (value Type) MarshalJSON() ([]byte, error) {
	str, err := (*value.GetEnumMap()).String(int(value))
	if err != nil {
		return nil, errors.Wrap(err, "Failed to get user generator type")
	}

	if str == "" {
		return nil, errors.Errorf("Unknown user generator type<%v>", value)
	}

	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// UnmarshalJSON unmarshal scheduler from json
func (value *UserGenerator) UnmarshalJSON(arg []byte) error {
	var gen generatorTmp
	if err := jsonit.Unmarshal(arg, &gen); err != nil {
		return errors.Wrap(err, "Failed to unmarshal user generator")
	}

	(*value).GeneratorType = gen.GeneratorType

	settings := UserGenHandler(gen.GeneratorType)
	if gen.GeneratorType == UserGeneratorNone {
		// Allow UserGenerator with no Settings for UserGeneratorNone type
		(*value).Settings = settings.(Settings)
		return nil
	}
	if err := jsonit.Unmarshal(gen.Settings, &settings); err != nil {
		return errors.Wrap(err, "Failed to unmarshal user generator settings")
	}

	if settings == nil {
		return errors.Errorf("Invalid user generator type")
	}

	var ok bool
	(*value).Settings, ok = settings.(Settings)
	if !ok {
		return errors.Errorf("Settings not of type Settings (%d)", gen.GeneratorType)
	}

	return nil
}

// GetNext user to simulate
func (value *UserGenerator) GetNext(counters *statistics.ExecutionCounters) *User {
	return value.Settings.Iterate(counters.Users.Inc())
}
