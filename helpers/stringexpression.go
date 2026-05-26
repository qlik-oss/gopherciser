package helpers

import (
	"strings"

	"github.com/goccy/go-json"
	"github.com/qlik-oss/enigma-go/v4"
)

type (
	stringExpression struct {
		Expression *enigma.StringExpression `json:"qStringExpression"`
	}

	StringExpression string
)

// unmarshals quoted(!) string or string expression object
func (expr *StringExpression) UnmarshalJSON(arg []byte) error {
	if expr == nil || len(arg) < 1 {
		return nil
	}
	switch arg[0] {
	case '{':
		var se stringExpression
		if err := json.Unmarshal(arg, &se); err != nil {
			*expr = StringExpression(arg) // fallback
		}
		*expr = StringExpression(se.Expression.Expr)
	case '"':
		*expr = StringExpression(strings.TrimPrefix(strings.TrimSuffix(string(arg), `"`), `"`))
	default:
		*expr = StringExpression(arg)
	}
	return nil
}

func (expr StringExpression) MarshalJSON() ([]byte, error) {
	if len(expr) < 1 {
		return []byte{}, nil
	}

	switch []rune(expr)[0] {
	case '=', '\'':
		return json.Marshal(stringExpression{
			Expression: &enigma.StringExpression{
				Expr: string(expr),
			},
		})
	default:
		return []byte(expr), nil
	}
}

// // MarshalJSON marshal HttpMethod
// func (method HttpMethod) MarshalJSON() ([]byte, error) {
// 	str, err := httpMethodEnum.String(int(method))
// 	if err != nil {
// 		return nil, errors.Errorf("Unknown HttpMethod<%d>", method)
// 	}
// 	return []byte(fmt.Sprintf(`"%s"`, str)), nil
// }
