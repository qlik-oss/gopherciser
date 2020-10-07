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
