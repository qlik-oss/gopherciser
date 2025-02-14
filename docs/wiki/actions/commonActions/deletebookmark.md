## DeleteBookmark action

Delete one or more bookmarks in the current app.

**Note:** Specify *either* `title` *or* `id`, not both.

* `title`: Name of the bookmark (supports the use of [variables](#session_variables)).
* `id`: ID of the bookmark.
* `mode`: 
    * `single`: Delete one bookmark that matches the specified `title` or `id` in the current app.
    * `matching`: Delete all bookmarks with the specified `title` in the current app.
    * `all`: Delete all bookmarks in the current app.

### Example

```json
{
    "action": "deletebookmark",
    "settings": {
        "mode": "single",
        "title": "My bookmark"
    }
}
```

