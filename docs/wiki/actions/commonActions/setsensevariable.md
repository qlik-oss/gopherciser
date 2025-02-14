## SetSenseVariable action

Sets a Qlik Sense variable on a sheet in the open app.

* `name`: Name of the Qlik Sense variable to set.
* `value`: Value to set the Qlik Sense variable to. (supports the use of [session variables](#session_variables))

### Example

Set a variable to 2000

```json
{
     "name": "vSampling",
     "value": "2000"
}
```
