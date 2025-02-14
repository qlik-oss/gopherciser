## PublishSheet action

Publish sheets in the current app.

* `mode`: 
    * `allsheets`: Publish all sheets in the app.
    * `sheetids`: Only publish the sheets specified by the `sheetIds` array.
* `sheetIds`: (optional) Array of sheet IDs for the `sheetids` mode.
* `includePublished`: Try to publish already published sheets.

### Example
```json
{
     "label": "PublishSheets",
     "action": "publishsheet",
     "settings": {
       "mode": "sheetids",
       "sheetIds": ["qmGcYS", "bKbmgT"]
     }
}
```

