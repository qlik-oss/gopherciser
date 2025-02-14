## SheetChanger action

Create and execute a `changesheet` action for each sheet in an app. This can be used to cache the inital state for all objects or, by chaining two subsequent `sheetchanger` actions, to measure how well the calculations in an app utilize the cache.


### Example

```json
{
    "label" : "Sheetchanger uncached",
    "action": "sheetchanger"
},
{
    "label" : "Sheetchanger cached",
    "action": "sheetchanger"
}
```

