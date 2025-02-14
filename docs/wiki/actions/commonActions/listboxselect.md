## ListBoxSelect action

Perform list object specific selectiontypes in listbox.


* `id`: ID of the listbox in which to select values.
* `type`: Selection type.
    * `all`: Select all values.
    * `alternative`: Select alternative values.
    * `excluded`: Select excluded values.
    * `possible`: Select possible values.
* `accept`: Accept or abort selection after selection (only used with `wrap`) (`true` / `false`).
* `wrap`: Wrap selection with Begin / End selection requests (`true` / `false`).

### Examples

```json
{
     "label": "ListBoxSelect",
     "action": "ListBoxSelect",
     "settings": {
         "id": "951e2eee-ad49-4f6a-bdfe-e9e3dddeb2cd",
         "type": "all",
         "wrap": true,
         "accept": true
     }
}
```

