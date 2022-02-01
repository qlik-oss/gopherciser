
### Examples

#### Search with one search term
```json
{
    "action": "smartsearch",
    "label": "one term search"
    "settings": {
        "searchtext": "term1"
    }
}
```

#### Search with two search terms
```json
{
    "action": "smartsearch",
    "label": "two term search"
    "settings": {
        "searchtext": "term1 term2"
    }
}
```

#### Search with one search term including spaces
```json
{
    "action": "smartsearch",
    "label": "one term search including spaces"
    "settings": {
        "searchtext": "\"word1 word2\""
    }
}
```

#### Search with two search terms, one of them including spaces
```json
{
    "action": "smartsearch",
    "label": "two term search, one including spaces"
    "settings": {
        "searchtext": "\"word1 word2\" term2"
    }
}
```

#### Search with one search term including double quote
```json
{
    "action": "smartsearch",
    "label": "one term search including spaces"
    "settings": {
        "searchtext": "\\\"hello"
    }
}
```
