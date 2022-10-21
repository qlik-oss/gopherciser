package connection

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enummap"
	"github.com/qlik-oss/gopherciser/helpers"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/users"
)

type (
	AuthenticationMode int

	ConnectionSettings struct {
		// Mode authentication mode, either JWT or WS
		Mode AuthenticationMode `json:"mode" doc-key:"config.connectionSettings.mode"`
		// JwtSettings JWT mode specific settings
		JwtSettings *ConnectJWTSettings `json:"jwtsettings,omitempty" doc-key:"config.connectionSettings.jwtsettings"`
		// WsSettings WS mode specific settings
		WsSettings *ConnectWsSettings `json:"wssettings,omitempty" doc-key:"config.connectionSettings.wssettings"`
		// Server remote host
		Server string `json:"server" doc-key:"config.connectionSettings.server"`
		// VirtualProxy sense virtual proxy used (added to connect path)
		VirtualProxy string `json:"virtualproxy" doc-key:"config.connectionSettings.virtualproxy"`
		// RawURL used to specify custom path for connection to sense app
		RawURL string `json:"rawurl,omitempty" doc-key:"config.connectionSettings.rawurl"`
		// Port port to be used, defaults to 80 or 443 depending on Security flag
		Port int `json:"port,omitempty" doc-key:"config.connectionSettings.port"`
		// Security use TLS
		Security bool `json:"security" doc-key:"config.connectionSettings.security"`
		// Allowuntrusted certificates
		Allowuntrusted bool `json:"allowuntrusted" doc-key:"config.connectionSettings.allowuntrusted"`
		// AppExt : By making this a pointer, we can check whether it was initialized
		// so that if omitted, it defaults to "app", but can be explicitly set to an empty string as well
		AppExt *string `json:"appext,omitempty" doc-key:"config.connectionSettings.appext"`
		// Header headers to add on the websocket connection
		Headers map[string]string `json:"headers" doc-key:"config.connectionSettings.headers"`

		syncTemplates sync.Once
		templates     map[string]*template.Template
	}

	// ConnectFunc connects to a sense environment, set reconnect to true if it's a reconnect and session in engine
	// is expected. Returns App GUID.
	ConnectFunc func(reconnect bool) (string, error)
)

const (
	// JWT connect using JWT
	JWT AuthenticationMode = iota
	// WS connect websocket without auth
	WS
)

var (
	funcMap = template.FuncMap{
		"now": time.Now,
	}
)

func (value AuthenticationMode) GetEnumMap() *enummap.EnumMap {
	enumMap, _ := enummap.NewEnumMap(map[string]int{
		"jwt": int(JWT),
		"ws":  int(WS),
	})
	return enumMap
}

// UnmarshalJSON unmarshal AuthenticationMode
func (value *AuthenticationMode) UnmarshalJSON(arg []byte) error {
	i, err := value.GetEnumMap().UnMarshal(arg)
	if err != nil {
		return errors.Wrap(err, "Failed to unmarshal AuthenticationMode")
	}

	*value = AuthenticationMode(i)
	return nil
}

// MarshalJSON marshal AuthenticationMode type
func (value AuthenticationMode) MarshalJSON() ([]byte, error) {
	str, err := value.GetEnumMap().String(int(value))
	if err != nil {
		return nil, errors.Errorf("Unknown AuthenticationMode<%d>", value)
	}
	return []byte(fmt.Sprintf(`"%s"`, str)), nil
}

// Validate connection setttings
func (connectionSettings *ConnectionSettings) Validate() error {
	switch connectionSettings.Mode {
	case JWT:
		if err := connectionSettings.JwtSettings.Validate(); err != nil {
			return errors.WithStack(err)
		}
	case WS:
		if err := connectionSettings.WsSettings.Validate(); err != nil {
			return errors.WithStack(err)
		}
	default:
		return errors.Errorf("Unknown connection mode <%d>", connectionSettings.Mode)
	}

	if connectionSettings.RawURL != "" {
		if strings.HasPrefix(connectionSettings.RawURL, "wss://") ||
			strings.HasPrefix(connectionSettings.RawURL, "ws://") {
			return nil
		}
		return errors.Errorf("Invalid raw url<%s>, must have protocol wss:// or ws://", connectionSettings.RawURL)
	}

	return nil
}

// GetConnectFunc Get function for connecting to sense
func (connectionSettings *ConnectionSettings) GetConnectFunc(state *session.State, appGUID, externalhost string, customHeaders http.Header) (ConnectFunc, error) {
	header, err := connectionSettings.GetHeaders(state, externalhost)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	switch connectionSettings.Mode {
	case JWT:
		return connectionSettings.JwtSettings.GetConnectFunc(state, connectionSettings, appGUID, externalhost, header, customHeaders), nil
	case WS:
		return connectionSettings.WsSettings.GetConnectFunc(state, connectionSettings, appGUID, externalhost, header, customHeaders), nil
	default:
		return nil, errors.Errorf("Unknown connection mode <%d>", connectionSettings.Mode)
	}
}

