### Examples

#### ThinkTime uniform

This simulates a think time of 10 to 15 seconds.

```json
{
     "label": "TimerDelay",
     "action": "thinktime",
     "settings": {
         "type": "uniform",
         "mean": 12.5,
         "dev": 2.5
     } 
} 
```

#### ThinkTime constant

This simulates a think time of 5 seconds.

```json
{
     "label": "TimerDelay",
     "action": "thinktime",
     "settings": {
         "type": "static",
         "delay": 5
     }
}
```
