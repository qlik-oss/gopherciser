## Select action

Select random values in an object.

See the [Limitations](README.md#limitations) section in the README.md file for limitations related to this action.
 
* `id`: ID of the object in which to select values.
* `type`: Selection type
    * `randomfromall`: Randomly select within all values of the symbol table.
    * `randomfromenabled`: Randomly select within the white and light grey values on the first data page.
    * `randomfromexcluded`: Randomly select within the dark grey values on the first data page.
    * `randomdeselect`: Randomly deselect values on the first data page.
    * `values`: Select specific element values, defined by `values` array.
* `accept`: Accept or abort selection after selection (only used with `wrap`) (`true` / `false`).
* `wrap`: Wrap selection with Begin / End selection requests (`true` / `false`).
* `min`: Minimum number of selections to make.
* `max`: Maximum number of selections to make.
* `dim`: Dimension / column in which to select.
* `values`: Array of element values to select when using selection type `values`. These are the element values for a selection, not the values seen by the user.

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

