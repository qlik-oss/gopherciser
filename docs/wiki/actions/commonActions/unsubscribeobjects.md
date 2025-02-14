## Unsubscribeobjects action

Unsubscribe to any currently subscribed object.

* `ids`: List of object IDs to unsubscribe from.
* `clear`: Remove any previously subscribed objects from the subscription list.

### Example

Unsubscribe from a single object (or a list of objects).

```json
{
    "action" : "unsubscribeobjects",
    "label" : "unsubscribe from object maVjt and its children",
    "disabled": false,
    "settings" : {
        "ids" : ["maVjt"]
    }
}
```

Unsubscribe from all currently subscribed objects.

```json
{
    "action" : "unsubscribeobjects",
    "label" : "unsubscribe from all objects",
    "disabled": false,
    "settings" : {
        "clear": true
    }
}
```
