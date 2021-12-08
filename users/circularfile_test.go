package users

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestUsersFromFile(t *testing.T) {
	f, err := ioutil.TempFile("", "Users*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

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
		{UserName: "testuser_1", Password: Password("DefaultPass"), Directory: "NOTDEF"},
		{UserName: "testuser_2", Password: Password("DefaultPass"), Directory: "DefaultDir"},
		{UserName: "testuser_3", Password: Password("DefaultPass"), Directory: "DefaultDir"},
		{UserName: "testuser_4", Password: Password("DefaultPass"), Directory: "NOTDEF"},
		{UserName: "testuser_5", Password: Password("PassWort"), Directory: "NOTDEF"},
		{UserName: "testuser_6", Password: Password("MYPass"), Directory: "DefaultDir"},
		{UserName: "testuser_7", Password: Password(";;2323;"), Directory: "MyDir"},
		{UserName: "testuser_8", Password: Password("Pass;;2323;"), Directory: "MyDir"},
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
