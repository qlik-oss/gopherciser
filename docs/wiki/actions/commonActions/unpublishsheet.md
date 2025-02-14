## UnpublishSheet action

Unpublish sheets in the current app.

* `mode`: 
    * `allsheets`: Unpublish all sheets in the app.
    * `sheetids`: Only unpublish the sheets specified by the `sheetIds` array.
* `sheetIds`: (optional) Array of sheet IDs for the `sheetids` mode.

### Example
```json
{
     "label": "UnpublishSheets",
     "action": "unpublishsheet",
     "settings": {
       "mode": "allsheets"        
     }
}
```

