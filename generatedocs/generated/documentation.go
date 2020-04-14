package generated

/*
	This file has been generated, do not edit the file directly.

	Generate with go run ./generatedocs/compile/main.go or by running go generate in gopherciser root project.
*/

import "github.com/qlik-oss/gopherciser/generatedocs/common"

var (
    
    Actions = map[string]common.DocEntry{ 
        "applybookmark": {
            Description: "## ApplyBookmark action\n\nApply a bookmark in the current app.\n\n**Note:** Specify *either* `title` *or* `id`, not both.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"applybookmark\",\n    \"settings\": {\n        \"title\": \"My bookmark\"\n    }\n}\n```\n",
        },
        "changesheet": {
            Description: "## ChangeSheet action\n\nChange to a new sheet, unsubscribe to the currently subscribed objects, and subscribe to all objects on the new sheet.\n\nThe action supports getting data from the following objects:\n\n* Listbox\n* Filter pane\n* Bar chart\n* Scatter plot\n* Map (only the first layer)\n* Combo chart\n* Table\n* Pivot table\n* Line chart\n* Pie chart\n* Tree map\n* Text-Image\n* KPI\n* Gauge\n* Box plot\n* Distribution plot\n* Histogram\n* Auto chart (including any support generated visualization from this list)\n* Waterfall chart\n",
            Examples: "### Example\n\n```json\n{\n     \"label\": \"Change Sheet Dashboard\",\n     \"action\": \"ChangeSheet\",\n     \"settings\": {\n         \"id\": \"TFJhh\"\n     }\n}\n```\n",
        },
        "clearall": {
            Description: "## ClearAll action\n\nClear all selections in an app.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"clearall\",\n    \"label\": \"Clear all selections (1)\"\n}\n```\n",
        },
        "createbookmark": {
            Description: "## CreateBookmark action\n\nCreate a bookmark from the current selection and selected sheet.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"createbookmark\",\n    \"settings\": {\n        \"title\": \"my bookmark\",\n        \"description\": \"This bookmark contains some interesting selections\"\n    }\n}\n```\n",
        },
        "createsheet": {
            Description: "## CreateSheet action\n\nCreate a new sheet in the current app.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"createsheet\",\n    \"settings\": {\n        \"title\" : \"Generated sheet\"\n    }\n}\n```\n",
        },
        "deletebookmark": {
            Description: "## DeleteBookmark action\n\nDelete one or more bookmarks in the current app.\n\n**Note:** Specify *either* `title` *or* `id`, not both.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"deletebookmark\",\n    \"settings\": {\n        \"mode\": \"single\",\n        \"title\": \"My bookmark\"\n    }\n}\n```\n",
        },
        "deletedata": {
            Description: "## DeleteData action\n\nDelete a data file from the Data manager.\n",
            Examples: "### Example\n\n```json\n{\n     \"action\": \"DeleteData\",\n     \"settings\": {\n         \"filename\": \"data.csv\",\n         \"path\": \"MyDataFiles\"\n     }\n}\n```\n",
        },
        "deleteodag": {
            Description: "## DeleteOdag action\n\nDelete all user-generated on-demand apps for the current user and the specified On-Demand App Generation (ODAG) link.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"DeleteOdag\",\n    \"settings\": {\n        \"linkname\": \"Drill to Template App\"\n    }\n}\n```\n",
        },
        "deletesheet": {
            Description: "## DeleteSheet action\n\nDelete one or more sheets in the current app.\n\n**Note:** Specify *either* `title` *or* `id`, not both.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"deletesheet\",\n    \"settings\": {\n        \"mode\": \"matching\",\n        \"title\": \"Test sheet\"\n    }\n}\n```\n",
        },
        "disconnectapp": {
            Description: "## DisconnectApp action\n\nDisconnect from an already connected app.\n",
            Examples: "### Example\n\n```json\n{\n    \"label\": \"Disconnect from server\",\n    \"action\" : \"disconnectapp\"\n}\n```\n",
        },
        "duplicatesheet": {
            Description: "## DuplicateSheet action\n\nDuplicate a sheet, including all objects.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"duplicatesheet\",\n    \"label\": \"Duplicate sheet1\",\n    \"settings\":{\n        \"id\" : \"mBshXB\",\n        \"save\": true,\n        \"changesheet\": true\n    }\n}\n```\n",
        },
        "elasticcreateapp": {
            Description: "## ElasticCreateApp action\n\nCreate an app in a QSEoK deployment. The app will be private to the user who creates it.\n",
            Examples: "### Example\n\n```json\n{\n     \"action\": \"ElasticCreateApp\",\n     \"label\": \"Create new app\",\n     \"settings\": {\n         \"title\": \"Created by script\",\n         \"stream\": \"Everyone\",\n         \"groups\": [\"Everyone\", \"cool kids\"]\n     }\n}\n```\n",
        },
        "elasticcreatecollection": {
            Description: "## ElasticCreateCollection action\n\nCreate a collection in a QSEoK deployment.\n",
            Examples: "### Example\n\n```json\n{\n   \"action\": \"ElasticCreateCollection\",\n   \"label\": \"Create collection\",\n   \"settings\": {\n       \"name\": \"Collection {{.Session}}\",\n       \"private\": false\n   }\n}\n```\n",
        },
        "elasticdeleteapp": {
            Description: "## ElasticDeleteApp action\n\nDelete an app from a QSEoK deployment.\n",
            Examples: "### Example\n\n```json\n{\n     \"action\": \"ElasticDeleteApp\",\n     \"label\": \"delete app myapp\",\n     \"settings\": {\n         \"mode\": \"single\",\n         \"appmode\": \"name\",\n         \"app\": \"myapp\"\n     }\n}\n```\n",
        },
        "elasticdeletecollection": {
            Description: "## ElasticDeleteCollection action\n\nDelete a collection in a QSEoK deployment.\n",
            Examples: "### Example\n\n```json\n{\n   \"action\": \"ElasticDeleteCollection\",\n   \"label\": \"Delete collection\",\n   \"settings\": {\n       \"name\": \"MyCollection\",\n       \"deletecontents\": true\n   }\n}\n```\n",
        },
        "elasticdeleteodag": {
            Description: "## ElasticDeleteOdag action\n\nDelete all user-generated on-demand apps for the current user and the specified On-Demand App Generation (ODAG) link.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"ElasticDeleteOdag\",\n    \"settings\": {\n        \"linkname\": \"Drill to Template App\"\n    }\n}\n```\n",
        },
        "elasticduplicateapp": {
            Description: "## ElasticDuplicateApp action\n\nDuplicate an app in a QSEoK deployment.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"ElasticDuplicateApp\",\n    \"settings\": {\n        \"appmode\": \"name\",\n        \"app\": \"myapp\",\n        \"title\": \"duplicated app {{.Session}}\"\n    }\n}\n```\n",
        },
        "elasticexplore": {
            Description: "## ElasticExplore action\n\nExplore the hub for apps and fill the artifact map with apps to be used by other actions in the script (for example, the `openapp` action with `appmode` set to `random` or `round`).\n",
            Examples: "### Examples\n\nThe following example shows how to clear the artifact map and fill it with apps having the tag \"mytag\" from the first page in the hub.\n\n```json\n{\n	\"action\": \"ElasticExplore\",\n	\"label\": \"\",\n	\"settings\": {\n		\"keepcurrent\": false,\n		\"tags\": [\"mytag\"]\n	}\n}\n```\n\nThe following example shows how to clear the artifact map, fill it with all apps from the space \"myspace\" and then add all apps from the space \"circles\".\n\n```json\n{\n	\"action\": \"ElasticExplore\",\n	\"label\": \"\",\n	\"settings\": {\n		\"keepcurrent\": false,\n		\"space\": \"myspace\",\n		\"paging\": true\n	}\n},\n{\n	\"action\": \"ElasticExplore\",\n	\"label\": \"\",\n	\"settings\": {\n		\"keepcurrent\": true,\n		\"space\": \"circles\",\n		\"paging\": true\n	}\n}\n```\n\nThe following example shows how to clear the artifact map and fill it with the apps from the first page of the space \"spaceX\". The apps must have the tag \"tag\" or \"team\" or a tag with id \"15172f9c-4a5f-4ee9-ae35-34c1edd78f8d\", but not be created by the simulated user. In addition, the apps are sorted by the time of modification.\n\n```json\n{\n	\"action\": \"ElasticExplore\",\n	\"label\": \"\",\n	\"settings\": {\n		\"keepcurrent\": false,\n		\"space\": \"spaceX\",\n		\"tags\": [\"tag\", \"team\"],\n		\"tagids\": [\"15172f9c-4a5f-4ee9-ae35-34c1edd78f8d\"],\n		\"owner\": \"others\",\n		\"sorting\": \"updated\",\n		\"paging\": false\n	}\n}\n```\n",
        },
        "elasticexportapp": {
            Description: "## ElasticExportApp action\n\nExport an app and, optionally, save it to file.\n",
            Examples: "### Example\n\n```json\n{\n	\"action\": \"elasticexportapp\",\n	\"label\": \"Export My App\",\n	\"settings\": {\n		\"appmode\": \"name\",\n		\"app\": \"My App\",\n		\"nodata\": false,\n		\"savetofile\": false\n	}\n}\n```\n",
        },
        "elasticgenerateodag": {
            Description: "## ElasticGenerateOdag action\n\nGenerate an on-demand app from an existing On-Demand App Generation (ODAG) link.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"ElasticGenerateOdag\",\n    \"settings\": {\n        \"linkname\": \"Drill to Template App\"\n    }\n}\n```\n",
        },
        "elastichubsearch": {
            Description: "## ElasticHubSearch action\n\nSearch the hub in a QSEoK deployment.\n",
            Examples: "### Example\n\n```json\n{\n	\"action\": \"ElasticHubSearch\",\n	\"settings\": {\n		\"searchfor\": \"apps\",\n		\"querysource\": \"fromfile\",\n		\"queryfile\": \"/MyQueries/Queries.txt\"\n	}\n}\n```\n",
        },
        "elasticmoveapp": {
            Description: "## ElasticMoveApp action\n\nMove an app from its existing space into the specified destination space.\n\n**Note:** Specify *either* `destinationspacename` *or* `destinationspaceid`, not both.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"elasticmoveapp\",\n    \"settings\": {\n        \"app\": \"AppForEveryone\",\n        \"appmode\": \"name\",\n        \"destinationspacename\": \"everyone\"\n    }\n}\n```\n",
        },
        "elasticopenhub": {
            Description: "## ElasticOpenHub action\n\nOpen the hub in a QSEoK deployment.\n",
            Examples: "### Example\n\n```json\n{\n	\"action\": \"ElasticOpenHub\",\n	\"label\": \"Open cloud hub with YourCollection and MyCollection\"\n}\n```\n",
        },
        "elasticpublishapp": {
            Description: "## ElasticPublishApp action\n\nPublish an app to a managed space.\n\n**Note:** Specify *either* `destinationspacename` *or* `destinationspaceid`, not both.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"elasticpublishapp\",\n    \"settings\": {\n        \"app\": \"Sales\",\n        \"appmode\": \"name\",\n        \"destinationspacename\": \"Finance\",\n        \"cleartags\": false\n    }\n}\n```\n",
        },
        "elasticreload": {
            Description: "## ElasticReload action\n\nReload an app by simulating selecting **Reload** in the app context menu in the hub.\n",
            Examples: "### Example\n\n```json\n{\n    \"label\": \"Reload MyApp\",\n    \"action\": \"elasticreload\",\n    \"settings\": {\n        \"appmode\": \"name\",\n        \"app\": \"MyApp\"\n    }\n}\n```\n",
        },
        "elasticshareapp": {
            Description: "## ElasticShareApp action\n\nShare an app with one or more groups.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\" : \"ElasticShareApp\",\n    \"label\": \"Share coolapp with Everyone group\",\n    \"settings\": {\n        \"title\": \"coolapp\",\n        \"groups\": [\"Everyone\"]\n    }\n}\n```\n",
        },
        "elasticuploadapp": {
            Description: "## ElasticUploadApp action\n\nUpload an app to a QSEoK deployment.\n",
            Examples: "### Example\n\n```json\n{\n     \"action\": \"ElasticUploadApp\",\n     \"label\": \"Upload myapp.qvf\",\n     \"settings\": {\n         \"title\": \"coolapp\",\n         \"filename\": \"/home/root/myapp.qvf\",\n         \"stream\": \"Everyone\",\n         \"spaceid\": \"2342798aaefcb23\",\n     }\n}\n```\n",
        },
        "generateodag": {
            Description: "## GenerateOdag action\n\nGenerate an on-demand app from an existing On-Demand App Generation (ODAG) link.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"GenerateOdag\",\n    \"settings\": {\n        \"linkname\": \"Drill to Template App\"\n    }\n}\n```\n",
        },
        "iterated": {
            Description: "## Iterated action\n\nLoop one or more actions.\n\n**Note:** This action does not require an app context (that is, it does not have to be prepended with an `openapp` action).\n",
            Examples: "### Example\n\n```json\n//Visit all sheets twice\n{\n     \"action\": \"iterated\",\n     \"label\": \"\",\n     \"settings\": {\n         \"iterations\" : 2,\n         \"actions\" : [\n            {\n                 \"action\": \"sheetchanger\"\n            },\n            {\n                \"action\": \"thinktime\",\n                \"settings\": {\n                    \"type\": \"static\",\n                    \"delay\": 5\n                }\n            }\n         ]\n     }\n}\n```\n",
        },
        "openapp": {
            Description: "## OpenApp action\n\nOpen an app.\n\n**Note:** If the app name is used to specify which app to open, this action cannot be the first action in the scenario. It must be preceded by an action that can populate the artifact map, such as `openhub`, `elasticopenhub` or `elasticexplore`.\n",
            Examples: "### Examples\n\n```json\n{\n     \"label\": \"OpenApp\",\n     \"action\": \"OpenApp\",\n     \"settings\": {\n         \"appmode\": \"guid\",\n         \"app\": \"7967af99-68b6-464a-86de-81de8937dd56\"\n     }\n}\n```\n```json\n{\n     \"label\": \"OpenApp\",\n     \"action\": \"OpenApp\",\n     \"settings\": {\n         \"appmode\": \"randomguidfromlist\",\n         \"list\": [\"7967af99-68b6-464a-86de-81de8937dd56\", \"ca1a9720-0f42-48e5-baa5-597dd11b6cad\"]\n     }\n}\n```\n",
        },
        "openhub": {
            Description: "## OpenHub action\n\nOpen the hub in a QSEoW environment.\n",
            Examples: "### Example\n\n```json\n{\n     \"action\": \"OpenHub\",\n     \"label\": \"Open the hub\"\n}\n```\n",
        },
        "productversion": {
            Description: "## ProductVersion action\n\nRequest the product version from the server and, optionally, save it to the log. This is a lightweight request that can be used as a keep-alive message in a loop.\n",
            Examples: "### Example\n\n```json\n//Keep-alive loop\n{\n    \"action\": \"iterated\",\n    \"settings\" : {\n        \"iterations\" : 10,\n        \"actions\" : [\n            {\n                \"action\" : \"productversion\"\n            },\n            {\n                \"action\": \"thinktime\",\n                \"settings\": {\n                    \"type\": \"static\",\n                    \"delay\": 30\n                }\n            }\n        ]\n    }\n}\n```\n",
        },
        "publishsheet": {
            Description: "## PublishSheet action\n\nPublish sheets in the current app.\n",
            Examples: "### Example\n```json\n{\n     \"label\": \"PublishSheets\",\n     \"action\": \"publishsheet\",\n     \"settings\": {\n       \"mode\": \"sheetids\",\n       \"sheetIds\": [\"qmGcYS\", \"bKbmgT\"]\n     }\n}\n```\n",
        },
        "randomaction": {
            Description: "## RandomAction action\n\nRandomly select other actions to perform. This meta-action can be used as a starting point for your testing efforts, to simplify script authoring or to add background load.\n\n`randomaction` accepts a list of action types between which to randomize. An execution of `randomaction` executes one or more of the listed actions (as determined by the `iterations` parameter), randomly chosen by a weighted probability. If nothing else is specified, each action has a default random mode that is used. An override is done by specifying one or more parameters of the original action.\n\nEach action executed by `randomaction` is followed by a customizable `thinktime`.\n\n**Note:** The recommended way to use this action is to prepend it with an `openapp` and a `changesheet` action as this ensures that a sheet is always in context.\n",
            Examples: "### Random action defaults\n\nThe following default values are used for the different actions:\n\n* `thinktime`: Mirrors the configuration of `thinktimesettings`\n* `sheetobjectselection`:\n\n```json\n{\n     \"settings\": \n     {\n         \"id\": <UNIFORMLY RANDOMIZED>,\n         \"type\": \"RandomFromAll\",\n         \"min\": 1,\n         \"max\": 2,\n         \"accept\": true\n     }\n}\n```\n\n* `changesheet`:\n\n```json\n{\n     \"settings\": \n     {\n         \"id\": <UNIFORMLY RANDOMIZED>\n     }\n}\n```\n\n* `clearall`:\n\n```json\n{\n     \"settings\": \n     {\n     }\n}\n```\n\n### Examples\n\n#### Generating a background load by executing 5 random actions\n\n```json\n{\n    \"action\": \"RandomAction\",\n    \"settings\": {\n        \"iterations\": 5,\n        \"actions\": [\n            {\n                \"type\": \"thinktime\",\n                \"weight\": 1\n            },\n            {\n                \"type\": \"sheetobjectselection\",\n                \"weight\": 3\n            },\n            {\n                \"type\": \"changesheet\",\n                \"weight\": 5\n            },\n            {\n                \"type\": \"clearall\",\n                \"weight\": 1\n            }\n        ],\n        \"thinktimesettings\": {\n            \"type\": \"uniform\",\n            \"mean\": 10,\n            \"dev\": 5\n        }\n    }\n}\n```\n\n#### Making random selections from excluded values\n\n```json\n{\n    \"action\": \"RandomAction\",\n    \"settings\": {\n        \"iterations\": 1,\n        \"actions\": [\n            {\n                \"type\": \"sheetobjectselection\",\n                \"weight\": 1,\n                \"overrides\": {\n                  \"type\": \"RandomFromExcluded\",\n                  \"min\": 1,\n                  \"max\": 5\n                }\n            }\n        ],\n        \"thinktimesettings\": {\n            \"type\": \"static\",\n            \"delay\": 1\n        }\n    }\n}\n```\n",
        },
        "reload": {
            Description: "## Reload action\n\nReload the current app by simulating selecting **Load data** in the Data load editor. To select an app, preceed this action with an `openapp` action.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"reload\",\n    \"settings\": {\n        \"mode\" : \"default\",\n        \"partial\": false\n    }\n}\n```\n",
        },
        "select": {
            Description: "## Select action\n\nSelect random values in an object.\n\nThe action supports:\n\n* Listbox\n* Bar chart\n* Scatter plot\n* Map (only the first layer)\n* Combo chart\n* Table\n* Line chart\n* Pie chart\n* Tree map\n* Box plot\n* Distribution plot\n* Histogram\n* Auto chart (including any support generated visualization from this list)\n",
            Examples: "### Example\n\n```json\n//Select Listbox RandomFromAll\n{\n     \"label\": \"ListBox Year\",\n     \"action\": \"Select\",\n     \"settings\": {\n         \"id\": \"RZmvzbF\",\n         \"type\": \"RandomFromAll\",\n         \"accept\": true,\n         \"wrap\": false,\n         \"min\": 1,\n         \"max\": 3,\n         \"dim\": 0\n     }\n}\n```\n",
        },
        "setscript": {
            Description: "## SetScript action\n\nSet the load script for the current app. To load the data from the script, use the `reload` action after the `setscript` action.\n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"setscript\",\n    \"settings\": {\n        \"script\" : \"Characters:\\nLoad Chr(RecNo()+Ord('A')-1) as Alpha, RecNo() as Num autogenerate 26;\"\n    }\n}\n```\n",
        },
        "sheetchanger": {
            Description: "## SheetChanger action\n\nCreate and execute a `changesheet` action for each sheet in an app. This can be used to cache the inital state for all objects or, by chaining two subsequent `sheetchanger` actions, to measure how well the calculations in an app utilize the cache.\n",
            Examples: "### Example\n\n```json\n{\n    \"label\" : \"Sheetchanger uncached\",\n    \"action\": \"sheetchanger\"\n},\n{\n    \"label\" : \"Sheetchanger cached\",\n    \"action\": \"sheetchanger\"\n}\n```\n",
        },
        "staticselect": {
            Description: "## StaticSelect action\n\nSelect values statically.\n\nThe action supports:\n\n* HyperCube: Normal hypercube\n* ListObject: Normal listbox\n",
            Examples: "### Examples\n\n#### StaticSelect Barchart\n\n```json\n{ \n\"label\": \"Chart Profit per year\",\n     \"action\": \"StaticSelect\",\n     \"settings\": {\n         \"id\": \"FERdyN\",\n	 \"path\": \"/qHyperCubeDef\",\n         \"type\": \"hypercubecells\",\n         \"accept\": true,\n         \"wrap\": false,\n         \"rows\": [2],\n	 \"cols\": [0]\n     }\n}\n```\n\n#### StaticSelect Listbox\n\n```json\n{		\n\"label\": \"ListBox Territory\",\n     \"action\": \"StaticSelect\",\n     \"settings\": {\n         \"id\": \"qpxmZm\",\n         \"path\": \"/qListObjectDef\",\n         \"type\": \"listobjectvalues\",\n         \"accept\": true,\n         \"wrap\": false,\n         \"rows\": [19,8],\n	 \"cols\": [0]\n     }\n}\n```\n",
        },
        "thinktime": {
            Description: "## ThinkTime action\n\nSimulate user think time.\n\n**Note:** This action does not require an app context (that is, it does not have to be prepended with an `openapp` action).\n",
            Examples: "### Examples\n\n#### ThinkTime uniform\n\nThis simulates a think time of 10 to 15 seconds.\n\n```json\n{\n     \"label\": \"TimerDelay\",\n     \"action\": \"thinktime\",\n     \"settings\": {\n         \"type\": \"uniform\",\n         \"mean\": 12.5,\n         \"dev\": 2.5\n     } \n} \n```\n\n#### ThinkTime constant\n\nThis simulates a think time of 5 seconds.\n\n```json\n{\n     \"label\": \"TimerDelay\",\n     \"action\": \"thinktime\",\n     \"settings\": {\n         \"type\": \"static\",\n         \"delay\": 5\n     }\n}\n```\n",
        },
        "unpublishsheet": {
            Description: "## UnpublishSheet action\n\nUnpublish sheets in the current app.\n",
            Examples: "### Example\n```json\n{\n     \"label\": \"UnpublishSheets\",\n     \"action\": \"unpublishsheet\",\n     \"settings\": {\n       \"mode\": \"allsheets\"        \n     }\n}\n```\n",
        },
        "uploaddata": {
            Description: "## UploadData action\n\nUpload a data file to the Data manager.\n",
            Examples: "### Example\n\n```json\n{\n     \"action\": \"UploadData\",\n     \"settings\": {\n         \"filename\": \"/home/root/data.csv\"\n     }\n}\n```\n",
        },
    }

    Params = map[string][]string{ 
        "applybookmark.id": { "(optional) GUID of the bookmark to apply."  },  
        "applybookmark.title": { "(optional) Name of the bookmark to apply."  },  
        "appselection.app": { "App name or app GUID (supports the use of [session variables](#session_variables)). Used with `appmode` set to `guid` or `name`."  },  
        "appselection.appmode": { "App selection mode","`current`: (default) Use the current app, selected by an app selection in a previous action, or set by the `elasticcreateapp`, `elasticduplicateapp` or `elasticuploadapp` action.","`guid`: Use the app GUID specified by the `app` parameter.","`name`: Use the app name specified by the `app` parameter.","`random`: Select a random app from the artifact map, which is filled by the `elasticopenhub` and/or the `elasticexplore` actions.","`randomnamefromlist`: Select a random app from a list of app names. The `list` parameter should contain a list of app names.","`randomguidfromlist`: Select a random app from a list of app GUIDs. The `list` parameter should contain a list of app GUIDs.","`randomnamefromfile`: Select a random app from a file with app names. The `filename` parameter should contain the path to a file in which each line represents an app name.","`randomguidfromfile`: Select a random app from a file with app GUIDs. The `filename` parameter should contain the path to a file in which each line represents an app GUID.","`round`: Select an app from the artifact map according to the round-robin principle.","`roundnamefromlist`: Select an app from a list of app names according to the round-robin principle. The `list` parameter should contain a list of app names.","`roundguidfromlist`: Select an app from a list of app GUIDs according to the round-robin principle. The `list` parameter should contain a list of app GUIDs.","`roundnamefromfile`: Select an app from a file with app names according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app name.","`roundguidfromfile`: Select an app from a file with app GUIDs according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app GUID."  },  
        "appselection.filename": { "Path to a file in which each line represents an app. Used with `appmode` set to `randomnamefromfile`, `randomguidfromfile`, `roundnamefromfile` or `roundguidfromfile`."  },  
        "appselection.list": { "List of apps. Used with `appmode` set to `randomnamefromlist`, `randomguidfromlist`, `roundnamefromlist` or `roundguidfromlist`."  },  
        "canaddtocollection.groups": { "DEPRECATED"  },  
        "changesheet.id": { "GUID of the sheet to change to."  },  
        "config.connectionSettings.allowuntrusted": { "Allow untrusted (for example, self-signed) certificates (`true` / `false`). Defaults to `false`, if omitted."  },  
        "config.connectionSettings.appext": { "Replace `app` in the connect URL for the `openapp` action. Defaults to `app`, if omitted."  },  
        "config.connectionSettings.headers": { "Headers to use in requests."  },  
        "config.connectionSettings.jwtsettings": { "(JWT only) Settings for the JWT connection."  },  
        "config.connectionSettings.jwtsettings.alg": { "The signing method used for the JWT. Defaults to `RS512`, if omitted.","For keyfiles in RSA format, supports `RS256`, `RS384` or `RS512`.","For keyfiles in EC format, supports `ES256`, `ES384` or `ES512`."  },  
        "config.connectionSettings.jwtsettings.claims": { "JWT claims as an escaped JSON string."  },  
        "config.connectionSettings.jwtsettings.jwtheader": { "JWT headers as an escaped JSON string. Custom headers to be added to the JWT header."  },  
        "config.connectionSettings.jwtsettings.keypath": { "Local path to the JWT key file."  },  
        "config.connectionSettings.mode": { "Authentication mode","`jwt`: JSON Web Token","`ws`: WebSocket"  },  
        "config.connectionSettings.port": { "Set another port than default (`80` for http and `443` for https)."  },  
        "config.connectionSettings.rawurl": { "Define the connect URL manually instead letting the `openapp` action do it. **Note**: The protocol must be `wss://` or `ws://`."  },  
        "config.connectionSettings.security": { "Use TLS (SSL) (`true` / `false`)."  },  
        "config.connectionSettings.server": { "Qlik Sense host."  },  
        "config.connectionSettings.virtualproxy": { "Prefix for the virtual proxy that handles the virtual users."  },  
        "config.connectionSettings.wssettings": { "(WebSocket only) Settings for the WebSocket connection."  },  
        "config.loginSettings": { "This section of the JSON file contains information on the login settings."  },  
        "config.loginSettings.settings": { "","`userList`: List of users for the `userlist` login request type. Directory and password can be specified per user or outside the list of usernames, which means that they are inherited by all users."  },  
        "config.loginSettings.settings.directory": { "Directory to set for the users."  },  
        "config.loginSettings.settings.prefix": { "Prefix to add to the username, so that it will be `prefix_{session}`."  },  
        "config.loginSettings.type": { "Type of login request","`prefix`: Add a prefix (specified by the `prefix` setting below) to the username, so that it will be `prefix_{session}`.","`userlist`: List of users as specified by the `userList` setting below.","`none`: Do not add a prefix to the username, so that it will be `{session}`."  },  
        "config.scenario": { "This section of the JSON file contains the actions that are performed in the load scenario."  },  
        "config.scenario.action": { "Name of the action to execute."  },  
        "config.scenario.disabled": { "(optional) Disable action (`true` / `false`). If set to `true`, the action is not executed."  },  
        "config.scenario.label": { "(optional) Custom string set by the user. This can be used to distinguish the action from other actions of the same type when analyzing the test results."  },  
        "config.scenario.settings": { "Most, but not all, actions have a settings section with action-specific settings."  },  
        "config.scheduler": { "This section of the JSON file contains scheduler settings for the users in the load scenario."  },  
        "config.scheduler.instance": { "Instance number for this instance. Use different instance numbers when running the same script in multiple instances to make sure the randomization is different in each instance. Defaults to 1."  },  
        "config.scheduler.iterationtimebuffer": { ""  },  
        "config.scheduler.iterationtimebuffer.duration": { "Duration of the time buffer (for example, `500ms`, `30s` or `1m10s`). Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, and `h`."  },  
        "config.scheduler.iterationtimebuffer.mode": { "Time buffer mode. Defaults to `nowait`, if omitted.","`nowait`: No time buffer in between the iterations.","`constant`: Add a constant time buffer after each iteration. Defined by `duration`.","`onerror`: Add a time buffer in case of an error. Defined by `duration`.","`minduration`: Add a time buffer if the iteration duration is less than `duration`."  },  
        "config.scheduler.settings": { ""  },  
        "config.scheduler.settings.concurrentusers": { "Number of concurrent users to simulate. Allowed values are positive integers."  },  
        "config.scheduler.settings.executiontime": { "Test execution time (seconds). The sessions are disconnected when the specified time has elapsed. Allowed values are positive integers. `-1` means an infinite execution time."  },  
        "config.scheduler.settings.iterations": { "Number of iterations for each 'concurrent' user to repeat. Allowed values are positive integers. `-1` means an infinite number of iterations."  },  
        "config.scheduler.settings.onlyinstanceseed": { "Disable session part of randomization seed. Defaults to `false`, if omitted.","`true`: All users and sessions have the same randomization sequence, which only changes if the `instance` flag is changed.","`false`: Normal randomization sequence, dependent on both the `instance` parameter and the current user session."  },  
        "config.scheduler.settings.rampupdelay": { "Time delay (seconds) scheduled in between each concurrent user during the startup period."  },  
        "config.scheduler.settings.reuseusers": { "","`true`: Every iteration for each concurrent user uses the same user and session.","`false`: Every iteration for each concurrent user uses a new user and session. The total number of users is the product of `concurrentusers` and `iterations`."  },  
        "config.scheduler.type": { "Type of scheduler","`simple`: Standard scheduler"  },  
        "config.settings": { "This section of the JSON file contains timeout and logging settings for the load scenario"  },  
        "config.settings.logs": { "Log settings"  },  
        "config.settings.logs.debug": { "Log debug information (`true` / `false`). Defaults to `false`, if omitted."  },  
        "config.settings.logs.filename": { "Name of the log file (supports the use of [variables](#session_variables))."  },  
        "config.settings.logs.format": { "Log format. Defaults to `tsvfile`, if omitted.","`tsvfile`: Log to file in TSV format and output status to console.","`tsvconsole`: Log to console in TSV format without any status output.","`jsonfile`: Log to file in JSON format and output status to console.","`jsonconsole`: Log to console in JSON format without any status output.","`console`: Log to console in color format without any status output.","`combined`: Log to file in TSV format and to console in JSON format.","`no`: Default logs and status output turned off.","`onlystatus`: Default logs turned off, but status output turned on."  },  
        "config.settings.logs.metrics": { "Log traffic metrics (`true` / `false`). Defaults to `false`, if omitted. **Note:** This should only be used for debugging purposes as traffic logging is resource-demanding."  },  
        "config.settings.logs.summary": { "Type of summary to display after the test run. Defaults to simple for minimal performance impact.","`0` or `undefined`: Simple, single-row summary","`1` or `none`: No summary","`2` or `simple`: Simple, single-row summary","`3` or `extended`: Extended summary that includes statistics on each unique combination of action, label and app GUID","`4` or `full`: Same as extended, but with statistics on each unique combination of method and endpoint added"  },  
        "config.settings.logs.traffic": { "Log traffic information (`true` / `false`). Defaults to `false`, if omitted. **Note:** This should only be used for debugging purposes as traffic logging is resource-demanding."  },  
        "config.settings.outputs": { "Used by some actions to save results to a file."  },  
        "config.settings.outputs.dir": { "Directory in which to save artifacts generated by the script (except log file)."  },  
        "config.settings.timeout": { "Timeout setting (seconds) for WebSocket requests."  },  
        "createbookmark.description": { "(optional) Description of the bookmark to create."  },  
        "createbookmark.id": { "(optional) ID to use with subsequent `applybookmark` or `deletebookmark` actions. **Note:** This ID is only used within the scenario."  },  
        "createbookmark.nosheet": { "Do not include sheet with bookmark."  },  
        "createbookmark.savelayout": { "Include layout with bookmark."  },  
        "createbookmark.title": { "Name of the bookmark to create."  },  
        "createsheet.description": { "(optional) Description of the sheet to create."  },  
        "createsheet.id": { "(optional) ID to be used to identify the sheet in any subsequent `changesheet`, `duplicatesheet`, `publishsheet` or `unpublishsheet` action."  },  
        "createsheet.title": { "Name of the sheet to create."  },  
        "deletebookmark.id": { "(optional) GUID of the bookmark to delete."  },  
        "deletebookmark.mode": { "","`single`: Delete one bookmark that matches the specified `title` or `id` in the current app.","`matching`: Delete all bookmarks with the specified `title` in the current app.","`all`: Delete all bookmarks in the current app."  },  
        "deletebookmark.title": { "(optional) Name of the bookmark to delete."  },  
        "deletedata.filename": { "Name of the file to delete."  },  
        "deletedata.path": { "(optional) Path in which to look for the file. Defaults to `MyDataFiles`, if omitted."  },  
        "deleteodag.linkname": { "Name of the ODAG link from which to delete generated apps. The name is displayed in the ODAG navigation bar at the bottom of the *selection app*."  },  
        "deletesheet.id": { "(optional) GUID of the sheet to delete."  },  
        "deletesheet.mode": { "","`single`: Delete one sheet that matches the specified `title` or `id` in the current app.","`matching`: Delete all sheets with the specified `title` in the current app.","`allunpublished`: Delete all unpublished sheets in the current app."  },  
        "deletesheet.title": { "(optional) Name of the sheet to delete."  },  
        "destinationspace.destinationspaceid": { "Specify destination space by ID."  },  
        "destinationspace.destinationspacename": { "Specify destination space by name."  },  
        "duplicatesheet.changesheet": { "Clear the objects currently subscribed to and then subribe to all objects on the cloned sheet (which essentially corresponds to using the `changesheet` action to go to the cloned sheet) (`true` / `false`). Defaults to `false`, if omitted."  },  
        "duplicatesheet.cloneid": { "(optional) ID to be used to identify the sheet in any subsequent `changesheet`, `duplicatesheet`, `publishsheet` or `unpublishsheet` action."  },  
        "duplicatesheet.id": { "ID of the sheet to clone."  },  
        "duplicatesheet.save": { "Execute `saveobjects` after the cloning operation to save all modified objects (`true` / `false`). Defaults to `false`, if omitted."  },  
        "elasticcreatecollection.description": { "(optional) Description of the collection to create."  },  
        "elasticcreatecollection.name": { "Name of the collection to create (supports the use of [session variables](#session_variables))."  },  
        "elasticcreatecollection.private": { "","`true`: Private collection","`false`: Public collection"  },  
        "elasticdeleteapp.collectionname": { "Name of the collection in which to delete apps."  },  
        "elasticdeleteapp.mode": { "","`single`: Delete the app specified explicitly by app GUID or app name.","`everything`: Delete all apps currently in the application context, as determined by the `elasticopenhub` action. **Note:** Use with care.","`clearcollection`: Delete all apps in the collection specified by `collectionname`."  },  
        "elasticdeletecollection.deletecontents": { "","`true`: Delete all apps in the collection before deleting the collection.","`false`: Delete the collection without doing anything to the apps in the collection."  },  
        "elasticdeletecollection.name": { "Name of the collection to delete."  },  
        "elasticdeleteodag.linkname": { "Name of the ODAG link from which to delete generated apps. The name is displayed in the ODAG navigation bar at the bottom of the *selection app*."  },  
        "elasticduplicateapp.spaceid": { "(optional) GUID of the shared space in which to publish the app."  },  
        "elasticexplore.keepcurrent": { "Keep the current artifact map and add the results from the `elasticexplore` action. Defaults to `false` (that is, empty the artifact map before adding the results from the `elasticexplore` action), if omitted."  },  
        "elasticexplore.owner": { "Filter apps by owner","`all`: Apps owned by anyone.","`me`: Apps owned by the simulated user.","`others`: Apps not owned by the simulated user."  },  
        "elasticexplore.paging": { "Go through all app pages in the hub. Defaults to `false` (that is, only include the first 30 apps that the user can see), if omitted."  },  
        "elasticexplore.sorting": { "Simulate selecting sort order in the drop-down menu in the hub","`default`: Default sort order (`created`).","`created`: Sort by the time of creation.","`updated`: Sort by the time of modification.","`name`: Sort by name."  },  
        "elasticexplore.space": { "Filter apps by space name (supports the use of [session variables](#session_variables)). **Note:** This filter cannot be used together with `spaceid`."  },  
        "elasticexplore.spaceid": { "Filter apps by space GUID. **Note:** This filter cannot be used together with `space`."  },  
        "elasticexplore.tagids": { "Filter apps by tag ids. This filter can be used together with `tags`."  },  
        "elasticexplore.tags": { "Filter apps by tag names. This filter can be used together with `tagids`."  },  
        "elasticexportapp.filename": { "Pattern for the filename when saving the exported app to a file, defaults to app title or app GUID. Supports the use of [session variables](#session_variables) and additionally `.Local.Title` can be used as a variable to add the title of the exported app."  },  
        "elasticexportapp.nodata": { "Export the app without data (`true`/`false`). Defaults to `false` (that is, export with data), if omitted."  },  
        "elasticexportapp.savetofile": { "Save the exported file in the specified directory (`true`/`false`). Defaults to `false`, if omitted."  },  
        "elasticgenerateodag.linkname": { "Name of the ODAG link from which to generate an app. The name is displayed in the ODAG navigation bar at the bottom of the *selection app*."  },  
        "elastichubsearch.query": { "(optional) Query string (in case of `querystring` as source)."  },  
        "elastichubsearch.queryfile": { "(optional) File from which to read a query (in case of `fromfile` as source)."  },  
        "elastichubsearch.querysource": { "","`querystring`: The query is provided as a string specified by `query`.","`fromfile`: The queries are read from the file specified by `queryfile`, where each line represents a query."  },  
        "elastichubsearch.searchfor": { "","`collections`: Search for collections only.","`apps`: Search for apps only.","`both`: Search for both collections and apps."  },  
        "elasticpublishapp.cleartags": { "Publish the app without its original tags."  },  
        "elasticreload.pollinterval": { "Reload status polling interval (seconds). Defaults to 5 seconds, if omitted."  },  
        "elasticshareapp.appguid": { "GUID of the app to share."  },  
        "elasticshareapp.groups": { "List of groups that should be given access to the app."  },  
        "elasticshareapp.title": { "Name of the app to share (supports the use of [session variables](#session_variables)). If `appguid` and `title` refer to different apps, `appguid` takes precedence."  },  
        "elasticuploadapp.chunksize": { "(optional) Upload chunk size (in bytes). Defaults to 300 MiB, if omitted or zero."  },  
        "elasticuploadapp.filename": { "Local file to send as payload."  },  
        "elasticuploadapp.mode": { "Upload mode. Defaults to `tus`, if omitted.","`tus`: Upload the file using the [tus](https://tus.io/) chunked upload protocol.","`legacy`: Upload the file using a single POST payload (legacy file upload mode)."  },  
        "elasticuploadapp.retries": { "(optional) Number of consecutive retries, if a chunk fails to upload. Defaults to 0 (no retries), if omitted. The first retry is issued instantly, the second with a one second back-off period, the third with a two second back-off period, and so on."  },  
        "elasticuploadapp.spaceid": { "DEPRECATED"  },  
        "elasticuploadapp.stream": { "(optional) Name of the private collection or public tag under which to publish the app (supports the use of [session variables](#session_variables))."  },  
        "elasticuploadapp.streamguid": { "(optional) GUID of the private collection or public tag under which to publish the app."  },  
        "elasticuploadapp.title": { "Name of the app to upload (supports the use of [session variables](#session_variables))."  },  
        "generateodag.linkname": { "Name of the ODAG link from which to generate an app. The name is displayed in the ODAG navigation bar at the bottom of the *selection app*."  },  
        "iterated.actions": { "Actions to iterate"  },  
        "iterated.iterations": { "Number of loops."  },  
        "productversion.log": { "Save the product version to the log (`true` / `false`). Defaults to `false`, if omitted."  },  
        "publishsheet.mode": { "","`allsheets`: Publish all sheets in the app.","`sheetids`: Only publish the sheets specified by the `sheetIds` array."  },  
        "publishsheet.sheetIds": { "(optional) Array of sheet IDs for the `sheetids` mode."  },  
        "randomaction.actions": { "List of actions from which to randomly pick an action to execute. Each item has a number of possible parameters."  },  
        "randomaction.actions.overrides": { "(optional) Static overrides to the action. The overrides can include any or all of the settings from the original action, as determined by the `type` field. If nothing is specified, the default values are used."  },  
        "randomaction.actions.type": { "Type of action","`thinktime`: See the `thinktime` action.","`sheetobjectselection`: Make random selections within objects visible on the current sheet. See the `select` action.","`changesheet`: See the `changesheet` action.","`clearall`: See the `clearall` action."  },  
        "randomaction.actions.weight": { "The probabilistic weight of the action, specified as an integer. This number is proportional to the likelihood of the specified action, and is used as a weight in a uniform random selection."  },  
        "randomaction.iterations": { "Number of random actions to perform."  },  
        "randomaction.thinktimesettings": { "Settings for the `thinktime` action, which is automatically inserted after every randomized action."  },  
        "reload.log": { "Save the reload log as a field in the output (`true` / `false`). Defaults to `false`, if omitted. **Note:** This should only be used when needed as the reload log can become very large."  },  
        "reload.mode": { "Error handling during the reload operation","`default`: Use the default error handling.","`abend`: Stop reloading the script, if an error occurs.","`ignore`: Continue reloading the script even if an error is detected in the script."  },  
        "reload.partial": { "Enable partial reload (`true` / `false`). This allows you to add data to an app without reloading all data. Defaults to `false`, if omitted."  },  
        "select.accept": { "Accept or abort selection after selection (only used with `wrap`) (`true` / `false`)."  },  
        "select.dim": { "Dimension / column in which to select."  },  
        "select.id": { "ID of the object in which to select values."  },  
        "select.max": { "Maximum number of selections to make."  },  
        "select.min": { "Minimum number of selections to make."  },  
        "select.type": { "Selection type","`randomfromall`: Randomly select within all values of the symbol table.","`randomfromenabled`: Randomly select within the white and light grey values on the first data page.","`randomfromexcluded`: Randomly select within the dark grey values on the first data page.","`randomdeselect`: Randomly deselect values on the first data page."  },  
        "select.wrap": { "Wrap selection with Begin / End selection requests (`true` / `false`)."  },  
        "setscript.script": { "Load script for the app (written as a string)."  },  
        "staticselect.accept": { "Accept or abort selection after selection (only used with `wrap`) (`true` / `false`)."  },  
        "staticselect.cols": { "Dimension / column in which to select."  },  
        "staticselect.id": { "ID of the object in which to select values."  },  
        "staticselect.path": { "Path to the hypercube or listobject (differs depending on object type)."  },  
        "staticselect.rows": { "Element values to select in the dimension / column."  },  
        "staticselect.type": { "Selection type","`hypercubecells`: Select in hypercube.","`listobjectvalues`: Select in listbox."  },  
        "staticselect.wrap": { "Wrap selection with Begin / End selection requests (`true` / `false`)."  },  
        "thinktime.delay": { "Delay (seconds), used with type `static`."  },  
        "thinktime.dev": { "Deviation (seconds) from `mean` value, used with type `uniform`."  },  
        "thinktime.mean": { "Mean (seconds), used with type `uniform`."  },  
        "thinktime.type": { "Type of think time","`static`: Static think time, defined by `delay`.","`uniform`: Random think time with uniform distribution, defined by `mean` and `dev`."  },  
        "unpublishsheet.mode": { "","`allsheets`: Unpublish all sheets in the app.","`sheetids`: Only unpublish the sheets specified by the `sheetIds` array."  },  
        "unpublishsheet.sheetIds": { "(optional) Array of sheet IDs for the `sheetids` mode."  },  
        "uploaddata.destinationpath": { "(optional) Path to which to upload the file. Defaults to `MyDataFiles`, if omitted."  },  
        "uploaddata.filename": { "Name of the local file to send as payload."  },  
    }
    
    Config = map[string]common.DocEntry{ 
        "connectionSettings" : {
            Description: "## Connection settings section\n\nThis section of the JSON file contains connection information.\n\nJSON Web Token (JWT), an open standard for creation of access tokens, or WebSocket can be used for authentication. When using JWT, the private key must be available in the path defined by `jwtsettings.keypath`.\n",
            Examples: "### Examples\n\n#### JWT authentication\n\n```json\n\"connectionSettings\": {\n    \"server\": \"myserver.com\",\n    \"mode\": \"jwt\",\n    \"virtualproxy\": \"jwt\",\n    \"security\": true,\n    \"allowuntrusted\": false,\n    \"jwtsettings\": {\n        \"keypath\": \"mock.pem\",\n        \"claims\": \"{\\\"user\\\":\\\"{{.UserName}}\\\",\\\"directory\\\":\\\"{{.Directory}}\\\"}\"\n    }\n}\n```\n\n* `jwtsettings`:\n\nThe strings for `reqheader`, `jwtheader` and `claims` are processed as a GO template where the `User` struct can be used as data:\n```golang\nstruct {\n	UserName  string\n	Password  string\n	Directory string\n	}\n```\nThere is also support for the `time.Now` method using the function `now`.\n\n* `jwtheader`:\n\nThe entries for message authentication code algorithm, `alg`, and token type, `typ`, are added automatically to the header and should not be included.\n    \n**Example:** To add a key ID header, `kid`, add the following string:\n```json\n{\n	\"jwtheader\": \"{\\\"kid\\\":\\\"myKeyId\\\"}\"\n}\n```\n\n* `claims`:\n\n**Example:** For on-premise JWT authentication (with the user and directory set as keys in the QMC), add the following string:\n```json\n{\n	\"claims\": \"{\\\"user\\\": \\\"{{.UserName}}\\\",\\\"directory\\\": \\\"{{.Directory}}\\\"}\"\n}\n```\n**Example:** To add the time at which the JWT was issued, `iat` (\"issued at\"), add the following string:\n```json\n{\n	\"claims\": \"{\\\"iat\\\":{{now.Unix}}\"\n}\n```\n**Example:** To add the expiration time, `exp`, with 5 hours expiration (time.Now uses nanoseconds), add the following string:\n```json\n{\n	\"claims\": \"{\\\"exp\\\":{{(now.Add 18000000000000).Unix}}}\"\n}\n```\n\n#### Static header authentication\n\n```json\nconnectionSettings\": {\n	\"server\": \"myserver.com\",\n	\"mode\": \"ws\",\n	\"security\": true,\n	\"virtualproxy\" : \"header\",\n	\"headers\" : {\n		\"X-Qlik-User-Header\" : \"{{.UserName}}\"\n}\n```\n",
        },
        "loginSettings" : {
            Description: "## Login settings section\n\nThis section of the JSON file contains information on the login settings.\n",
            Examples: "### Examples\n\n#### Prefix login request type\n\n```json\n\"loginSettings\": {\n   \"type\": \"prefix\",\n   \"settings\": {\n       \"directory\": \"anydir\",\n       \"prefix\": \"Nunit\"\n   }\n}\n```\n\n#### Userlist login request type\n\n```json\n  \"loginSettings\": {\n    \"type\": \"userlist\",\n    \"settings\": {\n      \"userList\": [\n        {\n          \"username\": \"sim1@myhost.example\",\n          \"directory\": \"anydir1\",\n          \"password\": \"MyPassword1\"\n        },\n        {\n          \"username\": \"sim2@myhost.example\"\n        }\n      ],\n      \"directory\": \"anydir2\",\n      \"password\": \"MyPassword2\"\n    }\n  }\n```\n",
        },
        "scenario" : {
            Description: "## Scenario section\n\nThis section of the JSON file contains the actions that are performed in the load scenario.\n\n### Structure of an action entry\n\nAll actions follow the same basic structure: \n",
            Examples: "### Example\n\n```json\n{\n    \"action\": \"actioname\",\n    \"label\": \"custom label for analysis purposes\",\n    \"disabled\": false,\n    \"settings\": {\n        \n    }\n}\n```\n",
        },
        "scheduler" : {
            Description: "## Scheduler section\n\nThis section of the JSON file contains scheduler settings for the users in the load scenario.\n",
            Examples: "### Example\n\n```json\n\"scheduler\": {\n   \"type\": \"simple\",\n   \"settings\": {\n       \"executiontime\": 120,\n       \"iterations\": -1,\n       \"rampupdelay\": 7.0,\n       \"concurrentusers\": 10\n   },\n   \"iterationtimebuffer\" : {\n       \"mode\": \"onerror\",\n       \"duration\" : \"5s\"\n   },\n   \"instance\" : 2\n}\n```\n",
        },
        "settings" : {
            Description: "## Settings section\n\nThis section of the JSON file contains timeout and logging settings for the load scenario.\n",
            Examples: "### Examples\n\n```json\n\"settings\": {\n	\"timeout\": 300,\n	\"logs\": {\n		\"traffic\": false,\n		\"debug\": false,\n		\"filename\": \"logs/{{.ConfigFile}}-{{timestamp}}.log\"\n	}\n}\n```\n\n```json\n\"settings\": {\n	\"timeout\": 300,\n	\"logs\": {\n		\"filename\": \"logs/scenario.log\"\n	},\n	\"outputs\" : {\n	    \"dir\" : \"./outputs\"\n	}\n}\n```\n",
        },
        "main" : {
            Description: "# Setting up load scenarios\n\nA load scenario is defined in a JSON file with a number of sections.\n",
            Examples: "\n## Example\n\n* [Load scenario example](./examples/configuration_example.json)\n",
        },
    }

    Groups = []common.GroupsEntry{ 
            {
                Name: "commonActions",
                Title: "Common actions",
                Actions: []string{ "applybookmark","changesheet","clearall","createbookmark","createsheet","deletebookmark","deletesheet","disconnectapp","duplicatesheet","iterated","openapp","productversion","publishsheet","randomaction","reload","select","setscript","sheetchanger","staticselect","thinktime","unpublishsheet" },
                DocEntry: common.DocEntry{
                    Description: "# Common actions\n\nThese actions are applicable to both Qlik Sense Enterprise for Windows (QSEfW) and Qlik Sense Enterprise on Kubernetes (QSEoK) deployments.\n\n**Note:** It is recommended to prepend the actions listed here with an `openapp` action as most of them perform operations in an app context (such as making selections or changing sheets).\n",
                    Examples: "",
                },
            },
            {
                Name: "qseowActions",
                Title: "Qlik Sense Enterprise on Windows (QSEoW) actions",
                Actions: []string{ "deleteodag","generateodag","openhub" },
                DocEntry: common.DocEntry{
                    Description: "## Qlik Sense Enterprise on Windows (QSEoW) actions\n\nThese actions are only applicable to Qlik Sense Enterprise on Windows (QSEoW) deployments.\n",
                    Examples: "",
                },
            },
            {
                Name: "qseokActions",
                Title: "Qlik Sense Enterprise on Kubernetes (QSEoK) / Elastic actions",
                Actions: []string{ "deletedata","elasticcreateapp","elasticcreatecollection","elasticdeleteapp","elasticdeletecollection","elasticdeleteodag","elasticduplicateapp","elasticexplore","elasticexportapp","elasticgenerateodag","elastichubsearch","elasticmoveapp","elasticopenhub","elasticpublishapp","elasticreload","elasticshareapp","elasticuploadapp","uploaddata" },
                DocEntry: common.DocEntry{
                    Description: "## Qlik Sense Enterprise on Kubernetes (QSEoK) / Elastic actions\n\nThese actions are only applicable to Qlik Sense Enterprise on Kubernetes (QSEoK) deployments.\n",
                    Examples: "",
                },
            },
    }

    Extra = map[string]common.DocEntry{ 
        "sessionvariables": {
            Description: "\n## Session variables\n\nThis section describes the session variables that can be used with some of the actions.\n\n<details>\n<summary><a name=\"session_variables\"></a>Session variables</summary>\n\nSome action parameters support session variables. A session variable is defined by putting the variable, prefixed by a dot, within double curly brackets, such as `{{.UserName}}`.\n\nThe following session variables are supported in actions:\n\n* `UserName`: The simulated username. This is not the same as the authenticated user, but rather how the username was defined by [Login settings](#login_settings).  \n* `Session`: The enumeration of the currently simulated session.\n* `Thread`: The enumeration of the currently simulated \"thread\" or \"concurrent user\".\n\nThe following variable is supported in the filename of the log file:\n\n* `ConfigFile`: The filename of the config file, without file extension.\n\nThe following functions are supported:\n\n* `now`: Evaluates Golang [time.Now()](https://golang.org/pkg/time/). \n* `hostname`: Hostname of the local machine.\n* `timestamp`: Timestamp in `yyyyMMddhhmmss` format.\n* `uuid`: Generate an uuid.\n\n### Example\n```json\n{\n    \"action\": \"ElasticCreateApp\",\n    \"label\": \"Create new app\",\n    \"settings\": {\n        \"title\": \"CreateApp {{.Thread}}-{{.Session}} ({{.UserName}})\",\n        \"stream\": \"mystream\",\n        \"groups\": [\n            \"mygroup\"\n        ]\n    }\n},\n{\n    \"label\": \"OpenApp\",\n    \"action\": \"OpenApp\",\n    \"settings\": {\n        \"appname\": \"CreateApp {{.Thread}}-{{.Session}} ({{.UserName}})\"\n    }\n},\n{\n    \"action\": \"elasticexportapp\",\n    \"label\": \"Export app\",\n    \"settings\": {\n        \"appmode\" : \"name\",\n        \"app\" : \"CreateApp {{.Thread}}-{{.Session}} ({{.UserName}})\",\n        \"savetofile\": true,\n        \"exportname\": \"Exported app {{.Thread}}-{{.Session}} {{now.UTC}}\"\n    }\n}\n\n```\n</details>\n",
            Examples: "",
        },
    }
)
