## AskHubAdvisor action

Perform a query in the Qlik Sense hub insight advisor.
* `querysource`: The source from which queries will be randomly picked.
    * `file`: Read queries from file defined by `file`.
    * `querylist`: Read queries from list defined by `querylist`.
* `querylist`: A list of queries. Plain strings are supported and will get a weight of `1`.
  * `weight`: A weight to set probablility of query being peformed.
  * `query`: A query sentence.
* `lang`: Query language.
* `maxfollowup`: The maximum depth of followup queries asked. A value of `0` means that a query from querysource is performed without followup queries.
* `file`: Path to query file.
* `app`: Optional name of app to pick in followup queries. If not set, a random app is picked.
* `saveimages`: Save images of charts to file.
* `saveimagefile`: File name of saved images. Defaults to server side file name. Supports [Session Variables](https://github.com/qlik-trial/gopherciser-oss/blob/master/docs/settingup.md#session-variables).
* `thinktime`: Settings for the `thinktime` action, which is automatically inserted before each followup. Defaults to a uniform distribution with mean=8 and deviation=4.
  * `type`: Type of think time
      * `static`: Static think time, defined by `delay`.
      * `uniform`: Random think time with uniform distribution, defined by `mean` and `dev`.
  * `delay`: Delay (seconds), used with type `static`.
  * `mean`: Mean (seconds), used with type `uniform`.
  * `dev`: Deviation (seconds) from `mean` value, used with type `uniform`.
* `followuptypes`: A list of followup types enabled for followup queries. If omitted, all types are enabled.
    * `app`: Enable followup queries which change app.
    * `measure`: Enable followups based on measures.
    * `dimension`: Enable followups based on dimensions.
    * `recommendation`: Enable followups based on recommendations.
    * `sentence`: Enable followup queries based on bare sentences.

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

#### Ask followups only based on app selection


```json
{
    "action": "AskHubAdvisor",
    "settings": {
        "querysource": "querylist",
        "querylist": [
            "what is the lowest price of shoes"
        ],
        "maxfollowup": 5,
        "followuptypes": ["app"]
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

