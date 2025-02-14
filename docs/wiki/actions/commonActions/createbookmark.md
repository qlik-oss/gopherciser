## CreateBookmark action

Create a bookmark from the current selection and selected sheet.

**Note:** Both `title` and `id` can be used to identify the bookmark in subsequent actions. 

* `title`: Name of the bookmark (supports the use of [variables](#session_variables)).
* `id`: ID of the bookmark.
* `description`: (optional) Description of the bookmark to create.
* `nosheet`: Do not include the sheet location in the bookmark.
* `savelayout`: Include the layout in the bookmark.

### Example

```json
{
    "action": "createbookmark",
    "settings": {
        "title": "my bookmark",
        "description": "This bookmark contains some interesting selections"
    }
}
```

