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

#### Fromfile login request type

Reads a user list from file. 1 User per row of the and with the format `username;directory;password`. `directory` and `password` are optional, if none are defined for a user it will use the default values on settings (i.e. `defaultdir` and `defaultpassword`). If the used authentication type doesn't use `directory` or `password` these can be omitted.

Definition with default values:

```json
"loginSettings": {
  "type": "fromfile",
  "settings": {
    "filename": "./myusers.txt",
    "directory": "defaultdir",
    "password": "defaultpassword"
  }
}
```

Definition without default values:

```json
"loginSettings": {
  "type": "fromfile",
  "settings": {
    "filename": "./myusers.txt"
  }
}
```

This is a valid format of a file.

```text
testuser1
testuser2;myspecialdirectory
testuser3;;somepassword
testuser4;specialdir;anotherpassword
testuser5;;A;d;v;a;n;c;e;d;;P;a;s;s;w;o;r;d;
```

*testuser1* will get default `directory` and `password`, *testuser3* and *testuser5* will get default `directory`.
