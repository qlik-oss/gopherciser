### Examples

#### Prefix login request type

```json
"loginSettings": {
   "type": "prefix",
   "settings": {
       "directory": "anydir",
       "prefix": "Nunit"
   }
}
```

#### Userlist login request type

```json
  "loginSettings": {
    "type": "userlist",
    "settings": {
      "userList": [
        {
          "username": "sim1@myhost.example",
          "directory": "anydir1",
          "password": "MyPassword1"
        },
        {
          "username": "sim2@myhost.example"
        }
      ],
      "directory": "anydir2",
      "password": "MyPassword2"
    }
  }
```
