# Setting up load scenarios

A load scenario is defined in a JSON file with a number of sections.

<details>
<summary>connectionSettings</summary>

## Connection settings section

This section of the JSON file contains connection information.

JSON Web Token (JWT), an open standard for creation of access tokens, or WebSocket can be used for authentication. When using JWT, the private key must be available in the path defined by `jwtsettings.keypath`.

* `mode`: Authentication mode
    * `jwt`: JSON Web Token
    * `ws`: WebSocket
* `jwtsettings`: (JWT only) Settings for the JWT connection.
  * `keypath`: Local path to the JWT key file.
  * `jwtheader`: JWT headers as an escaped JSON string. Custom headers to be added to the JWT header.
  * `claims`: JWT claims as an escaped JSON string.
  * `alg`: The signing method used for the JWT. Defaults to `RS512`, if omitted.
      * For keyfiles in RSA format, supports `RS256`, `RS384` or `RS512`.
      * For keyfiles in EC format, supports `ES256`, `ES384` or `ES512`.
* `wssettings`: (WebSocket only) Settings for the WebSocket connection.
* `server`: Qlik Sense host.
* `virtualproxy`: Prefix for the virtual proxy that handles the virtual users.
* `rawurl`: Define the connect URL manually instead letting the `openapp` action do it. **Note**: The protocol must be `wss://` or `ws://`.
* `port`: Set another port than default (`80` for http and `443` for https).
* `security`: Use TLS (SSL) (`true` / `false`).
* `allowuntrusted`: Allow untrusted (for example, self-signed) certificates (`true` / `false`). Defaults to `false`, if omitted.
* `appext`: Replace `app` in the connect URL for the `openapp` action. Defaults to `app`, if omitted.
* `headers`: Headers to use in requests.

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

</details><details>
<summary>loginSettings</summary>

## Login settings section

This section of the JSON file contains information on the login settings.

* `type`: Type of login request
    * `prefix`: Add a prefix (specified by the `prefix` setting below) to the username, so that it will be `prefix_{session}`.
    * `userlist`: List of users as specified by the `userList` setting below.
    * `none`: Do not add a prefix to the username, so that it will be `{session}`.
* `settings`: 
    * `userList`: List of users for the `userlist` login request type. Directory and password can be specified per user or outside the list of usernames, which means that they are inherited by all users.
  * `prefix`: Prefix to add to the username, so that it will be `prefix_{session}`.
  * `directory`: Directory to set for the users.

### Examples

#### Prefix login request type

```json
"loginSettings": {
   "type": "prefix",
   "settings": {
       "directory": "anydir",
       "prefix": "Nunit"
   }
}
```

#### Userlist login request type

```json
  "loginSettings": {
    "type": "userlist",
    "settings": {
      "userList": [
        {
          "username": "sim1@myhost.example",
          "directory": "anydir1",
          "password": "MyPassword1"
        },
        {
          "username": "sim2@myhost.example"
        }
      ],
      "directory": "anydir2",
      "password": "MyPassword2"
    }
  }
```

</details><details>
<summary>scheduler</summary>

## Scheduler section

This section of the JSON file contains scheduler settings for the users in the load scenario.

* `type`: Type of scheduler
    * `simple`: Standard scheduler
* `iterationtimebuffer`: 
  * `mode`: Time buffer mode. Defaults to `nowait`, if omitted.
      * `nowait`: No time buffer in between the iterations.
      * `constant`: Add a constant time buffer after each iteration. Defined by `duration`.
      * `onerror`: Add a time buffer in case of an error. Defined by `duration`.
      * `minduration`: Add a time buffer if the iteration duration is less than `duration`.
  * `duration`: Duration of the time buffer (for example, `500ms`, `30s` or `1m10s`). Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, and `h`.
* `instance`: Instance number for this instance. Use different instance numbers when running the same script in multiple instances to make sure the randomization is different in each instance. Defaults to 1.
* `settings`: 
  * `executionTime`: Test execution time (seconds). The sessions are disconnected when the specified time has elapsed. Allowed values are positive integers. `-1` means an infinite execution time.
  * `iterations`: Number of iterations for each 'concurrent' user to repeat. Allowed values are positive integers. `-1` means an infinite number of iterations.
  * `rampupDelay`: Time delay (seconds) scheduled in between each concurrent user during the startup period.
  * `concurrentUsers`: Number of concurrent users to simulate. Allowed values are positive integers.
  * `reuseUsers`: 
      * `true`: Every iteration for each concurrent user uses the same user and session.
      * `false`: Every iteration for each concurrent user uses a new user and session. The total number of users is the product of `concurrentusers` and `iterations`.
  * `onlyinstanceseed`: Disable session part of randomization seed. Defaults to `false`, if omitted.
      * `true`: All users and sessions have the same randomization sequence, which only changes if the `instance` flag is changed.
      * `false`: Normal randomization sequence, dependent on both the `instance` parameter and the current user session.

### Example

```json
"scheduler": {
   "type": "simple",
   "settings": {
       "executiontime": 120,
       "iterations": -1,
       "rampupdelay": 7.0,
       "concurrentusers": 10
   },
   "iterationtimebuffer" : {
       "mode": "onerror",
       "duration" : "5s"
   },
   "instance" : 2
}
```

</details><details>
<summary>settings</summary>

## Settings section

This section of the JSON file contains timeout and logging settings for the load scenario.

