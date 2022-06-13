package users

import (
	"fmt"
	"sync"
	"testing"

	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/stretchr/testify/assert"
)

var (
	someUserNames = []string{"one", "two", "three", "four"}
	somePassword  = "five"

	emptyCircularUsers = &CircularUsers{}
)

// createUserList matches user names to the only password
func createUserList(userNames []string, password string) []*User {
	if userNames == nil {
		return nil
	}

	users := make([]*User, len(userNames))
	for idx, uName := range userNames {
		users[idx] = &User{
			UserName: uName,
			Password: helpers.Password(password),
		}
	}
	return users
}

func TestNewCircularUsers(t *testing.T) {
	users := NewCircularUsers()
	users.UserList = createUserList(someUserNames, somePassword)
	assert.NotNil(t, users)
	assert.IsType(t, emptyCircularUsers, users)
	assert.Equal(t, len(users.UserList), len(someUserNames))
}

func TestNewCircularUsers_nilUserNames(t *testing.T) {
	users := NewCircularUsers()
	users.UserList = createUserList(nil, somePassword)
	assert.NotNil(t, users)
	assert.IsType(t, emptyCircularUsers, users)
	assert.Nil(t, users.UserList)
}

func TestNewCircularUsers_emptyPassword(t *testing.T) {
	users := NewCircularUsers()
	users.UserList = createUserList(someUserNames, somePassword)
	assert.NotNil(t, users)
	assert.IsType(t, emptyCircularUsers, users)
	assert.Equal(t, len(users.UserList), len(someUserNames))
}

func TestCircularUsers_Iterate(t *testing.T) {
	users := NewCircularUsers()
	users.UserList = createUserList(someUserNames, somePassword)
	assert.NotNil(t, users)
	u := users.Iterate(1)
	assert.NotNil(t, u)
}

func TestCircularUsers_Iterate_nil(t *testing.T) {
	users := NewCircularUsers()
	users.UserList = createUserList(someUserNames, somePassword)
	users.UserList = nil
	u := users.Iterate(1)
	assert.Nil(t, u)
	u = users.Iterate(0)
	assert.Nil(t, u)
}

func TestCircularUsers_Iterate_concurrent(t *testing.T) {
	users := NewCircularUsers()
	users.UserList = createUserList(someUserNames, somePassword)
	assert.NotNil(t, users)

	wg := sync.WaitGroup{}
	for i := 0; i < len(users.UserList)*2; i++ {
		wg.Add(1)
		iteration := uint64(i) + 1
		go func() {
			u := users.Iterate(iteration)
			assert.NotNil(t, u)
			assert.Equal(t, users.UserList[(int(iteration)-1)%len(users.UserList)], u)
			wg.Done()
		}()
	}
	wg.Wait()

	if t.Failed() {
		// log users list
		usersString := ""
		for i, u := range users.UserList {
			if i == 0 {
				usersString += "["
			}
			usersString += fmt.Sprintf("%v", u)
			if i == len(users.UserList) {
				usersString += "]"
			} else {
				usersString += " "
			}
		}
		t.Log(usersString)
	}
}
