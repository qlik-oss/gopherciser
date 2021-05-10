package users

import (
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	CircularUsersFile struct {
		Filename  helpers.RowFile `json:"filename" displayname:"Filename"`
		Password  Password        `json:"password,omitempty" displayname:"Password"`
		Directory string          `json:"directory,omitempty" displayname:"User directory"`
	}
)

// NewCircularUsersFromFile populate user list from file
func NewCircularUsersFromFile() *CircularUsersFile {
	return &CircularUsersFile{}
}

// Validate CircularUsersFile settings
func (users *CircularUsersFile) Validate() error {
	if users.Filename.IsEmpty() {
		return errors.Errorf("filename required for mode<%s>", UserGeneratorCircularFile)
	}
	if len(users.Filename.Rows()) < 1 {
		return errors.Errorf("filename<%s> has no user names", users.Filename)
	}

	// TODO validate each row conforms?
	return nil
}

// Iterate users from file
func (users *CircularUsersFile) Iterate(iteration uint64) *User {
	userRow := users.Filename.Rows()[(iteration-1)%uint64(len(users.Filename.Rows()))]

	return &User{
		UserName:  userRow,
		Password:  users.Password,
		Directory: users.Directory,
	}
}
