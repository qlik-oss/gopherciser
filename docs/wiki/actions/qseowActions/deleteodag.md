## DeleteOdag action

Delete all user-generated on-demand apps for the current user and the specified On-Demand App Generation (ODAG) link.

* `linkname`: Name of the ODAG link from which to delete generated apps. The name is displayed in the ODAG navigation bar at the bottom of the *selection app*.

### Example

```json
{
    "action": "DeleteOdag",
    "settings": {
        "linkname": "Drill to Template App"
    }
}
```

