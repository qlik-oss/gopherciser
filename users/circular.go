package users

import (
	"errors"
	"sync"

	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	User struct {
		UserName  string           `json:"username" displayname:"Username"`
		Password  helpers.Password `json:"password,omitempty" displayname:"Password"`
		Directory string           `json:"directory,omitempty" displayname:"User directory"`
	}

	CircularUsers struct {
		mtx       *sync.Mutex
		UserList  []*User          `json:"userlist" displayname:"User list"`
		Password  helpers.Password `json:"password,omitempty" displayname:"Password"`
		Directory string           `json:"directory,omitempty" displayname:"User directory"`
	}
)

// NewCircularUsers creates a circular list of users
func NewCircularUsers() *CircularUsers {
	return &CircularUsers{
		mtx:      &sync.Mutex{},
		UserList: []*User{{"", "", ""}},
	}
}

// Iterate returns the next user in a circular manner, iteration should always be > 0
func (users *CircularUsers) Iterate(iteration uint64) *User {
	if users == nil || users.UserList == nil || iteration < 1 {
		return nil
	}

	users.mtx.Lock()
	defer users.mtx.Unlock()

	user := users.UserList[(iteration-1)%uint64(len(users.UserList))]

	if user.Password == "" {
		user.Password = users.Password
	}
	if user.Directory == "" {
		user.Directory = users.Directory
	}
	return user
}

// Validate validates settings
func (users *CircularUsers) Validate() error {
	if len(users.UserList) < 1 {
		return errors.New("login type<userlist> requires a non empty userlist")
	}
	return nil
}
