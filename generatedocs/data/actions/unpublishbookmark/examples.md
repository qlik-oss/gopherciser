### Example

Unpublish bookmark previously created in the script with an id `bookmark1`.

```json
{
    "label" : "Unpublish bookmark 1",
    "action": "unpublishbookmark",
    "disabled" : false,
    "settings" : {
        "id" : "bookmark1"
    }
}
```

Unpublish bookmark with title "bookmark of testuser", where `testuser` is the username of the simulated user. 

```json
{
    "label" : "Unpublish bookmark 2",
    "action": "unpublishbookmark",
    "disabled" : false,
    "settings" : {
        "title" : "bookmark of {{.UserName}}"
    }
}
```
