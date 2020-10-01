### Example

```json
//Keep-alive loop
{
    "action": "iterated",
    "settings" : {
        "iterations" : 10,
        "actions" : [
            {
                "action" : "productversion"
            },
            {
                "action": "thinktime",
                "settings": {
                    "type": "static",
                    "delay": 30
                }
            }
        ]
    }
}
```
