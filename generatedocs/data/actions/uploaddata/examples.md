### Example

Upload data to personal space.

```json
{
     "action": "UploadData",
     "settings": {
         "filename": "/home/root/data.csv"
     }
}
```

Upload data to personal space, replacing existing file.

```json
{
     "action": "UploadData",
     "settings": {
         "filename": "/home/root/data.csv",
         "replace": true
     }
}
```

Upload data to space with space ID 25180576-755b-46e1-8683-12062584e52c.

```json
{
     "action": "UploadData",
     "settings": {
         "filename": "/home/root/data.csv",
         "spaceid": "25180576-755b-46e1-8683-12062584e52c"
     }
}
```
