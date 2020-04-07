### Examples

#### ThinkTime uniform

This would simulate a think time between 10 and 15 seconds.

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

This would simulate a think time of 5 seconds.

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
