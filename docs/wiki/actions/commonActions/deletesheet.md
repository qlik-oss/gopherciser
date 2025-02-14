## DeleteSheet action

Delete one or more sheets in the current app.

**Note:** Specify *either* `title` *or* `id`, not both.

* `mode`: 
    * `single`: Delete one sheet that matches the specified `title` or `id` in the current app.
    * `matching`: Delete all sheets with the specified `title` in the current app.
    * `allunpublished`: Delete all unpublished sheets in the current app.
* `title`: (optional) Name of the sheet to delete.
* `id`: (optional) GUID of the sheet to delete.

### Example

```json
{
    "action": "deletesheet",
    "settings": {
        "mode": "matching",
        "title": "Test sheet"
    }
}
```

