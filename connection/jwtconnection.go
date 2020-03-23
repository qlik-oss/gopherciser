package connection

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/session"
)

type (

	// ConnectJWTSettings app and server settings using JWT
	ConnectJWTSettings struct {
		// KeyPath path to jwt signing key
		KeyPath string `json:"keypath,omitempty" doc-key:"config.connectionSettings.jwtsettings.keypath"`
		// JwtHeader JWT headers as escaped json string. Custom headers to be added to the JWT header.
		// The strings for JwtHeader and Claims will be processed as a GO template
		// where User struct can be used
		//  struct {
		//	UserName  string
		// 	Password  string
		// 	Directory string
		// }
		// as well as the function "now" which corresponds to time.Now
		// The entries for "alg" and "typ" are added automatically to the header.
		// E.g. to add a key id header, "kid" add the following string
		// "{\"kid\":\"myKeyId\"}"
		JwtHeader session.SyncedTemplate `json:"jwtheader,omitempty" doc-key:"config.connectionSettings.jwtsettings.jwtheader"`
		// Claims JWT claims as escaped json string. E.g. for an on prem JWT auth (with user and directory set as keys in QMC):
		// "{\"user\": \"{{.UserName}}\",\"directory\": \"{{.Directory}}\"}"
		// to add "iat":
		// "{\"iat\":{{now.Unix}}}"
		// or to add "exp" with 5 hours expiration
		// "{\"exp\":{{(now.Add 18000000000000).Unix}}}"
		Claims session.SyncedTemplate `json:"claims,omitempty" doc-key:"config.connectionSettings.jwtsettings.claims"`

		// Alg is the signing method to be used for the JWT. Defaults to RS512 if omitted
		Alg string `json:"alg,omitempty" doc-key:"config.connectionSettings.jwtsettings.alg"`

		// handle jwt private key
		key     []byte
		readKey sync.Once
	}
)

// GetConnectFunc which establishes a connection to Qlik Sense
func (connectJWT *ConnectJWTSettings) GetConnectFunc(sessionState *session.State, connection *ConnectionSettings, appGUID string, headers http.Header) func() (string, error) {
	connectFunc := func() (string, error) {
		url, err := connection.GetURL(appGUID)

		if err != nil {
			return appGUID, errors.WithStack(err)
		}

		// Create sense uplink
		if sessionState.Connection == nil {
			sessionState.Connection = new(enigmahandlers.SenseConnection)
		} else {
			sessionState.Disconnect()
		}

		sense := enigmahandlers.NewSenseUplink(sessionState.BaseContext(), sessionState.LogEntry, sessionState.RequestMetrics, sessionState.TrafficLogger())
		sessionState.Connection.SetSense(sense)
		sense.OnUnexpectedDisconnect(sessionState.WSFailed)

		// Connect
		ctx, cancel := sessionState.ContextWithTimeout(sessionState.BaseContext())
		defer cancel()

		if sessionState.Cookies == nil {
			sessionState.Cookies, err = cookiejar.New(nil)
			if err != nil {
				return appGUID, errors.Wrap(err, "failed creating cookie jar")
			}
		}
		if err = sense.Connect(ctx, url, headers, sessionState.Cookies, connection.Allowuntrusted, sessionState.Timeout); err != nil {
			return appGUID, errors.WithStack(err)
		}

		return appGUID, nil
	}

	return connectFunc
}

// Validate connectJWTSettings
func (connectJWT *ConnectJWTSettings) Validate() error {
	// Do we have a key? (if so, also read into memory)
	key, err := connectJWT.getPrivateKey()
	if err != nil {
		return errors.Wrapf(err, "Error reading private key from file<%s>", connectJWT.KeyPath)
	}

	if len(key) < 1 {
		return errors.Errorf("No key in keyfile")
	}
	return nil
}

