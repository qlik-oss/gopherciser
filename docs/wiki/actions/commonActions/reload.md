## Reload action

Reload the current app by simulating selecting **Load data** in the Data load editor. To select an app, preceed this action with an `openapp` action.

* `mode`: Error handling during the reload operation
    * `default`: Use the default error handling.
    * `abend`: Stop reloading the script, if an error occurs.
    * `ignore`: Continue reloading the script even if an error is detected in the script.
* `partial`: Enable partial reload (`true` / `false`). This allows you to add data to an app without reloading all data. Defaults to `false`, if omitted.
* `log`: Save the reload log as a field in the output (`true` / `false`). Defaults to `false`, if omitted. **Note:** This should only be used when needed as the reload log can become very large.
* `nosave`: Do not send a save request for the app after the reload is done. Defaults to saving the app.

### Example

```json
{
    "action": "reload",
    "settings": {
        "mode" : "default",
        "partial": false
    }
}
```

