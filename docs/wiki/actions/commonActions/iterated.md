## Iterated action

Loop one or more actions.

**Note:** This action does not require an app context (that is, it does not have to be prepended with an `openapp` action).

* `iterations`: Number of loops.
* `actions`: Actions to iterate
  * `action`: Name of the action to execute.
  * `label`: (optional) Custom string set by the user. This can be used to distinguish the action from other actions of the same type when analyzing the test results.
  * `disabled`: (optional) Disable action (`true` / `false`). If set to `true`, the action is not executed.
  * `settings`: Most, but not all, actions have a settings section with action-specific settings.

### Example

```json
//Visit all sheets twice
{
     "action": "iterated",
     "label": "",
     "settings": {
         "iterations" : 2,
         "actions" : [
            {
                 "action": "sheetchanger"
            },
            {
                "action": "thinktime",
                "settings": {
                    "type": "static",
                    "delay": 5
                }
            }
         ]
     }
}
```

