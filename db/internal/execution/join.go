package execution

import (
	"minibank/internal/catalog"
	"minibank/internal/parser"
	"minibank/internal/storage"
)

type NestedLoopJoin struct {
	Left   Iterator
	Right  Iterator
	On     parser.BinaryExpr
	schema []catalog.Column

	outerTuple *storage.Tuple
}

func NewNestedLoopJoin(left, right Iterator, on parser.BinaryExpr) *NestedLoopJoin {
	schema := append(left.Schema(), right.Schema()...)
	return &NestedLoopJoin{
		Left:   left,
		Right:  right,
		On:     on,
		schema: schema,
	}
}

func (op *NestedLoopJoin) Open() error {
	if err := op.Left.Open(); err != nil {
		return err
	}
	if err := op.Right.Open(); err != nil {
		return err
	}

	var err error
	op.outerTuple, err = op.Left.Next()
	if err != nil {
		return err
	}
	return nil
}

func (op *NestedLoopJoin) Next() (*storage.Tuple, error) {

	for op.outerTuple != nil {
		rightTuple, err := op.Right.Next()
		if err != nil {
			return nil, err
		}

		if rightTuple == nil {
			op.outerTuple, err = op.Left.Next()
			if err != nil {
				return nil, err
			}
			if op.outerTuple == nil {
				return nil, nil
			}
			// Reset right
			op.Right.Close()
			op.Right.Open()
			continue
		}

		combined := &storage.Tuple{Cells: append(op.outerTuple.Cells, rightTuple.Cells...)}

		match, err := Evaluate(combined, &op.On, op.schema)
		if err != nil {
			return nil, err
		}

		if match {
			return combined, nil
		}
	}
	return nil, nil
}

func (op *NestedLoopJoin) Close() error {
	op.Left.Close()
	op.Right.Close()
	return nil
}

func (op *NestedLoopJoin) Schema() []catalog.Column {
	return op.schema
}
