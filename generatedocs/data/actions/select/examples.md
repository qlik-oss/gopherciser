### Example

Randomly select among all the values in object `RZmvzbF`.

```json
{
     "label": "ListBox Year",
     "action": "Select",
     "settings": {
         "id": "RZmvzbF",
         "type": "RandomFromAll",
         "accept": true,
         "wrap": false,
         "min": 1,
         "max": 3,
         "dim": 0
     }
}
```

Randomly select among all the enabled values (a.k.a "white" values) in object `RZmvzbF`.

```json
{
     "label": "ListBox Year",
     "action": "Select",
     "settings": {
         "id": "RZmvzbF",
         "type": "RandomFromEnabled",
         "accept": true,
         "wrap": false,
         "min": 1,
         "max": 3,
         "dim": 0
     }
}
```

#### Statically selecting specific values

This example selects specific element values in object `RZmvzbF`. These are the values which can be seen in a selection when e.g. inspecting traffic, it is not the data values presented to the user. E.g. when loading a table in the following order by a Sense loadscript:

```
Beta
Alpha
Gamma
```

which might be presented to the user sorted as

```
Alpha
Beta
Gamma
```

The element values will be Beta=0, Alpha=1 and Gamma=2.

To statically select "Gamma" in this case:

```json
{
     "label": "Select Gammma",
     "action": "Select",
     "settings": {
         "id": "RZmvzbF",
         "type": "values",
         "accept": true,
         "wrap": false,
         "values" : [2],
         "dim": 0
     }
}
```
