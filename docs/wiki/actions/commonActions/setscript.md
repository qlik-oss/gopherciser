## SetScript action

Set the load script for the current app. To load the data from the script, use the `reload` action after the `setscript` action.

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

