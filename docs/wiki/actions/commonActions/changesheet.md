## ChangeSheet action

Change to a new sheet, unsubscribe to the currently subscribed objects, and subscribe to all objects on the new sheet.

The action supports getting data from the following objects:

* Listbox
* Filter pane
* Bar chart
* Scatter plot
* Map (only the first layer)
* Combo chart
* Table
* Pivot table
* Line chart
* Pie chart
* Tree map
* Text-Image
* KPI
* Gauge
* Box plot
* Distribution plot
* Histogram
* Auto chart (including any support generated visualization from this list)
* Waterfall chart

* `id`: GUID of the sheet to change to.

### Example

```json
{
     "label": "Change Sheet Dashboard",
     "action": "ChangeSheet",
     "settings": {
         "id": "TFJhh"
     }
}
```

