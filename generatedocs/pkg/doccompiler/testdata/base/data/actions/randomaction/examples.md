### Random action defaults

The following default values are used for the different actions:

* `thinktime`: Mirrors the configuration of `thinktimesettings`
* `sheetobjectselection`:

```json
{
     "settings": 
     {
         "id": <UNIFORMLY RANDOMIZED>,
         "type": "RandomFromAll",
         "min": 1,
         "max": 2,
         "accept": true
     }
}
```

* `changesheet`:

```json
{
     "settings": 
     {
         "id": <UNIFORMLY RANDOMIZED>
     }
}
```

* `clearall`:

```json
{
     "settings": 
     {
     }
}
```

### Examples

#### Generating a background load by executing 5 random actions

```json
{
    "action": "RandomAction",
    "settings": {
        "iterations": 5,
        "actions": [
            {
                "type": "thinktime",
                "weight": 1
            },
            {
                "type": "sheetobjectselection",
                "weight": 3
            },
            {
                "type": "changesheet",
                "weight": 5
            },
            {
                "type": "clearall",
                "weight": 1
            }
        ],
        "thinktimesettings": {
            "type": "uniform",
            "mean": 10,
            "dev": 5
        }
    }
}
```

#### Making random selections from excluded values

```json
{
    "action": "RandomAction",
    "settings": {
        "iterations": 1,
        "actions": [
            {
                "type": "sheetobjectselection",
                "weight": 1,
                "overrides": {
                  "type": "RandomFromExcluded",
                  "min": 1,
                  "max": 5
                }
            }
        ],
        "thinktimesettings": {
            "type": "static",
            "delay": 1
        }
    }
}
```
