## OpenApp action

Open an app.

**Note:** If the app name is used to specify which app to open, this action cannot be the first action in the scenario. It must be preceded by an action that can populate the artifact map, such as `openhub`.

* `appmode`: App selection mode
    * `current`: (default) Use the current app, selected by an app selection in a previous action
    * `guid`: Use the app GUID specified by the `app` parameter.
    * `name`: Use the app name specified by the `app` parameter.
    * `random`: Select a random app from the artifact map, which is filled by e.g. `openhub`
    * `randomnamefromlist`: Select a random app from a list of app names. The `list` parameter should contain a list of app names.
    * `randomguidfromlist`: Select a random app from a list of app GUIDs. The `list` parameter should contain a list of app GUIDs.
    * `randomnamefromfile`: Select a random app from a file with app names. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `randomguidfromfile`: Select a random app from a file with app GUIDs. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
    * `round`: Select an app from the artifact map according to the round-robin principle.
    * `roundnamefromlist`: Select an app from a list of app names according to the round-robin principle. The `list` parameter should contain a list of app names.
    * `roundguidfromlist`: Select an app from a list of app GUIDs according to the round-robin principle. The `list` parameter should contain a list of app GUIDs.
    * `roundnamefromfile`: Select an app from a file with app names according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app name.
    * `roundguidfromfile`: Select an app from a file with app GUIDs according to the round-robin principle. The `filename` parameter should contain the path to a file in which each line represents an app GUID.
* `app`: App name or app GUID (supports the use of [session variables](#session_variables)). Used with `appmode` set to `guid` or `name`.
* `list`: List of apps. Used with `appmode` set to `randomnamefromlist`, `randomguidfromlist`, `roundnamefromlist` or `roundguidfromlist`.
* `filename`: Path to a file in which each line represents an app. Used with `appmode` set to `randomnamefromfile`, `randomguidfromfile`, `roundnamefromfile` or `roundguidfromfile`.
* `externalhost`: (optional) Sets an external host to be used instead of `server` configured in connection settings.
* `unique`: Create unqiue engine session not re-using session from previous connection with same user. Defaults to false.

### Examples

```json
{
     "label": "OpenApp",
     "action": "OpenApp",
     "settings": {
         "appmode": "guid",
         "app": "7967af99-68b6-464a-86de-81de8937dd56"
     }
}
```
```json
{
     "label": "OpenApp",
     "action": "OpenApp",
     "settings": {
         "appmode": "randomguidfromlist",
         "list": ["7967af99-68b6-464a-86de-81de8937dd56", "ca1a9720-0f42-48e5-baa5-597dd11b6cad"]
     }
}
```

