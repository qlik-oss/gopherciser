package users

import (
	"errors"
	"fmt"
)

type PrefixUsers struct {
	Prefix    string `json:"prefix" displayname:"Prefix" doc-key:"config.loginSettings.settings.prefix"`
	Directory string `json:"directory,omitempty" displayname:"User directory" doc-key:"config.loginSettings.settings.directory"`
}

// Iterate returns the next user in a circular manner
func (users *PrefixUsers) Iterate(iteration uint64) *User {
	return &User{
		fmt.Sprintf("%s_%d", users.Prefix, iteration),
		"",
		users.Directory,
	}
}

// Validate validates settings
func (users *PrefixUsers) Validate() error {
	if users.Prefix == "" {
		return errors.New("login type<prefix> requires prefix parameter to be set")
	}
	return nil
}
