package simpl

import "fmt"

type parser struct {
	tokens []Token
	idx    int
	diags  []Diagnostic
}

func parseProgram(tokens []Token) (*Program, []Diagnostic) {
	p := &parser{
		tokens: tokens,
		diags:  make([]Diagnostic, 0, 8),
	}

	prog := &Program{Statements: make([]Stmt, 0, 32)}
	for !p.isAtEnd() {
		stmt := p.parseStatement()
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
			continue
		}
		p.synchronize()
	}

	return prog, p.diags
}

func (p *parser) parseStatement() Stmt {
	switch p.peek().Type {
	case TokenVar:
		return p.parseDecl(false)
	case TokenConst:
		return p.parseDecl(true)
	case TokenRead:
		return p.parseRead()
	case TokenWrite:
		return p.parseWrite()
	case TokenPop:
		return p.parsePop()
	case TokenIf:
		return p.parseIf()
	case TokenWhile:
		return p.parseWhile()
	case TokenFor:
		return p.parseFor()
	case TokenLBrace:
		return p.parseBlock()
	case TokenIdentifier:
		return p.parseAssignment()
	case TokenEOF:
		return nil
	default:
		tok := p.peek()
		p.addDiag("PARSE_UNEXPECTED_TOKEN", fmt.Sprintf("Unexpected token '%s'.", tok.Type.String()), tok.Pos, "Start a statement with var/const/read/write/pop/if/while/for or an assignment.")
		p.advance()
		return nil
	}
}

func (p *parser) parseDecl(isConst bool) Stmt {
	kw := p.advance()
	nameTok, ok := p.consume(TokenIdentifier, "PARSE_EXPECTED_TOKEN", "Expected identifier after declaration keyword.", "Example: var count int")
	if !ok {
		return nil
	}

	typ, ok := p.parseType()
	if !ok {
		return nil
	}

	var init Expr
	if p.match(TokenAssign) {
		init = p.parseExpression()
	}

	if isConst && init == nil {
		p.addDiag("PARSE_EXPECTED_TOKEN", "Const declaration requires an initializer.", nameTok.Pos, "Example: const limit int = 10")
		return nil
	}

	return &DeclStmt{
		Pos:         kw.Pos,
		Name:        nameTok.Lexeme,
		DeclaredTyp: typ,
		Initializer: init,
		Const:       isConst,
	}
}

func (p *parser) parseType() (Type, bool) {
	tok := p.peek()
	switch tok.Type {
	case TokenTypeInt:
		p.advance()
		return Type{Kind: TypeInt}, true
	case TokenTypeFloat:
		p.advance()
		return Type{Kind: TypeFloat}, true
	case TokenTypeBool:
		p.advance()
		return Type{Kind: TypeBool}, true
	case TokenTypeString:
		p.advance()
		return Type{Kind: TypeString}, true
	case TokenTypeArray:
		p.advance()
		if _, ok := p.consume(TokenLBracket, "PARSE_EXPECTED_TOKEN", "Expected '[' after array type.", "Use array[int], array[string], etc."); !ok {
			return invalidType(), false
		}
		elem, ok := p.parseType()
		if !ok {
			return invalidType(), false
		}
		if _, ok := p.consume(TokenRBracket, "PARSE_EXPECTED_TOKEN", "Expected closing ']' for array type.", "Close array type with ']'."); !ok {
			return invalidType(), false
		}
		return Type{Kind: TypeArray, Elem: &elem}, true
	default:
		p.addDiag("PARSE_EXPECTED_TOKEN", "Expected a type name.", tok.Pos, "Allowed types: int, float, bool, string, array[...].")
		return invalidType(), false
	}
}

func (p *parser) parseRead() Stmt {
	kw := p.advance()
	ident, ok := p.consume(TokenIdentifier, "PARSE_EXPECTED_TOKEN", "Expected variable name after read.", "Example: read value")
	if !ok {
		return nil
	}
	return &ReadStmt{Pos: kw.Pos, Name: ident.Lexeme}
}

func (p *parser) parseWrite() Stmt {
	kw := p.advance()
	values := make([]Expr, 0, 4)
	values = append(values, p.parseExpression())
	for p.match(TokenComma) {
		values = append(values, p.parseExpression())
	}
	return &WriteStmt{Pos: kw.Pos, Values: values}
}

func (p *parser) parsePop() Stmt {
	kw := p.advance()
	target := p.parseAssignmentTarget()
	if target == nil {
		return nil
	}
	return &PopStmt{Pos: kw.Pos, Target: target}
}

