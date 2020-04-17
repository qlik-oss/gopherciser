### Example

Publish bookmark previously created in the script with an id `bookmark1`.

```json
{
    "label" : "Publish bookmark 1",
    "action": "publishbookmark",
    "disabled" : false,
    "settings" : {
        "id" : "bookmark1"
    }
}
```

Publish bookmark with title "bookmark of testuser", where `testuser` is the username of the simulated user. 

```json
{
    "label" : "Publish bookmark 2",
    "action": "publishbookmark",
    "disabled" : false,
    "settings" : {
        "title" : "bookmark of {{.UserName}}"
    }
}
```
