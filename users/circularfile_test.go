package users

import (
	"fmt"
	"os"
	"testing"

	"github.com/goccy/go-json"
	"github.com/qlik-oss/gopherciser/helpers"
)

func TestUsersFromFile(t *testing.T) {
	f, err := os.CreateTemp("", "Users*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()

	_, err = f.WriteString(`testuser_1;NOTDEF
testuser_2
testuser_3;
testuser_4;NOTDEF;
testuser_5;NOTDEF;PassWort
testuser_6;;MYPass
testuser_7;MyDir;;;2323;
testuser_8;MyDir;Pass;;2323;`)
	if err != nil {
		t.Fatal(err)
	}

	jsn := []byte(`{
		"filename" : "` + f.Name() + `",
		"Directory": "DefaultDir",
		"Password": "DefaultPass"
	}`)

	var usergen CircularUsersFile
	if err := json.Unmarshal(jsn, &usergen); err != nil {
		t.Fatal(err)
	}

	expects := []*User{
		{UserName: "testuser_1", Password: helpers.Password("DefaultPass"), Directory: "NOTDEF"},
		{UserName: "testuser_2", Password: helpers.Password("DefaultPass"), Directory: "DefaultDir"},
		{UserName: "testuser_3", Password: helpers.Password("DefaultPass"), Directory: "DefaultDir"},
		{UserName: "testuser_4", Password: helpers.Password("DefaultPass"), Directory: "NOTDEF"},
		{UserName: "testuser_5", Password: helpers.Password("PassWort"), Directory: "NOTDEF"},
		{UserName: "testuser_6", Password: helpers.Password("MYPass"), Directory: "DefaultDir"},
		{UserName: "testuser_7", Password: helpers.Password(";;2323;"), Directory: "MyDir"},
		{UserName: "testuser_8", Password: helpers.Password("Pass;;2323;"), Directory: "MyDir"},
	}

	if len(usergen.userList) != len(expects) {
		t.Fatalf("user list contains<%d> users not %d", len(usergen.userList), len(expects))
	}

	for i, user := range usergen.userList {
		if err := userEquals(user, expects[i]); err != nil {
			t.Errorf("validating user:%d failed: %v", i, err)
		}
	}
}

func userEquals(user1, user2 *User) error {
	if user1.UserName != user2.UserName {
		return fmt.Errorf("username %s != %s", user1.UserName, user2.UserName)
	}

	if user1.Directory != user2.Directory {
		return fmt.Errorf("directory %s != %s", user1.Directory, user2.Directory)
	}

	if string(user1.Password) != string(user2.Password) {
		return fmt.Errorf("passwrod %s != %s", string(user1.Password), string(user2.Password))
	}

	return nil
}