func (p *parser) parseIf() Stmt {
	kw := p.advance()
	cond := p.parseExpression()
	thenBlock := p.parseBlock()
	if thenBlock == nil {
		return nil
	}

	stmt := &IfStmt{
		Pos: kw.Pos,
		Primary: IfClause{
			Pos:       kw.Pos,
			Condition: cond,
			Block:     thenBlock,
		},
		ElseIfs: make([]IfClause, 0),
	}

	for p.match(TokenElse) {
		elseTok := p.previous()
		if p.match(TokenIf) {
			ifTok := p.previous()
			eCond := p.parseExpression()
			eBlock := p.parseBlock()
			if eBlock == nil {
				return nil
			}
			stmt.ElseIfs = append(stmt.ElseIfs, IfClause{Pos: ifTok.Pos, Condition: eCond, Block: eBlock})
			continue
		}
		b := p.parseBlock()
		if b == nil {
			p.addDiag("PARSE_EXPECTED_TOKEN", "Expected block after else.", elseTok.Pos, "Use: else { ... }")
			return nil
		}
		stmt.ElsePart = b
		break
	}
	return stmt
}

func (p *parser) parseWhile() Stmt {
	kw := p.advance()
	cond := p.parseExpression()
	body := p.parseBlock()
	if body == nil {
		return nil
	}
	return &WhileStmt{Pos: kw.Pos, Condition: cond, Body: body}
}

func (p *parser) parseFor() Stmt {
	kw := p.advance()
	nameTok, ok := p.consume(TokenIdentifier, "PARSE_EXPECTED_TOKEN", "Expected loop variable name after for.", "Example: for i from 0 until 10 step 1 { ... }")
	if !ok {
		return nil
	}
	if _, ok := p.consume(TokenFrom, "PARSE_EXPECTED_TOKEN", "Expected 'from' in for loop.", "Use: for i from A until B step S { ... }"); !ok {
		return nil
	}
	from := p.parseExpression()
	if _, ok := p.consume(TokenUntil, "PARSE_EXPECTED_TOKEN", "Expected 'until' in for loop.", "Use: for i from A until B step S { ... }"); !ok {
		return nil
	}
	until := p.parseExpression()
	if _, ok := p.consume(TokenStep, "PARSE_EXPECTED_TOKEN", "Expected 'step' in for loop.", "Use: for i from A until B step S { ... }"); !ok {
		return nil
	}
	step := p.parseExpression()
	body := p.parseBlock()
	if body == nil {
		return nil
	}
	return &ForStmt{Pos: kw.Pos, VarName: nameTok.Lexeme, From: from, Until: until, Step: step, Body: body}
}

func (p *parser) parseBlock() *BlockStmt {
	open, ok := p.consume(TokenLBrace, "PARSE_EXPECTED_TOKEN", "Expected '{' to start block.", "Use braces around block statements.")
	if !ok {
		return nil
	}
	b := &BlockStmt{Pos: open.Pos, Statements: make([]Stmt, 0, 8)}
	for !p.check(TokenRBrace) && !p.isAtEnd() {
		stmt := p.parseStatement()
		if stmt != nil {
			b.Statements = append(b.Statements, stmt)
			continue
		}
		p.synchronize()
	}
	if _, ok := p.consume(TokenRBrace, "PARSE_EXPECTED_TOKEN", "Expected '}' to close block.", "Close the block with '}'."); !ok {
		return nil
	}
	return b
}

func (p *parser) parseAssignment() Stmt {
	start := p.peek()
	target := p.parseAssignmentTarget()
	if target == nil {
		return nil
	}
	if !p.match(TokenAssign) {
		p.addDiag("PARSE_EXPECTED_TOKEN", "Expected '=' in assignment.", start.Pos, "Use: name = expression")
		return nil
	}
	value := p.parseExpression()
	return &AssignStmt{Pos: start.Pos, Target: target, Value: value}
}

func (p *parser) parseAssignmentTarget() Expr {
	name, ok := p.consume(TokenIdentifier, "PARSE_EXPECTED_TOKEN", "Expected variable name.", "Use a variable or indexed target like name[0].")
	if !ok {
		return nil
	}
	var expr Expr = &IdentifierExpr{Pos: name.Pos, Name: name.Lexeme}
	for p.match(TokenLBracket) {
		idxExpr := p.parseExpression()
		if _, ok := p.consume(TokenRBracket, "PARSE_EXPECTED_TOKEN", "Expected closing ']' in index expression.", "Close index expression with ']'."); !ok {
			return nil
		}
		expr = &IndexExpr{Pos: expr.Position(), Collection: expr, Index: idxExpr}
	}
	return expr
}

func (p *parser) parseExpression() Expr {
	return p.parseEquality()
}

func (p *parser) parseEquality() Expr {
	expr := p.parseComparison()
	for p.match(TokenEqual, TokenNotEqual) {
		op := p.previous()
		right := p.parseComparison()
		expr = &BinaryExpr{Pos: op.Pos, Left: expr, Operator: op.Type, Right: right}
	}
	return expr
}

func (p *parser) parseComparison() Expr {
	expr := p.parseTerm()
	for p.match(TokenGreater, TokenGreaterEq, TokenLess, TokenLessEq) {
		op := p.previous()
		right := p.parseTerm()
		expr = &BinaryExpr{Pos: op.Pos, Left: expr, Operator: op.Type, Right: right}
	}
	return expr
}

