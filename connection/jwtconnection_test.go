package connection

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/qlik-oss/gopherciser/logger"
	"github.com/qlik-oss/gopherciser/session"
	"github.com/qlik-oss/gopherciser/synced"
	"github.com/qlik-oss/gopherciser/users"
)

type TestAlgos struct {
	Alg        string
	PKeyWriter func(keyfile *os.File) error
}

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

	SigningTest(t, sessionState, &settings)
}

func SigningTest(t *testing.T, sessionState *session.State, settings *ConnectJWTSettings) {
	algoTests := []TestAlgos{
		{
			Alg:        "",
			PKeyWriter: func(keyfile *os.File) error { return writeRSAKey(2048, keyfile) },
		},
		{
			Alg:        "RS256",
			PKeyWriter: func(keyfile *os.File) error { return writeRSAKey(2048, keyfile) },
		},
		{
			Alg:        "RS384",
			PKeyWriter: func(keyfile *os.File) error { return writeRSAKey(2048, keyfile) },
		},
		{
			Alg:        "RS512",
			PKeyWriter: func(keyfile *os.File) error { return writeRSAKey(2048, keyfile) },
		},
		{
			Alg:        "none",
			PKeyWriter: func(keyfile *os.File) error { return os.WriteFile(keyfile.Name(), []byte{}, 0600) },
		},
		{
			Alg:        "EdDSA",
			PKeyWriter: func(keyfile *os.File) error { return writeEdDSAKey(keyfile) },
		},
		{
			Alg:        "ES256",
			PKeyWriter: func(keyfile *os.File) error { return writeECKey("ES256", keyfile) },
		},
		{
			Alg:        "ES384",
			PKeyWriter: func(keyfile *os.File) error { return writeECKey("ES384", keyfile) },
		},
		{
			Alg:        "ES512",
			PKeyWriter: func(keyfile *os.File) error { return writeECKey("ES512", keyfile) },
		},
		{
			Alg:        "PS256",
			PKeyWriter: func(keyfile *os.File) error { return writeRSAKey(2048, keyfile) },
		},
		{
			Alg:        "PS384",
			PKeyWriter: func(keyfile *os.File) error { return writeRSAKey(2048, keyfile) },
		},
		{
			Alg:        "PS512",
			PKeyWriter: func(keyfile *os.File) error { return writeRSAKey(2048, keyfile) },
		},
	}

	for _, test := range algoTests {
		keyfile, err := os.CreateTemp("", "PrivateKey")
		defer func() {
			_ = keyfile.Close()
		}()
		if err != nil {
			t.Fatal(err)
		}
		settings.KeyPath = keyfile.Name()
		settings.Alg = test.Alg
		if err := test.PKeyWriter(keyfile); err != nil {
			t.Error(err)
			continue
		}
		// rfile, _ := os.ReadFile(keyfile.Name())
		// t.Logf("key<%s>:\n%s\n", settings.Alg, rfile)
		settings.readKey = sync.Once{} // make sure to re-read key
		header, err := settings.GetJwtHeader(sessionState, nil)
		if err != nil {
			t.Errorf("GetJwtHeader failed algo<%s>: %v", test.Alg, err)
		}
		t.Logf("alg<%s> bearer: %s", test.Alg, header)
	}
}

func writeRSAKey(bits int, keyfile *os.File) error {
	genKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(genKey)
	if err != nil {
		return err
	}
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	return pem.Encode(keyfile, privateKeyBlock)
}

func writeEdDSAKey(keyfile *os.File) error {
	_, genKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}

	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(genKey)
	if err != nil {
		return err
	}

	privateKeyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	return pem.Encode(keyfile, privateKeyBlock)
}

func writeECKey(alg string, keyfile *os.File) error {
	var curve elliptic.Curve
	switch alg {
	case "ES256": // prime256v1
		curve = elliptic.P256()
	case "ES384": // secp384r1
		curve = elliptic.P384()
	case "ES512": // secp521r1
		curve = elliptic.P521()
	default:
		return fmt.Errorf("unsupported alg<%s>", alg)
	}

	genKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return err
	}

	privateKeyBytes, err := x509.MarshalECPrivateKey(genKey)
	if err != nil {
		return err
	}

	privateKeyBlock := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	return pem.Encode(keyfile, privateKeyBlock)
}

func validate(t *testing.T, key, value, expected string) {
	t.Helper()

	if value != expected {
		t.Errorf("key<%s> expected<%s> got<%s>", key, expected, value)
	}
}