* `timeout`: Timeout setting (seconds) for WebSocket requests.
* `logs`: Log settings
  * `traffic`: Log traffic information (`true` / `false`). Defaults to `false`, if omitted. **Note:** This should only be used for debugging purposes as traffic logging is resource-demanding.
  * `debug`: Log debug information (`true` / `false`). Defaults to `false`, if omitted.
  * `metrics`: Log traffic metrics (`true` / `false`). Defaults to `false`, if omitted. **Note:** This should only be used for debugging purposes as traffic logging is resource-demanding.
  * `filename`: Name of the log file (supports the use of [variables](#session_variables)).
  * `format`: Log format. Defaults to `tsvfile`, if omitted.
      * `tsvfile`: Log to file in TSV format and output status to console.
      * `tsvconsole`: Log to console in TSV format without any status output.
      * `jsonfile`: Log to file in JSON format and output status to console.
      * `jsonconsole`: Log to console in JSON format without any status output.
      * `console`: Log to console in color format without any status output.
      * `combined`: Log to file in TSV format and to console in JSON format.
      * `no`: Default logs and status output turned off.
      * `onlystatus`: Default logs turned off, but status output turned on.
  * `summary`: Type of summary to display after the test run. Defaults to simple for minimal performance impact.
      * `0` or `undefined`: Simple, single-row summary
      * `1` or `none`: No summary
      * `2` or `simple`: Simple, single-row summary
      * `3` or `extended`: Extended summary that includes statistics on each unique combination of action, label and app GUID
      * `4` or `full`: Same as extended, but with statistics on each unique combination of method and endpoint added
* `outputs`: Used by some actions to save results to a file.
  * `dir`: Directory in which to save artifacts generated by the script (except log file).

### Examples

```json
"settings": {
	"timeout": 300,
	"logs": {
		"traffic": false,
		"debug": false,
		"filename": "logs/{{.ConfigFile}}-{{timestamp}}.log"
	}
}
```

```json
"settings": {
	"timeout": 300,
	"logs": {
		"filename": "logs/scenario.log"
	},
	"outputs" : {
	    "dir" : "./outputs"
	}
}
```

</details><details>
<summary>scenario</summary>

## Scenario section

This section of the JSON file contains the actions that are performed in the load scenario.

### Structure of an action entry

All actions follow the same basic structure: 

* `action`: Name of the action to execute.
* `label`: (optional) Custom string set by the user. This can be used to distinguish the action from other actions of the same type when analyzing the test results.
* `disabled`: (optional) Disable action (`true` / `false`). If set to `true`, the action is not executed.
* `settings`: Most, but not all, actions have a settings section with action-specific settings.
### Example

```json
{
    "action": "actioname",
    "label": "custom label for analysis purposes",
    "disabled": false,
    "settings": {
        
    }
}
```

<details>
<summary>Common actions</summary>

# Common actions

These actions are applicable to both Qlik Sense Enterprise for Windows (QSEfW) and Qlik Sense Enterprise on Kubernetes (QSEoK) deployments.

**Note:** It is recommended to prepend the actions listed here with an `openapp` action as most of them perform operations in an app context (such as making selections or changing sheets).


<details>
<summary>applybookmark</summary>

## ApplyBookmark action

Apply a bookmark in the current app.

**Note:** Specify *either* `title` *or* `id`, not both.

### Settings

* `title`: Name of the bookmark.
* `id`: GUID of the bookmark.

### Example

```json
{
    "action": "applybookmark",
    "settings": {
        "title": "My bookmark"
    }
}
```

</details><details>
<summary>changesheet</summary>

## ChangeSheet action

Change to a new sheet, unsubscribe to the currently subscribed objects, and subscribe to all objects on the new sheet.

The action supports getting data from the following objects:

* Listbox
* Filter pane
* Bar chart
* Scatter plot
* Map (only the first layer)
* Combo chart
* Table
* Pivot table
* Line chart
* Pie chart
* Tree map
* Text-Image
* KPI
* Gauge
* Box plot
* Distribution plot
* Histogram
* Auto chart (including any support generated visualization from this list)
* Waterfall chart

### Settings

* `id`: GUID of the sheet to change to.

### Example

```json
{
     "label": "Change Sheet Dashboard",
     "action": "ChangeSheet",
     "settings": {
         "id": "TFJhh"
     }
}
```

</details><details>
<summary>clearall</summary>

## ClearAll action

Clear all selections in an app.


### Example

```json
{
    "action": "clearall",
    "label": "Clear all selections (1)"
}
```

</details><details>
<summary>createbookmark</summary>

## CreateBookmark action

Create a bookmark from the current selection and selected sheet.

### Settings

* `title`: Name of the bookmark.
* `id`: GUID of the bookmark.
* `description`: (optional) Description of the bookmark to create.
* `nosheet`: Do not include the sheet location in the bookmark.
* `savelayout`: Include the layout in the bookmark.

### Example

```json
{
    "action": "createbookmark",
    "settings": {
        "title": "my bookmark",
        "description": "This bookmark contains some interesting selections"
    }
}
```

</details><details>
<summary>createsheet</summary>

## CreateSheet action

Create a new sheet in the current app.

### Settings

* `id`: (optional) ID to be used to identify the sheet in any subsequent `changesheet`, `duplicatesheet`, `publishsheet` or `unpublishsheet` action.
* `title`: Name of the sheet to create.
* `description`: (optional) Description of the sheet to create.

### Example

```json
{
    "action": "createsheet",
    "settings": {
        "title" : "Generated sheet"
    }
}
```

</details><details>
<summary>deletebookmark</summary>

## DeleteBookmark action

Delete one or more bookmarks in the current app.

**Note:** Specify *either* `title` *or* `id`, not both.

### Settings

* `title`: Name of the bookmark.
* `id`: GUID of the bookmark.
* `mode`: 
    * `single`: Delete one bookmark that matches the specified `title` or `id` in the current app.
    * `matching`: Delete all bookmarks with the specified `title` in the current app.
    * `all`: Delete all bookmarks in the current app.

### Example

```json
{
    "action": "deletebookmark",
    "settings": {
        "mode": "single",
        "title": "My bookmark"
    }
}
```

</details><details>
<summary>deletesheet</summary>

## DeleteSheet action

Delete one or more sheets in the current app.

**Note:** Specify *either* `title` *or* `id`, not both.

### Settings

* `mode`: 
    * `single`: Delete one sheet that matches the specified `title` or `id` in the current app.
    * `matching`: Delete all sheets with the specified `title` in the current app.
    * `allunpublished`: Delete all unpublished sheets in the current app.
* `title`: (optional) Name of the sheet to delete.
* `id`: (optional) GUID of the sheet to delete.

### Example

```json
{
    "action": "deletesheet",
    "settings": {
        "mode": "matching",
        "title": "Test sheet"
    }
}
```

</details><details>
<summary>disconnectapp</summary>

## DisconnectApp action

Disconnect from an already connected app.


### Example

```json
{
    "label": "Disconnect from server",
    "action" : "disconnectapp"
}
```

</details><details>
<summary>duplicatesheet</summary>

## DuplicateSheet action

Duplicate a sheet, including all objects.

### Settings

* `id`: ID of the sheet to clone.
* `changesheet`: Clear the objects currently subscribed to and then subribe to all objects on the cloned sheet (which essentially corresponds to using the `changesheet` action to go to the cloned sheet) (`true` / `false`). Defaults to `false`, if omitted.
* `save`: Execute `saveobjects` after the cloning operation to save all modified objects (`true` / `false`). Defaults to `false`, if omitted.
* `cloneid`: (optional) ID to be used to identify the sheet in any subsequent `changesheet`, `duplicatesheet`, `publishsheet` or `unpublishsheet` action.

### Example

```json
{
    "action": "duplicatesheet",
    "label": "Duplicate sheet1",
    "settings":{
        "id" : "mBshXB",
        "save": true,
        "changesheet": true
    }
}
```

</details><details>
<summary>iterated</summary>

## Iterated action

Loop one or more actions.

**Note:** This action does not require an app context (that is, it does not have to be prepended with an `openapp` action).

### Settings

* `iterations`: Number of loops.
* `actions`: Actions to iterate
  * `action`: Name of the action to execute.
  * `label`: (optional) Custom string set by the user. This can be used to distinguish the action from other actions of the same type when analyzing the test results.
  * `disabled`: (optional) Disable action (`true` / `false`). If set to `true`, the action is not executed.
  * `settings`: Most, but not all, actions have a settings section with action-specific settings.

### Example

```json
//Visit all sheets twice
{
     "action": "iterated",
     "label": "",
     "settings": {
         "iterations" : 2,
         "actions" : [
            {
                 "action": "sheetchanger"
            },
            {
                "action": "thinktime",
                "settings": {
                    "type": "static",
                    "delay": 5
                }
            }
         ]
     }
}
```

</details><details>
<summary>openapp</summary>

## OpenApp action

Open an app.

**Note:** If the app name is used to specify which app to open, this action cannot be the first action in the scenario. It must be preceded by an action that can populate the artifact map, such as `openhub`, `elasticopenhub` or `elasticexplore`.

### Settings

* `appmode`: App selection mode
    * `current`: (default) Use the current app, selected by an app selection in a previous action, or set by the `elasticcreateapp`, `elasticduplicateapp` or `elasticuploadapp` action.
    * `guid`: Use the app GUID specified by the `app` parameter.
    * `name`: Use the app name specified by the `app` parameter.
    * `random`: Select a random app from the artifact map, which is filled by the `elasticopenhub` and/or the `elasticexplore` actions.
    * `randomnamefromlist`: Select a random app from a list of app names. The `list` parameter should contain a list of app names.
    * `randomguidfromlist`: Select a random app from a list of app GUIDs. The `list` parameter should contain a list of app GUIDs.
    * `randomnamefromfile`: Select a random app from a file with app names. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `randomguidfromfile`: Select a random app from a file with app GUIDs. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
    * `round`: Select an app from the artifact map according to the round-robin principle.
    * `roundnamefromlist`: Select an app from a list of app names according to the round-robin principle. The `list` parameter should contain a list of app names.
    * `roundguidfromlist`: Select an app from a list of app GUIDs according to the round-robin principle. The `list` parameter should contain a list of app GUIDs.
    * `roundnamefromfile`: Select an app from a file with app names according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `roundguidfromfile`: Select an app from a file with app GUIDs according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
* `app`: App name or app GUID (supports the use of [session variables](#session_variables)). Used with `appmode` set to `guid` or `name`.
* `list`: List of apps. Used with `appmode` set to `randomnamefromlist`, `randomguidfromlist`, `roundnamefromlist` or `roundguidfromlist`.
* `filename`: Path to a file in which each line represents an app. Used with `appmode` set to `randomnamefromfile`, `randomguidfromfile`, `roundnamefromfile` or `roundguidfromfile`.

### Examples

```json
{
     "label": "OpenApp",
     "action": "OpenApp",
     "settings": {
         "appmode": "guid",
         "app": "7967af99-68b6-464a-86de-81de8937dd56"
     }
}
```
```json
{
     "label": "OpenApp",
     "action": "OpenApp",
     "settings": {
         "appmode": "randomguidfromlist",
         "list": ["7967af99-68b6-464a-86de-81de8937dd56", "ca1a9720-0f42-48e5-baa5-597dd11b6cad"]
     }
}
```

</details><details>
<summary>productversion</summary>

## ProductVersion action

Request the product version from the server and, optionally, save it to the log. This is a lightweight request that can be used as a keep-alive message in a loop.

### Settings

* `log`: Save the product version to the log (`true` / `false`). Defaults to `false`, if omitted.

### Example

```json
//Keep-alive loop
{
    "action": "iterated",
    "settings" : {
        "iterations" : 10,
        "actions" : [
            {
                "action" : "productversion"
            },
            {
                "action": "thinktime",
                "settings": {
                    "type": "static",
                    "delay": 30
                }
            }
        ]
    }
}
```

</details><details>
<summary>publishsheet</summary>

## PublishSheet action

Publish sheets in the current app.

### Settings

* `mode`: 
    * `allsheets`: Publish all sheets in the app.
    * `sheetids`: Only publish the sheets specified by the `sheetIds` array.
* `sheetIds`: (optional) Array of sheet IDs for the `sheetids` mode.

### Example
```json
{
     "label": "PublishSheets",
     "action": "publishsheet",
     "settings": {
       "mode": "sheetids",
       "sheetIds": ["qmGcYS", "bKbmgT"]
     }
}
```

</details><details>
<summary>randomaction</summary>

## RandomAction action

Randomly select other actions to perform. This meta-action can be used as a starting point for your testing efforts, to simplify script authoring or to add background load.

`randomaction` accepts a list of action types between which to randomize. An execution of `randomaction` executes one or more of the listed actions (as determined by the `iterations` parameter), randomly chosen by a weighted probability. If nothing else is specified, each action has a default random mode that is used. An override is done by specifying one or more parameters of the original action.

Each action executed by `randomaction` is followed by a customizable `thinktime`.

**Note:** The recommended way to use this action is to prepend it with an `openapp` and a `changesheet` action as this ensures that a sheet is always in context.

### Settings

* `actions`: List of actions from which to randomly pick an action to execute. Each item has a number of possible parameters.
  * `type`: Type of action
      * `thinktime`: See the `thinktime` action.
      * `sheetobjectselection`: Make random selections within objects visible on the current sheet. See the `select` action.
      * `changesheet`: See the `changesheet` action.
      * `clearall`: See the `clearall` action.
  * `weight`: The probabilistic weight of the action, specified as an integer. This number is proportional to the likelihood of the specified action, and is used as a weight in a uniform random selection.
  * `overrides`: (optional) Static overrides to the action. The overrides can include any or all of the settings from the original action, as determined by the `type` field. If nothing is specified, the default values are used.
* `thinktimesettings`: Settings for the `thinktime` action, which is automatically inserted after every randomized action.
  * `type`: Type of think time
      * `static`: Static think time, defined by `delay`.
      * `uniform`: Random think time with uniform distribution, defined by `mean` and `dev`.
  * `delay`: Delay (seconds), used with type `static`.
  * `mean`: Mean (seconds), used with type `uniform`.
  * `dev`: Deviation (seconds) from `mean` value, used with type `uniform`.
* `iterations`: Number of random actions to perform.

### Random action defaults

The following default values are used for the different actions:

* `thinktime`: Mirrors the configuration of `thinktimesettings`
* `sheetobjectselection`:

```json
{
     "settings": 
     {
         "id": <UNIFORMLY RANDOMIZED>,
         "type": "RandomFromAll",
         "min": 1,
         "max": 2,
         "accept": true
     }
}
```

* `changesheet`:

```json
{
     "settings": 
     {
         "id": <UNIFORMLY RANDOMIZED>
     }
}
```

* `clearall`:

```json
{
     "settings": 
     {
     }
}
```

### Examples

#### Generating a background load by executing 5 random actions

```json
{
    "action": "RandomAction",
    "settings": {
        "iterations": 5,
        "actions": [
            {
                "type": "thinktime",
                "weight": 1
            },
            {
                "type": "sheetobjectselection",
                "weight": 3
            },
            {
                "type": "changesheet",
                "weight": 5
            },
            {
                "type": "clearall",
                "weight": 1
            }
        ],
        "thinktimesettings": {
            "type": "uniform",
            "mean": 10,
            "dev": 5
        }
    }
}
```

#### Making random selections from excluded values

```json
{
    "action": "RandomAction",
    "settings": {
        "iterations": 1,
        "actions": [
            {
                "type": "sheetobjectselection",
                "weight": 1,
                "overrides": {
                  "type": "RandomFromExcluded",
                  "min": 1,
                  "max": 5
                }
            }
        ],
        "thinktimesettings": {
            "type": "static",
            "delay": 1
        }
    }
}
```

</details><details>
<summary>reload</summary>

## Reload action

Reload the current app by simulating selecting **Load data** in the Data load editor. To select an app, preceed this action with an `openapp` action.

### Settings

* `mode`: Error handling during the reload operation
    * `default`: Use the default error handling.
    * `abend`: Stop reloading the script, if an error occurs.
    * `ignore`: Continue reloading the script even if an error is detected in the script.
* `partial`: Enable partial reload (`true` / `false`). This allows you to add data to an app without reloading all data. Defaults to `false`, if omitted.
* `log`: Save the reload log as a field in the output (`true` / `false`). Defaults to `false`, if omitted. **Note:** This should only be used when needed as the reload log can become very large.

### Example

```json
{
    "action": "reload",
    "settings": {
        "mode" : "default",
        "partial": false
    }
}
```

</details><details>
<summary>select</summary>

## Select action

Select random values in an object.

See the [Limitations](README.md#limitations) section in the README.md file for limitations related to this action.
 
### Settings

* `id`: ID of the object in which to select values.
* `type`: Selection type
    * `randomfromall`: Randomly select within all values of the symbol table.
    * `randomfromenabled`: Randomly select within the white and light grey values on the first data page.
    * `randomfromexcluded`: Randomly select within the dark grey values on the first data page.
    * `randomdeselect`: Randomly deselect values on the first data page.
* `accept`: Accept or abort selection after selection (only used with `wrap`) (`true` / `false`).
* `wrap`: Wrap selection with Begin / End selection requests (`true` / `false`).
* `min`: Minimum number of selections to make.
* `max`: Maximum number of selections to make.
* `dim`: Dimension / column in which to select.

### Example

```json
//Select Listbox RandomFromAll
{
     "label": "ListBox Year",
     "action": "Select",
     "settings": {
         "id": "RZmvzbF",
         "type": "RandomFromAll",
         "accept": true,
         "wrap": false,
         "min": 1,
         "max": 3,
         "dim": 0
     }
}
```

</details><details>
<summary>setscript</summary>

## SetScript action

Set the load script for the current app. To load the data from the script, use the `reload` action after the `setscript` action.

### Settings

* `script`: Load script for the app (written as a string).

### Example

```json
{
    "action": "setscript",
    "settings": {
        "script" : "Characters:\nLoad Chr(RecNo()+Ord('A')-1) as Alpha, RecNo() as Num autogenerate 26;"
    }
}
```

</details><details>
<summary>sheetchanger</summary>

## SheetChanger action

Create and execute a `changesheet` action for each sheet in an app. This can be used to cache the inital state for all objects or, by chaining two subsequent `sheetchanger` actions, to measure how well the calculations in an app utilize the cache.


### Example

```json
{
    "label" : "Sheetchanger uncached",
    "action": "sheetchanger"
},
{
    "label" : "Sheetchanger cached",
    "action": "sheetchanger"
}
```

</details><details>
<summary>staticselect</summary>

## StaticSelect action

Select values statically.

The action supports:

* HyperCube: Normal hypercube
* ListObject: Normal listbox

### Settings

* `id`: ID of the object in which to select values.
* `path`: Path to the hypercube or listobject (differs depending on object type).
* `rows`: Element values to select in the dimension / column.
* `cols`: Dimension / column in which to select.
* `type`: Selection type
    * `hypercubecells`: Select in hypercube.
    * `listobjectvalues`: Select in listbox.
* `accept`: Accept or abort selection after selection (only used with `wrap`) (`true` / `false`).
* `wrap`: Wrap selection with Begin / End selection requests (`true` / `false`).

### Examples

#### StaticSelect Barchart

```json
{ 
"label": "Chart Profit per year",
     "action": "StaticSelect",
     "settings": {
         "id": "FERdyN",
	 "path": "/qHyperCubeDef",
         "type": "hypercubecells",
         "accept": true,
         "wrap": false,
         "rows": [2],
	 "cols": [0]
     }
}
```

#### StaticSelect Listbox

```json
{		
"label": "ListBox Territory",
     "action": "StaticSelect",
     "settings": {
         "id": "qpxmZm",
         "path": "/qListObjectDef",
         "type": "listobjectvalues",
         "accept": true,
         "wrap": false,
         "rows": [19,8],
	 "cols": [0]
     }
}
```

</details><details>
<summary>thinktime</summary>

## ThinkTime action

Simulate user think time.

**Note:** This action does not require an app context (that is, it does not have to be prepended with an `openapp` action).

### Settings

* `type`: Type of think time
    * `static`: Static think time, defined by `delay`.
    * `uniform`: Random think time with uniform distribution, defined by `mean` and `dev`.
* `delay`: Delay (seconds), used with type `static`.
* `mean`: Mean (seconds), used with type `uniform`.
* `dev`: Deviation (seconds) from `mean` value, used with type `uniform`.

### Examples

#### ThinkTime uniform

This simulates a think time of 10 to 15 seconds.

```json
{
     "label": "TimerDelay",
     "action": "thinktime",
     "settings": {
         "type": "uniform",
         "mean": 12.5,
         "dev": 2.5
     } 
} 
```

#### ThinkTime constant

This simulates a think time of 5 seconds.

```json
{
     "label": "TimerDelay",
     "action": "thinktime",
     "settings": {
         "type": "static",
         "delay": 5
     }
}
```

</details><details>
<summary>unpublishsheet</summary>

## UnpublishSheet action

Unpublish sheets in the current app.

### Settings

* `mode`: 
    * `allsheets`: Unpublish all sheets in the app.
    * `sheetids`: Only unpublish the sheets specified by the `sheetIds` array.
* `sheetIds`: (optional) Array of sheet IDs for the `sheetids` mode.

### Example
```json
{
     "label": "UnpublishSheets",
     "action": "unpublishsheet",
     "settings": {
       "mode": "allsheets"        
     }
}
```

</details>
</details><details>
<summary>Qlik Sense Enterprise on Windows (QSEoW) actions</summary>

## Qlik Sense Enterprise on Windows (QSEoW) actions

These actions are only applicable to Qlik Sense Enterprise on Windows (QSEoW) deployments.


<details>
<summary>deleteodag</summary>

## DeleteOdag action

Delete all user-generated on-demand apps for the current user and the specified On-Demand App Generation (ODAG) link.

### Settings

* `linkname`: Name of the ODAG link from which to delete generated apps. The name is displayed in the ODAG navigation bar at the bottom of the *selection app*.

### Example

```json
{
    "action": "DeleteOdag",
    "settings": {
        "linkname": "Drill to Template App"
    }
}
```

</details><details>
<summary>generateodag</summary>

## GenerateOdag action

Generate an on-demand app from an existing On-Demand App Generation (ODAG) link.

### Settings

* `linkname`: Name of the ODAG link from which to generate an app. The name is displayed in the ODAG navigation bar at the bottom of the *selection app*.

### Example

```json
{
    "action": "GenerateOdag",
    "settings": {
        "linkname": "Drill to Template App"
    }
}
```

</details><details>
<summary>openhub</summary>

## OpenHub action

Open the hub in a QSEoW environment.


### Example

```json
{
     "action": "OpenHub",
     "label": "Open the hub"
}
```

</details>
</details><details>
<summary>Qlik Sense Enterprise on Kubernetes (QSEoK) / Elastic actions</summary>

## Qlik Sense Enterprise on Kubernetes (QSEoK) / Elastic actions

These actions are only applicable to Qlik Sense Enterprise on Kubernetes (QSEoK) deployments.


<details>
<summary>deletedata</summary>

## DeleteData action

Delete a data file from the Data manager.

### Settings

* `filename`: Name of the file to delete.
* `path`: (optional) Path in which to look for the file. Defaults to `MyDataFiles`, if omitted.

### Example

```json
{
     "action": "DeleteData",
     "settings": {
         "filename": "data.csv",
         "path": "MyDataFiles"
     }
}
```

</details><details>
<summary>elasticcreateapp</summary>

## ElasticCreateApp action

Create an app in a QSEoK deployment. The app will be private to the user who creates it.

### Settings

* `title`: Name of the app to upload (supports the use of [session variables](#session_variables)).
* `stream`: (optional) Name of the private collection or public tag under which to publish the app (supports the use of [session variables](#session_variables)).
* `streamguid`: (optional) GUID of the private collection or public tag under which to publish the app.

### Example

```json
{
     "action": "ElasticCreateApp",
     "label": "Create new app",
     "settings": {
         "title": "Created by script",
         "stream": "Everyone",
         "groups": ["Everyone", "cool kids"]
     }
}
```

</details><details>
<summary>elasticcreatecollection</summary>

## ElasticCreateCollection action

Create a collection in a QSEoK deployment.

### Settings

* `name`: Name of the collection to create (supports the use of [session variables](#session_variables)).
* `description`: (optional) Description of the collection to create.
* `private`: 
    * `true`: Private collection
    * `false`: Public collection

### Example

```json
{
   "action": "ElasticCreateCollection",
   "label": "Create collection",
   "settings": {
       "name": "Collection {{.Session}}",
       "private": false
   }
}
```

</details><details>
<summary>elasticdeleteapp</summary>

## ElasticDeleteApp action

Delete an app from a QSEoK deployment.

### Settings

* `appmode`: App selection mode
    * `current`: (default) Use the current app, selected by an app selection in a previous action, or set by the `elasticcreateapp`, `elasticduplicateapp` or `elasticuploadapp` action.
    * `guid`: Use the app GUID specified by the `app` parameter.
    * `name`: Use the app name specified by the `app` parameter.
    * `random`: Select a random app from the artifact map, which is filled by the `elasticopenhub` and/or the `elasticexplore` actions.
    * `randomnamefromlist`: Select a random app from a list of app names. The `list` parameter should contain a list of app names.
    * `randomguidfromlist`: Select a random app from a list of app GUIDs. The `list` parameter should contain a list of app GUIDs.
    * `randomnamefromfile`: Select a random app from a file with app names. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `randomguidfromfile`: Select a random app from a file with app GUIDs. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
    * `round`: Select an app from the artifact map according to the round-robin principle.
    * `roundnamefromlist`: Select an app from a list of app names according to the round-robin principle. The `list` parameter should contain a list of app names.
    * `roundguidfromlist`: Select an app from a list of app GUIDs according to the round-robin principle. The `list` parameter should contain a list of app GUIDs.
    * `roundnamefromfile`: Select an app from a file with app names according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `roundguidfromfile`: Select an app from a file with app GUIDs according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
* `app`: App name or app GUID (supports the use of [session variables](#session_variables)). Used with `appmode` set to `guid` or `name`.
* `list`: List of apps. Used with `appmode` set to `randomnamefromlist`, `randomguidfromlist`, `roundnamefromlist` or `roundguidfromlist`.
* `filename`: Path to a file in which each line represents an app. Used with `appmode` set to `randomnamefromfile`, `randomguidfromfile`, `roundnamefromfile` or `roundguidfromfile`.
* `mode`: 
    * `single`: Delete the app specified explicitly by app GUID or app name.
    * `everything`: Delete all apps currently in the application context, as determined by the `elasticopenhub` action. **Note:** Use with care.
    * `clearcollection`: Delete all apps in the collection specified by `collectionname`.
* `collectionname`: Name of the collection in which to delete apps.

### Example

```json
{
     "action": "ElasticDeleteApp",
     "label": "delete app myapp",
     "settings": {
         "mode": "single",
         "appmode": "name",
         "app": "myapp"
     }
}
```

</details><details>
<summary>elasticdeletecollection</summary>

## ElasticDeleteCollection action

Delete a collection in a QSEoK deployment.

### Settings

* `name`: Name of the collection to delete.
* `deletecontents`: 
    * `true`: Delete all apps in the collection before deleting the collection.
    * `false`: Delete the collection without doing anything to the apps in the collection.

### Example

```json
{
   "action": "ElasticDeleteCollection",
   "label": "Delete collection",
   "settings": {
       "name": "MyCollection",
       "deletecontents": true
   }
}
```

</details><details>
<summary>elasticdeleteodag</summary>

## ElasticDeleteOdag action

Delete all user-generated on-demand apps for the current user and the specified On-Demand App Generation (ODAG) link.

### Settings

* `linkname`: Name of the ODAG link from which to delete generated apps. The name is displayed in the ODAG navigation bar at the bottom of the *selection app*.

### Example

```json
{
    "action": "ElasticDeleteOdag",
    "settings": {
        "linkname": "Drill to Template App"
    }
}
```

</details><details>
<summary>elasticduplicateapp</summary>

## ElasticDuplicateApp action

Duplicate an app in a QSEoK deployment.

### Settings

* `appmode`: App selection mode
    * `current`: (default) Use the current app, selected by an app selection in a previous action, or set by the `elasticcreateapp`, `elasticduplicateapp` or `elasticuploadapp` action.
    * `guid`: Use the app GUID specified by the `app` parameter.
    * `name`: Use the app name specified by the `app` parameter.
    * `random`: Select a random app from the artifact map, which is filled by the `elasticopenhub` and/or the `elasticexplore` actions.
    * `randomnamefromlist`: Select a random app from a list of app names. The `list` parameter should contain a list of app names.
    * `randomguidfromlist`: Select a random app from a list of app GUIDs. The `list` parameter should contain a list of app GUIDs.
    * `randomnamefromfile`: Select a random app from a file with app names. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `randomguidfromfile`: Select a random app from a file with app GUIDs. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
    * `round`: Select an app from the artifact map according to the round-robin principle.
    * `roundnamefromlist`: Select an app from a list of app names according to the round-robin principle. The `list` parameter should contain a list of app names.
    * `roundguidfromlist`: Select an app from a list of app GUIDs according to the round-robin principle. The `list` parameter should contain a list of app GUIDs.
    * `roundnamefromfile`: Select an app from a file with app names according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `roundguidfromfile`: Select an app from a file with app GUIDs according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
* `app`: App name or app GUID (supports the use of [session variables](#session_variables)). Used with `appmode` set to `guid` or `name`.
* `list`: List of apps. Used with `appmode` set to `randomnamefromlist`, `randomguidfromlist`, `roundnamefromlist` or `roundguidfromlist`.
* `filename`: Path to a file in which each line represents an app. Used with `appmode` set to `randomnamefromfile`, `randomguidfromfile`, `roundnamefromfile` or `roundguidfromfile`.
* `title`: Name of the app to upload (supports the use of [session variables](#session_variables)).
* `stream`: (optional) Name of the private collection or public tag under which to publish the app (supports the use of [session variables](#session_variables)).
* `streamguid`: (optional) GUID of the private collection or public tag under which to publish the app.
* `spaceid`: (optional) GUID of the shared space in which to publish the app.

### Example

```json
{
    "action": "ElasticDuplicateApp",
    "settings": {
        "appmode": "name",
        "app": "myapp",
        "title": "duplicated app {{.Session}}"
    }
}
```

</details><details>
<summary>elasticexplore</summary>

## ElasticExplore action

Explore the hub for apps and fill the artifact map with apps to be used by other actions in the script (for example, the `openapp` action with `appmode` set to `random` or `round`).

### Settings

* `keepcurrent`: Keep the current artifact map and add the results from the `elasticexplore` action. Defaults to `false` (that is, empty the artifact map before adding the results from the `elasticexplore` action), if omitted.
* `paging`: Go through all app pages in the hub. Defaults to `false` (that is, only include the first 30 apps that the user can see), if omitted.
* `sorting`: Simulate selecting sort order in the drop-down menu in the hub
    * `default`: Default sort order (`created`).
    * `created`: Sort by the time of creation.
    * `updated`: Sort by the time of modification.
    * `name`: Sort by name.
* `owner`: Filter apps by owner
    * `all`: Apps owned by anyone.
    * `me`: Apps owned by the simulated user.
    * `others`: Apps not owned by the simulated user.
* `space`: Filter apps by space name (supports the use of [session variables](#session_variables)). **Note:** This filter cannot be used together with `spaceid`.
* `spaceid`: Filter apps by space GUID. **Note:** This filter cannot be used together with `space`.
* `tagids`: Filter apps by tag ids. This filter can be used together with `tags`.
* `tags`: Filter apps by tag names. This filter can be used together with `tagids`.

### Examples

The following example shows how to clear the artifact map and fill it with apps having the tag "mytag" from the first page in the hub.

```json
{
	"action": "ElasticExplore",
	"label": "",
	"settings": {
		"keepcurrent": false,
		"tags": ["mytag"]
	}
}
```

The following example shows how to clear the artifact map, fill it with all apps from the space "myspace" and then add all apps from the space "circles".

```json
{
	"action": "ElasticExplore",
	"label": "",
	"settings": {
		"keepcurrent": false,
		"space": "myspace",
		"paging": true
	}
},
{
	"action": "ElasticExplore",
	"label": "",
	"settings": {
		"keepcurrent": true,
		"space": "circles",
		"paging": true
	}
}
```

The following example shows how to clear the artifact map and fill it with the apps from the first page of the space "spaceX". The apps must have the tag "tag" or "team" or a tag with id "15172f9c-4a5f-4ee9-ae35-34c1edd78f8d", but not be created by the simulated user. In addition, the apps are sorted by the time of modification.

```json
{
	"action": "ElasticExplore",
	"label": "",
	"settings": {
		"keepcurrent": false,
		"space": "spaceX",
		"tags": ["tag", "team"],
		"tagids": ["15172f9c-4a5f-4ee9-ae35-34c1edd78f8d"],
		"owner": "others",
		"sorting": "updated",
		"paging": false
	}
}
```

</details><details>
<summary>elasticexportapp</summary>

## ElasticExportApp action

Export an app and, optionally, save it to file.

### Settings

* `appmode`: App selection mode
    * `current`: (default) Use the current app, selected by an app selection in a previous action, or set by the `elasticcreateapp`, `elasticduplicateapp` or `elasticuploadapp` action.
    * `guid`: Use the app GUID specified by the `app` parameter.
    * `name`: Use the app name specified by the `app` parameter.
    * `random`: Select a random app from the artifact map, which is filled by the `elasticopenhub` and/or the `elasticexplore` actions.
    * `randomnamefromlist`: Select a random app from a list of app names. The `list` parameter should contain a list of app names.
    * `randomguidfromlist`: Select a random app from a list of app GUIDs. The `list` parameter should contain a list of app GUIDs.
    * `randomnamefromfile`: Select a random app from a file with app names. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `randomguidfromfile`: Select a random app from a file with app GUIDs. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
    * `round`: Select an app from the artifact map according to the round-robin principle.
    * `roundnamefromlist`: Select an app from a list of app names according to the round-robin principle. The `list` parameter should contain a list of app names.
    * `roundguidfromlist`: Select an app from a list of app GUIDs according to the round-robin principle. The `list` parameter should contain a list of app GUIDs.
    * `roundnamefromfile`: Select an app from a file with app names according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `roundguidfromfile`: Select an app from a file with app GUIDs according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
* `app`: App name or app GUID (supports the use of [session variables](#session_variables)). Used with `appmode` set to `guid` or `name`.
* `list`: List of apps. Used with `appmode` set to `randomnamefromlist`, `randomguidfromlist`, `roundnamefromlist` or `roundguidfromlist`.
* `filename`: Path to a file in which each line represents an app. Used with `appmode` set to `randomnamefromfile`, `randomguidfromfile`, `roundnamefromfile` or `roundguidfromfile`.
* `nodata`: Export the app without data (`true`/`false`). Defaults to `false` (that is, export with data), if omitted.
* `exportname`: Pattern for the filename when saving the exported app to a file, defaults to app title or app GUID. Supports the use of [session variables](#session_variables) and additionally `.Local.Title` can be used as a variable to add the title of the exported app.
* `savetofile`: Save the exported file in the specified directory (`true`/`false`). Defaults to `false`, if omitted.

### Example

```json
{
	"action": "elasticexportapp",
	"label": "Export My App",
	"settings": {
		"appmode": "name",
		"app": "My App",
		"nodata": false,
		"savetofile": false
	}
}
```

</details><details>
<summary>elasticgenerateodag</summary>

## ElasticGenerateOdag action

Generate an on-demand app from an existing On-Demand App Generation (ODAG) link.

### Settings

* `linkname`: Name of the ODAG link from which to generate an app. The name is displayed in the ODAG navigation bar at the bottom of the *selection app*.

### Example

```json
{
    "action": "ElasticGenerateOdag",
    "settings": {
        "linkname": "Drill to Template App"
    }
}
```

</details><details>
<summary>elastichubsearch</summary>

## ElasticHubSearch action

Search the hub in a QSEoK deployment.

### Settings

* `searchfor`: 
    * `collections`: Search for collections only.
    * `apps`: Search for apps only.
    * `both`: Search for both collections and apps.
* `querysource`: 
    * `querystring`: The query is provided as a string specified by `query`.
    * `fromfile`: The queries are read from the file specified by `queryfile`, where each line represents a query.
* `query`: (optional) Query string (in case of `querystring` as source).
* `queryfile`: (optional) File from which to read a query (in case of `fromfile` as source).

### Example

```json
{
	"action": "ElasticHubSearch",
	"settings": {
		"searchfor": "apps",
		"querysource": "fromfile",
		"queryfile": "/MyQueries/Queries.txt"
	}
}
```

</details><details>
<summary>elasticmoveapp</summary>

## ElasticMoveApp action

Move an app from its existing space into the specified destination space.

**Note:** Specify *either* `destinationspacename` *or* `destinationspaceid`, not both.

### Settings

* `appmode`: App selection mode
    * `current`: (default) Use the current app, selected by an app selection in a previous action, or set by the `elasticcreateapp`, `elasticduplicateapp` or `elasticuploadapp` action.
    * `guid`: Use the app GUID specified by the `app` parameter.
    * `name`: Use the app name specified by the `app` parameter.
    * `random`: Select a random app from the artifact map, which is filled by the `elasticopenhub` and/or the `elasticexplore` actions.
    * `randomnamefromlist`: Select a random app from a list of app names. The `list` parameter should contain a list of app names.
    * `randomguidfromlist`: Select a random app from a list of app GUIDs. The `list` parameter should contain a list of app GUIDs.
    * `randomnamefromfile`: Select a random app from a file with app names. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `randomguidfromfile`: Select a random app from a file with app GUIDs. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
    * `round`: Select an app from the artifact map according to the round-robin principle.
    * `roundnamefromlist`: Select an app from a list of app names according to the round-robin principle. The `list` parameter should contain a list of app names.
    * `roundguidfromlist`: Select an app from a list of app GUIDs according to the round-robin principle. The `list` parameter should contain a list of app GUIDs.
    * `roundnamefromfile`: Select an app from a file with app names according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `roundguidfromfile`: Select an app from a file with app GUIDs according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
* `app`: App name or app GUID (supports the use of [session variables](#session_variables)). Used with `appmode` set to `guid` or `name`.
* `list`: List of apps. Used with `appmode` set to `randomnamefromlist`, `randomguidfromlist`, `roundnamefromlist` or `roundguidfromlist`.
* `filename`: Path to a file in which each line represents an app. Used with `appmode` set to `randomnamefromfile`, `randomguidfromfile`, `roundnamefromfile` or `roundguidfromfile`.
* `destinationspaceid`: Specify destination space by ID.
* `destinationspacename`: Specify destination space by name.

### Example

```json
{
    "action": "elasticmoveapp",
    "settings": {
        "app": "AppForEveryone",
        "appmode": "name",
        "destinationspacename": "everyone"
    }
}
```

</details><details>
<summary>elasticopenhub</summary>

## ElasticOpenHub action

Open the hub in a QSEoK deployment.


### Example

```json
{
	"action": "ElasticOpenHub",
	"label": "Open cloud hub with YourCollection and MyCollection"
}
```

</details><details>
<summary>elasticpublishapp</summary>

## ElasticPublishApp action

Publish an app to a managed space.

**Note:** Specify *either* `destinationspacename` *or* `destinationspaceid`, not both.

### Settings

* `appmode`: App selection mode
    * `current`: (default) Use the current app, selected by an app selection in a previous action, or set by the `elasticcreateapp`, `elasticduplicateapp` or `elasticuploadapp` action.
    * `guid`: Use the app GUID specified by the `app` parameter.
    * `name`: Use the app name specified by the `app` parameter.
    * `random`: Select a random app from the artifact map, which is filled by the `elasticopenhub` and/or the `elasticexplore` actions.
    * `randomnamefromlist`: Select a random app from a list of app names. The `list` parameter should contain a list of app names.
    * `randomguidfromlist`: Select a random app from a list of app GUIDs. The `list` parameter should contain a list of app GUIDs.
    * `randomnamefromfile`: Select a random app from a file with app names. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `randomguidfromfile`: Select a random app from a file with app GUIDs. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
    * `round`: Select an app from the artifact map according to the round-robin principle.
    * `roundnamefromlist`: Select an app from a list of app names according to the round-robin principle. The `list` parameter should contain a list of app names.
    * `roundguidfromlist`: Select an app from a list of app GUIDs according to the round-robin principle. The `list` parameter should contain a list of app GUIDs.
    * `roundnamefromfile`: Select an app from a file with app names according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `roundguidfromfile`: Select an app from a file with app GUIDs according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
* `app`: App name or app GUID (supports the use of [session variables](#session_variables)). Used with `appmode` set to `guid` or `name`.
* `list`: List of apps. Used with `appmode` set to `randomnamefromlist`, `randomguidfromlist`, `roundnamefromlist` or `roundguidfromlist`.
* `filename`: Path to a file in which each line represents an app. Used with `appmode` set to `randomnamefromfile`, `randomguidfromfile`, `roundnamefromfile` or `roundguidfromfile`.
* `destinationspaceid`: Specify destination space by ID.
* `destinationspacename`: Specify destination space by name.
* `cleartags`: Publish the app without its original tags.

### Example

```json
{
    "action": "elasticpublishapp",
    "settings": {
        "app": "Sales",
        "appmode": "name",
        "destinationspacename": "Finance",
        "cleartags": false
    }
}
```

</details><details>
<summary>elasticreload</summary>

## ElasticReload action

Reload an app by simulating selecting **Reload** in the app context menu in the hub.

### Settings

* `appmode`: App selection mode
    * `current`: (default) Use the current app, selected by an app selection in a previous action, or set by the `elasticcreateapp`, `elasticduplicateapp` or `elasticuploadapp` action.
    * `guid`: Use the app GUID specified by the `app` parameter.
    * `name`: Use the app name specified by the `app` parameter.
    * `random`: Select a random app from the artifact map, which is filled by the `elasticopenhub` and/or the `elasticexplore` actions.
    * `randomnamefromlist`: Select a random app from a list of app names. The `list` parameter should contain a list of app names.
    * `randomguidfromlist`: Select a random app from a list of app GUIDs. The `list` parameter should contain a list of app GUIDs.
    * `randomnamefromfile`: Select a random app from a file with app names. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `randomguidfromfile`: Select a random app from a file with app GUIDs. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
    * `round`: Select an app from the artifact map according to the round-robin principle.
    * `roundnamefromlist`: Select an app from a list of app names according to the round-robin principle. The `list` parameter should contain a list of app names.
    * `roundguidfromlist`: Select an app from a list of app GUIDs according to the round-robin principle. The `list` parameter should contain a list of app GUIDs.
    * `roundnamefromfile`: Select an app from a file with app names according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `roundguidfromfile`: Select an app from a file with app GUIDs according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
* `app`: App name or app GUID (supports the use of [session variables](#session_variables)). Used with `appmode` set to `guid` or `name`.
* `list`: List of apps. Used with `appmode` set to `randomnamefromlist`, `randomguidfromlist`, `roundnamefromlist` or `roundguidfromlist`.
* `filename`: Path to a file in which each line represents an app. Used with `appmode` set to `randomnamefromfile`, `randomguidfromfile`, `roundnamefromfile` or `roundguidfromfile`.
* `pollinterval`: Reload status polling interval (seconds). Defaults to 5 seconds, if omitted.

### Example

```json
{
    "label": "Reload MyApp",
    "action": "elasticreload",
    "settings": {
        "appmode": "name",
        "app": "MyApp"
    }
}
```

</details><details>
<summary>elasticshareapp</summary>

## ElasticShareApp action

Share an app with one or more groups.

### Settings

* `title`: Name of the app to share (supports the use of [session variables](#session_variables)). If `appguid` and `title` refer to different apps, `appguid` takes precedence.
* `appguid`: GUID of the app to share.
* `groups`: List of groups that should be given access to the app.

### Example

```json
{
    "action" : "ElasticShareApp",
    "label": "Share coolapp with Everyone group",
    "settings": {
        "title": "coolapp",
        "groups": ["Everyone"]
    }
}
```

</details><details>
<summary>elasticuploadapp</summary>

## ElasticUploadApp action

Upload an app to a QSEoK deployment.

### Settings

* `chunksize`: (optional) Upload chunk size (in bytes). Defaults to 300 MiB, if omitted or zero.
* `retries`: (optional) Number of consecutive retries, if a chunk fails to upload. Defaults to 0 (no retries), if omitted. The first retry is issued instantly, the second with a one second back-off period, the third with a two second back-off period, and so on.
* `mode`: Upload mode. Defaults to `tus`, if omitted.
    * `tus`: Upload the file using the [tus](https://tus.io/) chunked upload protocol.
    * `legacy`: Upload the file using a single POST payload (legacy file upload mode).
* `filename`: Local file to send as payload.
* `spaceid`: DEPRECATED
* `destinationspaceid`: Specify destination space by ID.
* `destinationspacename`: Specify destination space by name.
* `title`: Name of the app to upload (supports the use of [session variables](#session_variables)).
* `stream`: (optional) Name of the private collection or public tag under which to publish the app (supports the use of [session variables](#session_variables)).
* `streamguid`: (optional) GUID of the private collection or public tag under which to publish the app.

### Example

```json
{
     "action": "ElasticUploadApp",
     "label": "Upload myapp.qvf",
     "settings": {
         "title": "coolapp",
         "filename": "/home/root/myapp.qvf",
         "stream": "Everyone",
         "spaceid": "2342798aaefcb23",
     }
}
```

</details><details>
<summary>uploaddata</summary>

## UploadData action

Upload a data file to the Data manager.

### Settings

* `filename`: Name of the local file to send as payload.
* `destinationpath`: (optional) Path to which to upload the file. Defaults to `MyDataFiles`, if omitted.

### Example

```json
{
     "action": "UploadData",
     "settings": {
         "filename": "/home/root/data.csv"
     }
}
```

</details>
</details>
<details>
<summary>Ungrouped actions</summary>


* `title`: Name of the bookmark.
* `id`: GUID of the bookmark.


</details>

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

</details>

## Example

* [Load scenario example](./examples/configuration_example.json)

