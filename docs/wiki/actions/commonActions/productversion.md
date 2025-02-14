## ProductVersion action

Request the product version from the server and, optionally, save it to the log. This is a lightweight request that can be used as a keep-alive message in a loop.

* `log`: Save the product version to the log (`true` / `false`). Defaults to `false`, if omitted.

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

