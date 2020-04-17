### Example

Publish the bookmark with `id` "bookmark1" that was created earlier on in the script.

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

Publish the bookmark with the `title` "bookmark of testuser", where "testuser" is the username of the simulated user.

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
