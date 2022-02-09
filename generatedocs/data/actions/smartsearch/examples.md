
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
    "label": "one term search",
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
    "label": "one term search",
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
    "label": "one term search",
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
    "label": "one term search",
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
    "label": "one term search",
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
