## ClickActionButton action

A `ClickActionButton`-action simulates clicking an _action-button_. An _action-button_ is a chart item which, when clicked, executes a series of actions. The series of actions contained by an action-button begins with any number _generic button-actions_ and ends with an optional _navigation button-action_.

**Note:** Clicking an action-button may have side effects such as changing sheet and locking selections, which highly affect the outcome of following actions.

### Supported button-actions
#### Generic button-actions
- Apply bookmark
- Move backward in all selections
- Move forward in all selections
- Lock all selections
- Clear all selections
- Lock field
- Unlock field
- Select all in field
- Select alternatives in field
- Select excluded in field
- Select possible in field
- Select values matching search criteria in field
- Clear selection in field
- Toggle selection in field
- Set value of variable

#### Navigation button-actions
- Change to first sheet
- Change to last sheet
- Change to previous sheet
- Change sheet by name
- Change sheet by ID