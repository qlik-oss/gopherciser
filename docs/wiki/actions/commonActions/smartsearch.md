## SmartSearch action

Perform a Smart Search in Sense app to find suggested selections.

* `searchtextsource`: Source for list of strings used for searching.
    * `searchtextlist` (default)
    * `searchtextfile`
* `searchtextlist`: List of of strings used for searching.
* `searchtextfile`: File path to file with one search string per line.
* `pastesearchtext`: 
    * `true`: Simulate pasting search text.
    * `false`: Simulate typing at normal speed (default).
* `makeselection`: Select a random search result.
    * `true`
    * `false`
* `selectionthinktime`: Think time before selection if `makeselection` is `true`, defaults to a 1 second delay.
  * `type`: Type of think time
      * `static`: Static think time, defined by `delay`.
      * `uniform`: Random think time with uniform distribution, defined by `mean` and `dev`.
  * `delay`: Delay (seconds), used with type `static`.
  * `mean`: Mean (seconds), used with type `uniform`.
  * `dev`: Deviation (seconds) from `mean` value, used with type `uniform`.


### Examples

#### Search with one search term
```json
{
    "action": "smartsearch",
    "label": "one term search",
    "settings": {
        "searchtextlist": [
            "term1"
        ]
    }
}
```

#### Search with two search terms
```json
{
    "action": "smartsearch",
    "label": "two term search",
    "settings": {
        "searchtextlist": [
            "term1 term2"
        ]
    }
}
```

#### Search with random selection of search text from list
```json
{
    "action": "smartsearch",
    "settings": {
        "searchtextlist": [
            "text1",
            "text2",
            "text3"
        ]
    }
}
```

#### Search with random selection of search text from file
```json
{
    "action": "smartsearch",
    "settings": {
        "searchtextsource": "searchtextfile",
        "searchtextfile": "data/searchtexts.txt"
    }
}
```
##### `data/searchtexts.txt`
```
search text
"quoted search text"
another search text
```

#### Simulate pasting search text

The default behavior is to simulate typing at normal speed.
```json
{
    "action": "smartsearch",
    "settings": {
        "pastesearchtext": true,
        "searchtextlist": [
            "text1"
        ]
    }
}
```

#### Make a random selection from search results
```json
{
    "action": "smartsearch",
    "settings": {
        "searchtextlist": [
            "term1"
        ],
        "makeselection": true,
        "selectionthinktime": {
            "type": "static",
            "delay": 2
        }
    }
}
```

#### Search with one search term including spaces
```json
{
    "action": "smartsearch",
    "settings": {
        "searchtextlist": [
            "\"word1 word2\""
        ]
    }
}
```

#### Search with two search terms, one of them including spaces
```json
{
    "action": "smartsearch",
    "label": "two term search, one including spaces",
    "settings": {
        "searchtextlist": [
            "\"word1 word2\" term2"
        ]
    }
}
```

#### Search with one search term including double quote
```json
{
    "action": "smartsearch",
    "label": "one term search including spaces",
    "settings": {
        "searchtext":
        "searchtextlist": [
            "\\\"hello"
        ]
    }
}
```

