## Connection settings section

This section of the JSON file contains connection information.

JSON Web Token (JWT), an open standard for creation of access tokens, or WebSocket can be used for authentication. When using JWT, the private key must be available in the path defined by `jwtsettings.keypath`.

### Creating private / public key pair

Keypairs are most easily created using `openssl`. The private key is used by gopherciser and the public key used to when configuring the Sense environment. If no `Alg` is defined it will default to `RS512`.

Supported signing algorithms in QSEoW Virtual proxy are: RS256, RS384, RS512. Elliptical curve algorithms are not supported in QSEoW virtual proxies.

```bash
# Generate a 4096 bit private key
openssl genrsa -out privatekey.pem 4096
# Generates a certificate valid for one year
openssl req -new -x509 -key ./keyfiles/rsa.key -out ./keyfiles/rsa.cer -days 365 
```

The generated rsa.cer is what's used when creating the virtual proxy with `JWT` _Authentication Method_ in QSEoW.
