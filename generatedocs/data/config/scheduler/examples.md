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

Simple scheduler set to attempt reconnect on unexpected websocket disconnection 

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
