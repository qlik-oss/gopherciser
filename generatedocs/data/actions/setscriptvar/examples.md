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

A more advanced example.

Create a bookmark "BookmarkX" for each iteration in a loop, and add this to an array "MyArrayVar". After the first `iterated` action this will look like "Bookmark1,Bookmark2,Bookmark3". The second `iterated` action then deletes these bookmarks using the created array.

Dissecting the first array construction action. The `join` command takes the elements `.ScriptVars.MyArrayVar` and joins them together into a string separated by the separtor `,`. So with an array of [ elem1 elem2 ] this becomes a string as `elem1,elem2`. The `if` statement checks if the value of `.ScriptVars.BookmarkCounter` is 0, if it is 0 (i.e. the first iteration) it sets the string to `Bookmark1`. If it is not 0, it executes the join command on .ScriptVars.MyArrayVar, on iteration 3, the result of this would be `Bookmark1,Bookmark2` then it appends the fixed string `,Bookmark`, so far the string is `Bookmark1,Bookmark2,Bookmark`. Lastly it takes the value of `.ScriptVars.BookmarkCounter`, which is now 2, and adds 1 too it and appends, making the entire string `Bookmark1,Bookmark2,Bookmark3`.

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
    "disabled": false,
    "settings": {
        "iterations": 3,
        "actions": [
            {
                "action": "setscriptvar",
                "settings": {
                    "name": "MyArrayVar",
                    "type": "array",
                    "value": "{{ if eq 0 .ScriptVars.BookmarkCounter }}Bookmark1{{ else }}{{ join .ScriptVars.MyArrayVar \",\" }},Bookmark{{ .ScriptVars.BookmarkCounter | add 1 }}{{ end }}",
                    "sep": ","
                }
            },
            {
                "action": "createbookmark",
                "settings": {
                    "title": "{{ index .ScriptVars.MyArrayVar .ScriptVars.BookmarkCounter }}",
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
                "action": "deletebookmark",
                "settings": {
                    "mode": "single",
                    "title": "{{ index .ScriptVars.MyArrayVar .ScriptVars.BookmarkCounter }}"
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