### Examples

Open app using specific `guid`

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

Open app using random `guid` from a list

```json
{
     "label": "OpenApp",
     "action": "OpenApp",
     "settings": {
         "appmode": "randomguidfromlist",
         "list": [
            "7967af99-68b6-464a-86de-81de8937dd56", "ca1a9720-0f42-48e5-baa5-597dd11b6cad"
         ]
     }
}
```

Open apps round robin from list using app name (requires preceeding action which has filled the artifact map with app names)

```json
{
     "label": "OpenApp",
     "action": "OpenApp",
     "settings": {
         "appmode": "roundnamefromlist",
         "list": [
            "MyApp1",
            "MyApp2"
         ]
     }
}
```

Open app with custom timeouts for connecting to engine and opening app in engine. This can be used for particularly slow apps to open.

```json
{
     "label": "OpenApp",
     "action": "OpenApp",
     "settings": {
         "appmode": "guid",
         "app": "7967af99-68b6-464a-86de-81de8937dd56",
         "timeouts" : {
            "connect": "6m",
            "open" : "7m30s"
         }
     }
}
```

Open app with no data

```json
{
     "label": "OpenApp",
     "action": "OpenApp",
     "settings": {
         "appmode": "guid",
         "app": "7967af99-68b6-464a-86de-81de8937dd56",
         "nodata": true
     }
}
```