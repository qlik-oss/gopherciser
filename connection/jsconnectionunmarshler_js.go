//go:build js
// +build js

// This code exists to avoid decoding key when doing an UnmarshalJSON in GUI

package connection

import "github.com/goccy/go-json"

func (connectJWT *ConnectJWTSettings) UnmarshalJSON(arg []byte) error {
	if err := json.Unmarshal(arg, &connectJWT.ConnectJWTSettingsCore); err != nil {
		return err
	}

	return nil
}
