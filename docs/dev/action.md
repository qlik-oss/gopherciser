# Adding a new action

## VSCode snippets for gopherciser development in VSCode

Documentation how to use snippets to help with development of gopherciser actions when using VSCodium or VSCode can be found [here](./docs/vscode/Readme.md).

## Action interface

An action primarily needs to implement the `ActionSettings` interface.

```golang
// ActionSettings scenario action interface for mandatory methods
ActionSettings interface {
    // Execute action
    Execute(sessionState *session.State, actionState *action.State, connectionSettings *connection.ConnectionSettings, label string /* Label */, reset func() /* reset action start */)
    // Validate action []string are validation warnings to be reported to user
    Validate() ([]string, error)
}
```

Where `Execute` is the main logic for the action and `Validate` validates the settings of action on startup.

## Registering the action

An action needs to be registered to be be able to be found by the action handler. For a default action in github.com/qlik-oss/gopherciser repo, this is done by adding the action to the list in `ResetDefaultActions` function in [actionhandler.go](../../scenario/actionhandler.go).

If either the action is in an extension repo and not in github.com/qlik-oss/gopherciser, or added to a different package other than `scenario`, then it should be registered using `RegisterAction` during `init`. This can either be done for one action:

```golang
func init() {
    if err := scenario.RegisterAction("myaction", MyActionSettings{}); err != nil {
        panic(fmt.Sprintf("failed to register custom action:\n %+v", err))
    }
}
```

Multiple actions can also be registered simultanously, here's an example how to register two custom actions:

```golang
const(
    MyAction1 = "MyAction1"
    MyAction2 = "MyAction2"
)

func init() {
	err := scenario.RegisterActions(
		map[string]scenario.ActionSettings{
			MyAction1:   MyAction1Settings{},
			MyAction2:   MyAction2Settings{},
		})
	if err != nil {
		panic(fmt.Sprintf("failed to register custom actions:\n %+v", err))
	}
}
```

When extending github.com/qlik-oss/gopherciser, core actions could be also be overriden using `RegisterActionOverride` or `RegisterActionsOverride`.

## Optional interfaces

There's also a few optional interfaces which could be added to add certain specfic functionality to the action.

### IsContainerAction

If an action implements this interface on action settings to mark an action as a container action containing other actions. A container action will not log result as a normal action, instead result will be logged as level=info, infotype: containeractionend. This can be used when an action divides it's logic into sub actions, but does not itself contain any logic which should be measured.

```golang
func (settings MyActionSettings) IsContainerAction() {}
```

### AppStructureAction

AppStructureAction returns if this action should be included when doing an "get app structure" from script, most actions should not implement this.

The returns are:

`AppStructureInfo`
* `IsAppAction`: Tells the scenario to insert a "getappstructure" action after that action using data from sessionState.CurrentApp, should only be used for actions "opening" new apps such as `openapp`.
* `Include`: Tells if this action itself should be included. Set to false when action should not be included, but interface is implement to return a list of subactions (see e.g. `iterated` action). 

`[]Action`: List of sub actions to be evaluated for inclusion (see e.g. `iterated` action).

```golang
func (settings MyActionSettings) AppStructureAction() (*AppStructureInfo, []Action) {
	return &AppStructureInfo{
		IsAppAction: false,
		Include:     true,
	}, nil
}
```

### AffectsAppObjectsAction

AffectsAppObjectsAction is an interface that should be implemented by all actions that affect the availability of selectable objects for app structure consumption. App structure of the currentapp is passed as an argument. The returns are:

* `*config.AppStructurePopulatedObjects` - objects to be added to the selectable list by this action
* `[]string` - ids of objects that are removed (including any children) by this action
* `bool` - clears all objects except bookmarks and sheets

```golang
func (settings MyActionSettings) AffectsAppObjectsAction(structure appstructure.AppStructure) ([]*appstructure.AppStructurePopulatedObjects, []string, bool) {
    newObjs := settings.GetSelectableObjects(structure)
    return []*appstructure.AppStructurePopulatedObjects{newObjs}, nil, true
}
```

### IsActionValidForScheduler

This can be implemented on a action in a scenario to validate if scheduler type is allowed to use the action, implement this to give an error or warning when a scheduler type contains the action in it's scenario. Returns list of warnings and possible error.

For example, to only allow `MyActionSettings` to be run when using the scheduler `MyCustomScheduler`:

```golang
func (settings MyActionSettings)  IsActionValidForScheduler(string) ([]string, error) {
	if schedType == MyCustomScheduler {
		return nil, nil
	}
	return nil, errors.Errorf("scheduler type<%s> not compatible with MyAction action", schedType)
}
```
