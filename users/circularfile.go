package users

import (
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/helpers"
)

type (
	CircularUsersFileCore struct {
		Filename  helpers.RowFile `json:"filename" displayname:"Filename"`
		Password  Password        `json:"password,omitempty" displayname:"Password"`
		Directory string          `json:"directory,omitempty" displayname:"User directory"`
	}

	CircularUsersFile struct {
		CircularUsersFileCore

		userList []*User
		mu       sync.Mutex
		fill     sync.Once
	}
)

const (
	DefaultUserfileSeparator = ";"
)

// NewCircularUsersFromFile populate user list from file
func NewCircularUsersFromFile() *CircularUsersFile {
	return &CircularUsersFile{}
}

// UnmarshalJSON CircularUsersFile
func (users *CircularUsersFile) UnmarshalJSON(arg []byte) error {
	if err := jsonit.Unmarshal(arg, &users.CircularUsersFileCore); err != nil {
		return err
	}
	if err := users.parseUserList(); err != nil {
		return err
	}

	return nil
}

// Validate CircularUsersFile settings
func (users *CircularUsersFile) Validate() error {
	if users.Filename.IsEmpty() {
		return errors.Errorf("filename required for mode<%s>", UserGeneratorCircularFile)
	}
	if len(users.userList) < 1 && len(users.Filename.Rows()) < 1 {
		return errors.Errorf("filename<%s> has no user names", users.Filename)
	}
	return users.parseUserList()
}

// Iterate users from file
func (users *CircularUsersFile) Iterate(iteration uint64) *User {
	return users.userList[(iteration-1)%uint64(len(users.userList))]
}

func (users *CircularUsersFile) parseUserList() error {
	var errParse error
	users.fill.Do(func() {
		users.mu.Lock()
		defer users.mu.Unlock()
		rows := users.Filename.Rows()
		users.userList = make([]*User, 0, len(rows))
		for i, row := range rows {
			user, err := parseRow(row, users.Directory, users.Password)
			if err != nil {
				errParse = errors.Wrapf(err, "row:%d<%s> not correctly formated", i+1, row)
				return
			}
			users.userList = append(users.userList, user)
		}
		users.Filename.PurgeRows()
	})
	return errParse
}

func parseRow(row, defaultDirectory string, defaultPassword Password) (*User, error) {
	if row == "" {
		return nil, errors.New("row is empty")
	}

	subStrs := strings.SplitN(row, DefaultUserfileSeparator, 3)
	user := &User{
		UserName: subStrs[0],
	}

	if len(subStrs) > 1 && len(subStrs[1]) > 0 {
		user.Directory = subStrs[1]
	} else {
		user.Directory = defaultDirectory
	}

	if len(subStrs) > 2 && len(subStrs[2]) > 0 {
		user.Password = Password(subStrs[2])
	} else {
		user.Password = defaultPassword
	}

	return user, nil
}
