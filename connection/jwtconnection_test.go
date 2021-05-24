package connection

import (
	"fmt"
	"testing"

	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/synced"
	"github.com/qlik-oss/gopherciser/users"
)

func TestParsing(t *testing.T) {
	// simple claims
	stClaims, err := synced.New(`{
			"user": "{{.UserName}}",
			"directory": "{{.Directory}}"
		}`)
	if err != nil {
		t.Fatal(err)
	}

	settings := ConnectJWTSettings{
		Claims: *stClaims,
	}

	user := &users.User{
		UserName:  "mytestuser_1",
		Directory: "mydirectory",
	}

	sessionState := &session.State{
		User: user,
		LogEntry: &logger.LogEntry{
			Session: &logger.SessionEntry{},
			Action:  &logger.ActionEntry{},
		},
	}

	claims, err := settings.executeClaimsTemplates(sessionState)
	if err != nil {
		t.Fatal(err)
	}

	expected := user.UserName
	key := "user"
	value := fmt.Sprintf("%v", claims[key])
	validate(t, key, value, expected)

	expected = user.Directory
	key = "directory"
	value = fmt.Sprintf("%v", claims[key])
	validate(t, key, value, expected)

	stClaims, err = synced.New(`{
			"iat": {{now.Unix}},
			"exp": {{(now.Add 18000000000000).Unix}},
			"iss" : "https://qlik.api.interal",
			"aud" : "qlik.api",
			"sub": "custom",
			"name": "{{.UserName}}",
			"groups": ["group1", "group for user {{.UserName}}"]
		}`)
	if err != nil {
		t.Fatal(err)
	}

	// advanced claims
	settings = ConnectJWTSettings{
		Claims: *stClaims,
	}

	claims, err = settings.executeClaimsTemplates(sessionState)
	if err != nil {
		t.Fatal(err)
	}

	expected = "https://qlik.api.interal"
	key = "iss"
	value = fmt.Sprintf("%v", claims[key])
	validate(t, key, value, expected)

	expected = "qlik.api"
	key = "aud"
	value = fmt.Sprintf("%v", claims[key])
	validate(t, key, value, expected)

	expected = "custom"
	key = "sub"
	value = fmt.Sprintf("%v", claims[key])
	validate(t, key, value, expected)

	expected = user.UserName
	key = "name"
	value = fmt.Sprintf("%v", claims[key])
	validate(t, key, value, expected)

	expected = fmt.Sprintf("%v", []string{"group1", fmt.Sprintf("group for user %s", user.UserName)})
	key = "groups"
	value = fmt.Sprintf("%v", claims[key])
	validate(t, key, value, expected)

	key = "iat"
	if claims[key] == nil {
		t.Error(key, "not set")
	} else {
		v, ok := claims[key].(float64)
		if !ok {
			t.Error(key, "not a number")
		}
		if v < 1 {
			t.Error(key, "not set correctly, value:", v)
		}
	}

	key = "exp"
	if claims[key] == nil {
		t.Error(key, "not set")
	} else {
		v, ok := claims[key].(float64)
		if !ok {
			t.Error(key, "not a number")
		}
		if v < 1 {
			t.Error(key, "not set correctly, value:", v)
		}
	}

	stJWTHeader, err := synced.New("{\"kid\":\"{{.UserName}}-Key\"}")
	if err != nil {
		t.Fatal(err)
	}
	// test jwt header
	settings.JwtHeader = *stJWTHeader
	jwtHeader, err := settings.executeJWTHeaderTemplates(sessionState)
	if err != nil {
		t.Fatal("failed parsing jwtheader", err)
	}

	expected = fmt.Sprintf("%s-Key", user.UserName)
	key = "kid"
	value = fmt.Sprintf("%v", jwtHeader[key])
	validate(t, key, value, expected)
}

func validate(t *testing.T, key, value, expected string) {
	t.Helper()

	if value != expected {
		t.Errorf("key<%s> expected<%s> got<%s>", key, expected, value)
	}
}
