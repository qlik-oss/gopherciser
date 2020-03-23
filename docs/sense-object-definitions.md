# Supporting extensions and overriding defaults

Gopherciser supports custom definitions for how to get data and make selections in standard Qlik Sense® objects. Defaults can be overridden and custom extensions added when it comes to getting data and making selections.

## Importing custom object definitions

Definitions are imported from a JSON file using the `--definitions` (or `-d`) flag with the `execute` command. Only the definitions to be overridden or added need to be included in the definitions file. 

```bash
./gopherciser execute -c localtests/QlikCoreCtrl00Select.json -d defs.json
```

## Objdef command

The `objdef` command handles export and validation of definitions files.

```bash
╰─➤  ./gopherciser objdef
Handles sense object data definitions for gopherciser.
Use to export default values or examples or to validate custom definitions or overrides.

Usage:
  gopherciser objdef [flags]
  gopherciser objdef [command]

Aliases:
  objdef, od

Available Commands:
  generate    Generate object definitions from default values.
  validate    Validate object definitions in file.

Flags:
  -h, --help   help for objdef

Use "gopherciser objdef [command] --help" for more information about a command.
```

### Exporting default definitions

```bash
╰─➤  ./gopherciser objdef generate
Generate object definitions from default values, either a full json with all defaults or defined objects.

Usage:
  gopherciser objdef generate [flags]

Aliases:
  generate, gen, g

Flags:
  -d, --definitions string   (mandatory) definitions file.
  -f, --force                overwrite definitions file if existing.
  -h, --help                 help for generate
  -o, --objects strings      (optional) list of objects, defaults to all.
```

The `generate` command generates a JSON definitions file (defined by the `--definitions` flag) with all default object definitions.

```bash
╰─➤  ./gopherciser objdef generate -d defs.json
defs.json written successfully.
```

If only one or a few objects are to be overridden, they can be defined in a list. Only the objects to be overridden or added need to be in the definitions file. If the definitions file already exists, the `--force` flag is used to overwrite the file. 
 
```bash
╰─➤  ./gopherciser objdef generate -d defs.json -o listbox,barchart -f
defs.json written successfully.
```

The above produces a definitions file similar to the following:

```JSON
{
  "barchart": {
    "datadef": {
      "type": "hypercube",
      "path": "/qHyperCube"
    },
    "data": [
      {
        "requests": [
          {
            "type": "hypercubereduceddata",
            "path": "/qHyperCubeDef",
            "height": 500
          }
        ]
      }
    ],
    "select": {
      "type": "hypercubevalues",
      "path": "/qHyperCubeDef"
    }
  },
  "listbox": {
    "datadef": {
      "type": "listobject",
      "path": "/qListObject"
    },
    "data": [
      {
        "requests": [
          {
            "type": "listobjectdata",
            "path": "/qListObjectDef",
            "height": 500
          }
        ]
      }
    ],
    "select": {
      "type": "listobjectvalues",
      "path": "/qListObjectDef"
    }
  }
}
```

### Validating a definitions file

The `validate` command is used to perform basic structural validation of definitions files.

```bash
╰─➤  ./gopherciser objdef validate
Validate object definitions from provided JSON file. Will print how many definitions where found,
it's recommended to use to -v verbose flag and verify all parameters where interpreted correctly'

Usage:
  gopherciser objdef validate [flags]

Aliases:
  validate, val, v

Flags:
  -d, --definitions string   (mandatory) definitions file.
  -h, --help                 help for validate
  -v, --verbose              print summary of definitions.
```

It is recommended to use the `--verbose` (or `-v`) flag to display a summary when validating a definitions file.

```bash
╰─➤  ./gopherciser objdef validate -d defs.json -v
2 object definitions found

[barchart]
/ 1 data constraint entry.
|
|   Constraint: Default
|     1 data request:
|     [0]: Type:   hypercubereduceddata
|          Path:   /qHyperCubeDef
|          Height: 500
|
|
/ DataDef Type: hypercube
|         Path: /qHyperCube
|
/ Select  Type: hypercubevalues
|         Path: /qHyperCubeDef
*

[listbox]
/ 1 data constraint entry.
|
|   Constraint: Default
|     1 data request:
|     [0]: Type:   listobjectdata
|          Path:   /qListObjectDef
|          Height: 500
|
|
/ DataDef Type: listobject
|         Path: /qListObject
|
/ Select  Type: listobjectvalues
|         Path: /qListObjectDef
*
```

