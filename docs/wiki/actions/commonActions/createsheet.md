## CreateSheet action

Create a new sheet in the current app.

* `id`: (optional) ID to be used to identify the sheet in any subsequent `changesheet`, `duplicatesheet`, `publishsheet` or `unpublishsheet` action.
* `title`: Name of the sheet to create.
* `description`: (optional) Description of the sheet to create.

### Example

```json
{
    "action": "createsheet",
    "settings": {
        "title" : "Generated sheet"
    }
}
```

