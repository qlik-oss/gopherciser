### Example

Create a variable containing a string and use it in openapp.

```json
{
    "action": "setscriptvar",
    "settings": {
        "name": "mylocalvar",
        "type": "string",
        "value": "My app Name with number for session {{ .Session }}"
    }
},
{
    "action": "openapp",
    "settings": {
        "appmode": "name",
        "app": "{{ .ScriptVars.mylocalvar }}"
    }
}
```

Create a variable containing an integer and use it in a loop creating bookmarks numbered 1 to 5. Then in a different loop reset variable and delete the bookmarks.

```json
{
    "action": "setscriptvar",
    "settings": {
        "name": "BookmarkCounter",
        "type": "int",
        "value": "0"
    }
},
{
    "action": "iterated",
    "settings": {
        "iterations": 5,
        "actions": [
            {
                "action": "setscriptvar",
                "settings": {
                    "name": "BookmarkCounter",
                    "type": "int",
                    "value": "{{ add .ScriptVars.BookmarkCounter 1 }}"
                }
            },
            {
                "action": "createbookmark",
                "settings": {
                    "title": "Bookmark {{ .ScriptVars.BookmarkCounter }}",
                    "description": "This bookmark contains some interesting selections"
                }
            }
            
        ]
    }
},
{
    "action": "setscriptvar",
    "settings": {
        "name": "BookmarkCounter",
        "type": "int",
        "value": "0"
    }
},
{
    "action": "iterated",
    "disabled": false,
    "settings": {
        "iterations": 5,
        "actions": [
            {
                "action": "setscriptvar",
                "settings": {
                    "name": "BookmarkCounter",
                    "type": "int",
                    "value": "{{ .ScriptVars.BookmarkCounter | add 1}}"
                }
            },
            {
                "action": "deletebookmark",
                "settings": {
                    "mode": "single",
                    "title": "Bookmark {{ .ScriptVars.BookmarkCounter }}"
                }
            }
        ]
    }
}
```

