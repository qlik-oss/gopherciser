package connection

import (
	"net/http"
	"testing"

	"github.com/goccy/go-json"
	"github.com/qlik-oss/gopherciser/users"
)

func TestWsConnection(t *testing.T) {

	raw := `{
			"server" : "myhost",
			"mode" : "ws",
			"headers" : {
				"static" : "staticstuff",
				"user" : "{{.UserName}}",
				"combined" : "{{.Directory}}\\\\{{.UserName}}"
			}
		}`

	var connection ConnectionSettings
	if err := json.Unmarshal([]byte(raw), &connection); err != nil {
		t.Fatal("failed to unmarshal connectionsettings:", err)
	}

	if connection.Mode != WS {
		t.Errorf("expected mode %d got %d", WS, connection.Mode)
	}

	if connection.Server != "myhost" {
		t.Errorf("expected host<myhost>, got host<%s>", connection.Server)
	}

	data := &users.User{
		UserName:  "user1",
		Password:  "password1",
		Directory: "directory1",
	}

	header, err := connection.addReqHeaders(data, nil)
	if err != nil {
		t.Error(err)
	}
	validateHeaders(t, header)

	header = make(http.Header, 1)
	header.Set("extra", "extra1")
	header, err = connection.addReqHeaders(data, header)
	if err != nil {
		t.Error(err)
	}
	validateHeaders(t, header)
	extra := header.Get("extra")
	if extra != "extra1" {
		t.Errorf("unexpected extra header<%s>, expected<extra1>", extra)
	}
}

func validateHeaders(t *testing.T, header http.Header) {
	t.Helper()

	static := header.Get("static")
	if static != "staticstuff" {
		t.Errorf("unexpected static header<%s>, expected<staticstuff>", static)
	}

	user := header.Get("user")
	if user != "user1" {
		t.Errorf("unexpected user header<%s>, expected<user1>", user)
	}

	combined := header.Get("combined")
	if combined != "directory1\\\\user1" {
		t.Errorf("unexpected combined header<%s>, expected<directory1\\\\user1>", combined)
	}
}
