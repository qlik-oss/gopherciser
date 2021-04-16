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
        "iterations": 3,
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
                    "title": "Bookmark {{ $element:=range.ScriptVars.BookmarkCounter }} {{ $element }}{{ end }}"
                }
            }
        ]
    }
}
```

Combine two variables `MyArrayVar` and `BookmarkCounter` to create 3 bookmarks with the names `Bookmark one`, `Bookmark two` and `Bookmark three`.

```json
{
    "action": "setscriptvar",
    "settings": {
        "name": "MyArrayVar",
        "type": "array",
        "value": "one,two,three,four,five",
        "sep": ","
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
        "iterations": 3,
        "actions": [
            {
                "action": "createbookmark",
                "settings": {
                    "title": "Bookmark {{ index .ScriptVars.MyArrayVar .ScriptVars.BookmarkCounter }}",
                    "description": "This bookmark contains some interesting selections"
                }
            },
            {
                "action": "setscriptvar",
                "settings": {
                    "name": "BookmarkCounter",
                    "type": "int",
                    "value": "{{ .ScriptVars.BookmarkCounter | add 1}}"
                }
            }
        ]
    }
}
 ```