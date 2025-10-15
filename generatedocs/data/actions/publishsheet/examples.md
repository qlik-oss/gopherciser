### Example

Publish two sheets

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

Publish all unpublished sheets with 5 seconds think time in between each one

```json
{
     "label": "PublishSheets",
     "action": "publishsheet",
     "settings": {
       "mode": "allsheets",
       "thinktime": "5s"
     }
}
```
