## ThinkTime action

Simulate user think time.

**Note:** This action does not require an app context (that is, it does not have to be prepended with an `openapp` action).

* `type`: Type of think time
    * `static`: Static think time, defined by `delay`.
    * `uniform`: Random think time with uniform distribution, defined by `mean` and `dev`.
* `delay`: Delay (seconds), used with type `static`.
* `mean`: Mean (seconds), used with type `uniform`.
* `dev`: Deviation (seconds) from `mean` value, used with type `uniform`.

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

