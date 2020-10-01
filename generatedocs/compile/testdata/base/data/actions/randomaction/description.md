## RandomAction action

Randomly select other actions to perform. This meta-action can be used as a starting point for your testing efforts, to simplify script authoring or to add background load.

`randomaction` accepts a list of action types between which to randomize. An execution of `randomaction` executes one or more of the listed actions (as determined by the `iterations` parameter), randomly chosen by a weighted probability. If nothing else is specified, each action has a default random mode that is used. An override is done by specifying one or more parameters of the original action.

Each action executed by `randomaction` is followed by a customizable `thinktime`.

**Note:** The recommended way to use this action is to prepend it with an `openapp` and a `changesheet` action as this ensures that a sheet is always in context.
