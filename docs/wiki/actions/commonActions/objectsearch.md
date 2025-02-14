## ObjectSearch action

Perform a search select in a listbox, field or master dimension.


* `id`: Identifier for the object, this would differ depending on `type`.
    * `listbox`: Use the ID of listbox object
    * `field`: Use the name of the field
    * `dimension`: Use the title of the dimension masterobject.
* `searchterms`: List of search terms to search for.
* `type`: Type of object to search
    * `listbox`: (Default) `id` is the ID of a listbox.
    * `field`: `id` is the name of a field.
    * `dimension`: `id` is the title of a master object dimension.
* `source`: Source of search terms
    * `fromlist`: (Default) Use search terms from `searchterms` array.
    * `fromfile`: Use search term from file defined by `searchtermsfile`
* `erroronempty`: If set to true and the object search yields an empty result, the action will result in an error. Defaults to false.
* `searchtermsfile`: Path to search terms file when using `source` of type `fromfile`. File should contain one term per row.

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

Search a field. Users use one random search term from the `searchterms` list.

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
        "searchtermsfile": "./resources/objectsearchterms.txt"
    }
}
```

