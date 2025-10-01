package connection

import (
	"fmt"
	"maps"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/enigmahandlers"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/synced"
)

type (

	// ConnectJWTSettings app and server settings using JWT
	ConnectJWTSettingsCore struct {
		// KeyPath path to jwt signing key
		KeyPath string `json:"keypath,omitempty" doc-key:"config.connectionSettings.jwtsettings.keypath" displayname:"Key Path"`
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
		JwtHeader synced.Template `json:"jwtheader" doc-key:"config.connectionSettings.jwtsettings.jwtheader"  displayname:"JWT Header"`
		// Claims JWT claims as escaped json string. E.g. for an on prem JWT auth (with user and directory set as keys in QMC):
		// "{\"user\": \"{{.UserName}}\",\"directory\": \"{{.Directory}}\"}"
		// to add "iat":
		// "{\"iat\":{{now.Unix}}}"
		// or to add "exp" with 5 hours expiration
		// "{\"exp\":{{(now.Add 18000000000000).Unix}}}"
		Claims synced.Template `json:"claims" doc-key:"config.connectionSettings.jwtsettings.claims" displayname:"Claims"`

		// Alg is the signing method to be used for the JWT. Defaults to RS512 if omitted
		Alg string `json:"alg,omitempty" doc-key:"config.connectionSettings.jwtsettings.alg" displayname:"Algoritm"`
	}

	ConnectJWTSettings struct {
		ConnectJWTSettingsCore
		key           any // parsed private key
		signingMethod jwt.SigningMethod
	}
)

func (connectJWT *ConnectJWTSettings) UnmarshalJSON(arg []byte) error {
	if err := json.Unmarshal(arg, &connectJWT.ConnectJWTSettingsCore); err != nil {
		return err
	}

	key, err := os.ReadFile(connectJWT.KeyPath)
	if err != nil {
		return errors.Wrapf(err, "error reading private key from file<%s>", connectJWT.KeyPath)
	}

	if connectJWT.Alg != "" {
		connectJWT.signingMethod = jwt.GetSigningMethod(connectJWT.Alg)
		if connectJWT.signingMethod == nil {
			return errors.Errorf("unknown signing method<%s>", connectJWT.Alg)
		}
		var err error
		switch connectJWT.signingMethod {
		case jwt.SigningMethodES256, jwt.SigningMethodES384, jwt.SigningMethodES512:
			connectJWT.key, err = jwt.ParseECPrivateKeyFromPEM(key)
		case jwt.SigningMethodEdDSA:
			connectJWT.key, err = jwt.ParseEdPrivateKeyFromPEM(key)
		case jwt.SigningMethodRS256, jwt.SigningMethodRS384, jwt.SigningMethodRS512, jwt.SigningMethodPS256, jwt.SigningMethodPS384, jwt.SigningMethodPS512:
			connectJWT.key, err = jwt.ParseRSAPrivateKeyFromPEM(key)
		case jwt.SigningMethodNone:
		default:
			err = errors.Errorf("alg<%s> not supported", connectJWT.signingMethod.Alg())
		}
		if err != nil {
			return errors.Wrap(err, "Error parsing private key")
		}
	} else {
		// Discover from key
		connectJWT.signingMethod = jwt.SigningMethodRS512
		connectJWT.key, err = jwt.ParseRSAPrivateKeyFromPEM(key)
		if err != nil {
			connectJWT.signingMethod = jwt.SigningMethodEdDSA
			connectJWT.key, err = jwt.ParseEdPrivateKeyFromPEM(key)
			if err != nil {
				key, err := jwt.ParseECPrivateKeyFromPEM(key)
				if err != nil {
					return errors.Errorf("no alg defined and could not autodetect private key type")
				}
				connectJWT.key = key
				switch key.Curve.Params().Name {
				case "P-256":
					connectJWT.signingMethod = jwt.SigningMethodES256
				case "P-384":
					connectJWT.signingMethod = jwt.SigningMethodES384
				case "P-521":
					connectJWT.signingMethod = jwt.SigningMethodES512
				}
			}
		}
	}

	return nil
}

