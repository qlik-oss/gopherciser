### Examples

The following example shows how to clear the artifact map and fill it with apps having the tag "mytag" from the first page in the hub.

```json
{
	"action": "ElasticExplore",
	"label": "",
	"settings": {
		"keepcurrent": false,
		"tags": ["mytag"]
	}
}
```

The following example shows how to clear the artifact map, fill it with all apps from the space "myspace" and then add all apps from the space "circles".

```json
{
	"action": "ElasticExplore",
	"label": "",
	"settings": {
		"keepcurrent": false,
		"space": "myspace",
		"paging": true
	}
},
{
	"action": "ElasticExplore",
	"label": "",
	"settings": {
		"keepcurrent": true,
		"space": "circles",
		"paging": true
	}
}
```

The following example shows how to clear the artifact map and fill it with the apps from the first page of the space "spaceX". The apps must have the tag "tag" or "team" or a tag with id "15172f9c-4a5f-4ee9-ae35-34c1edd78f8d", but not be created by the simulated user. In addition, the apps are sorted by the time of modification.

```json
{
	"action": "ElasticExplore",
	"label": "",
	"settings": {
		"keepcurrent": false,
		"space": "spaceX",
		"tags": ["tag", "team"],
		"tagids": ["15172f9c-4a5f-4ee9-ae35-34c1edd78f8d"],
		"owner": "others",
		"sorting": "updated",
		"paging": false
	}
}
```
