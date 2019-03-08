package query

import (
	"errors"
	"fmt"
	"io"
)

type Statement interface {
	kind() string
	String() string
}

type AndStatement struct {
	Left, Right Statement
}

func (*AndStatement) kind() string {
	return "AndStatement"
}

func (s *AndStatement) String() string {
	return "(" + s.Left.String() + " AND " + s.Right.String() + ")"
}

type OrStatement struct {
	Left, Right Statement
}

func (*OrStatement) kind() string {
	return "OrStatement"
}

func (s *OrStatement) String() string {
	return "(" + s.Left.String() + " OR " + s.Right.String() + ")"
}

type NotStatement struct {
	S Statement
}

func (*NotStatement) kind() string {
	return "NotStatement"
}

func (s *NotStatement) String() string {
	return "NOT(" + s.S.String() + ")"
}

type Expression struct {
	Value string
}

func (*Expression) kind() string {
	return "Expression"
}

func (e *Expression) String() string {
	return e.Value
}

type Visitor interface {
	Visit(Statement) Visitor
}

func Walk(v Visitor, stmt Statement) {
	if v = v.Visit(stmt); v == nil {
		return
	}
	switch n := stmt.(type) {
	case *AndStatement:
		Walk(v, n.Left)
		Walk(v, n.Right)
	case *OrStatement:
		Walk(v, n.Left)
		Walk(v, n.Right)
	case *NotStatement:
		Walk(v, n.S)
	case *Expression:
	default:
		panic(fmt.Sprintf("query.Walk: unexpected node type %T", n))

	}
	v.Visit(nil)
}

type inspector func(Statement) bool

func (f inspector) Visit(node Statement) Visitor {
	if f(node) {
		return f
	}
	return nil
}

// Inspect traverses an AST in depth-first order: It starts by calling
// f(node); node must not be nil. If f returns true, Inspect invokes f
// recursively for each of the non-nil children of node, followed by a
// call of f(nil).
//
func Inspect(node Statement, f func(Statement) bool) {
	Walk(inspector(f), node)
}

// Parser represents a parser.
type Parser struct {
	s   *Scanner
	buf struct {
		tok       Token  // last read token
		lit       string // last read literal
		haveToken bool   // Whether or not the buffer has a token to read.
	}
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	return &Parser{s: NewScanner(r)}
}

// scan returns the next token from the underlying scanner.
// If a token has been unscanned then read that instead.
func (p *Parser) scan() (tok Token, lit string) {
	// If we have a token on the buffer, then return it.
	if p.buf.haveToken {
		p.buf.haveToken = false
		return p.buf.tok, p.buf.lit
	}

	// Otherwise read the next token from the scanner.
	tok, lit = p.s.Scan()

	// Save it to the buffer in case we unscan later.
	p.buf.tok, p.buf.lit = tok, lit

	return
}

// unscan pushes the previously read token back onto the buffer.
func (p *Parser) unscan() { p.buf.haveToken = true }

// scanIgnoreWhitespace scans the next non-whitespace token.
func (p *Parser) scanIgnoreWhitespace() (tok Token, lit string) {
	tok, lit = p.scan()
	if tok == WS {
		tok, lit = p.scan()
	}
	return
}

func (p *Parser) Parse() (Statement, error) {
	type data struct {
		lit string
		tok Token
	}

	var exprQueue []data
	var opStack []data
	var statementStack []Statement
	for tok, lit := p.scanIgnoreWhitespace(); tok != EOF; tok, lit = p.scanIgnoreWhitespace() {
		switch tok {
		case IDENT:
			exprQueue = append(exprQueue, data{lit, tok})
		case NOT:
			opStack = append(opStack, data{lit, tok})
		case AND, OR:
			for len(opStack) != 0 && ((opStack[len(opStack)-1].tok.isFunction() || opStack[len(opStack)-1].tok.isOperator()) && opStack[len(opStack)-1].tok != LEFT_PAREN) {
				var op data
				op, opStack = opStack[len(opStack)-1], opStack[:len(opStack)-1]
				exprQueue = append(exprQueue, op)
			}
			opStack = append(opStack, data{lit, tok})
		case LEFT_PAREN:
			opStack = append(opStack, data{lit, tok})
		case RIGHT_PAREN:
			for len(opStack) != 0 && opStack[len(opStack)-1].tok != LEFT_PAREN {
				var op data
				op, opStack = opStack[len(opStack)-1], opStack[:len(opStack)-1]
				exprQueue = append(exprQueue, op)
			}
			if len(opStack) != 0 && opStack[len(opStack)-1].tok == LEFT_PAREN {
				_, opStack = opStack[len(opStack)-1], opStack[:len(opStack)-1]
			} else if len(opStack) == 0 {
				return nil, errors.New("unmatched parens right paren")
			}
		}
	}
	for len(opStack) > 0 {
		var op data
		op, opStack = opStack[len(opStack)-1], opStack[:len(opStack)-1]
		if op.lit == "(" {
			return nil, errors.New("mismatched parentheses found")
		}
		exprQueue = append(exprQueue, op)
	}

	for _, expr := range exprQueue {
		switch expr.tok {
		case IDENT:
			statementStack = append(statementStack, &Expression{Value: expr.lit})
		case NOT:
			var inner Statement
			inner, statementStack = statementStack[len(statementStack)-1], statementStack[:len(statementStack)-1]
			not := &NotStatement{S: inner}
			statementStack = append(statementStack, not)
		case AND:
			var left, right Statement
			right, left, statementStack = statementStack[len(statementStack)-1], statementStack[len(statementStack)-2], statementStack[:len(statementStack)-2]
			and := &AndStatement{Left: left, Right: right}
			statementStack = append(statementStack, and)
		case OR:
			var left, right Statement
			right, left, statementStack = statementStack[len(statementStack)-1], statementStack[len(statementStack)-2], statementStack[:len(statementStack)-2]
			and := &OrStatement{Left: left, Right: right}
			statementStack = append(statementStack, and)
		default:
			return nil, fmt.Errorf("Unknown token type: %d - %s", expr.tok, expr.lit)
		}
	}

	if len(statementStack) != 1 {
		return nil, fmt.Errorf("Expected only 1 statement, but stack was %v", statementStack)
	}

	return statementStack[0], nil
}