// GetConnectFunc which establishes a connection to Qlik Sense
func (connectJWT *ConnectJWTSettings) GetConnectFunc(sessionState *session.State, connectionSettings *ConnectionSettings, appGUID, externalhost string, headers, customHeaders http.Header) ConnectFunc {
	connectFunc := func(reconnect bool) (string, error) {
		url, err := connectionSettings.GetURL(appGUID, externalhost)
		if err != nil {
			return appGUID, errors.WithStack(err)
		}

		// Create sense uplink
		if sessionState.Connection == nil {
			sessionState.Connection = new(enigmahandlers.SenseConnection)
		} else {
			sessionState.Disconnect()
		}

		sense := enigmahandlers.NewSenseUplink(sessionState.BaseContext(), sessionState.LogEntry, sessionState.RequestMetrics, sessionState.TrafficLogger(), connectionSettings.MaxFrameSize)
		sessionState.Connection.SetSense(sense)

		// Connect
		ctx, cancel := sessionState.ContextWithTimeout(sessionState.BaseContext())
		defer cancel()

		if sessionState.Cookies == nil {
			sessionState.Cookies, err = cookiejar.New(nil)
			if err != nil {
				return appGUID, errors.Wrap(err, "failed creating cookie jar")
			}
		}

		// combine headers for connection
		connectHeaders := make(http.Header)
		for k, v := range headers {
			connectHeaders[k] = v
		}
		for k, v := range customHeaders {
			connectHeaders[k] = v
		}
		if err = sense.Connect(ctx, url, connectHeaders, sessionState.Cookies, connectionSettings.Allowuntrusted, sessionState.Timeout, reconnect); err != nil {
			return appGUID, errors.WithStack(err)
		}
		sense.OnUnexpectedDisconnect(sessionState.WSFailed)

		return appGUID, nil
	}

	return connectFunc
}

// Validate connectJWTSettings
func (connectJWT *ConnectJWTSettings) Validate() error {
	if connectJWT == nil {
		return errors.New("no JWT settings defined")
	}

	if connectJWT.key == nil {
		return errors.Errorf("No private key found")
	}
	return nil
}

// GetJwtHeader get Authorization header
func (connectJWT *ConnectJWTSettings) GetJwtHeader(sessionState *session.State, header http.Header) (http.Header, error) {

	if connectJWT.signingMethod == nil {
		return nil, errors.Errorf("no signing method set")
	}

	if connectJWT.signingMethod != jwt.SigningMethodNone && connectJWT.key == nil {
		return nil, errors.Errorf("no private key set")
	}

	// replace variables in jwt claims and create token
	claims, errClaims := connectJWT.executeClaimsTemplates(sessionState)
	if errClaims != nil {
		return nil, errors.WithStack(errClaims)
	}
	token := jwt.NewWithClaims(connectJWT.signingMethod, jwt.MapClaims(claims))

	// replace variables and set jwt headers
	jwtHeader, errJwtHeader := connectJWT.executeJWTHeaderTemplates(sessionState)
	if errJwtHeader != nil {
		return nil, errors.WithStack(errJwtHeader)
	}
	maps.Copy(token.Header, jwtHeader)

	// sign JWT
	signedToken, err := GetSignedJwtToken(connectJWT.key, token)
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

func (connectJWT *ConnectJWTSettings) executeClaimsTemplates(sessionState *session.State) (map[string]interface{}, error) {
	if connectJWT.Claims.String() == "" {
		return nil, nil
	}

	claims, errClaims := sessionState.ReplaceSessionVariables(&connectJWT.Claims)
	if errClaims != nil {
		return nil, errors.WithStack(errClaims)
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(claims), &m); err != nil {
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
	if err := json.Unmarshal([]byte(jwtHeader), &m); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal JWT Header<%s>", jwtHeader)
	}
	return m, nil
}

// GetSignedJwtToken Sign token with key
func GetSignedJwtToken(privKey any, token *jwt.Token) (string, error) {
	if token.Method == jwt.SigningMethodNone {
		return token.SigningString()
	}
	if privKey == nil {
		return "", errors.Errorf("No private key provided")
	}

	signedToken, err := token.SignedString(privKey)
	if err != nil {
		return "", errors.Wrapf(err, "Error getting signed token of key")
	}

	return signedToken, nil
}
