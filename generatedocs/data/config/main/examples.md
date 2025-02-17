
## Example

```json
{
  "settings": {
    "timeout": 300,
    "logs": {
      "filename": "scenarioresult.tsv"
    },
    "outputs": {
      "dir": ""
    }
  },
  "loginSettings": {
    "type": "prefix",
    "settings": {
      "prefix": "testuser"
    }
  },
  "connectionSettings": {
    "mode": "ws",
    "server": "localhost",
    "virtualproxy": "header",
    "security": true,
    "allowuntrusted": true,
    "headers": {
      "Qlik-User-Header": "{{.UserName}}"
    }
  },
  "scheduler": {
    "type": "simple",
    "iterationtimebuffer": {
      "mode": "onerror",
      "duration": "10s"
    },
    "instance": 1,
    "reconnectsettings": {
      "reconnect": false,
      "backoff": null
    },
    "settings": {
      "executionTime": -1,
      "iterations": 10,
      "rampupDelay": 7,
      "concurrentUsers": 10,
      "reuseUsers": false,
      "onlyinstanceseed": false
    }
  },
  "scenario": [
    {
      "action": "openhub",
      "label": "open hub",
      "disabled": false,
      "settings": {}
    },
    {
      "action": "thinktime",
      "label": "think for 10-15s",
      "disabled": false,
      "settings": {
        "type": "uniform",
        "mean": 15,
        "dev": 5
      }
    },
    {
      "action": "openapp",
      "label": "open app",
      "disabled": false,
      "settings": {
        "appmode": "name",
        "app": "myapp",
        "filename": "",
        "unique": false
      }
    },
    {
      "action": "thinktime",
      "label": "think for 10-15s",
      "disabled": false,
      "settings": {
        "type": "uniform",
        "mean": 15,
        "dev": 5
      }
    },
    {
      "action": "changesheet",
      "label": "change sheet to analysis sheet",
      "disabled": false,
      "settings": {
        "id": "QWERTY"
      }
    },
    {
      "action": "thinktime",
      "label": "think for 10-15s",
      "disabled": false,
      "settings": {
        "type": "uniform",
        "mean": 15,
        "dev": 5
      }
    },
    {
      "action": "select",
      "label": "select 1-10 values in object uvxyz",
      "disabled": false,
      "settings": {
        "id": "uvxyz",
        "type": "randomfromenabled",
        "accept": false,
        "wrap": false,
        "min": 1,
        "max": 10,
        "dim": 0,
        "values": null
      }
    }
  ]
}
```

