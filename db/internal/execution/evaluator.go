package execution

import (
	"fmt"
	"math/big"
	"minibank/internal/catalog"
	"minibank/internal/errors"
	"minibank/internal/parser"
	"minibank/internal/storage"
	"strings"
)

func Evaluate(t *storage.Tuple, expr parser.Expression, schema []catalog.Column) (bool, error) {
	val, err := evalExpr(t, expr, schema)
	if err != nil {
		return false, err
	}
	b, ok := val.(bool)
	if !ok {
		return false, errors.New(errors.ErrTypeMismatch,
			fmt.Sprintf("WHERE clause must evaluate to BOOL, got %T", val),
			"Use comparison operators (=, <, >) or logical operators (AND, OR) to return a boolean.")
	}
	return b, nil
}

func evalExpr(t *storage.Tuple, expr parser.Expression, schema []catalog.Column) (interface{}, error) {
	switch e := expr.(type) {
	case *parser.BinaryExpr:
		left, err := evalExpr(t, e.Left, schema)
		if err != nil {
			return nil, err
		}
		right, err := evalExpr(t, e.Right, schema)
		if err != nil {
			return nil, err
		}
		return compare(left, right, e.Op)
	case *parser.LiteralExpr:
		return e.Value, nil
	case *parser.IdentifierExpr:
		for i, col := range schema {
			if strings.Contains(e.Name, ".") {
				fqn := col.TableName + "." + col.Name
				if strings.EqualFold(fqn, e.Name) {
					return t.Cells[i].Value, nil
				}
			} else {
				if strings.EqualFold(col.Name, e.Name) {
					return t.Cells[i].Value, nil
				}
			}
		}
		return nil, fmt.Errorf("column %s not found", e.Name)
	}
	return nil, fmt.Errorf("unknown expression type")
}

func compare(left, right interface{}, op parser.Operator) (bool, error) {
	if isNumeric(left) && isNumeric(right) {
		lRat, err := toRat(left)
		if err != nil {
			return false, err
		}
		rRat, err := toRat(right)
		if err != nil {
			return false, err
		}
		return cmpRat(lRat, rRat, op)
	}

	// Mixed Numeric/String (e.g. Decimal as string vs RawNumber)
	if (isNumeric(left) && isString(right)) || (isString(left) && isNumeric(right)) {
		lRat, err1 := toRat(left)
		rRat, err2 := toRat(right)
		if err1 == nil && err2 == nil {
			return cmpRat(lRat, rRat, op)
		}
		// If conversion fails, fall through to type mismatch (or could wrap error)
	}

	if l, ok := left.(string); ok {
		if r, ok := right.(string); ok {
			switch op {
			case parser.OpEq:
				return l == r, nil
			case parser.OpNeq:
				return l != r, nil
			case parser.OpLt:
				return l < r, nil
			case parser.OpGt:
				return l > r, nil
			case parser.OpLte:
				return l <= r, nil
			case parser.OpGte:
				return l >= r, nil
			}
			return false, fmt.Errorf("invalid operator for string comparison: %s", op)
		}
	}

	if l, ok := left.(bool); ok {
		if r, ok := right.(bool); ok {
			switch op {
			case parser.OpEq:
				return l == r, nil
			case parser.OpNeq:
				return l != r, nil
			case parser.OpAnd:
				return l && r, nil
			case parser.OpOr:
				return l || r, nil
			}
			return false, fmt.Errorf("invalid operator for boolean comparison: %s", op)
		}
	}

	return false, fmt.Errorf("type mismatch or unsupported comparison: %T %s %T", left, op, right)
}

func isNumeric(v interface{}) bool {
	switch v.(type) {
	case int, int64, float64, parser.RawNumber:
		return true
	}
	return false
}

func isString(v interface{}) bool {
	_, ok := v.(string)
	return ok
}

func toRat(v interface{}) (*big.Rat, error) {
	switch val := v.(type) {
	case int:
		return big.NewRat(int64(val), 1), nil
	case int64:
		return big.NewRat(val, 1), nil
	case float64:
		return new(big.Rat).SetFloat64(val), nil
	case parser.RawNumber:
		r, ok := new(big.Rat).SetString(string(val))
		if !ok {
			return nil, fmt.Errorf("invalid number format: %s", val)
		}
		return r, nil
	case string:
		r, ok := new(big.Rat).SetString(val)
		if !ok {
			return nil, fmt.Errorf("invalid number format: %s", val)
		}
		return r, nil
	}
	return nil, fmt.Errorf("cannot convert %T to key", v)
}
