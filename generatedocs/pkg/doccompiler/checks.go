package doccompiler

import (
	"github.com/pkg/errors"
	"github.com/qlik-oss/gopherciser/generatedocs/pkg/common"
)

type check func(data *docData) []error

var checks = []check{
	checkAllActionsDocumented,
	checkAllConfigFieldsDocumented,
	checkAllActionsInGroup,
}

// func checkAndWarn(data *docData) {
// 	for _, finding := range checkAll(data) {
// 		fmt.Printf("WARNING: %v\n", finding)
// 	}
// }

func checkAll(data *docData) []error {
	findings := []error{}
	for _, check := range checks {
		findings = append(findings, check(data)...)
	}
	return findings
}

func checkAllActionsDocumented(data *docData) []error {
	findings := []error{}
	allActions := common.ActionStrings()
	for _, a := range allActions {
		docEntry, ok := data.ActionMap[a]
		if !ok {
			findings = append(findings, errors.Errorf(`action "%s" is not documented`, a))
			continue
		}

		if docEntry.Description == "" {
			findings = append(findings, errors.Errorf(`action "%s" has no description`, a))
		}
		if docEntry.Examples == "" {
			findings = append(findings, errors.Errorf(`action "%s" has no examples`, a))
		}
	}
	return findings
}

func checkAllConfigFieldsDocumented(data *docData) []error {
	findings := []error{}
	// Get all config fields
	expectedConfigFields, err := common.FieldsString()
	if err != nil {
		common.Exit(err, ExitCodeFailedConfigFields)
	}
	// Add documentation wrapping entire document as "main" entry into config map
	expectedConfigFields = append(expectedConfigFields, "main")
	for _, field := range expectedConfigFields {
		docEntry, ok := data.ConfigMap[field]
		if !ok {
			findings = append(findings, errors.Errorf(`config field "%s" is not documented`, field))
			continue
		}

		if docEntry.Description == "" {
			findings = append(findings, errors.Errorf(`config field "%s" has no description`, field))
		}
		if docEntry.Examples == "" {
			findings = append(findings, errors.Errorf(`config field "%s" has no examples`, field))
		}
	}
	return findings
}

func checkAllActionsInGroup(data *docData) []error {
	findings := []error{}

	// map actions to groups
	actionToGroups := map[string][]string{}
	for _, action := range data.Actions {
		actionToGroups[action] = []string{}
	}
	for _, group := range data.Groups {
		for _, action := range group.Actions {
			actionToGroups[action] = append(actionToGroups[action], group.Name)
		}
	}

	// check action belongs to one and only one group
	for action, groups := range actionToGroups {
		lenGroups := len(groups)
		switch {
		case lenGroups == 0:
			findings = append(findings, errors.Errorf(`action "%s" does not belong to a group`, action))
		case lenGroups > 1:
			findings = append(findings, errors.Errorf(`action "%s" belong to %d groups %v`, action, lenGroups, groups))
		}
	}

	return findings
}
