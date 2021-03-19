### Examples

#### Pick queries from file

```json
{
    "action": "AskHubAdvisor",
    "settings": {
        "querysource": "file",
        "file": "queries.txt"
    }
}
```

The file `queries.txt` contains one query and an optional weight per line. The line format is `[WEIGHT;]QUERY`.
```txt
show sales per country
5; what is the lowest price of shoes
```

#### Pick queries from list

```json
{
    "action": "AskHubAdvisor",
    "settings": {
        "querysource": "querylist",
        "querylist": ["show sales per country", "what is the lowest price of shoes"]
    }
}
```

#### Perform followup queries if possible (default: 0)

```json
{
    "action": "AskHubAdvisor",
    "settings": {
        "querysource": "querylist",
        "querylist": ["show sales per country", "what is the lowest price of shoes"],
        "maxfollowup": 3
    }
}
```

#### Change lanuage (default: "en")

```json
{
    "action": "AskHubAdvisor",
    "settings": {
        "querysource": "querylist",
        "querylist": ["show sales per country", "what is the lowest price of shoes"],
        "lang": "fr"
    }
}
```

#### Weights in querylist

```json
{
    "action": "AskHubAdvisor",
    "settings": {
        "querysource": "querylist",
        "querylist": [
            {
                "query": "show sales per country",
                "weight": 5,
            },
            "what is the lowest price of shoes"
        ]
    }
}
```

#### Thinktime before followup queries

See detailed examples of settings in the documentation for thinktime action.

```json
{
    "action": "AskHubAdvisor",
    "settings": {
        "querysource": "querylist",
        "querylist": [
            "what is the lowest price of shoes"
        ],
        "maxfollowup": 5,
        "thinktime": {
            "type": "static",
            "delay": 5
        }
    }
}
```

#### Save chart images to file

```json
{
    "action": "AskHubAdvisor",
    "settings": {
        "querysource": "querylist",
        "querylist": [
            "show price per shoe type"
        ],
        "maxfollowup": 5,
        "saveimages": true
    }
}
```

#### Save chart images to file with custom name

The `saveimagefile` file name template setting supports
[Session Variables](https://github.com/qlik-trial/gopherciser-oss/blob/master/docs/settingup.md#session-variables).
You can apart from session variables include the following action local variables in the `saveimagefile` file name template:
- .Local.ImageCount - _the number of images written to file_
- .Local.ServerFileName - _the server side name of image file_
- .Local.Query - _the query sentence_
- .Local.AppName - _the name of app, if any app, where query is asked_
- .Local.AppID - _the id of app, if any app, where query is asked_

```json
{
    "action": "AskHubAdvisor",
    "settings": {
        "querysource": "querylist",
        "querylist": [
            "show price per shoe type"
        ],
        "maxfollowup": 5,
        "saveimages": true,
        "saveimagefile": "{{.Local.Query}}--app-{{.Local.AppName}}--user-{{.UserName}}--thread-{{.Thread}}--session-{{.Session}}"
    }
}
```
