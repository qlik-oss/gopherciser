package generated

/*
	This file has been generated, do not edit the file directly.

	Generate with go run ./generatedocs/compile/main.go or by running go generate in gopherciser root project.
*/

import "github.com/qlik-oss/gopherciser/generatedocs/pkg/common"

var (
	Actions = map[string]common.DocEntry{
		"applybookmark": {
			Description: "## ApplyBookmark action\n\nApply a bookmark in the current app.\n\n**Note:** Specify *either* `title` *or* `id`, not both.\n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"applybookmark\",\n    \"settings\": {\n        \"title\": \"My bookmark\"\n    }\n}\n```\n",
		},
		"askhubadvisor": {
			Description: "## AskHubAdvisor action\n\nPerform a query in the Qlik Sense hub insight advisor.",
			Examples:    "### Examples\n\n#### Pick queries from file\n\n```json\n{\n    \"action\": \"AskHubAdvisor\",\n    \"settings\": {\n        \"querysource\": \"file\",\n        \"file\": \"queries.txt\"\n    }\n}\n```\n\nThe file `queries.txt` contains one query and an optional weight per line. The line format is `[WEIGHT;]QUERY`.\n```txt\nshow sales per country\n5; what is the lowest price of shoes\n```\n\n#### Pick queries from list\n\n```json\n{\n    \"action\": \"AskHubAdvisor\",\n    \"settings\": {\n        \"querysource\": \"querylist\",\n        \"querylist\": [\"show sales per country\", \"what is the lowest price of shoes\"]\n    }\n}\n```\n\n#### Perform followup queries if possible (default: 0)\n\n```json\n{\n    \"action\": \"AskHubAdvisor\",\n    \"settings\": {\n        \"querysource\": \"querylist\",\n        \"querylist\": [\"show sales per country\", \"what is the lowest price of shoes\"],\n        \"maxfollowup\": 3\n    }\n}\n```\n\n#### Change lanuage (default: \"en\")\n\n```json\n{\n    \"action\": \"AskHubAdvisor\",\n    \"settings\": {\n        \"querysource\": \"querylist\",\n        \"querylist\": [\"show sales per country\", \"what is the lowest price of shoes\"],\n        \"lang\": \"fr\"\n    }\n}\n```\n\n#### Weights in querylist\n\n```json\n{\n    \"action\": \"AskHubAdvisor\",\n    \"settings\": {\n        \"querysource\": \"querylist\",\n        \"querylist\": [\n            {\n                \"query\": \"show sales per country\",\n                \"weight\": 5,\n            },\n            \"what is the lowest price of shoes\"\n        ]\n    }\n}\n```\n\n#### Thinktime before followup queries\n\nSee detailed examples of settings in the documentation for thinktime action.\n\n```json\n{\n    \"action\": \"AskHubAdvisor\",\n    \"settings\": {\n        \"querysource\": \"querylist\",\n        \"querylist\": [\n            \"what is the lowest price of shoes\"\n        ],\n        \"maxfollowup\": 5,\n        \"thinktime\": {\n            \"type\": \"static\",\n            \"delay\": 5\n        }\n    }\n}\n```\n\n#### Save chart images to file\n\n```json\n{\n    \"action\": \"AskHubAdvisor\",\n    \"settings\": {\n        \"querysource\": \"querylist\",\n        \"querylist\": [\n            \"show price per shoe type\"\n        ],\n        \"maxfollowup\": 5,\n        \"saveimages\": true\n    }\n}\n```\n\n#### Save chart images to file with custom name\n\nThe `saveimagefile` file name template setting supports\n[Session Variables](https://github.com/qlik-trial/gopherciser-oss/blob/master/docs/settingup.md#session-variables).\nYou can apart from session variables include the following action local variables in the `saveimagefile` file name template:\n- .Local.ImageCount - _the number of images written to file_\n- .Local.ServerFileName - _the server side name of image file_\n- .Local.Query - _the query sentence_\n- .Local.AppName - _the name of app, if any app, where query is asked_\n- .Local.AppID - _the id of app, if any app, where query is asked_\n\n```json\n{\n    \"action\": \"AskHubAdvisor\",\n    \"settings\": {\n        \"querysource\": \"querylist\",\n        \"querylist\": [\n            \"show price per shoe type\"\n        ],\n        \"maxfollowup\": 5,\n        \"saveimages\": true,\n        \"saveimagefile\": \"{{.Local.Query}}--app-{{.Local.AppName}}--user-{{.UserName}}--thread-{{.Thread}}--session-{{.Session}}\"\n    }\n}\n```\n",
		},
		"changesheet": {
			Description: "## ChangeSheet action\n\nChange to a new sheet, unsubscribe to the currently subscribed objects, and subscribe to all objects on the new sheet.\n\nThe action supports getting data from the following objects:\n\n* Listbox\n* Filter pane\n* Bar chart\n* Scatter plot\n* Map (only the first layer)\n* Combo chart\n* Table\n* Pivot table\n* Line chart\n* Pie chart\n* Tree map\n* Text-Image\n* KPI\n* Gauge\n* Box plot\n* Distribution plot\n* Histogram\n* Auto chart (including any support generated visualization from this list)\n* Waterfall chart\n",
			Examples:    "### Example\n\n```json\n{\n     \"label\": \"Change Sheet Dashboard\",\n     \"action\": \"ChangeSheet\",\n     \"settings\": {\n         \"id\": \"TFJhh\"\n     }\n}\n```\n",
		},
		"clearall": {
			Description: "## ClearAll action\n\nClear all selections in an app.\n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"clearall\",\n    \"label\": \"Clear all selections (1)\"\n}\n```\n",
		},
		"clearfield": {
			Description: "## ClearField action\n\nClear selections in a field.\n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"clearfield\",\n    \"label\": \"Clear selections in Alpha\",\n    \"settings\" : {\n        \"name\": \"Alpha\"\n    }\n}\n```\n",
		},
		"clickactionbutton": {
			Description: "## ClickActionButton action\n\nA `ClickActionButton`-action simulates clicking an _action-button_. An _action-button_ is a sheet item which, when clicked, executes a series of actions. The series of actions contained by an action-button begins with any number _generic button-actions_ and ends with an optional _navigation button-action_.\n\n### Supported button-actions\n#### Generic button-actions\n- Apply bookmark\n- Move backward in all selections\n- Move forward in all selections\n- Lock all selections\n- Clear all selections\n- Lock field\n- Unlock field\n- Select all in field\n- Select alternatives in field\n- Select excluded in field\n- Select possible in field\n- Select values matching search criteria in field\n- Clear selection in field\n- Toggle selection in field\n- Set value of variable\n\n#### Navigation button-actions\n- Change to first sheet\n- Change to last sheet\n- Change to previous sheet\n- Change sheet by name\n- Change sheet by ID",
			Examples:    "### Examples\n\n```json\n{\n     \"label\": \"ClickActionButton\",\n     \"action\": \"ClickActionButton\",\n     \"settings\": {\n         \"id\": \"951e2eee-ad49-4f6a-bdfe-e9e3dddeb2cd\"\n     }\n}\n```\n",
		},
		"containertab": {
			Description: "## Containertab action\n\nA `Containertab` action simulates switching the active object in a `container` object.\n",
			Examples:    "### Examples\n\n```json\n{\n  \"label\": \"Switch to object qwerty in container object XYZ\",\n  \"action\": \"containertab\",\n  \"settings\": {\n    \"containerid\": \"xyz\",\n    \"mode\": \"id\",\n    \"objectid\" : \"qwerty\"\n  }\n}\n```\n\n```json\n{\n  \"label\": \"Switch to random object in container object XYZ\",\n  \"action\": \"containertab\",\n  \"settings\": {\n    \"containerid\": \"xyz\",\n    \"mode\": \"random\"\n  }\n}\n```\n\n```json\n{\n  \"label\": \"Switch to object in first tab in container object XYZ\",\n  \"action\": \"containertab\",\n  \"settings\": {\n    \"containerid\": \"xyz\",\n    \"mode\": \"index\",\n    \"index\": 0\n  }\n}\n```\n",
		},
		"createbookmark": {
			Description: "## CreateBookmark action\n\nCreate a bookmark from the current selection and selected sheet.\n\n**Note:** Both `title` and `id` can be used to identify the bookmark in subsequent actions. \n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"createbookmark\",\n    \"settings\": {\n        \"title\": \"my bookmark\",\n        \"description\": \"This bookmark contains some interesting selections\"\n    }\n}\n```\n",
		},
		"createsheet": {
			Description: "## CreateSheet action\n\nCreate a new sheet in the current app.\n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"createsheet\",\n    \"settings\": {\n        \"title\" : \"Generated sheet\"\n    }\n}\n```\n",
		},
		"deletebookmark": {
			Description: "## DeleteBookmark action\n\nDelete one or more bookmarks in the current app.\n\n**Note:** Specify *either* `title` *or* `id`, not both.\n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"deletebookmark\",\n    \"settings\": {\n        \"mode\": \"single\",\n        \"title\": \"My bookmark\"\n    }\n}\n```\n",
		},
		"deleteodag": {
			Description: "## DeleteOdag action\n\nDelete all user-generated on-demand apps for the current user and the specified On-Demand App Generation (ODAG) link.\n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"DeleteOdag\",\n    \"settings\": {\n        \"linkname\": \"Drill to Template App\"\n    }\n}\n```\n",
		},
		"deletesheet": {
			Description: "## DeleteSheet action\n\nDelete one or more sheets in the current app.\n\n**Note:** Specify *either* `title` *or* `id`, not both.\n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"deletesheet\",\n    \"settings\": {\n        \"mode\": \"matching\",\n        \"title\": \"Test sheet\"\n    }\n}\n```\n",
		},
		"disconnectapp": {
			Description: "## DisconnectApp action\n\nDisconnect from an already connected app.\n",
			Examples:    "### Example\n\n```json\n{\n    \"label\": \"Disconnect from server\",\n    \"action\" : \"disconnectapp\"\n}\n```\n",
		},
		"disconnectenvironment": {
			Description: "## DisconnectEnvironment action\n\nDisconnect from an environment. This action will disconnect open websockets towards sense and events. The action is not needed for most scenarios, however if a scenario mixes different types of environmentsor uses custom actions towards external environment, it should be used directly after the last action towards the environment.\n\nSince the action also disconnects any open websocket to Sense apps, it does not need to be preceeded with a `disconnectapp` action.\n",
			Examples:    "### Example\n\n```json\n{\n    \"label\": \"Disconnect from environment\",\n    \"action\" : \"disconnectenvironment\"\n}\n```\n",
		},
		"dosave": {
			Description: "## DoSave action\n\n`DoSave` issues a command to engine to save the currently open app. If the simulated user does not have permission to save the app it will result in an error.",
			Examples:    "### Example\n\n```json\n{\n    \"label\": \"Save MyApp\",\n    \"action\" : \"dosave\"\n}\n```\n",
		},
		"duplicatesheet": {
			Description: "## DuplicateSheet action\n\nDuplicate a sheet, including all objects.\n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"duplicatesheet\",\n    \"label\": \"Duplicate sheet1\",\n    \"settings\":{\n        \"id\" : \"mBshXB\",\n        \"save\": true,\n        \"changesheet\": true\n    }\n}\n```\n",
		},
		"generateodag": {
			Description: "## GenerateOdag action\n\nGenerate an on-demand app from an existing On-Demand App Generation (ODAG) link.\n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"GenerateOdag\",\n    \"settings\": {\n        \"linkname\": \"Drill to Template App\"\n    }\n}\n```\n",
		},
		"iterated": {
			Description: "## Iterated action\n\nLoop one or more actions.\n\n**Note:** This action does not require an app context (that is, it does not have to be prepended with an `openapp` action).\n",
			Examples:    "### Example\n\n```json\n//Visit all sheets twice\n{\n     \"action\": \"iterated\",\n     \"label\": \"\",\n     \"settings\": {\n         \"iterations\" : 2,\n         \"actions\" : [\n            {\n                 \"action\": \"sheetchanger\"\n            },\n            {\n                \"action\": \"thinktime\",\n                \"settings\": {\n                    \"type\": \"static\",\n                    \"delay\": 5\n                }\n            }\n         ]\n     }\n}\n```\n",
		},
		"listboxselect": {
			Description: "## ListBoxSelect action\n\nPerform list object specific selectiontypes in listbox.\n\n",
			Examples:    "### Examples\n\n```json\n{\n     \"label\": \"ListBoxSelect\",\n     \"action\": \"ListBoxSelect\",\n     \"settings\": {\n         \"id\": \"951e2eee-ad49-4f6a-bdfe-e9e3dddeb2cd\",\n         \"type\": \"all\",\n         \"wrap\": true,\n         \"accept\": true\n     }\n}\n```\n",
		},
		"openapp": {
			Description: "## OpenApp action\n\nOpen an app.\n\n**Note:** If the app name is used to specify which app to open, this action cannot be the first action in the scenario. It must be preceded by an action that can populate the artifact map, such as `openhub`.\n",
			Examples:    "### Examples\n\n```json\n{\n     \"label\": \"OpenApp\",\n     \"action\": \"OpenApp\",\n     \"settings\": {\n         \"appmode\": \"guid\",\n         \"app\": \"7967af99-68b6-464a-86de-81de8937dd56\"\n     }\n}\n```\n```json\n{\n     \"label\": \"OpenApp\",\n     \"action\": \"OpenApp\",\n     \"settings\": {\n         \"appmode\": \"randomguidfromlist\",\n         \"list\": [\"7967af99-68b6-464a-86de-81de8937dd56\", \"ca1a9720-0f42-48e5-baa5-597dd11b6cad\"]\n     }\n}\n```\n",
		},
		"openhub": {
			Description: "## OpenHub action\n\nOpen the hub in a QSEoW environment.\n",
			Examples:    "### Example\n\n```json\n{\n     \"action\": \"OpenHub\",\n     \"label\": \"Open the hub\"\n}\n```\n",
		},
		"productversion": {
			Description: "## ProductVersion action\n\nRequest the product version from the server and, optionally, save it to the log. This is a lightweight request that can be used as a keep-alive message in a loop.\n",
			Examples:    "### Example\n\n```json\n//Keep-alive loop\n{\n    \"action\": \"iterated\",\n    \"settings\" : {\n        \"iterations\" : 10,\n        \"actions\" : [\n            {\n                \"action\" : \"productversion\"\n            },\n            {\n                \"action\": \"thinktime\",\n                \"settings\": {\n                    \"type\": \"static\",\n                    \"delay\": 30\n                }\n            }\n        ]\n    }\n}\n```\n",
		},
		"publishbookmark": {
			Description: "## PublishBookmark action\n\nPublish a bookmark.\n\n**Note:** Specify *either* `title` *or* `id`, not both.\n",
			Examples:    "### Example\n\nPublish the bookmark with `id` \"bookmark1\" that was created earlier on in the script.\n\n```json\n{\n    \"label\" : \"Publish bookmark 1\",\n    \"action\": \"publishbookmark\",\n    \"disabled\" : false,\n    \"settings\" : {\n        \"id\" : \"bookmark1\"\n    }\n}\n```\n\nPublish the bookmark with the `title` \"bookmark of testuser\", where \"testuser\" is the username of the simulated user.\n\n```json\n{\n    \"label\" : \"Publish bookmark 2\",\n    \"action\": \"publishbookmark\",\n    \"disabled\" : false,\n    \"settings\" : {\n        \"title\" : \"bookmark of {{.UserName}}\"\n    }\n}\n```\n",
		},
		"publishsheet": {
			Description: "## PublishSheet action\n\nPublish sheets in the current app.\n",
			Examples:    "### Example\n```json\n{\n     \"label\": \"PublishSheets\",\n     \"action\": \"publishsheet\",\n     \"settings\": {\n       \"mode\": \"sheetids\",\n       \"sheetIds\": [\"qmGcYS\", \"bKbmgT\"]\n     }\n}\n```\n",
		},
		"randomaction": {
			Description: "## RandomAction action\n\nRandomly select other actions to perform. This meta-action can be used as a starting point for your testing efforts, to simplify script authoring or to add background load.\n\n`randomaction` accepts a list of action types between which to randomize. An execution of `randomaction` executes one or more of the listed actions (as determined by the `iterations` parameter), randomly chosen by a weighted probability. If nothing else is specified, each action has a default random mode that is used. An override is done by specifying one or more parameters of the original action.\n\nEach action executed by `randomaction` is followed by a customizable `thinktime`.\n\n**Note:** The recommended way to use this action is to prepend it with an `openapp` and a `changesheet` action as this ensures that a sheet is always in context.\n",
			Examples:    "### Random action defaults\n\nThe following default values are used for the different actions:\n\n* `thinktime`: Mirrors the configuration of `thinktimesettings`\n* `sheetobjectselection`:\n\n```json\n{\n     \"settings\": \n     {\n         \"id\": <UNIFORMLY RANDOMIZED>,\n         \"type\": \"RandomFromAll\",\n         \"min\": 1,\n         \"max\": 2,\n         \"accept\": true\n     }\n}\n```\n\n* `changesheet`:\n\n```json\n{\n     \"settings\": \n     {\n         \"id\": <UNIFORMLY RANDOMIZED>\n     }\n}\n```\n\n* `clearall`:\n\n```json\n{\n     \"settings\": \n     {\n     }\n}\n```\n\n### Examples\n\n#### Generating a background load by executing 5 random actions\n\n```json\n{\n    \"action\": \"RandomAction\",\n    \"settings\": {\n        \"iterations\": 5,\n        \"actions\": [\n            {\n                \"type\": \"thinktime\",\n                \"weight\": 1\n            },\n            {\n                \"type\": \"sheetobjectselection\",\n                \"weight\": 3\n            },\n            {\n                \"type\": \"changesheet\",\n                \"weight\": 5\n            },\n            {\n                \"type\": \"clearall\",\n                \"weight\": 1\n            }\n        ],\n        \"thinktimesettings\": {\n            \"type\": \"uniform\",\n            \"mean\": 10,\n            \"dev\": 5\n        }\n    }\n}\n```\n\n#### Making random selections from excluded values\n\n```json\n{\n    \"action\": \"RandomAction\",\n    \"settings\": {\n        \"iterations\": 1,\n        \"actions\": [\n            {\n                \"type\": \"sheetobjectselection\",\n                \"weight\": 1,\n                \"overrides\": {\n                  \"type\": \"RandomFromExcluded\",\n                  \"min\": 1,\n                  \"max\": 5\n                }\n            }\n        ],\n        \"thinktimesettings\": {\n            \"type\": \"static\",\n            \"delay\": 1\n        }\n    }\n}\n```\n",
		},
		"reload": {
			Description: "## Reload action\n\nReload the current app by simulating selecting **Load data** in the Data load editor. To select an app, preceed this action with an `openapp` action.\n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"reload\",\n    \"settings\": {\n        \"mode\" : \"default\",\n        \"partial\": false\n    }\n}\n```\n",
		},
		"select": {
			Description: "## Select action\n\nSelect random values in an object.\n\nSee the [Limitations](README.md#limitations) section in the README.md file for limitations related to this action.\n ",
			Examples:    "### Example\n\nRandomly select among all the values in object `RZmvzbF`.\n\n```json\n{\n     \"label\": \"ListBox Year\",\n     \"action\": \"Select\",\n     \"settings\": {\n         \"id\": \"RZmvzbF\",\n         \"type\": \"RandomFromAll\",\n         \"accept\": true,\n         \"wrap\": false,\n         \"min\": 1,\n         \"max\": 3,\n         \"dim\": 0\n     }\n}\n```\n\nRandomly select among all the enabled values (a.k.a \"white\" values) in object `RZmvzbF`.\n\n```json\n{\n     \"label\": \"ListBox Year\",\n     \"action\": \"Select\",\n     \"settings\": {\n         \"id\": \"RZmvzbF\",\n         \"type\": \"RandomFromEnabled\",\n         \"accept\": true,\n         \"wrap\": false,\n         \"min\": 1,\n         \"max\": 3,\n         \"dim\": 0\n     }\n}\n```\n\n#### Statically selecting specific values\n\nThis example selects specific element values in object `RZmvzbF`. These are the values which can be seen in a selection when e.g. inspecting traffic, it is not the data values presented to the user. E.g. when loading a table in the following order by a Sense loadscript:\n\n```\nBeta\nAlpha\nGamma\n```\n\nwhich might be presented to the user sorted as\n\n```\nAlpha\nBeta\nGamma\n```\n\nThe element values will be Beta=0, Alpha=1 and Gamma=2.\n\nTo statically select \"Gamma\" in this case:\n\n```json\n{\n     \"label\": \"Select Gammma\",\n     \"action\": \"Select\",\n     \"settings\": {\n         \"id\": \"RZmvzbF\",\n         \"type\": \"values\",\n         \"accept\": true,\n         \"wrap\": false,\n         \"values\" : [2],\n         \"dim\": 0\n     }\n}\n```\n",
		},
		"setscript": {
			Description: "## SetScript action\n\nSet the load script for the current app. To load the data from the script, use the `reload` action after the `setscript` action.\n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"setscript\",\n    \"settings\": {\n        \"script\" : \"Characters:\\nLoad Chr(RecNo()+Ord('A')-1) as Alpha, RecNo() as Num autogenerate 26;\"\n    }\n}\n```\n",
		},
		"setscriptvar": {
			Description: "## SetScriptVar action\n\nSets a variable which can be used within the same session. Cannot be accessed across different simulated users.\n",
			Examples:    "### Example\n\nCreate a variable containing a string and use it in openapp.\n\n```json\n{\n    \"action\": \"setscriptvar\",\n    \"settings\": {\n        \"name\": \"mylocalvar\",\n        \"type\": \"string\",\n        \"value\": \"My app Name with number for session {{ .Session }}\"\n    }\n},\n{\n    \"action\": \"openapp\",\n    \"settings\": {\n        \"appmode\": \"name\",\n        \"app\": \"{{ .ScriptVars.mylocalvar }}\"\n    }\n}\n```\n\nCreate a variable containing an integer and use it in a loop creating bookmarks numbered 1 to 5. Then in a different loop reset variable and delete the bookmarks.\n\n```json\n{\n    \"action\": \"setscriptvar\",\n    \"settings\": {\n        \"name\": \"BookmarkCounter\",\n        \"type\": \"int\",\n        \"value\": \"0\"\n    }\n},\n{\n    \"action\": \"iterated\",\n    \"settings\": {\n        \"iterations\": 5,\n        \"actions\": [\n            {\n                \"action\": \"setscriptvar\",\n                \"settings\": {\n                    \"name\": \"BookmarkCounter\",\n                    \"type\": \"int\",\n                    \"value\": \"{{ add .ScriptVars.BookmarkCounter 1 }}\"\n                }\n            },\n            {\n                \"action\": \"createbookmark\",\n                \"settings\": {\n                    \"title\": \"Bookmark {{ .ScriptVars.BookmarkCounter }}\",\n                    \"description\": \"This bookmark contains some interesting selections\"\n                }\n            }\n            \n        ]\n    }\n},\n{\n    \"action\": \"setscriptvar\",\n    \"settings\": {\n        \"name\": \"BookmarkCounter\",\n        \"type\": \"int\",\n        \"value\": \"0\"\n    }\n},\n{\n    \"action\": \"iterated\",\n    \"disabled\": false,\n    \"settings\": {\n        \"iterations\": 3,\n        \"actions\": [\n            {\n                \"action\": \"setscriptvar\",\n                \"settings\": {\n                    \"name\": \"BookmarkCounter\",\n                    \"type\": \"int\",\n                    \"value\": \"{{ .ScriptVars.BookmarkCounter | add 1}}\"\n                }\n            },\n            {\n                \"action\": \"deletebookmark\",\n                \"settings\": {\n                    \"mode\": \"single\",\n                    \"title\": \"Bookmark {{ $element:=range.ScriptVars.BookmarkCounter }} {{ $element }}{{ end }}\"\n                }\n            }\n        ]\n    }\n}\n```\n\nCombine two variables `MyArrayVar` and `BookmarkCounter` to create 3 bookmarks with the names `Bookmark one`, `Bookmark two` and `Bookmark three`.\n\n```json\n{\n    \"action\": \"setscriptvar\",\n    \"settings\": {\n        \"name\": \"MyArrayVar\",\n        \"type\": \"array\",\n        \"value\": \"one,two,three,four,five\",\n        \"sep\": \",\"\n    }           \n},\n{\n    \"action\": \"setscriptvar\",\n    \"settings\": {\n        \"name\": \"BookmarkCounter\",\n        \"type\": \"int\",\n        \"value\": \"0\"\n    }\n},\n{\n    \"action\": \"iterated\",\n    \"disabled\": false,\n    \"settings\": {\n        \"iterations\": 3,\n        \"actions\": [\n            {\n                \"action\": \"createbookmark\",\n                \"settings\": {\n                    \"title\": \"Bookmark {{ index .ScriptVars.MyArrayVar .ScriptVars.BookmarkCounter }}\",\n                    \"description\": \"This bookmark contains some interesting selections\"\n                }\n            },\n            {\n                \"action\": \"setscriptvar\",\n                \"settings\": {\n                    \"name\": \"BookmarkCounter\",\n                    \"type\": \"int\",\n                    \"value\": \"{{ .ScriptVars.BookmarkCounter | add 1}}\"\n                }\n            }\n        ]\n    }\n}\n ```\n\nA more advanced example.\n\nCreate a bookmark \"BookmarkX\" for each iteration in a loop, and add this to an array \"MyArrayVar\". After the first `iterated` action this will look like \"Bookmark1,Bookmark2,Bookmark3\". The second `iterated` action then deletes these bookmarks using the created array.\n\nDissecting the first array construction action. The `join` command takes the elements `.ScriptVars.MyArrayVar` and joins them together into a string separated by the separtor `,`. So with an array of [ elem1 elem2 ] this becomes a string as `elem1,elem2`. The `if` statement checks if the value of `.ScriptVars.BookmarkCounter` is 0, if it is 0 (i.e. the first iteration) it sets the string to `Bookmark1`. If it is not 0, it executes the join command on .ScriptVars.MyArrayVar, on iteration 3, the result of this would be `Bookmark1,Bookmark2` then it appends the fixed string `,Bookmark`, so far the string is `Bookmark1,Bookmark2,Bookmark`. Lastly it takes the value of `.ScriptVars.BookmarkCounter`, which is now 2, and adds 1 too it and appends, making the entire string `Bookmark1,Bookmark2,Bookmark3`.\n\n ```json\n{\n    \"action\": \"setscriptvar\",\n    \"settings\": {\n        \"name\": \"BookmarkCounter\",\n        \"type\": \"int\",\n        \"value\": \"0\"\n    }\n},\n{\n    \"action\": \"iterated\",\n    \"disabled\": false,\n    \"settings\": {\n        \"iterations\": 3,\n        \"actions\": [\n            {\n                \"action\": \"setscriptvar\",\n                \"settings\": {\n                    \"name\": \"MyArrayVar\",\n                    \"type\": \"array\",\n                    \"value\": \"{{ if eq 0 .ScriptVars.BookmarkCounter }}Bookmark1{{ else }}{{ join .ScriptVars.MyArrayVar \\\",\\\" }},Bookmark{{ .ScriptVars.BookmarkCounter | add 1 }}{{ end }}\",\n                    \"sep\": \",\"\n                }\n            },\n            {\n                \"action\": \"createbookmark\",\n                \"settings\": {\n                    \"title\": \"{{ index .ScriptVars.MyArrayVar .ScriptVars.BookmarkCounter }}\",\n                    \"description\": \"This bookmark contains some interesting selections\"\n                }\n            },\n            {\n                \"action\": \"setscriptvar\",\n                \"settings\": {\n                    \"name\": \"BookmarkCounter\",\n                    \"type\": \"int\",\n                    \"value\": \"{{ .ScriptVars.BookmarkCounter | add 1}}\"\n                }\n            }\n        ]\n    }\n},\n{\n    \"action\": \"setscriptvar\",\n    \"settings\": {\n        \"name\": \"BookmarkCounter\",\n        \"type\": \"int\",\n        \"value\": \"0\"\n    }\n},\n{\n    \"action\": \"iterated\",\n    \"disabled\": false,\n    \"settings\": {\n        \"iterations\": 3,\n        \"actions\": [\n            {\n                \"action\": \"deletebookmark\",\n                \"settings\": {\n                    \"mode\": \"single\",\n                    \"title\": \"{{ index .ScriptVars.MyArrayVar .ScriptVars.BookmarkCounter }}\"\n                }\n            },\n            {\n                \"action\": \"setscriptvar\",\n                \"settings\": {\n                    \"name\": \"BookmarkCounter\",\n                    \"type\": \"int\",\n                    \"value\": \"{{ .ScriptVars.BookmarkCounter | add 1}}\"\n                }\n            }\n        ]\n    }\n}\n ```",
		},
		"setsensevariable": {
			Description: "## SetSenseVariable action\n\nSets a Qlik Sense variable on a sheet in the open app.\n",
			Examples:    "### Example\n\nSet a variable to 2000\n\n```json\n{\n     \"name\": \"vSampling\",\n     \"value\": \"2000\"\n}\n```",
		},
		"sheetchanger": {
			Description: "## SheetChanger action\n\nCreate and execute a `changesheet` action for each sheet in an app. This can be used to cache the inital state for all objects or, by chaining two subsequent `sheetchanger` actions, to measure how well the calculations in an app utilize the cache.\n",
			Examples:    "### Example\n\n```json\n{\n    \"label\" : \"Sheetchanger uncached\",\n    \"action\": \"sheetchanger\"\n},\n{\n    \"label\" : \"Sheetchanger cached\",\n    \"action\": \"sheetchanger\"\n}\n```\n",
		},
		"subscribeobjects": {
			Description: "## Subscribeobjects action\n\nSubscribe to any object in the currently active app.\n",
			Examples:    "### Example\n\nSubscribe to two objects in the currently active app and remove any previous subscriptions. \n\n```json\n{\n    \"action\" : \"subscribeobjects\",\n    \"label\" : \"clear subscriptions and subscribe to mBshXB and f2a50cb3-a7e1-40ac-a015-bc4378773312\",\n     \"disabled\": false,\n    \"settings\" : {\n        \"clear\" : true,\n        \"ids\" : [\"mBshXB\", \"f2a50cb3-a7e1-40ac-a015-bc4378773312\"]\n    }\n}\n```\n\nSubscribe to an additional single object (or a list of objects) in the currently active app, adding the new subscription to any previous subscriptions.\n\n```json\n{\n    \"action\" : \"subscribeobjects\",\n    \"label\" : \"add c430d8e2-0f05-49f1-aa6f-7234e325dc35 to currently subscribed objects\",\n     \"disabled\": false,\n    \"settings\" : {\n        \"clear\" : false,\n        \"ids\" : [\"c430d8e2-0f05-49f1-aa6f-7234e325dc35\"]\n    }\n}\n```",
		},
		"thinktime": {
			Description: "## ThinkTime action\n\nSimulate user think time.\n\n**Note:** This action does not require an app context (that is, it does not have to be prepended with an `openapp` action).\n",
			Examples:    "### Examples\n\n#### ThinkTime uniform\n\nThis simulates a think time of 10 to 15 seconds.\n\n```json\n{\n     \"label\": \"TimerDelay\",\n     \"action\": \"thinktime\",\n     \"settings\": {\n         \"type\": \"uniform\",\n         \"mean\": 12.5,\n         \"dev\": 2.5\n     } \n} \n```\n\n#### ThinkTime constant\n\nThis simulates a think time of 5 seconds.\n\n```json\n{\n     \"label\": \"TimerDelay\",\n     \"action\": \"thinktime\",\n     \"settings\": {\n         \"type\": \"static\",\n         \"delay\": 5\n     }\n}\n```\n",
		},
		"unpublishbookmark": {
			Description: "## UnpublishBookmark action\n\nUnpublish a bookmark.\n\n**Note:** Specify *either* `title` *or* `id`, not both.\n",
			Examples:    "### Example\n\nUnpublish the bookmark with `id` \"bookmark1\" that was created earlier on in the script.\n\n```json\n{\n    \"label\" : \"Unpublish bookmark 1\",\n    \"action\": \"unpublishbookmark\",\n    \"disabled\" : false,\n    \"settings\" : {\n        \"id\" : \"bookmark1\"\n    }\n}\n```\n\nUnpublish the bookmark with the `title` \"bookmark of testuser\", where \"testuser\" is the username of the simulated user.\n\n```json\n{\n    \"label\" : \"Unpublish bookmark 2\",\n    \"action\": \"unpublishbookmark\",\n    \"disabled\" : false,\n    \"settings\" : {\n        \"title\" : \"bookmark of {{.UserName}}\"\n    }\n}\n```\n",
		},
		"unpublishsheet": {
			Description: "## UnpublishSheet action\n\nUnpublish sheets in the current app.\n",
			Examples:    "### Example\n```json\n{\n     \"label\": \"UnpublishSheets\",\n     \"action\": \"unpublishsheet\",\n     \"settings\": {\n       \"mode\": \"allsheets\"        \n     }\n}\n```\n",
		},
		"unsubscribeobjects": {
			Description: "## Unsubscribeobjects action\n\nUnsubscribe to any currently subscribed object.\n",
			Examples:    "### Example\n\nUnsubscribe from a single object (or a list of objects).\n\n```json\n{\n    \"action\" : \"unsubscribeobjects\",\n    \"label\" : \"unsubscribe from object maVjt and its children\",\n    \"disabled\": false,\n    \"settings\" : {\n        \"ids\" : [\"maVjt\"]\n    }\n}\n```\n\nUnsubscribe from all currently subscribed objects.\n\n```json\n{\n    \"action\" : \"unsubscribeobjects\",\n    \"label\" : \"unsubscribe from all objects\",\n    \"disabled\": false,\n    \"settings\" : {\n        \"clear\": true\n    }\n}\n```",
		},
	}

	Params = map[string][]string{
		"applybookmark.selectionsonly":                    {"Apply selections only."},
		"appselection.app":                                {"App name or app GUID (supports the use of [session variables](#session_variables)). Used with `appmode` set to `guid` or `name`."},
		"appselection.appmode":                            {"App selection mode", "`current`: (default) Use the current app, selected by an app selection in a previous action", "`guid`: Use the app GUID specified by the `app` parameter.", "`name`: Use the app name specified by the `app` parameter.", "`random`: Select a random app from the artifact map, which is filled by e.g. `openhub`", "`randomnamefromlist`: Select a random app from a list of app names. The `list` parameter should contain a list of app names.", "`randomguidfromlist`: Select a random app from a list of app GUIDs. The `list` parameter should contain a list of app GUIDs.", "`randomnamefromfile`: Select a random app from a file with app names. The `filename` parameter should contain the path to a file in which each line represents an app name.", "`randomguidfromfile`: Select a random app from a file with app GUIDs. The `filename` parameter should contain the path to a file in which each line represents an app GUID.", "`round`: Select an app from the artifact map according to the round-robin principle.", "`roundnamefromlist`: Select an app from a list of app names according to the round-robin principle. The `list` parameter should contain a list of app names.", "`roundguidfromlist`: Select an app from a list of app GUIDs according to the round-robin principle. The `list` parameter should contain a list of app GUIDs.", "`roundnamefromfile`: Select an app from a file with app names according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app name.", "`roundguidfromfile`: Select an app from a file with app GUIDs according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app GUID."},
		"appselection.filename":                           {"Path to a file in which each line represents an app. Used with `appmode` set to `randomnamefromfile`, `randomguidfromfile`, `roundnamefromfile` or `roundguidfromfile`."},
		"appselection.list":                               {"List of apps. Used with `appmode` set to `randomnamefromlist`, `randomguidfromlist`, `roundnamefromlist` or `roundguidfromlist`."},
		"askhubadvisor.app":                               {"Optional name of app to pick in followup queries. If not set, a random app is picked."},
		"askhubadvisor.file":                              {"Path to query file."},
		"askhubadvisor.lang":                              {"Query language."},
		"askhubadvisor.maxfollowup":                       {"The maximum depth of followup queries asked. A value of `0` means that a query from querysource is performed without followup queries."},
		"askhubadvisor.querylist":                         {"A list of queries. Plain strings are supported and will get a weight of `1`."},
		"askhubadvisor.querylist.query":                   {"A query sentence."},
		"askhubadvisor.querylist.weight":                  {"A weight to set probablility of query being peformed."},
		"askhubadvisor.querysource":                       {"The source from which queries will be randomly picked.", "`file`: Read queries from file defined by `file`.", "`querylist`: Read queries from list defined by `querylist`."},
		"askhubadvisor.saveimagefile":                     {"File name of saved images. Defaults to server side file name. Supports [Session Variables](https://github.com/qlik-trial/gopherciser-oss/blob/master/docs/settingup.md#session-variables)."},
		"askhubadvisor.saveimages":                        {"Save images of charts to file."},
		"askhubadvisor.thinktime":                         {"Settings for the `thinktime` action, which is automatically inserted before each followup."},
		"bookmark.id":                                     {"ID of the bookmark."},
		"bookmark.title":                                  {"Name of the bookmark (supports the use of [variables](#session_variables))."},
		"changesheet.id":                                  {"GUID of the sheet to change to."},
		"clearfield.name":                                 {"Name of field to clear."},
		"clickactionbutton.id":                            {"ID of the action-button to click."},
		"config.connectionSettings.allowuntrusted":        {"Allow untrusted (for example, self-signed) certificates (`true` / `false`). Defaults to `false`, if omitted."},
		"config.connectionSettings.appext":                {"Replace `app` in the connect URL for the `openapp` action. Defaults to `app`, if omitted."},
		"config.connectionSettings.headers":               {"Headers to use in requests."},
		"config.connectionSettings.jwtsettings":           {"(JWT only) Settings for the JWT connection."},
		"config.connectionSettings.jwtsettings.alg":       {"The signing method used for the JWT. Defaults to `RS512`, if omitted.", "For keyfiles in RSA format, supports `RS256`, `RS384` or `RS512`.", "For keyfiles in EC format, supports `ES256`, `ES384` or `ES512`."},
		"config.connectionSettings.jwtsettings.claims":    {"JWT claims as an escaped JSON string."},
		"config.connectionSettings.jwtsettings.jwtheader": {"JWT headers as an escaped JSON string. Custom headers to be added to the JWT header."},
		"config.connectionSettings.jwtsettings.keypath":   {"Local path to the JWT key file."},
		"config.connectionSettings.mode":                  {"Authentication mode", "`jwt`: JSON Web Token", "`ws`: WebSocket"},
		"config.connectionSettings.port":                  {"Set another port than default (`80` for http and `443` for https)."},
		"config.connectionSettings.rawurl":                {"Define the connect URL manually instead letting the `openapp` action do it. **Note**: The protocol must be `wss://` or `ws://`."},
		"config.connectionSettings.security":              {"Use TLS (SSL) (`true` / `false`)."},
		"config.connectionSettings.server":                {"Qlik Sense host."},
		"config.connectionSettings.virtualproxy":          {"Prefix for the virtual proxy that handles the virtual users."},
		"config.connectionSettings.wssettings":            {"(WebSocket only) Settings for the WebSocket connection."},
		"config.loginSettings":                            {"This section of the JSON file contains information on the login settings."},
		"config.loginSettings.settings":                   {"", "`userList`: List of users for the `userlist` login request type. Directory and password can be specified per user or outside the list of usernames, which means that they are inherited by all users."},
		"config.loginSettings.settings.directory":         {"Directory to set for the users."},
		"config.loginSettings.settings.prefix":            {"Prefix to add to the username, so that it will be `prefix_{session}`."},
		"config.loginSettings.type":                       {"Type of login request", "`prefix`: Add a prefix (specified by the `prefix` setting below) to the username, so that it will be `prefix_{session}`.", "`userlist`: List of users as specified by the `userList` setting below.", "`none`: Do not add a prefix to the username, so that it will be `{session}`."},
		"config.scenario":                                 {"This section of the JSON file contains the actions that are performed in the load scenario."},
		"config.scenario.action":                          {"Name of the action to execute."},
		"config.scenario.disabled":                        {"(optional) Disable action (`true` / `false`). If set to `true`, the action is not executed."},
		"config.scenario.label":                           {"(optional) Custom string set by the user. This can be used to distinguish the action from other actions of the same type when analyzing the test results."},
		"config.scenario.settings":                        {"Most, but not all, actions have a settings section with action-specific settings."},
		"config.scheduler":                                {"This section of the JSON file contains scheduler settings for the users in the load scenario."},
		"config.scheduler.instance":                       {"Instance number for this instance. Use different instance numbers when running the same script in multiple instances to make sure the randomization is different in each instance. Defaults to 1."},
		"config.scheduler.iterationtimebuffer":            {""},
		"config.scheduler.iterationtimebuffer.duration":   {"Duration of the time buffer (for example, `500ms`, `30s` or `1m10s`). Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, and `h`."},
		"config.scheduler.iterationtimebuffer.mode":       {"Time buffer mode. Defaults to `nowait`, if omitted.", "`nowait`: No time buffer in between the iterations.", "`constant`: Add a constant time buffer after each iteration. Defined by `duration`.", "`onerror`: Add a time buffer in case of an error. Defined by `duration`.", "`minduration`: Add a time buffer if the iteration duration is less than `duration`."},
		"config.scheduler.reconnectsettings":              {"Settings for enabling re-connection attempts in case of unexpected disconnects."},
		"config.scheduler.settings":                       {""},
		"config.scheduler.settings.concurrentusers":       {"Number of concurrent users to simulate. Allowed values are positive integers."},
		"config.scheduler.settings.executiontime":         {"Test execution time (seconds). The sessions are disconnected when the specified time has elapsed. Allowed values are positive integers. `-1` means an infinite execution time."},
		"config.scheduler.settings.iterations":            {"Number of iterations for each 'concurrent' user to repeat. Allowed values are positive integers. `-1` means an infinite number of iterations."},
		"config.scheduler.settings.onlyinstanceseed":      {"Disable session part of randomization seed. Defaults to `false`, if omitted.", "`true`: All users and sessions have the same randomization sequence, which only changes if the `instance` flag is changed.", "`false`: Normal randomization sequence, dependent on both the `instance` parameter and the current user session."},
		"config.scheduler.settings.rampupdelay":           {"Time delay (seconds) scheduled in between each concurrent user during the startup period."},
		"config.scheduler.settings.reuseusers":            {"", "`true`: Every iteration for each concurrent user uses the same user and session.", "`false`: Every iteration for each concurrent user uses a new user and session. The total number of users is the product of `concurrentusers` and `iterations`."},
		"config.scheduler.type":                           {"Type of scheduler", "`simple`: Standard scheduler"},
		"config.settings":                                 {"This section of the JSON file contains timeout and logging settings for the load scenario"},
		"config.settings.logs":                            {"Log settings"},
		"config.settings.logs.debug":                      {"Log debug information (`true` / `false`). Defaults to `false`, if omitted."},
		"config.settings.logs.filename":                   {"Name of the log file (supports the use of [variables](#session_variables))."},
		"config.settings.logs.format":                     {"Log format. Defaults to `tsvfile`, if omitted.", "`tsvfile`: Log to file in TSV format and output status to console.", "`tsvconsole`: Log to console in TSV format without any status output.", "`jsonfile`: Log to file in JSON format and output status to console.", "`jsonconsole`: Log to console in JSON format without any status output.", "`console`: Log to console in color format without any status output.", "`combined`: Log to file in TSV format and to console in JSON format.", "`no`: Default logs and status output turned off.", "`onlystatus`: Default logs turned off, but status output turned on."},
		"config.settings.logs.metrics":                    {"Log traffic metrics (`true` / `false`). Defaults to `false`, if omitted. **Note:** This should only be used for debugging purposes as traffic logging is resource-demanding."},
		"config.settings.logs.regression":                 {"Log regression data (`true` / `false`). Defaults to `false`, if omitted. **Note:** Do not log regression data when testing performance. **Note** With regression logging enabled, the the scheduler is implicitly set to execute the scenario as one user for one iteration."},
		"config.settings.logs.summary":                    {"Type of summary to display after the test run. Defaults to simple for minimal performance impact.", "`0` or `undefined`: Simple, single-row summary", "`1` or `none`: No summary", "`2` or `simple`: Simple, single-row summary", "`3` or `extended`: Extended summary that includes statistics on each unique combination of action, label and app GUID", "`4` or `full`: Same as extended, but with statistics on each unique combination of method and endpoint added"},
		"config.settings.logs.traffic":                    {"Log traffic information (`true` / `false`). Defaults to `false`, if omitted. **Note:** This should only be used for debugging purposes as traffic logging is resource-demanding."},
		"config.settings.outputs":                         {"Used by some actions to save results to a file."},
		"config.settings.outputs.dir":                     {"Directory in which to save artifacts generated by the script (except log file)."},
		"config.settings.timeout":                         {"Timeout setting (seconds) for WebSocket requests."},
		"containertab.containerid":                        {"ID of the container object."},
		"containertab.index":                              {"Zero based index of tab to switch to, used with mode `index`."},
		"containertab.mode":                               {"Mode for container tab switching, one of: `objectid`, `random` or `index`.", "`objectid`: Switch to tab with object defined by `objectid`.", "`random`: Switch to a random visible tab within the container.", "`index`: Switch to tab with zero based index defined but `index`."},
		"containertab.objectid":                           {"ID of the object to set as active, used with mode `objectid`."},
		"createbookmark.description":                      {"(optional) Description of the bookmark to create."},
		"createbookmark.nosheet":                          {"Do not include the sheet location in the bookmark."},
		"createbookmark.savelayout":                       {"Include the layout in the bookmark."},
		"createsheet.description":                         {"(optional) Description of the sheet to create."},
		"createsheet.id":                                  {"(optional) ID to be used to identify the sheet in any subsequent `changesheet`, `duplicatesheet`, `publishsheet` or `unpublishsheet` action."},
		"createsheet.title":                               {"Name of the sheet to create."},
		"deletebookmark.mode":                             {"", "`single`: Delete one bookmark that matches the specified `title` or `id` in the current app.", "`matching`: Delete all bookmarks with the specified `title` in the current app.", "`all`: Delete all bookmarks in the current app."},
		"deleteodag.linkname":                             {"Name of the ODAG link from which to delete generated apps. The name is displayed in the ODAG navigation bar at the bottom of the *selection app*."},
		"deletesheet.id":                                  {"(optional) GUID of the sheet to delete."},
		"deletesheet.mode":                                {"", "`single`: Delete one sheet that matches the specified `title` or `id` in the current app.", "`matching`: Delete all sheets with the specified `title` in the current app.", "`allunpublished`: Delete all unpublished sheets in the current app."},
		"deletesheet.title":                               {"(optional) Name of the sheet to delete."},
		"destinationspace.destinationspaceid":             {"Specify destination space by ID."},
		"destinationspace.destinationspacename":           {"Specify destination space by name."},
		"duplicatesheet.changesheet":                      {"Clear the objects currently subscribed to and then subribe to all objects on the cloned sheet (which essentially corresponds to using the `changesheet` action to go to the cloned sheet) (`true` / `false`). Defaults to `false`, if omitted."},
		"duplicatesheet.cloneid":                          {"(optional) ID to be used to identify the sheet in any subsequent `changesheet`, `duplicatesheet`, `publishsheet` or `unpublishsheet` action."},
		"duplicatesheet.id":                               {"ID of the sheet to clone."},
		"duplicatesheet.save":                             {"Execute `saveobjects` after the cloning operation to save all modified objects (`true` / `false`). Defaults to `false`, if omitted."},
		"generateodag.linkname":                           {"Name of the ODAG link from which to generate an app. The name is displayed in the ODAG navigation bar at the bottom of the *selection app*."},
		"iterated.actions":                                {"Actions to iterate"},
		"iterated.iterations":                             {"Number of loops."},
		"listboxselect.accept":                            {"Accept or abort selection after selection (only used with `wrap`) (`true` / `false`)."},
		"listboxselect.id":                                {"ID of the listbox in which to select values."},
		"listboxselect.type":                              {"Selection type.", "`all`: Select all values.", "`alternative`: Select alternative values.", "`excluded`: Select excluded values.", "`possible`: Select possible values."},
		"listboxselect.wrap":                              {"Wrap selection with Begin / End selection requests (`true` / `false`)."},
		"openapp.unique":                                  {"Create unqiue engine session not re-using session from previous connection with same user. Defaults to false."},
		"productversion.log":                              {"Save the product version to the log (`true` / `false`). Defaults to `false`, if omitted."},
		"publishsheet.mode":                               {"", "`allsheets`: Publish all sheets in the app.", "`sheetids`: Only publish the sheets specified by the `sheetIds` array."},
		"publishsheet.sheetIds":                           {"(optional) Array of sheet IDs for the `sheetids` mode."},
		"randomaction.actions":                            {"List of actions from which to randomly pick an action to execute. Each item has a number of possible parameters."},
		"randomaction.actions.overrides":                  {"(optional) Static overrides to the action. The overrides can include any or all of the settings from the original action, as determined by the `type` field. If nothing is specified, the default values are used."},
		"randomaction.actions.type":                       {"Type of action", "`thinktime`: See the `thinktime` action.", "`sheetobjectselection`: Make random selections within objects visible on the current sheet. See the `select` action.", "`changesheet`: See the `changesheet` action.", "`clearall`: See the `clearall` action."},
		"randomaction.actions.weight":                     {"The probabilistic weight of the action, specified as an integer. This number is proportional to the likelihood of the specified action, and is used as a weight in a uniform random selection."},
		"randomaction.iterations":                         {"Number of random actions to perform."},
		"randomaction.thinktimesettings":                  {"Settings for the `thinktime` action, which is automatically inserted after every randomized action."},
		"reconnectsettings.backoff":                       {"Re-connection backoff scheme. Defaults to `[0.0, 2.0, 2.0, 2.0, 2.0, 2.0]`, if left empty. An example backoff scheme could be `[0.0, 1.0, 10.0, 20.0]`:", "`0.0`: If the WebSocket is disconnected, wait 0.0s before attempting to re-connect", "`1.0`: If the previous attempt to re-connect failed, wait 1.0s before attempting again", "`10.0`: If the previous attempt to re-connect failed, wait 10.0s before attempting again", "`20.0`: If the previous attempt to re-connect failed, wait 20.0s before attempting again"},
		"reconnectsettings.reconnect":                     {"Enable re-connection attempts if the WebSocket is disconnected. Defaults to `false`."},
		"reload.log":                                      {"Save the reload log as a field in the output (`true` / `false`). Defaults to `false`, if omitted. **Note:** This should only be used when needed as the reload log can become very large."},
		"reload.mode":                                     {"Error handling during the reload operation", "`default`: Use the default error handling.", "`abend`: Stop reloading the script, if an error occurs.", "`ignore`: Continue reloading the script even if an error is detected in the script."},
		"reload.partial":                                  {"Enable partial reload (`true` / `false`). This allows you to add data to an app without reloading all data. Defaults to `false`, if omitted."},
		"select.accept":                                   {"Accept or abort selection after selection (only used with `wrap`) (`true` / `false`)."},
		"select.dim":                                      {"Dimension / column in which to select."},
		"select.id":                                       {"ID of the object in which to select values."},
		"select.max":                                      {"Maximum number of selections to make."},
		"select.min":                                      {"Minimum number of selections to make."},
		"select.type":                                     {"Selection type", "`randomfromall`: Randomly select within all values of the symbol table.", "`randomfromenabled`: Randomly select within the white and light grey values on the first data page.", "`randomfromexcluded`: Randomly select within the dark grey values on the first data page.", "`randomdeselect`: Randomly deselect values on the first data page.", "`values`: Select specific element values, defined by `values` array."},
		"select.values":                                   {"Array of element values to select when using selection type `values`. These are the element values for a selection, not the values seen by the user."},
		"select.wrap":                                     {"Wrap selection with Begin / End selection requests (`true` / `false`)."},
		"setscript.script":                                {"Load script for the app (written as a string)."},
		"setscriptvar.name":                               {"Name of variable to set. Will overwrite any existing variable with same name."},
		"setscriptvar.type":                               {"Type of the variable.", "`string`: Variable of type string e.g. `my var value`.", "`int`: Variable of type integer e.g. `6`.", "`array`: Variable of type arrat e.g. `[1, 2, 3]`."},
		"setscriptvar.value":                              {"Value to set to variable (supports the use of [session variables](#session_variables))."},
		"setsensevariable.name":                           {"Name of the Qlik Sense variable to set."},
		"setsensevariable.value":                          {"Value to set the Qlik Sense variable to."},
		"subscribeobjects.clear":                          {"Remove any previously subscribed objects from the subscription list."},
		"subscribeobjects.ids":                            {"List of object IDs to subscribe to."},
		"thinktime.delay":                                 {"Delay (seconds), used with type `static`."},
		"thinktime.dev":                                   {"Deviation (seconds) from `mean` value, used with type `uniform`."},
		"thinktime.mean":                                  {"Mean (seconds), used with type `uniform`."},
		"thinktime.type":                                  {"Type of think time", "`static`: Static think time, defined by `delay`.", "`uniform`: Random think time with uniform distribution, defined by `mean` and `dev`."},
		"tus.chunksize":                                   {"Upload chunk size (in bytes). Defaults to 300 MiB, if omitted or zero."},
		"tus.retries":                                     {"Number of consecutive retries, if a chunk fails to upload. Defaults to 0 (no retries), if omitted. The first retry is issued instantly, the second with a one second back-off period, the third with a two second back-off period, and so on."},
		"tus.timeout":                                     {"Duration after which the upload times out (for example, `1h`, `30s` or `1m10s`). Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, and `h`."},
		"unpublishsheet.mode":                             {"", "`allsheets`: Unpublish all sheets in the app.", "`sheetids`: Only unpublish the sheets specified by the `sheetIds` array."},
		"unpublishsheet.sheetIds":                         {"(optional) Array of sheet IDs for the `sheetids` mode."},
		"unsubscribeobjects.clear":                        {"Remove any previously subscribed objects from the subscription list."},
		"unsubscribeobjects.ids":                          {"List of object IDs to unsubscribe from."},
		"uploaddata.filename":                             {"Name of the local file to send as payload."},
		"uploaddata.replace":                              {"Set to true to replace existing file. If set to false, a warning of existing file will be reported and file will not be replaced."},
		"uploaddata.spaceid":                              {"(optional) Space ID of space where to upload the data. Leave blank to upload to personal space."},
		"uploaddata.title":                                {"(optional) Set custom title of file on upload. Defaults to file name (excluding extension). (supports the use of [session variables](#session_variables))"},
	}

	Config = map[string]common.DocEntry{
		"connectionSettings": {
			Description: "## Connection settings section\n\nThis section of the JSON file contains connection information.\n\nJSON Web Token (JWT), an open standard for creation of access tokens, or WebSocket can be used for authentication. When using JWT, the private key must be available in the path defined by `jwtsettings.keypath`.\n",
			Examples: "### Examples\n\n#### JWT authentication\n\n```json\n\"connectionSettings\": {\n    \"server\": \"myserver.com\",\n    \"mode\": \"jwt\",\n    \"virtualproxy\": \"jwt\",\n    \"security\": true,\n    \"allowuntrusted\": false,\n    \"jwtsettings\": {\n        \"keypath\": \"mock.pem\",\n        \"claims\": \"{\\\"user\\\":\\\"{{.UserName}}\\\",\\\"directory\\\":\\\"{{.Directory}}\\\"}\"\n    }\n}\n```\n\n* `jwtsettings`:\n\nThe strings for `reqheader`, `jwtheader` and `claims` are processed as a GO template where the `User` struct can be used as data:\n```golang\nstruct {\n	UserName  string\n	Password  string\n	Directory string\n	}\n```\nThere is also support for the `time.Now` method using the function `now`.\n\n* `jwtheader`:\n\nThe entries for message authentication code algorithm, `alg`, and token type, `typ`, are added automatically to the header and should not be included.\n    \n**Example:** To add a key ID header, `kid`, add the following string:\n```json\n{\n	\"jwtheader\": \"{\\\"kid\\\":\\\"myKeyId\\\"}\"\n}\n```\n\n* `claims`:\n\n**Example:** For on-premise JWT authentication (with the user and directory set as keys in the QMC), add the following string:\n```json\n{\n	\"claims\": \"{\\\"user\\\": \\\"{{.UserName}}\\\",\\\"directory\\\": \\\"{{.Directory}}\\\"}\"\n}\n```\n**Example:** To add the time at which the JWT was issued, `iat` (\"issued at\"), add the following string:\n```json\n{\n	\"claims\": \"{\\\"iat\\\":{{now.Unix}}\"\n}\n```\n**Example:** To add the expiration time, `exp`, with 5 hours expiration (time.Now uses nanoseconds), add the following string:\n```json\n{\n	\"claims\": \"{\\\"exp\\\":{{(now.Add 18000000000000).Unix}}}\"\n}\n```\n\n#### Static header authentication\n\n```json\nconnectionSettings\": {\n	\"server\": \"myserver.com\",\n	\"mode\": \"ws\",\n	\"security\": true,\n	\"virtualproxy\" : \"header\",\n	\"headers\" : {\n		\"X-Qlik-User-Header\" : \"{{.UserName}}\"\n}\n```\n",
		},
		"loginSettings": {
			Description: "## Login settings section\n\nThis section of the JSON file contains information on the login settings.\n",
			Examples:    "### Examples\n\n#### Prefix login request type\n\n```json\n\"loginSettings\": {\n   \"type\": \"prefix\",\n   \"settings\": {\n       \"directory\": \"anydir\",\n       \"prefix\": \"Nunit\"\n   }\n}\n```\n\n#### Userlist login request type\n\n```json\n  \"loginSettings\": {\n    \"type\": \"userlist\",\n    \"settings\": {\n      \"userList\": [\n        {\n          \"username\": \"sim1@myhost.example\",\n          \"directory\": \"anydir1\",\n          \"password\": \"MyPassword1\"\n        },\n        {\n          \"username\": \"sim2@myhost.example\"\n        }\n      ],\n      \"directory\": \"anydir2\",\n      \"password\": \"MyPassword2\"\n    }\n  }\n```\n",
		},
		"main": {
			Description: "# Setting up load scenarios\n\nA load scenario is defined in a JSON file with a number of sections.\n",
			Examples:    "\n## Example\n\n* [Load scenario example](./examples/configuration_example.json)\n",
		},
		"scenario": {
			Description: "## Scenario section\n\nThis section of the JSON file contains the actions that are performed in the load scenario.\n\n### Structure of an action entry\n\nAll actions follow the same basic structure: \n",
			Examples:    "### Example\n\n```json\n{\n    \"action\": \"actioname\",\n    \"label\": \"custom label for analysis purposes\",\n    \"disabled\": false,\n    \"settings\": {\n        \n    }\n}\n```\n",
		},
		"scheduler": {
			Description: "## Scheduler section\n\nThis section of the JSON file contains scheduler settings for the users in the load scenario.\n",
			Examples:    "### Using `reconnectsettings`\n\nIf `reconnectsettings.reconnect` is enabled, the following is attempted:\n\n1. Re-connect the WebSocket.\n2. Get the currently opened app in the re-attached engine session.\n3. Re-subscribe to the same object as before the disconnection.\n4. If successful, the action during which the re-connect happened is logged as a successful action with `action` and `label` changed to `Reconnect(action)` and `Reconnect(label)`.\n5. Restart the action that was executed when the disconnection occurred (unless it is a `thinktime` action, which will not be restarted).\n6. Log an info row with info type `WebsocketReconnect` and with a semicolon-separated `details` section as follows: \"success=`X`;attempts=`Y`;TimeSpent=`Z`\"\n    * `X`: True/false\n    * `Y`: An integer representing the number of re-connection attempts\n    * `Z`: The time spent re-connecting (ms)\n\n### Example\n\nSimple scheduler settings:\n\n```json\n\"scheduler\": {\n   \"type\": \"simple\",\n   \"settings\": {\n       \"executiontime\": 120,\n       \"iterations\": -1,\n       \"rampupdelay\": 7.0,\n       \"concurrentusers\": 10\n   },\n   \"iterationtimebuffer\" : {\n       \"mode\": \"onerror\",\n       \"duration\" : \"5s\"\n   },\n   \"instance\" : 2\n}\n```\n\nSimple scheduler set to attempt re-connection in case of an unexpected WebSocket disconnection: \n\n```json\n\"scheduler\": {\n   \"type\": \"simple\",\n   \"settings\": {\n       \"executiontime\": 120,\n       \"iterations\": -1,\n       \"rampupdelay\": 7.0,\n       \"concurrentusers\": 10\n   },\n   \"iterationtimebuffer\" : {\n       \"mode\": \"onerror\",\n       \"duration\" : \"5s\"\n   },\n    \"reconnectsettings\" : {\n      \"reconnect\" : true\n    }\n}\n```\n",
		},
		"settings": {
			Description: "## Settings section\n\nThis section of the JSON file contains timeout and logging settings for the load scenario.\n",
			Examples: "### Examples\n\n```json\n\"settings\": {\n	\"timeout\": 300,\n	\"logs\": {\n		\"traffic\": false,\n		\"debug\": false,\n		\"filename\": \"logs/{{.ConfigFile}}-{{timestamp}}.log\"\n	}\n}\n```\n\n```json\n\"settings\": {\n	\"timeout\": 300,\n	\"logs\": {\n		\"filename\": \"logs/scenario.log\"\n	},\n	\"outputs\" : {\n	    \"dir\" : \"./outputs\"\n	}\n}\n```\n",
		},
	}

	Groups = []common.GroupsEntry{
		{
			Name:    "commonActions",
			Title:   "Common actions",
			Actions: []string{"applybookmark", "askhubadvisor", "changesheet", "clearall", "clearfield", "clickactionbutton", "containertab", "createbookmark", "createsheet", "deletebookmark", "deletesheet", "disconnectapp", "disconnectenvironment", "dosave", "duplicatesheet", "iterated", "listboxselect", "openapp", "productversion", "publishbookmark", "publishsheet", "randomaction", "reload", "select", "setscript", "setscriptvar", "setsensevariable", "sheetchanger", "subscribeobjects", "thinktime", "unpublishbookmark", "unpublishsheet", "unsubscribeobjects"},
			DocEntry: common.DocEntry{
				Description: "# Common actions\n\nThese actions are applicable for most types of Qlik Sense deployments.\n\n**Note:** It is recommended to prepend the actions listed here with an `openapp` action as most of them perform operations in an app context (such as making selections or changing sheets).\n",
				Examples:    "",
			},
		},
		{
			Name:    "qseowActions",
			Title:   "Qlik Sense Enterprise on Windows (QSEoW) actions",
			Actions: []string{"deleteodag", "generateodag", "openhub"},
			DocEntry: common.DocEntry{
				Description: "## Qlik Sense Enterprise on Windows (QSEoW) actions\n\nThese actions are only applicable to Qlik Sense Enterprise on Windows (QSEoW) deployments.\n",
				Examples:    "",
			},
		},
	}

	Extra = map[string]common.DocEntry{
		"sessionvariables": {
			Description: "\n## Session variables\n\nThis section describes the session variables that can be used with some of the actions.\n\n<details>\n<summary><a name=\"session_variables\"></a>Session variables</summary>\n\nSome action parameters support session variables. A session variable is defined by putting the variable, prefixed by a dot, within double curly brackets, such as `{{.UserName}}`.\n\nThe following session variables are supported in actions:\n\n* `UserName`: The simulated username. This is not the same as the authenticated user, but rather how the username was defined by [Login settings](#login_settings).  \n* `Session`: The enumeration of the currently simulated session.\n* `Thread`: The enumeration of the currently simulated \"thread\" or \"concurrent user\".\n* `ScriptVars`: A map containing script variables added by the action `setscriptvar`.\n\nThe following variable is supported in the filename of the log file:\n\n* `ConfigFile`: The filename of the config file, without file extension.\n\nThe following functions are supported:\n\n* `now`: Evaluates Golang [time.Now()](https://golang.org/pkg/time/). \n* `hostname`: Hostname of the local machine.\n* `timestamp`: Timestamp in `yyyyMMddhhmmss` format.\n* `uuid`: Generate an uuid.\n* `env`: Retrieve a specific environment variable. Takes one argument - the name of the environment variable to expand.\n* `add`: Adds two integer values together and outputs the sum. E.g. `{{ add 1 2 }}`.\n* `join`: Joins array elements together to a string separated by defined separator. E.g. `{{ join .ScriptVars.MyArray \\\",\\\" }}`.\n\n### Example\n\n```json\n{\n    \"label\" : \"Create bookmark\",\n    \"action\": \"createbookmark\",\n    \"settings\": {\n        \"title\": \"my bookmark {{.Thread}}-{{.Session}} ({{.UserName}})\",\n        \"description\": \"This bookmark contains some interesting selections\"\n    }\n},\n{\n    \"label\" : \"Publish created bookmark\",\n    \"action\": \"publishbookmark\",\n    \"disabled\" : false,\n    \"settings\" : {\n        \"title\": \"my bookmark {{.Thread}}-{{.Session}} ({{.UserName}})\",\n    }\n}\n\n```\n\n```json\n{\n  \"action\": \"createbookmark\",\n  \"settings\": {\n    \"title\": \"{{env \\\"TITLE\\\"}}\",\n    \"description\": \"This bookmark contains some interesting selections\"\n  }\n}\n```\n\n```json\n{\n    \"action\": \"setscriptvar\",\n    \"settings\": {\n        \"name\": \"BookmarkCounter\",\n        \"type\": \"int\",\n        \"value\": \"1\"\n    }\n},\n{\n  \"action\": \"createbookmark\",\n  \"settings\": {\n    \"title\": \"Bookmark no {{ add .ScriptVars.BookmarkCounter 1 }}\",\n    \"description\": \"This bookmark will have the title Bookmark no 2\"\n  }\n}\n```\n\n</details>\n",
			Examples:    "",
		},
	}
)
