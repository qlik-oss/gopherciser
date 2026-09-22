### Example

Change to all sheets once to calculate states, then once more to get response times with cached initial states

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

Add a 5-15s think time inbetween each changesheet action.

```json
{
    "label" : "Sheetchanger uncached",
    "action": "sheetchanger",
    "settings": {
        "thinktimesettings": {
            "type": "uniform",
            "mean": 10,
            "dev": 5
        }
    }
}
```
