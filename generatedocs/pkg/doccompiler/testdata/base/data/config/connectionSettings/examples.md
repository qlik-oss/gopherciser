### Examples

#### JWT authentication

```json
"connectionSettings": {
    "server": "myserver.com",
    "mode": "jwt",
    "virtualproxy": "jwt",
    "security": true,
    "allowuntrusted": false,
    "jwtsettings": {
        "keypath": "mock.pem",
        "claims": "{\"user\":\"{{.UserName}}\",\"directory\":\"{{.Directory}}\"}"
    }
}
```

* `jwtsettings`:

The strings for `reqheader`, `jwtheader` and `claims` are processed as a GO template where the `User` struct can be used as data:
```golang
struct {
	UserName  string
	Password  string
	Directory string
	}
```
There is also support for the `time.Now` method using the function `now`.

* `jwtheader`:

The entries for message authentication code algorithm, `alg`, and token type, `typ`, are added automatically to the header and should not be included.
    
**Example:** To add a key ID header, `kid`, add the following string:
```json
{
	"jwtheader": "{\"kid\":\"myKeyId\"}"
}
```

* `claims`:

**Example:** For on-premise JWT authentication (with the user and directory set as keys in the QMC), add the following string:
```json
{
	"claims": "{\"user\": \"{{.UserName}}\",\"directory\": \"{{.Directory}}\"}"
}
```
**Example:** To add the time at which the JWT was issued, `iat` ("issued at"), add the following string:
```json
{
	"claims": "{\"iat\":{{now.Unix}}"
}
```
**Example:** To add the expiration time, `exp`, with 5 hours expiration (time.Now uses nanoseconds), add the following string:
```json
{
	"claims": "{\"exp\":{{(now.Add 18000000000000).Unix}}}"
}
```

#### Static header authentication

```json
connectionSettings": {
	"server": "myserver.com",
	"mode": "ws",
	"security": true,
	"virtualproxy" : "header",
	"headers" : {
		"X-Qlik-User-Header" : "{{.UserName}}"
}
```
