## ApplyBookmark action

Apply a bookmark in the current app.

**Note:** Specify *either* `title` *or* `id`, not both.

* `title`: Name of the bookmark (supports the use of [variables](#session_variables)).
* `id`: ID of the bookmark.
* `selectionsonly`: Apply selections only.

### Example

```json
{
    "action": "applybookmark",
    "settings": {
        "title": "My bookmark"
    }
}
```

