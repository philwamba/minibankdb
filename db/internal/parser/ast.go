package parser

import "minibank/internal/catalog"

type NodeType int

const (
	NodeSelect NodeType = iota
	NodeInsert
	NodeCreateTable
	NodeUpdate
	NodeDelete
	NodeCreateIndex
)

type RawNumber string

type ASTNode interface {
	Type() NodeType
}

type CreateTableStmt struct {
	TableName string
	Columns   []catalog.Column
}

func (n *CreateTableStmt) Type() NodeType { return NodeCreateTable }

type InsertStmt struct {
	TableName string
	Columns   []string
	Values    []interface{}
}

func (n *InsertStmt) Type() NodeType { return NodeInsert }

type SelectStmt struct {
	TableName string
	Fields    []string
	Where     *WhereClause
	Join      *JoinClause
}

func (n *SelectStmt) Type() NodeType { return NodeSelect }

type UpdateStmt struct {
	TableName string
	SetPairs  map[string]interface{}
	Where     *WhereClause
}

func (n *UpdateStmt) Type() NodeType { return NodeUpdate }

type DeleteStmt struct {
	TableName string
	Where     *WhereClause
}

func (n *DeleteStmt) Type() NodeType { return NodeDelete }

type CreateIndexStmt struct {
	IndexName string
	TableName string
	Column    string
}

func (n *CreateIndexStmt) Type() NodeType { return NodeCreateIndex }

type WhereClause struct {
	Left  string
	Op    Operator
	Right interface{}
	Expr  Expression
}

type JoinClause struct {
	Table string
	On    BinaryExpr
}

type ExprType int

const (
	ExprBinary ExprType = iota
	ExprLiteral
	ExprIdentifier
)

type Expression interface {
	ExprType() ExprType
}

type Operator string

const (
	OpEq  Operator = "="
	OpNeq Operator = "!="
	OpLt  Operator = "<"
	OpGte Operator = ">="
	OpGt  Operator = ">"
	OpLte Operator = "<="
	OpAnd Operator = "AND"
	OpOr  Operator = "OR"
)

type BinaryExpr struct {
	Left  Expression
	Op    Operator
	Right Expression
}

func (b *BinaryExpr) ExprType() ExprType { return ExprBinary }

type LiteralExpr struct {
	Value interface{}
}

func (l *LiteralExpr) ExprType() ExprType { return ExprLiteral }

type IdentifierExpr struct {
	Name string
}

func (i *IdentifierExpr) ExprType() ExprType { return ExprIdentifier }
