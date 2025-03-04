# Generating Gopherciser documentation

This document describes how to generate the Gopherciser documentation.

## What

The tools for generating the documentation first combine the documentation data and then generate a `documentation.go` file and subsequently a `settingup.md` file:

* `documentation.go`: Contains a set of variables that can be used for accessing the documentation programmatically.
* `settingup.md`: A markdown formatted file to be rendered by a markdown reader or the GitHub project page.

## Why

The reason for having an intermediate "programmatically readable" step is to allow other projects to include parts of the documentation at compile time. Allowing the documentation to be read as code is easier than having other projects importing the documentation files.

## How: Generating the documentation

To generate new `documentation.go` and `settingup.md` files after updating the documentation data, use the following command in the project root:

```bash
go generate
```

The Gopherciser `main.go` file contains the following commands for `go generate`:

```
//go:generate go run ./generatedocs/cmd/compiledocs
//go:generate go run ./generatedocs/cmd/generatemarkdown --output ./docs/settingup.md
```

### Compiling documentation data to be used by the GUI and for markdown generation

To only generate a new `documentation.go` file, use the following command:

```
go run ./generatedocs/cmd/compiledocs
```

#### Optional flags

* `--output string`: Filepath to the generated file. Defaults to `generatedocs/generated/documentation.go`.
* `--data`: Comma separated filepaths to the data to read. Filepaths Defaults to `generatedocs/data`.

### Generating markdown files

To only generate a new `settingup.md` file, use the following command:

```
go run ./generatedocs/generate/generate.go --output ./docs/settingup.md
```

#### Optional flags

* `--output`: Defaults to `generatedocs/generated/settingup.md`.

## How: Updating/adding data

The structure of the `data` folder is as follows:

```
data
	-> actions
		-> action folders
	-> config
		-> config sections folders
	-> extra
		-> extra folders
	-> groups
		-> groups folders
		-> groups.json
	-> documentation.template
	-> params.json
	-> settingup.md.template
```

Data can be overloaded by passing a comma separated list to the `--data` flag, e.g. `--data=path/to/data1,path/to/data2`. The overload precedence goes from low to high within the list, meaning `data1` will be overloaded by `data2`.

### Documentation structure for actions, groups and config sections

The documentation structure for actions, groups and config sections follows the same pattern. They are all documented in three separate parts:

* Description
* Parameters
* Examples

The reason for the structure is to make the description and each parameter individually accessible when accessing the documentation programmatically. The description and examples parts are written directly as  markdown files, located in the appropriate action, config or groups subfolder. The following example shows the folder structure for the `changesheet` action:

```
data
	-> actions
		-> changesheet
			-> description.md
			-> examples.md
```

The parameters part is documented in the `params.json` file. The following example shows the parameters for the `changesheet` action:

```json
{
	"changesheet.id" : ["GUID of the sheet to change to."]
}
```

In the code for each parameter there is a `doc-key` tag that connects the parameter to its documentation in the `params.json` file. The following example shows the tags for `ChangeSheetSettings`:

```golang
// ChangeSheetSettings settings for change sheet action
	ChangeSheetSettings struct {
		ID string `json:"id" displayname:"Sheet ID" doc-key:"changesheet.id" appstructure:"active:sheet"`
	}
```

### Documenting enums

As enums cannot be reflected, they are documented as lists. The entries in the `params.json` file are set to an array of strings, where the first string is the description of the parameter and the subsequent strings are sub items to the parameter. The following example shows the `doc-key` for the `type` parameter in the `select` action:

```golang
Type SelectionType `json:"type" displayname:"Selection type" doc-key:"select.type"`
```

The `doc-key` connects to the following entry in the `params.json` file:

```json
{
    "select.type": [
        "Selection type",
        "`randomfromall`: Randomly select within all values of the symbol table.",
        "`randomfromenabled`: Randomly select within the white and light grey values on the first data page.",
        "`randomfromexcluded`: Randomly select within the dark grey values on the first data page.",
        "`randomdeselect`: Randomly deselect values on the first data page.",
        "`values`: Select specific element values, defined by `values` array."
    ]
}
```

The entry above renders the following markdown result:

```markdown
* `type`: Selection type
    * `randomfromall`: Randomly select within all values of the symbol table.
    * `randomfromenabled`: Randomly select within the white and light grey values on the first data page.
    * `randomfromexcluded`: Randomly select within the dark grey values on the first data page.
    * `randomdeselect`: Randomly deselect values on the first data page.
    * `values`: Select specific element values, defined by `values` array.
```

### Groups of actions

Groups are defined in the `groups.json` file, which has the following structure:

* name: Name of the group
* title: Full title of the group
* actions: List of actions in the group

Example:

```json
{
	"name": "qseowActions",
	"title": "Qlik Sense Enterprise on Windows (QSEoW) actions",
	"actions": [
		"deleteodag",
		"generateodag",
		"openhub"
	]
}
```

If an action does not belong to a group, it is added to an `Ungrouped actions` section.

### Extra folders

Any subfolder in the `extra` subfolder is added as a DocEntry in the `Extra` map in `documentation.go`.

## How: Extending existing documetation

### Extending `documentation.go`

The [doccompiler](pkg/doccompiler) packge has support for extending the generated and programmatically readable documentation in `documentation.go`. The simplest way to extend the `documentation.go` in this repository is to use the [extenddocs](pkg/extenddocs) package (which makes use of the extend functionality in the `doccompiler`).


In a project which extends the `github.com/qlik-oss/gopherciser` the following `main` would implement a documentation extender with the same flags (`--data` and `--output`) as `compiledocs`:
```go
package main

import (
	"fmt"
	"os"

	"github.com/qlik-oss/gopherciser/generatedocs/pkg/extenddocs"

	// Make sure to register any new actions that shall be included in the
	// extended documentation. In this case they are registered in the `init()`
	// function of the `registeractions` package below.
	_ "github.com/my-user/extended-gopherciser/registeractions"
)

func main() {
	if err := extenddocs.ExtendOSSDocs(); err != nil {
		fmt.Printf("Errors:\n%v", err)
		os.Exit(1)
	}
}
```

### Extending `settingup.md`

The extended `settingup.md` shall then import the extended programatically readble documentation and use the [genmd](pkg/genmd) package to generate the markdown documentation.

```go
package main

import (
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/genmd"
	"github.com/my-user/extended-gopherciser/generatedocs/generated"

	// Once again, make sure to register any new actions that shall be included in the
	// extended documentation.
	_ "github.com/my-user/extended-gopherciser/registeractions"
)

func main() {
	genmd.GenerateMarkdown(&genmd.CompiledDocs{
		Actions: generated.Actions,
		Params:  generated.Params,
		Config:  generated.Config,
		Groups:  generated.Groups,
		Extra:   generated.Extra,
	})
}
```