## Object definition

An object definition consists of the following sections:

* `datadef`: Defines the type of data and where in the object structure to find the data
* `data`: Defines the types of data requests to send and under which circumstances
* `select`: Defines the select method and path to use for the object

### Datadef

* `type`: Type of data
    * `listobject`: Data carrier is a list object (for example, a listbox).
    * `hypercube`: Data carrier is a hypercube (used for most charts).
    * `nodata`: Object does not contain any data.
* `path`: Path to the data carrier within the object structure.

```json
"datadef": {
  "type": "hypercube",
  "path": "/qHyperCube"
}
```

### Data

A list of data requests to send for an object, with the possibility to send different requests depending on different  constraints. The list of constraints is evaluated top-down, where the first constraint fulfilled is used. "Nil" requests are always considered to be true (that is, if the list of constraints starts with an entry without a constraint, the subsequent constraints are not used). It is therefore important to start the list with the constraint to be evaluated first. In the above `scatterplot` example, the constraint "qcy > 1000" is evaluated before performing the default operation.

* `constraint`: Constraint for sending the defined set of data requests. An empty or omitted constraint is always considered to be  true.
    * `path`: Path to the value to evaluate in the object structure.
    * `value`: Value constraint definition. The first character must be `<`, `>`, `=` or `!` followed by a number or the words `true` / `false`. 
    * `required`: Require the constraint to be evaluated and return an error if the evaluation fails (for example, if the path in the object structure is not traversable). Defaults to `false`.
* `requests`: List of data requests to send if the constraint is successfully evaluated. A request is defined as:
    * `type`: Data request type
        * `layout`: Get data from layout.
        * `listobjectdata`: Get data from listobject data.
        * `hypercubedata`: Get data from hypercube.
        * `hypercubereduceddata`: Get hypercube reduced data.
        * `hypercubebinneddata`: Get hypercube binned data.
        * `hypercubestackdata`: Get hypercube stacked data.
        * `hypercubedatacolumns`: Get data from hypercube "as columns".
        * `hypercubecontinuousdata`: Get hypercube continuous data.
    * `path`: Path to be sent in the get data request.
    * `height`: Height of data. The default height is used, if omitted or set to `0`.

### Select

* `type`: Type of select request
    * `listobjectvalues`: Send a `SelectListObjectValues` request.
    * `hypercubevalues`: Send a `SelectHyperCubeValues` request unless the data is binned, in which case a `MultiRangeSelectHyperCubeValues` request is sent.
    * `hypercubecolumnvalues`: Same as `hypercubevalues`, except that the data is considered to be in columned format when using select methods such as `RandomFromEnabled`, `RandomFromExcluded` etc.
* `path`: Path to send in the select request.

### Examples

Simple example of an object with the data in Layout message only:

```json
{
  "qlik-word-cloud": {
    "datadef": {
      "type": "hypercube",
      "path": "/qHyperCube"
    },
    "data": [
      {
        "requests": [
          {
            "type": "layout"
          }
        ]
      }
    ]
  }
}
```

An object that sends different data requests depending on the size of the data:

```json
{
  "scatterplot": {
    "datadef": {
      "type": "hypercube",
      "path": "/qHyperCube"
    },
    "data": [
      {
        "constraint": {
          "path": "/qHyperCube/qSize/qcy",
          "value": ">1000",
          "required": true
        },
        "requests": [
          {
            "type": "hypercubebinneddata",
            "path": "/qHyperCubeDef"
          }
        ]
      },
      {
        "requests": [
          {
            "type": "hypercubedata",
            "path": "/qHyperCubeDef",
            "height": 1000
          }
        ]
      }
    ],
    "select": {
      "type": "hypercubevalues",
      "path": "/qHyperCubeDef"
    }
  }
}
```
