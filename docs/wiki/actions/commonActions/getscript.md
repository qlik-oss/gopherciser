## GetScript action

Get the load script for the app.


* `savelog`: Save load script to log file under the INFO log labelled *LoadScript*

### Example

Get the load script for the app

```json
{
    "action": "getscript"
}
```

Get the load script for the app and save to log file

```json
{
    "action": "getscript",
    "settings": {
        "savelog" : true
    }
}
```

