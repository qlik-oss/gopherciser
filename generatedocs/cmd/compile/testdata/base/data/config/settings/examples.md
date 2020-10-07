### Examples

```json
"settings": {
	"timeout": 300,
	"logs": {
		"traffic": false,
		"debug": false,
		"filename": "logs/{{.ConfigFile}}-{{timestamp}}.log"
	}
}
```

```json
"settings": {
	"timeout": 300,
	"logs": {
		"filename": "logs/scenario.log"
	},
	"outputs" : {
	    "dir" : "./outputs"
	}
}
```
