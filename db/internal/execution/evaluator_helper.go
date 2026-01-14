package execution

import (
	"fmt"
	"math/big"
	"minibank/internal/parser"
)

func cmpRat(lRat, rRat *big.Rat, op parser.Operator) (bool, error) {
	cmp := lRat.Cmp(rRat)
	switch op {
	case parser.OpEq:
		return cmp == 0, nil
	case parser.OpNeq:
		return cmp != 0, nil
	case parser.OpLt:
		return cmp < 0, nil
	case parser.OpGt:
		return cmp > 0, nil
	case parser.OpLte:
		return cmp <= 0, nil
	case parser.OpGte:
		return cmp >= 0, nil
	}
	return false, fmt.Errorf("invalid operator for numeric comparison: %s", op)
}
