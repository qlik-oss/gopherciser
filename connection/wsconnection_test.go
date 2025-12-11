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

func TestConnectionStrings(t *testing.T) {
	type ExpectedValues struct {
		Server    string
		Mode      AuthenticationMode
		Url       string
		EngineUrl string
	}
	tests := []struct {
		Name         string
		Raw          []byte
		AppGUID      string
		ExternalHost string
		Expected     ExpectedValues
	}{
		{
			Name:    "non clean server",
			Raw:     []byte(`{"server" : "https://myhost.example","mode" : "ws"}`),
			AppGUID: "appGUID1234",
			Expected: ExpectedValues{
				Server:    "myhost.example",
				Mode:      WS,
				Url:       "http://myhost.example",
				EngineUrl: "ws://myhost.example:80/app/appGUID1234",
			},
		},
		{
			Name:    "basic host with domain",
			Raw:     []byte(`{"server" : "myhost.example","mode" : "ws"}`),
			AppGUID: "appGUID1234",
			Expected: ExpectedValues{
				Server:    "myhost.example",
				Mode:      WS,
				Url:       "http://myhost.example",
				EngineUrl: "ws://myhost.example:80/app/appGUID1234",
			},
		},
		{
			Name:    "basic host",
			Raw:     []byte(`{"server" : "myhost","mode" : "ws"}`),
			AppGUID: "appGUID1234",
			Expected: ExpectedValues{
				Server:    "myhost",
				Mode:      WS,
				Url:       "http://myhost",
				EngineUrl: "ws://myhost:80/app/appGUID1234",
			},
		},
		{
			Name:    "https host",
			Raw:     []byte(`{"server" : "myhost","mode" : "ws", "security": true}`),
			AppGUID: "appGUID1234",
			Expected: ExpectedValues{
				Server:    "myhost",
				Mode:      WS,
				Url:       "https://myhost",
				EngineUrl: "wss://myhost:443/app/appGUID1234",
			},
		},
		{
			Name:    "https host with port",
			Raw:     []byte(`{"server" : "myhost:4321","mode" : "ws", "security": true, "port": 1234}`),
			AppGUID: "appGUID1234",
			Expected: ExpectedValues{
				Server:    "myhost",
				Mode:      WS,
				Url:       "https://myhost:1234",
				EngineUrl: "wss://myhost:1234/app/appGUID1234",
			},
		},
		{
			Name:         "external host",
			Raw:          []byte(`{"server" : "myhost","mode" : "ws", "security": true, "port": 1234}`),
			AppGUID:      "appGUID1234",
			ExternalHost: "myexternalserver",
			Expected: ExpectedValues{
				Server:    "myhost",
				Mode:      WS,
				Url:       "https://myhost:1234",
				EngineUrl: "wss://myexternalserver/app/appGUID1234",
			},
		},
		{
			Name:         "external host with porst",
			Raw:          []byte(`{"server" : "myhost","mode" : "ws", "security": true, "port": 1234}`),
			AppGUID:      "appGUID1234",
			ExternalHost: "myexternalserver:5678",
			Expected: ExpectedValues{
				Server:    "myhost",
				Mode:      WS,
				Url:       "https://myhost:1234",
				EngineUrl: "wss://myexternalserver:5678/app/appGUID1234",
			},
		},
		{
			Name:    "virtual proxy",
			Raw:     []byte(`{"server" : "myhost","mode" : "ws", "security": true, "virtualproxy":"myvp"}`),
			AppGUID: "appGUID1234",
			Expected: ExpectedValues{
				Server:    "myhost",
				Mode:      WS,
				Url:       "https://myhost/myvp",
				EngineUrl: "wss://myhost:443/myvp/app/appGUID1234",
			},
		},
		{
			Name:    "custom app extension",
			Raw:     []byte(`{"server" : "myhost","mode" : "ws", "security": true, "appext": "customapp"}`),
			AppGUID: "appGUID1234",
			Expected: ExpectedValues{
				Server:    "myhost",
				Mode:      WS,
				Url:       "https://myhost",
				EngineUrl: "wss://myhost:443/customapp/appGUID1234",
			},
		},
		{
			Name:    "no app extension",
			Raw:     []byte(`{"server" : "myhost","mode" : "ws", "security": true, "appext": ""}`),
			AppGUID: "appGUID1234",
			Expected: ExpectedValues{
				Server:    "myhost",
				Mode:      WS,
				Url:       "https://myhost",
				EngineUrl: "wss://myhost:443/appGUID1234",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var connectionSettings ConnectionSettings
			if err := json.Unmarshal([]byte(test.Raw), &connectionSettings); err != nil {
				t.Fatalf("failed to unmarshal connectionsettings<%s>: %v", string(test.Raw), err)
			}
			if connectionSettings.Mode != test.Expected.Mode {
				t.Errorf("expected mode %d got %d", test.Expected.Mode, connectionSettings.Mode)
			}
			if connectionSettings.Server != test.Expected.Server {
				t.Errorf("expected host<%s>, got host<%s>", test.Expected.Server, connectionSettings.Server)
			}
			restUrl, err := connectionSettings.GetRestUrl()
			if err != nil {
				t.Fatal(err)
			}
			if restUrl != test.Expected.Url {
				t.Errorf("expected url<%s> got<%s>", test.Expected.Url, restUrl)
			}
			engineUrl, err := connectionSettings.GetEngineUrl(test.AppGUID, test.ExternalHost)
			if err != nil {
				t.Fatal(err)
			}
			if engineUrl.String() != test.Expected.EngineUrl {
				t.Errorf("expected engine url<%s> got<%s>", test.Expected.EngineUrl, engineUrl.String())
			}
		})
	}

	// TODO multiple externalhost
	var connectionSettings ConnectionSettings
	if err := json.Unmarshal([]byte(`{"server" : "myhost","mode" : "ws", "security": true, "port": 1234}`), &connectionSettings); err != nil {
		t.Fatal("failed to unmarshal connectionsettings: ", err)
	}
	engineUrl, err := connectionSettings.GetEngineUrl("appGUID", "externalhost1:4321")
	if err != nil {
		t.Fatal(err)
	}
	expectedUrl := "wss://externalhost1:4321/app/appGUID"
	if engineUrl.String() != expectedUrl {
		t.Errorf("external host engineUrl<%s> != expected<%s>", engineUrl, expectedUrl)
	}

	// verify changes with another external host
	engineUrl, err = connectionSettings.GetEngineUrl("appGUID", "externalhost2:5678")
	if err != nil {
		t.Fatal(err)
	}
	expectedUrl = "wss://externalhost2:5678/app/appGUID"
	if engineUrl.String() != expectedUrl {
		t.Errorf("external host engineUrl<%s> != expected<%s>", engineUrl, expectedUrl)
	}
}
