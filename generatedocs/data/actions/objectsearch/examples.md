### Examples

Search a listbox object, all users searches for same thing and gets an error if no result found

```json
{
    "label": "Search and select Sweden in listbox",
    "action": "objectsearch",
    "settings": {
        "id": "maesVjgte",
        "searchterms": ["Sweden"],
        "type": "listbox",
        "erroronempty": true
    }
}
```

Search a field, users randomly uses one search term from the `searchterms` list.

```json
{
    "label": "Search field",
    "action": "objectsearch",
    "disabled": false,
    "settings": {
        "id": "Countries",
        "searchterms": [
            "Sweden",
            "Germany",
            "Liechtenstein"
        ],
        "type": "field"
    }
}
```

Search a master object dimension using search terms from a file.

```json
{
    "label": "Search dimension",
    "action": "objectsearch",
    "disabled": false,
    "settings": {
        "id": "Dim1M",
        "type": "dimension",
        "erroronempty": true,
        "source": "fromfile",
        "searchtermsfile": "./resources/objectsearchterms.txt",
    }
}
```