// GetHeaders Get auth headers
func (connectionSettings *ConnectionSettings) GetHeaders(state *session.State, externalhost string) (http.Header, error) {
	var host string
	var err error
	if externalhost == "" {
		host, err = connectionSettings.GetHost()
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		host = externalhost
	}

	header := state.HeaderJar.GetHeader(host)
	if header != nil {
		return header, nil
	}

	header, err = connectionSettings.addReqHeaders(state.User, header)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	switch connectionSettings.Mode {
	case JWT:
		header, err = connectionSettings.JwtSettings.GetJwtHeader(state, header)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	case WS:
	default:
		return nil, errors.Errorf("Unknown connection mode <%d>", connectionSettings.Mode)
	}

	state.HeaderJar.SetHeader(host, header)

	return header, nil
}

// GetHost get hostname
func (connectionSettings *ConnectionSettings) GetHost() (string, error) {
	urlObj, err := url.Parse(connectionSettings.Server)
	if err != nil {
		return "", err
	}
	host := strings.Split(urlObj.Host+urlObj.Path, "/")[0]
	if host == "" {
		return "", errors.Errorf("Failed to extract hostname from <%v>", connectionSettings.Server)
	}

	return host, nil
}

func (connection *ConnectionSettings) GetRestUrl() (string, error) {
	restProtocol := "http://"
	if connection.Security {
		restProtocol = "https://"
	}

	host, err := connection.GetHost()
	if err != nil {
		return "", err
	}

	if connection.Port > 0 {
		return fmt.Sprintf("%v%v:%d", restProtocol, host, connection.Port), nil
	} else {
		return fmt.Sprintf("%v%v", restProtocol, host), nil
	}
}

// GetURL get websocket URL
func (connection *ConnectionSettings) GetURL(appGUID, externalhost string) (string, error) {
	if connection.RawURL != "" {
		return connection.RawURL, nil
	}

	// Remove protocol
	var url string
	if externalhost == "" {
		url = connection.Server
	} else {
		url = externalhost
	}

	splitUrl := strings.Split(url, "://")
	if len(splitUrl) > 1 {
		url = splitUrl[1]
	}

	// Remove trailing path
	pathIndex := strings.IndexRune(url, '/')
	if pathIndex > -1 {
		url = url[0:pathIndex]
	}

	// Set protocol
	if connection.Security {
		url = "wss://" + url
		if connection.Port < 1 {
			connection.Port = 443
		}
	} else {
		url = "ws://" + url
		if connection.Port < 1 {
			connection.Port = 80
		}
	}

	// Add port
	url += ":" + strconv.Itoa(connection.Port)

	// Add virtual proxy
	if connection.VirtualProxy != "" {
		url += "/" + connection.VirtualProxy
	}

	// Add path to app
	if connection.AppExt == nil {
		AppExt := "app"
		connection.AppExt = &AppExt
	}
	AppExt := *connection.AppExt
	if *connection.AppExt != "" {
		AppExt = *connection.AppExt + "/"
	}
	if appGUID != "" {
		url += "/" + AppExt + appGUID
	} else {
		url += "/" + AppExt
	}

	return url, nil
}

func (connectionSettings *ConnectionSettings) addReqHeaders(data *users.User, header http.Header) (http.Header, error) {
	if len(connectionSettings.Headers) < 1 {
		return header, nil
	}
	if err := connectionSettings.parseTemplates(); err != nil {
		return header, errors.WithStack(err)
	}

	if header == nil {
		header = make(http.Header, len(connectionSettings.Headers))
	}

	for k := range connectionSettings.Headers {
		tmpl := connectionSettings.templates[fmt.Sprintf("reqHead-%s", k)]
		if tmpl == nil {
			continue
		}

		buf := helpers.GlobalBufferPool.Get()
		defer helpers.GlobalBufferPool.Put(buf)
		if err := tmpl.Execute(buf, data); err != nil {
			return header, errors.Wrapf(err, "failed executing %s template", tmpl.Name())
		}

		header.Set(k, buf.String()) // todo decide to use add or set?
	}

	return header, nil
}

func (connectionSettings *ConnectionSettings) parseTemplates() error {
	var parseErr error

	connectionSettings.syncTemplates.Do(func() {
		if connectionSettings != nil && len(connectionSettings.Headers) > 0 {
			connectionSettings.templates = make(map[string]*template.Template, len(connectionSettings.Headers))
			for k, v := range connectionSettings.Headers {
				tmplKey := fmt.Sprintf("reqHead-%s", k)
				tmpl := template.New(tmplKey)
				if _, err := tmpl.Funcs(funcMap).Parse(v); err != nil {
					parseErr = errors.Wrapf(err, "error parsing header<%s> template", k)
					return
				}
				connectionSettings.templates[tmplKey] = tmpl
			}
		}

		// todo add templates for more connection parameters
	})

	return parseErr
}

// AllowUntrusted implements session.ConnectionSettings interface
func (connectionSettings *ConnectionSettings) AllowUntrusted() bool {
	return connectionSettings.Allowuntrusted
}
