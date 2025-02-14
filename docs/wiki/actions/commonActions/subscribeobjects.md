## Subscribeobjects action

Subscribe to any object in the currently active app.

* `clear`: Remove any previously subscribed objects from the subscription list.
* `ids`: List of object IDs to subscribe to.

### Example

Subscribe to two objects in the currently active app and remove any previous subscriptions. 

```json
{
    "action" : "subscribeobjects",
    "label" : "clear subscriptions and subscribe to mBshXB and f2a50cb3-a7e1-40ac-a015-bc4378773312",
     "disabled": false,
    "settings" : {
        "clear" : true,
        "ids" : ["mBshXB", "f2a50cb3-a7e1-40ac-a015-bc4378773312"]
    }
}
```

Subscribe to an additional single object (or a list of objects) in the currently active app, adding the new subscription to any previous subscriptions.

```json
{
    "action" : "subscribeobjects",
    "label" : "add c430d8e2-0f05-49f1-aa6f-7234e325dc35 to currently subscribed objects",
     "disabled": false,
    "settings" : {
        "clear" : false,
        "ids" : ["c430d8e2-0f05-49f1-aa6f-7234e325dc35"]
    }
}
```
