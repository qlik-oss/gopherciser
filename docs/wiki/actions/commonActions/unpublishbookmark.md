## UnpublishBookmark action

Unpublish a bookmark.

**Note:** Specify *either* `title` *or* `id`, not both.

* `title`: Name of the bookmark (supports the use of [variables](#session_variables)).
* `id`: ID of the bookmark.

### Example

Unpublish the bookmark with `id` "bookmark1" that was created earlier on in the script.

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

Unpublish the bookmark with the `title` "bookmark of testuser", where "testuser" is the username of the simulated user.

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

