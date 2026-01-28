//go:build !js
// +build !js

package connection

import (
	"os"

	"github.com/goccy/go-json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

func (connectJWT *ConnectJWTSettings) UnmarshalJSON(arg []byte) error {
	if err := json.Unmarshal(arg, &connectJWT.ConnectJWTSettingsCore); err != nil {
		return err
	}

	if connectJWT.KeyPath == "" {
		return nil // don't give unmarshal error when no key is set, let validate take care of it
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
