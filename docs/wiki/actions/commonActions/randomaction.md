## RandomAction action

Randomly select other actions to perform. This meta-action can be used as a starting point for your testing efforts, to simplify script authoring or to add background load.

`randomaction` accepts a list of action types between which to randomize. An execution of `randomaction` executes one or more of the listed actions (as determined by the `iterations` parameter), randomly chosen by a weighted probability. If nothing else is specified, each action has a default random mode that is used. An override is done by specifying one or more parameters of the original action.

Each action executed by `randomaction` is followed by a customizable `thinktime`.

**Note:** The recommended way to use this action is to prepend it with an `openapp` and a `changesheet` action as this ensures that a sheet is always in context.

* `actions`: List of actions from which to randomly pick an action to execute. Each item has a number of possible parameters.
  * `type`: Type of action
      * `thinktime`: See the `thinktime` action.
      * `sheetobjectselection`: Make random selections within objects visible on the current sheet. See the `select` action.
      * `changesheet`: See the `changesheet` action.
      * `clearall`: See the `clearall` action.
  * `weight`: The probabilistic weight of the action, specified as an integer. This number is proportional to the likelihood of the specified action, and is used as a weight in a uniform random selection.
  * `overrides`: (optional) Static overrides to the action. The overrides can include any or all of the settings from the original action, as determined by the `type` field. If nothing is specified, the default values are used.
* `thinktimesettings`: Settings for the `thinktime` action, which is automatically inserted after every randomized action.
  * `type`: Type of think time
      * `static`: Static think time, defined by `delay`.
      * `uniform`: Random think time with uniform distribution, defined by `mean` and `dev`.
  * `delay`: Delay (seconds), used with type `static`.
  * `mean`: Mean (seconds), used with type `uniform`.
  * `dev`: Deviation (seconds) from `mean` value, used with type `uniform`.
* `iterations`: Number of random actions to perform.

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

