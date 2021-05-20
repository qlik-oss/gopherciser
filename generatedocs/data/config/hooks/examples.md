### Example

#### Send a request to slack that a test is starting.

```json
"hooks": {
    "preexecute": {
        "url": "https://hooks.slack.com/services/XXXXXXXXX/YYYYYYYYYYY/ZZZZZZZZZZZZZZZZZZZZZZZZ",
        "method": "POST",
        "payload": "{ \"text\": \"Running test with {{ .Scheduler.ConcurrentUsers }} concurrent users and {{ .Scheduler.Iterations }} iterations towards {{ .ConnectionSettings.Server }}.\"}",
        "contenttype": "application/json"
    },
    "postexecute": {
        "url": "https://hooks.slack.com/services/XXXXXXXXX/YYYYYYYYYYY/ZZZZZZZZZZZZZZZZZZZZZZZZ",
        "method": "POST",
        "payload": "{ \"text\": \"Test finished with {{ .Counters.Errors }} errors and {{ .Counters.Warnings }} warnings. Total Sessions: {{ .Counters.Sessions }}\"}"
    }
}
```

This will send a message on test startup such as:

```text
Running test with 10 concurrent users and 2 iterations towards MyServer.com.
```

And a message on test finished such as:

```text
Test finished with 4 errors and 12 warnings. Total Sessions: 20.
```

#### Ask an endpoint before execution if test is ok to run

```json
"hooks": {
    "preexecute": {
        "url": "http://myserver:8080/oktoexecute",
        "method": "POST",
        "headers": [
            {
                "name" : "someheader",
                "value": "headervalue"
            }
        ],
        "payload": "{\"testID\": \"12345\",\"startAt\": \"{{now.Format \"2006-01-02T15:04:05Z07:00\"}}\"}",
        "extractors": [
            {
                "name": "oktorun",
                "path" : "/oktorun",
                "faillevel": "error",
                "validator" : {
                    "type": "bool",
                    "value": "true"
                }
            }
        ]
    }
}
```

This will POST a request to `http://myserver:8080/oktoexecute` with the body:

```json
{
    "testID": "12345",
    "startAt": "2021-05-06T08:00:00Z01:00"
}
```

For a test started at `2021-05-06T08:00:00` in timezone UTC+1.

Let's assume the response from this endpoint is:

```json
{
    "oktorun": false
}
```

The validator with path `/oktorun` will extract the value `false` and compare to the value defined in the validator, in this case `true`. Since the they are not equal the test will stop with error before starting exection.
