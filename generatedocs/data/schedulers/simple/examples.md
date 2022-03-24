### Using `reconnectsettings`

If `reconnectsettings.reconnect` is enabled, the following is attempted:

1. Re-connect the WebSocket.
2. Get the currently opened app in the re-attached engine session.
3. Re-subscribe to the same object as before the disconnection.
4. If successful, the action during which the re-connect happened is logged as a successful action with `action` and `label` changed to `Reconnect(action)` and `Reconnect(label)`.
5. Restart the action that was executed when the disconnection occurred (unless it is a `thinktime` action, which will not be restarted).
6. Log an info row with info type `WebsocketReconnect` and with a semicolon-separated `details` section as follows: "success=`X`;attempts=`Y`;TimeSpent=`Z`"
    * `X`: True/false
    * `Y`: An integer representing the number of re-connection attempts
    * `Z`: The time spent re-connecting (ms)

### Example

Simple scheduler settings:

```json
"scheduler": {
   "type": "simple",
   "settings": {
       "executiontime": 120,
       "iterations": -1,
       "rampupdelay": 7.0,
       "concurrentusers": 10
   },
   "iterationtimebuffer" : {
       "mode": "onerror",
       "duration" : "5s"
   },
   "instance" : 2
}
```

Simple scheduler set to attempt re-connection in case of an unexpected WebSocket disconnection: 

```json
"scheduler": {
   "type": "simple",
   "settings": {
       "executiontime": 120,
       "iterations": -1,
       "rampupdelay": 7.0,
       "concurrentusers": 10
   },
   "iterationtimebuffer" : {
       "mode": "onerror",
       "duration" : "5s"
   },
    "reconnectsettings" : {
      "reconnect" : true
    }
}
```
