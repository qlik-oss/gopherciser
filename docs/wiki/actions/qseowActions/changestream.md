## ChangeStream action

Change to specified stream. This makes the apps in the specified stream selectable by actions such as `openapp`.
* `mode`: Decides what kind of value the `stream` field contains. Defaults to `name`.
    * `name`: `stream` is the name of the stream.
    * `id`: `stream` is the ID if the stream.
* `stream`: 

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

