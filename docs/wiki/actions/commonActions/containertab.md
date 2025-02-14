## Containertab action

A `Containertab` action simulates switching the active object in a `container` object.

* `mode`: Mode for container tab switching, one of: `objectid`, `random` or `index`.
    * `objectid`: Switch to tab with object defined by `objectid`.
    * `random`: Switch to a random visible tab within the container.
    * `index`: Switch to tab with zero based index defined but `index`.
* `containerid`: ID of the container object.
* `objectid`: ID of the object to set as active, used with mode `objectid`.
* `index`: Zero based index of tab to switch to, used with mode `index`.

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

