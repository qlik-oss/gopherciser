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
		// MaxFrameSize (Default 0 - No limit). Max size in bytes to be read on sense websocket. Limit exceeded yields an error.
		MaxFrameSize int64 `json:"maxframesize" doc-key:"config.connectionSettings.maxframesize"`

		syncTemplates sync.Once
		templates     map[string]*template.Template

		syncEngineUrl sync.Once
		engineUrl     *url.URL

		syncRestUrl sync.Once
		restUrl     *url.URL

		csrfToken string
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
func (connectionSettings *ConnectionSettings) GetConnectFunc(state *session.State, appGUID, externalhost string, customHeaders http.Header, timeout time.Duration) (ConnectFunc, error) {
	header, err := connectionSettings.GetHeaders(state, externalhost)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	switch connectionSettings.Mode {
	case JWT:
		return connectionSettings.JwtSettings.GetConnectFunc(state, connectionSettings, appGUID, externalhost, header, customHeaders, timeout), nil
	case WS:
		return connectionSettings.WsSettings.GetConnectFunc(state, connectionSettings, appGUID, externalhost, header, customHeaders, timeout), nil
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

func (connectionSettings *ConnectionSettings) parseRestUrl() error {
	var parseError error
	connectionSettings.syncRestUrl.Do(func() {
		server := connectionSettings.Server
		if !strings.Contains(connectionSettings.Server, "://") {
			server = "http"
			if connectionSettings.Security {
				server += "s"
			}
			server += "://" + connectionSettings.Server
		}
		urlObj, err := url.Parse(server)
		if err != nil {
			parseError = err
			return
		}
		host := strings.Split(urlObj.Host+urlObj.Path, "/")[0]
		if host == "" {
			parseError = errors.Errorf("Failed to extract hostname from <%v>", connectionSettings.Server)
			return
		}

		restProtocol := "http://"
		if connectionSettings.Security {
			restProtocol = "https://"
		}

		var restUrlString string
		if connectionSettings.Port > 0 {
			restUrlString = fmt.Sprintf("%v%v:%d", restProtocol, host, connectionSettings.Port)
		} else {
			restUrlString = fmt.Sprintf("%v%v", restProtocol, host)
		}
		// TODO rewrite above code

		connectionSettings.restUrl, parseError = url.Parse(restUrlString)
	})
	return parseError
}

// Host, returns host or host:port
func (connectionSettings *ConnectionSettings) Host() (string, error) {
	if err := connectionSettings.parseRestUrl(); err != nil {
		return "", err
	}

	return connectionSettings.restUrl.Host, nil
}

// GetHost get hostname
// deprecated: use Host()
func (connectionSettings *ConnectionSettings) GetHost() (string, error) {
	return connectionSettings.Host()
}

// GetRestUrl
func (connectionSettings *ConnectionSettings) GetRestUrl() (string, error) {
	if err := connectionSettings.parseRestUrl(); err != nil {
		return "", err
	}
	return connectionSettings.restUrl.String(), nil
}

// GetEngineUrl get websocket URL
func (connectionSettings *ConnectionSettings) GetEngineUrl(appGUID, externalhost string) (*url.URL, error) {

	var err error
	connectionSettings.syncEngineUrl.Do(func() {
		if connectionSettings.RawURL != "" {
			connectionSettings.engineUrl, err = url.Parse(connectionSettings.RawURL)
			return
		}

		// Remove protocol
		var buildUrl string
		if externalhost == "" {
			buildUrl = connectionSettings.Server
		} else {
			buildUrl = externalhost
		}

		splitUrl := strings.Split(buildUrl, "://")
		if len(splitUrl) > 1 {
			buildUrl = splitUrl[1]
		}

		// Remove trailing path
		pathIndex := strings.IndexRune(buildUrl, '/')
		if pathIndex > -1 {
			buildUrl = buildUrl[0:pathIndex]
		}

		// Set protocol
		port := connectionSettings.Port
		if connectionSettings.Security {
			buildUrl = "wss://" + buildUrl
			if port < 1 {
				port = 443
			}
		} else {
			buildUrl = "ws://" + buildUrl
			if port < 1 {
				port = 80
			}
		}

		if connectionSettings.AppExt == nil {
			AppExt := "app"
			connectionSettings.AppExt = &AppExt
		}

		// Add port
		buildUrl += ":" + strconv.Itoa(port)

		// TODO rewrite above code
		connectionSettings.engineUrl, err = url.Parse(buildUrl)
	})

	if connectionSettings.engineUrl == nil || err != nil {
		return nil, err
	}

	// clone url
	engineUrl, err := url.Parse(connectionSettings.engineUrl.String())
	if err != nil {
		return nil, err
	}

	if externalhost != "" {
		engineUrl.Host = externalhost // TODO verify port part, maybe even parse verify externalhost?
	}

	engineUrl = engineUrl.JoinPath(connectionSettings.VirtualProxy, *connectionSettings.AppExt, appGUID)
	// Add virtual proxy
	// if connectionSettings.VirtualProxy != "" {
	// 	buildUrl += "/" + connectionSettings.VirtualProxy
	// }

	// Add path to app
	// AppExt := *connectionSettings.AppExt
	// if *connectionSettings.AppExt != "" {
	// 	AppExt = *connectionSettings.AppExt + "/"
	// }
	// if appGUID != "" {
	// 	buildUrl += "/" + AppExt + appGUID
	// } else {
	// 	buildUrl += "/" + AppExt
	// }

	if connectionSettings.csrfToken != "" {
		query := engineUrl.Query()
		query.Add("qlik-csrf-token", connectionSettings.csrfToken)
		engineUrl.RawQuery = query.Encode()
		// buildUrl += "?qlik-csrf-token=" + connectionSettings.csrfToken
	}

	return engineUrl, err
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

		header.Set(k, buf.String())
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

// SetCSRFToken setting token that will be added to connect url when connecting to engine
func (connectionSettings *ConnectionSettings) SetCSRFToken(token string) {
	connectionSettings.csrfToken = token
}
