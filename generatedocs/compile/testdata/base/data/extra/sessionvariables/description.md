
## Session variables

This section describes the session variables that can be used with some of the actions.

<details>
<summary><a name="session_variables"></a>Session variables</summary>

Some action parameters support session variables. A session variable is defined by putting the variable, prefixed by a dot, within double curly brackets, such as `{{.UserName}}`.

The following session variables are supported in actions:

* `UserName`: The simulated username. This is not the same as the authenticated user, but rather how the username was defined by [Login settings](#login_settings).  
* `Session`: The enumeration of the currently simulated session.
* `Thread`: The enumeration of the currently simulated "thread" or "concurrent user".

The following variable is supported in the filename of the log file:

* `ConfigFile`: The filename of the config file, without file extension.

The following functions are supported:

* `now`: Evaluates Golang [time.Now()](https://golang.org/pkg/time/). 
* `hostname`: Hostname of the local machine.
* `timestamp`: Timestamp in `yyyyMMddhhmmss` format.
* `uuid`: Generate an uuid.

### Example
```json
{
    "action": "ElasticCreateApp",
    "label": "Create new app",
    "settings": {
        "title": "CreateApp {{.Thread}}-{{.Session}} ({{.UserName}})",
        "stream": "mystream",
        "groups": [
            "mygroup"
        ]
    }
},
{
    "label": "OpenApp",
    "action": "OpenApp",
    "settings": {
        "appname": "CreateApp {{.Thread}}-{{.Session}} ({{.UserName}})"
    }
},
{
    "action": "elasticexportapp",
    "label": "Export app",
    "settings": {
        "appmode" : "name",
        "app" : "CreateApp {{.Thread}}-{{.Session}} ({{.UserName}})",
        "savetofile": true,
        "exportname": "Exported app {{.Thread}}-{{.Session}} {{now.UTC}}"
    }
}

```
</details>
