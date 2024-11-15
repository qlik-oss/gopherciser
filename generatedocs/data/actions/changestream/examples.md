### Example

Make apps in stream `Everyone` selectable by subsequent actions.

```json
{
     "label": "ChangeStream Everyone",
     "action": "changestream",
     "settings": {
         "mode": "name",
         "stream" : "Everyone"
     }
}
```

Make  apps in stream with id `ABSCDFSDFSDFO1231234` selectable subsequent actions.

```json
{
     "label": "ChangeStream Test1",
     "action": "changestream",
     "settings": {
         "mode": "id",
         "stream" : "ABSCDFSDFSDFO1231234"
     }
}
```
