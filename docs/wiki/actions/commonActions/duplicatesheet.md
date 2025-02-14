## DuplicateSheet action

Duplicate a sheet, including all objects.

* `id`: ID of the sheet to clone.
* `changesheet`: Clear the objects currently subscribed to and then subribe to all objects on the cloned sheet (which essentially corresponds to using the `changesheet` action to go to the cloned sheet) (`true` / `false`). Defaults to `false`, if omitted.
* `save`: Execute `saveobjects` after the cloning operation to save all modified objects (`true` / `false`). Defaults to `false`, if omitted.
* `cloneid`: (optional) ID to be used to identify the sheet in any subsequent `changesheet`, `duplicatesheet`, `publishsheet` or `unpublishsheet` action.

### Example

```json
{
    "action": "duplicatesheet",
    "label": "Duplicate sheet1",
    "settings":{
        "id" : "mBshXB",
        "save": true,
        "changesheet": true
    }
}
```

