package parser

import (
	"fmt"
	"minibank/internal/catalog"
)

type Parser struct {
	lexer     *Lexer
	curToken  Token
	peekToken Token
}

func NewParser(lexer *Lexer) *Parser {
	p := &Parser{lexer: lexer}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.lexer.NextToken()
}

func (p *Parser) Parse() (ASTNode, error) {
	switch p.curToken.Type {
	case TokenKeyword:
		switch p.curToken.Value {
		case "CREATE":
			return p.parseCreate()
		case "INSERT":
			return p.parseInsert()
		case "SELECT":
			return p.parseSelect()
		case "UPDATE":
			return p.parseUpdate()
		case "DELETE":
			return p.parseDelete()
		default:
			return nil, fmt.Errorf("unexpected token: %v", p.curToken)
		}
	default:
		return nil, fmt.Errorf("unexpected token: %v", p.curToken)
	}
}

// CREATE TABLE or CREATE INDEX
func (p *Parser) parseCreate() (ASTNode, error) {
	p.nextToken() // skip CREATE
	switch p.curToken.Value {
	case "TABLE":
		return p.parseCreateTable()
	case "INDEX":
		return p.parseCreateIndex()
	default:
		return nil, fmt.Errorf("expected TABLE or INDEX after CREATE")
	}
}

func (p *Parser) parseCreateTable() (*CreateTableStmt, error) {
	p.nextToken() // skip TABLE
	name := p.curToken.Value
	if p.curToken.Type != TokenIdentifier {
		return nil, fmt.Errorf("expected table name")
	}
	p.nextToken()

	if p.curToken.Value != "(" {
		return nil, fmt.Errorf("expected (")
	}
	p.nextToken()

	stmt := &CreateTableStmt{TableName: name}
	for p.curToken.Value != ")" {
		colName := p.curToken.Value
		if p.curToken.Type != TokenIdentifier {
			return nil, fmt.Errorf("expected column name")
		}
		p.nextToken()

		typeToken := p.curToken.Value
		var colType catalog.ColumnType
		switch typeToken {
		case "INT":
			colType = catalog.TypeInt
		case "STRING":
			colType = catalog.TypeString
		case "DECIMAL":
			colType = catalog.TypeDecimal
		case "BOOL":
			colType = catalog.TypeBool
		case "TIMESTAMP":
			colType = catalog.TypeTimestamp
		default:
			return nil, fmt.Errorf("unknown type: %s", typeToken)
		}
		p.nextToken()

		col := catalog.Column{Name: colName, Type: colType}

		// Constraints
		for p.curToken.Value == "PRIMARY" || p.curToken.Value == "UNIQUE" {
			switch p.curToken.Value {
			case "PRIMARY":
				p.nextToken()
				if p.curToken.Value != "KEY" {
					return nil, fmt.Errorf("expected KEY after PRIMARY")
				}
				p.nextToken()
				col.IsPrimary = true
			case "UNIQUE":
				p.nextToken()
				col.IsUnique = true
			}
		}

		stmt.Columns = append(stmt.Columns, col)

		if p.curToken.Value == "," {
			p.nextToken()
		}
	}
	p.nextToken()
	return stmt, nil
}

func (p *Parser) parseCreateIndex() (*CreateIndexStmt, error) {
	p.nextToken()
	idxName := p.curToken.Value
	p.nextToken()

	if p.curToken.Value != "ON" {
		return nil, fmt.Errorf("expected ON")
	}
	p.nextToken()

	tableName := p.curToken.Value
	p.nextToken()

	if p.curToken.Value != "(" {
		return nil, fmt.Errorf("expected (")
	}
	p.nextToken()

	colName := p.curToken.Value
	p.nextToken()

	if p.curToken.Value != ")" {
		return nil, fmt.Errorf("expected )")
	}
	p.nextToken()

	return &CreateIndexStmt{
		IndexName: idxName,
		TableName: tableName,
		Column:    colName,
	}, nil
}

func (p *Parser) parseInsert() (*InsertStmt, error) {
	p.nextToken()
	if p.curToken.Value != "INTO" {
		return nil, fmt.Errorf("expected INTO")
	}
	p.nextToken()
	tableName := p.curToken.Value
	p.nextToken()

	stmt := &InsertStmt{TableName: tableName}

	if p.curToken.Value == "(" {
		p.nextToken()
		for p.curToken.Value != ")" {
			stmt.Columns = append(stmt.Columns, p.curToken.Value)
			p.nextToken()
			if p.curToken.Value == "," {
				p.nextToken()
			}
		}
		p.nextToken()
	}

	if p.curToken.Value != "VALUES" {
		return nil, fmt.Errorf("expected VALUES")
	}
	p.nextToken()

	if p.curToken.Value != "(" {
		return nil, fmt.Errorf("expected (")
	}
	p.nextToken()

	for p.curToken.Value != ")" {
		// Literal
		val, err := p.parseLiteral()
		if err != nil {
			return nil, err
		}
		stmt.Values = append(stmt.Values, val)
		p.nextToken()
		if p.curToken.Value == "," {
			p.nextToken()
		}
	}
	p.nextToken()

	return stmt, nil
}