func (p *parser) parseTerm() Expr {
	expr := p.parseFactor()
	for p.match(TokenPlus, TokenMinus) {
		op := p.previous()
		right := p.parseFactor()
		expr = &BinaryExpr{Pos: op.Pos, Left: expr, Operator: op.Type, Right: right}
	}
	return expr
}

func (p *parser) parseFactor() Expr {
	expr := p.parseUnary()
	for p.match(TokenStar, TokenSlash, TokenPercent) {
		op := p.previous()
		right := p.parseUnary()
		expr = &BinaryExpr{Pos: op.Pos, Left: expr, Operator: op.Type, Right: right}
	}
	return expr
}

func (p *parser) parseUnary() Expr {
	if p.match(TokenSize) {
		op := p.previous()
		value := p.parseUnary()
		return &SizeExpr{Pos: op.Pos, Value: value}
	}
	if p.match(TokenMinus, TokenBang) {
		op := p.previous()
		right := p.parseUnary()
		return &UnaryExpr{Pos: op.Pos, Operator: op.Type, Right: right}
	}
	return p.parsePostfix()
}

func (p *parser) parsePostfix() Expr {
	expr := p.parsePrimary()
	for p.match(TokenLBracket) {
		idx := p.parseExpression()
		if _, ok := p.consume(TokenRBracket, "PARSE_EXPECTED_TOKEN", "Expected closing ']' in index expression.", "Close index expression with ']'."); !ok {
			return expr
		}
		expr = &IndexExpr{Pos: expr.Position(), Collection: expr, Index: idx}
	}
	return expr
}

func (p *parser) parsePrimary() Expr {
	tok := p.peek()
	switch tok.Type {
	case TokenInt, TokenFloat, TokenString, TokenBool:
		p.advance()
		return &LiteralExpr{Pos: tok.Pos, Value: tok.Literal}
	case TokenIdentifier:
		p.advance()
		return &IdentifierExpr{Pos: tok.Pos, Name: tok.Lexeme}
	case TokenLParen:
		open := p.advance()
		expr := p.parseExpression()
		if _, ok := p.consume(TokenRParen, "PARSE_EXPECTED_TOKEN", "Expected ')' after expression.", "Close grouped expression with ')'."); !ok {
			return expr
		}
		return &GroupExpr{Pos: open.Pos, Inner: expr}
	case TokenLBracket:
		return p.parseArrayLiteral()
	default:
		p.addDiag("PARSE_UNEXPECTED_TOKEN", fmt.Sprintf("Unexpected token '%s' in expression.", tok.Type.String()), tok.Pos, "Use a literal, identifier, indexed value, or grouped expression.")
		p.advance()
		return &LiteralExpr{Pos: tok.Pos, Value: nil}
	}
}

func (p *parser) parseArrayLiteral() Expr {
	open := p.advance()
	elements := make([]Expr, 0, 4)
	if p.match(TokenRBracket) {
		return &ArrayLiteralExpr{Pos: open.Pos, Elements: elements}
	}

	for {
		elements = append(elements, p.parseExpression())
		if p.match(TokenComma) {
			continue
		}
		if _, ok := p.consume(TokenRBracket, "PARSE_EXPECTED_TOKEN", "Expected ']' to close array literal.", "Close array literal with ']'."); !ok {
			return &ArrayLiteralExpr{Pos: open.Pos, Elements: elements}
		}
		break
	}
	return &ArrayLiteralExpr{Pos: open.Pos, Elements: elements}
}

func (p *parser) isAtEnd() bool {
	return p.peek().Type == TokenEOF
}

func (p *parser) peek() Token {
	return p.tokens[p.idx]
}

func (p *parser) previous() Token {
	return p.tokens[p.idx-1]
}

func (p *parser) check(tt TokenType) bool {
	if p.isAtEnd() {
		return tt == TokenEOF
	}
	return p.peek().Type == tt
}

func (p *parser) advance() Token {
	if !p.isAtEnd() {
		p.idx++
	}
	return p.tokens[p.idx-1]
}

func (p *parser) match(types ...TokenType) bool {
	for _, tt := range types {
		if p.check(tt) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *parser) consume(tt TokenType, code, message, hint string) (Token, bool) {
	if p.check(tt) {
		return p.advance(), true
	}
	tok := p.peek()
	p.addDiag(code, message, tok.Pos, hint)
	return Token{}, false
}

func (p *parser) addDiag(code, message string, pos Position, hint string) {
	p.diags = append(p.diags, newDiagnostic(code, "parser", message, pos, hint))
}

func (p *parser) synchronize() {
	if p.isAtEnd() {
		return
	}
	p.advance()
	for !p.isAtEnd() {
		switch p.peek().Type {
		case TokenVar, TokenConst, TokenRead, TokenWrite, TokenPop, TokenIf, TokenWhile, TokenFor, TokenRBrace:
			return
		default:
			p.advance()
		}
	}
}
