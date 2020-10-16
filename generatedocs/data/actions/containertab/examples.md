### Examples

```json
{
  "label": "Switch to object qwerty in container object XYZ",
  "action": "containertab",
  "settings": {
    "containerid": "xyz",
    "mode": "id",
    "objectid" : "qwerty"
  }
}
```

```json
{
  "label": "Switch to random object in container object XYZ",
  "action": "containertab",
  "settings": {
    "containerid": "xyz",
    "mode": "random"
  }
}
```

```json
{
  "label": "Switch to object in first tab in container object XYZ",
  "action": "containertab",
  "settings": {
    "containerid": "xyz",
    "mode": "index",
    "index": 0
  }
}
```