// GetJwtHeader get Authorization header
func (connectJWT *ConnectJWTSettings) GetJwtHeader(sessionState *session.State, header http.Header) (http.Header, error) {
	// Do we have a key? (then read into memory)
	key, err := connectJWT.getPrivateKey()
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading private key from file<%s>", connectJWT.KeyPath)
	}

	// replace variables in jwt claims and create token
	claims, errClaims := connectJWT.executeClaimsTemplates(sessionState)
	if errClaims != nil {
		return nil, errors.WithStack(errClaims)
	}
	alg := connectJWT.Alg
	if alg == "" {
		alg = "RS512"
	}
	signingMethod := jwt.GetSigningMethod(alg)
	if signingMethod == nil {
		return nil, errors.Errorf("Unknown signing method<%s>", alg)
	}
	token := jwt.NewWithClaims(signingMethod, jwt.MapClaims(claims))

	// replace variables and set jwt headers
	jwtHeader, errJwtHeader := connectJWT.executeJWTHeaderTemplates(sessionState)
	if errJwtHeader != nil {
		return nil, errors.WithStack(errJwtHeader)
	}
	for k, v := range jwtHeader {
		token.Header[k] = v
	}

	// sign JWT
	signedToken, err := GetSignedJwtToken(key, token)
	if err != nil {
		return nil, errors.Wrapf(err, "Error signing token with key from file<%s>", connectJWT.KeyPath)
	}

	// set request headers
	if header == nil {
		header = make(http.Header, 1)
	}
	header.Set("Authorization", fmt.Sprintf("Bearer %s", signedToken))

	return header, err
}

func parseAlgo(key []byte) (string, error) {
	str := fmt.Sprintf("%s", key)
	startMarker := "BEGIN "
	endMarker := " PRIVATE"
	startIndex := strings.Index(str, startMarker) + len(startMarker)
	endIndex := strings.Index(str, endMarker)
	diff := endIndex - startIndex
	if startIndex == -1 || endIndex == -1 || diff < 1 {
		return "", errors.New("algorithm not declared")
	}
	algo := str[startIndex:endIndex]
	return algo, nil
}

func (connectJWT *ConnectJWTSettings) executeClaimsTemplates(sessionState *session.State) (map[string]interface{}, error) {
	if connectJWT.Claims.String() == "" {
		return nil, nil
	}

	claims, errClaims := sessionState.ReplaceSessionVariables(&connectJWT.Claims)
	if errClaims != nil {
		return nil, errors.WithStack(errClaims)
	}

	var m map[string]interface{}
	if err := jsonit.Unmarshal([]byte(claims), &m); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal JWT Claims<%s>", claims)
	}
	return m, nil
}

func (connectJWT *ConnectJWTSettings) executeJWTHeaderTemplates(sessionState *session.State) (map[string]interface{}, error) {
	if connectJWT.JwtHeader.String() == "" {
		return nil, nil
	}

	jwtHeader, errJWTHeader := sessionState.ReplaceSessionVariables(&connectJWT.JwtHeader)
	if errJWTHeader != nil {
		return nil, errors.WithStack(errJWTHeader)
	}

	var m map[string]interface{}
	if err := jsonit.Unmarshal([]byte(jwtHeader), &m); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal JWT Header<%s>", jwtHeader)
	}
	return m, nil
}

func (connectJWT *ConnectJWTSettings) getPrivateKey() ([]byte, error) {
	// read private key into memory
	var readKeyErr error
	connectJWT.readKey.Do(func() {
		connectJWT.key, readKeyErr = ioutil.ReadFile(connectJWT.KeyPath)
	})

	if readKeyErr != nil {
		return nil, errors.Wrapf(readKeyErr, "Error reading private key from file<%s>", connectJWT.KeyPath)
	}

	return connectJWT.key, nil
}

// GetSignedJwtToken Sign token with key
func GetSignedJwtToken(key []byte, token *jwt.Token) (string, error) {
	if key == nil {
		return "", errors.Errorf("No jwt key provided")
	}

	parsedKeyFormat, err := parseAlgo(key)
	var privKey interface{}
	if err == nil && parsedKeyFormat == "EC" {
		privKey, err = jwt.ParseECPrivateKeyFromPEM(key)
	} else { // Key is either RSA *or* fallback
		privKey, err = jwt.ParseRSAPrivateKeyFromPEM(key)
	}

	if err != nil {
		return "", errors.Wrap(err, "Error parsing private key")
	}

	signedToken, err := token.SignedString(privKey)
	if err != nil {
		return "", errors.Wrapf(err, "Error getting signed token of key")
	}

	return signedToken, nil
}