func (p *Parser) parseSelect() (*SelectStmt, error) {
	p.nextToken()
	stmt := &SelectStmt{}

	// Fields
	// Fields
	for p.curToken.Value != "FROM" {
		name := p.curToken.Value
		p.nextToken()
		if p.curToken.Value == "." {
			p.nextToken() // consume dot
			if p.curToken.Type != TokenIdentifier {
				return nil, fmt.Errorf("expected column name after dot")
			}
			name = name + "." + p.curToken.Value
			p.nextToken()
		}
		stmt.Fields = append(stmt.Fields, name)

		if p.curToken.Value == "," {
			p.nextToken()
		} else if p.curToken.Value != "FROM" {
			return nil, fmt.Errorf("expected comma or FROM, got %s", p.curToken.Value)
		}
	}

	p.nextToken()
	stmt.TableName = p.curToken.Value
	p.nextToken()

	// JOIN
	if p.curToken.Value == "JOIN" {
		p.nextToken()
		stmt.Join = &JoinClause{}
		stmt.Join.Table = p.curToken.Value
		p.nextToken()

		if p.curToken.Value != "ON" {
			return nil, fmt.Errorf("expected ON")
		}
		p.nextToken()

		left, err := p.parseSimpleExpr()
		if err != nil {
			return nil, err
		}
		op := Operator(p.curToken.Value)
		p.nextToken()
		right, err := p.parseSimpleExpr()
		if err != nil {
			return nil, err
		}
		stmt.Join.On = BinaryExpr{Left: left, Op: op, Right: right}
	}

	// WHERE
	if p.curToken.Value == "WHERE" {
		p.nextToken()
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		stmt.Where = &WhereClause{Expr: expr}
	}

	return stmt, nil
}

func (p *Parser) parseExpression() (Expression, error) {

	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	for p.curToken.Value == "AND" || p.curToken.Value == "OR" {
		op := Operator(p.curToken.Value)
		p.nextToken()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Left: left, Op: op, Right: right}
	}
	return left, nil
}

func (p *Parser) parseTerm() (Expression, error) {
	left, err := p.parseSimpleExpr()
	if err != nil {
		return nil, err
	}

	if isOperator(p.curToken.Value) {
		op := Operator(p.curToken.Value)
		p.nextToken()
		right, err := p.parseSimpleExpr()
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Left: left, Op: op, Right: right}, nil
	}

	return left, nil
}

func (p *Parser) parseSimpleExpr() (Expression, error) {
	switch p.curToken.Type {
	case TokenIdentifier:
		name := p.curToken.Value
		p.nextToken()
		if p.curToken.Value == "." {
			p.nextToken()
			if p.curToken.Type != TokenIdentifier {
				return nil, fmt.Errorf("expected column name after .")
			}
			name = name + "." + p.curToken.Value
			p.nextToken()
		}
		return &IdentifierExpr{Name: name}, nil
	case TokenString:
		val := p.curToken.Value
		p.nextToken()
		return &LiteralExpr{Value: val}, nil
	case TokenNumber:
		val := p.curToken.Value
		p.nextToken()
		return &LiteralExpr{Value: RawNumber(val)}, nil
	default:
		return nil, fmt.Errorf("unexpected token in expression: %v", p.curToken)
	}
}

func (p *Parser) parseLiteral() (interface{}, error) {
	switch p.curToken.Type {
	case TokenString:
		return p.curToken.Value, nil
	case TokenNumber:
		return RawNumber(p.curToken.Value), nil
	}
	return nil, fmt.Errorf("expected literal")
}

func (p *Parser) parseUpdate() (*UpdateStmt, error) {
	p.nextToken()
	stmt := &UpdateStmt{SetPairs: make(map[string]interface{})}
	stmt.TableName = p.curToken.Value
	p.nextToken()

	if p.curToken.Value != "SET" {
		return nil, fmt.Errorf("expected SET")
	}
	p.nextToken()

	for {
		col := p.curToken.Value
		p.nextToken()
		if p.curToken.Value != "=" {
			return nil, fmt.Errorf("expected =")
		}
		p.nextToken()
		val, _ := p.parseLiteral()
		stmt.SetPairs[col] = val
		p.nextToken()

		if p.curToken.Value != "," {
			break
		}
		p.nextToken()
	}

	if p.curToken.Value == "WHERE" {
		p.nextToken()
		expr, _ := p.parseExpression()
		stmt.Where = &WhereClause{Expr: expr}
	}
	return stmt, nil
}

func (p *Parser) parseDelete() (*DeleteStmt, error) {
	p.nextToken()
	if p.curToken.Value != "FROM" {
		return nil, fmt.Errorf("expected FROM")
	}
	p.nextToken()

	stmt := &DeleteStmt{TableName: p.curToken.Value}
	p.nextToken()

	if p.curToken.Value == "WHERE" {
		p.nextToken()
		expr, _ := p.parseExpression()
		stmt.Where = &WhereClause{Expr: expr}
	}
	return stmt, nil
}

func isOperator(s string) bool {
	return s == "=" || s == "!=" || s == "<" || s == ">" || s == "<=" || s == ">="
}
