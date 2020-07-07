## Scheduler section

This section of the JSON file contains scheduler settings for the users in the load scenario.

### Using "reconnectsettings"

If `reconnectsettings.reconnect` is enabled the tool will attempt the following:

* Re-connect the websocket.
* Get the currently opened app in re-attached engine session.
* Re-subscribe to the same object has before getting disconnected.
* Restart the action which was ongoing during the disconnect.
* If successful the action where the re-connect happened will be logged as a successful action with changed `action` and `label`, these will instead be `Reconnect(action)` and `Reconnect(label)`.
* Logs an info row with the info type `WebsocketReconnect` and with a semicolon separated `details` section as follows: "success=`X`;attempts=`Y`;TimeSpent=`Z`" where:
    * `X` is true/false
    * `Y` is an integer with the amount of re-connection attempts
    * `Z` is the time spent re-connecting in milliseconds.   