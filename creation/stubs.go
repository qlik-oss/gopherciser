package creation

import "github.com/qlik-oss/enigma-go"

// StubMetaDef creates a MetaDef template to be used when creating objects
func StubMetaDef(title string, description string) map[string]interface{} {
	metaDef := make(map[string]interface{})
	if description != "" {
		metaDef["description"] = description
	}
	metaDef["title"] = title
	return metaDef
}

// StubNxInfo creates a NxInfo template to be used when creating objects
func StubNxInfo(typeString string) enigma.NxInfo {
	return enigma.NxInfo{
		Type: typeString,
	}
}
