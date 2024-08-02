
## Session variables

This section describes the session variables that can be used with some of the actions.

<details>
<summary><a name="session_variables"></a>Session variables</summary>

Some action parameters support session variables. A session variable is defined by putting the variable, prefixed by a dot, within double curly brackets, such as `{{.UserName}}`.

The following session variables are supported in actions:

* `UserName`: The simulated username. This is not the same as the authenticated user, but rather how the username was defined by [Login settings](#login_settings).  
* `Session`: The enumeration of the currently simulated session.
* `Thread`: The enumeration of the currently simulated "thread" or "concurrent user".
* `ScriptVars`: A map containing script variables added by the action `setscriptvar`.
* `Artifacts`:
  * `GetIDByTypeAndName`: A function that accepts the two string arguments,
    `artifactType` and `artifactName`, and returns the resource id of the artifact.
  * `GetNameByTypeAndID`: A function that accepts the two string arguments,
    `artifactType` and `artifactID`, and returns the name of the artifact.


The following variable is supported in the filename of the log file:

* `ConfigFile`: The filename of the config file, without file extension.

The following functions are supported:

* `now`: Evaluates Golang [time.Now()](https://golang.org/pkg/time/). 
* `hostname`: Hostname of the local machine.
* `timestamp`: Timestamp in `yyyyMMddhhmmss` format.
* `uuid`: Generate an uuid.
* `env`: Retrieve a specific environment variable. Takes one argument - the name of the environment variable to expand.
* `add`: Adds two integer values together and outputs the sum. E.g. `{{ add 1 2 }}`.
* `join`: Joins array elements together to a string separated by defined separator. E.g. `{{ join .ScriptVars.MyArray \",\" }}`.
* `modulo`: Returns modulo of two integer values and output the result. E.g. `{{ modulo 10 4 }}` (will return 2)

### Example

```json
{
    "label" : "Create bookmark",
    "action": "createbookmark",
    "settings": {
        "title": "my bookmark {{.Thread}}-{{.Session}} ({{.UserName}})",
        "description": "This bookmark contains some interesting selections"
    }
},
{
    "label" : "Publish created bookmark",
    "action": "publishbookmark",
    "disabled" : false,
    "settings" : {
        "title": "my bookmark {{.Thread}}-{{.Session}} ({{.UserName}})",
    }
}

```

```json
{
  "action": "createbookmark",
  "settings": {
    "title": "{{env \"TITLE\"}}",
    "description": "This bookmark contains some interesting selections"
  }
}
```

```json
{
    "action": "setscriptvar",
    "settings": {
        "name": "BookmarkCounter",
        "type": "int",
        "value": "1"
    }
},
{
  "action": "createbookmark",
  "settings": {
    "title": "Bookmark no {{ add .ScriptVars.BookmarkCounter 1 }}",
    "description": "This bookmark will have the title Bookmark no 2"
  }
}
```

```json
{
  "action": "setscriptvar",
  "settings": {
    "name": "MyAppId",
    "type": "string",
    "value": "{{.Artifacts.GetIDByTypeAndName \"app\" (print \"an-app-\" .Session)}}"
  }
}
```

Let's assume the case there are 4 apps to be used in the test, all ending with number 0 to 3. The use of modulo in the example will cycle through the app suffix number in following order: 1, 2, 3, 0.

```json
{
  "action": "elastictriggersubscription",
  "label": "trigger reporting task",
  "settings": {
    "subscriptiontype": "template-sharing",
    "limitperpage": 100,
    "appname": "PS-18566_Test_Levels_Pages- {{ modulo .Session 4}}",
    "subscriptionmode": "random",
  }
}
```

Very similar case as above but apps have number suffix from 1 to 4. This can be habdled combining `modulo` and `add` functions. The cycle through the suffix number will be done in following order: 2, 3, 4, 1.
```json
{
  "action": "elastictriggersubscription",
  "label": "trigger reporting task",
  "settings": {
    "subscriptiontype": "template-sharing",
    "limitperpage": 100,
    "appname": "PS-18566_Test_Levels_Pages- {{ modulo .Session 4 | add 1 }}",
    "subscriptionmode": "random",
  }
}
```

</details>
