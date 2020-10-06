package helpers

import (
	"github.com/pkg/errors"
	"strings"
)

// HasDeprecatedFields check json if keys exist at provided paths, if any exist report error
func HasDeprecatedFields(rawJson []byte, deprecatedPaths []string) error {
	hasPaths := make([]string, 0, len(deprecatedPaths))
	for _, path := range deprecatedPaths {
		dp := DataPath(path)
		if _, err := dp.Lookup(rawJson); err == nil {
			hasPaths = append(hasPaths, path)
		}
	}
	if len(hasPaths) > 0 {
		return errors.Errorf("has deprecated fields: %s", strings.Join(hasPaths, ","))
	}
	return nil
}
