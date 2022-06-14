package helpers

import (
	"fmt"
	"runtime"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"
)

// This helper will confuscate a password field when unmarshaled for logfiles.
type (
	// Password
	Password string
)

// UnmarshalJSON unmarshal password from json
func (passwd *Password) UnmarshalJSON(arg []byte) error {
	var s string
	if err := json.Unmarshal(arg, &s); err != nil {
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
