package helpers

import (
	"fmt"
	"strings"
)

// Params map with URL parameters
type Params map[string]string

// String returns parameters as string with the format ?param1=val1&param2=val2, parameter order is not guaranteed
func (params Params) String() string {
	if len(params) < 1 {
		return ""
	}

	p := make([]string, 0, len(params))

	for k, v := range params {
		p = append(p, fmt.Sprintf("%s=%s", k, v))
	}

	return fmt.Sprintf("?%s", strings.Join(p, "&"))
}
