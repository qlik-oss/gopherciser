### Example

Unpublish the bookmark with id "bookmark1" that was created earlier on in the script.

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

Unpublish the bookmark with the title "bookmark of testuser", where "testuser" is the username of the simulated user.

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
