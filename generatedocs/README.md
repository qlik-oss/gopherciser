# Generating Gopherciser documentation

This document describes how to generate the Gopherciser documentation.

## What

The tools for generating the documentation first combine the documentation data and then generate a `documentation.go` file and subsequently a `settingup.md` file: 

* `documentation.go`: Contains a set of variables that can be used for accessing the documentation programmatically.
* `settingup.md`: A markdown formatted file to be rendered by a markdown reader or the GitHub project page.

## Why

The reason for having an intermediate "programmatically readable" step is to allow other projects to include parts of the documentation at compile time. Allowing the documentation to be read as code is easier than having other projects importing the documentation files. 

## How: Generating the documentation

### Compiling documentation data to be used by the GUI and for markdown generation

```
go run ./generatedocs/compile/compile.go
```

#### Optional flags

* `--output string`: Filepath to the generated file. Defaults to `generatedocs/generated/documentation.go`.
* `--data`: Filepath to the data to read. Defaults to `generatedocs/data`.  

### Generating markdown files

```
run ./generatedocs/generate/generate.go --output ./docs/settingup.md
```

#### Optional flags

* `--template`:  Defaults to `generatedocs/data/settingup.md.template`.
* `--output`: Defaults to `generatedocs/generated/settingup.md`.

The Gopherciser `main.go` file contains the following commands for `go generate`:

```
//go:generate go run ./generatedocs/compile/compile.go
//go:generate go run ./generatedocs/generate/generate.go --output ./docs/settingup.md
```

To generate new `documentation.go` and `settingup.md` files after updating the documentation data, use the following command in the project root:

```bash
go generate
```

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

### Documentation structure for actions, groups and config sections

The documentation structure for actions, groups and config sections follows the same pattern. They are all documented in three separate parts:

* Description
* Parameters
* Examples

The reason for the structure is to make the description and each parameter individually accessible when accessing the documentation programmatically. The description and examples parts are written directly as  markdown files, located in the appropriate action, config or groups subfolder. The following example shows the folder structure for the `applybookmark` action:

```
data
    -> actions
        -> applybookmark
            -> description.md
            -> examples.md
``` 

The parameters part is documented in the `params.json` file. The following example shows the parameters for the `applybookmark` action:

```json
{
	"applybookmark.title" : ["(optional) Name of the bookmark to apply."],
	"applybookmark.id" : ["(optional) GUID of the bookmark to apply."]
}
```

In the code for each parameter there is a `doc-key` tag that connects the parameter to its documentation in the `params.json` file. The following example shows the tags for `ApplyBookmarkSettings`:

```golang
//ApplyBookmarkSettings apply bookmark settings
ApplyBookmarkSettings struct {
    Title string `json:"title" displayname:"Bookmark title" doc-key:"applybookmark.title"`
    Id    string `json:"id" displayname:"Bookmark ID" doc-key:"applybookmark.id"`
}
```

### Documenting enums

As enums cannot be reflected, they are documented as lists. The entries in the `params.json` file are set to an array of strings, where the first string is the description of the parameter and the subsequent strings are sub items to the parameter. The following example shows the `doc-key` for the `owner` parameter in the `elasticexplore` action:

```golang
Owner OwnerMode `json:"owner" displayname:"Owner mode" doc-key:"elasticexplore.owner"`
```

The `doc-key` connects to the following entry in the `params.json` file:

```json
{
	"elasticexplore.owner" : ["Filter apps by owner",
	    "`all`: Apps owned by anyone.",
	    "`me`: Apps owned by the simulated user.",
	    "`others`: Apps not owned by the simulated user."]
}
```

The entry above renders the following markdown result:

```markdown
* `owner`: Filter apps by owner
    * `all`: Apps owned by anyone.
    * `me`: Apps owned by the simulated user.
    * `others`: Apps not owned by the simulated user.
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

If an action does not belong to a group, it is added to an `Ungrouped actions` section as defined in `settingup.md.template`.

### Extra folders

Any subfolder in the `extra` subfolder is added as a DocEntry in the `Extra` map in `documentation.go`. These can later be accessed individually in `settungup.md.template`.

## Templates

The following templates are used by `compile` and `generate`, respectively:

* `documentation.template`: Go template that generates the `documentation.go` file. The default output location is `generatedocs/generated/documentation.go`.
* `settingup.md.template`: Go template that generates a markdown file from the data in `documentation.go`. The default output location is `docs/settingup.md`.