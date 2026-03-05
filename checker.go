package simpl

import "fmt"

type symbol struct {
	typ     Type
	mutable bool
}

type checker struct {
	scopes []map[string]symbol
	diags  []Diagnostic
}

func checkProgram(prog *Program) []Diagnostic {
	c := &checker{
		scopes: make([]map[string]symbol, 0, 8),
		diags:  make([]Diagnostic, 0, 16),
	}
	c.pushScope()
	for _, stmt := range prog.Statements {
		c.checkStmt(stmt)
	}
	c.popScope()
	return c.diags
}

func (c *checker) checkStmt(stmt Stmt) {
	switch s := stmt.(type) {
	case *DeclStmt:
		if _, exists := c.scopes[len(c.scopes)-1][s.Name]; exists {
			c.addDiag("TYPE_REDECLARED_IDENTIFIER", fmt.Sprintf("Variable '%s' is already declared in this scope.", s.Name), s.Pos, "Use a different name or remove the duplicate declaration.")
			return
		}
		if s.Initializer != nil {
			got := c.checkExpr(s.Initializer, &s.DeclaredTyp)
			if !got.Equals(s.DeclaredTyp) && got.Kind != TypeInvalid {
				c.addDiag("TYPE_MISMATCH", fmt.Sprintf("Cannot assign %s to %s '%s'.", got.String(), s.DeclaredTyp.String(), s.Name), s.Initializer.Position(), "Match the declared type exactly.")
			}
		}
		c.scopes[len(c.scopes)-1][s.Name] = symbol{typ: s.DeclaredTyp, mutable: !s.Const}
	case *AssignStmt:
		targetType := c.checkExpr(s.Target, nil)
		if targetType.Kind == TypeInvalid {
			return
		}
		if !c.isMutableTarget(s.Target) {
			c.addDiag("TYPE_CONST_REASSIGN", "Cannot assign to constant target.", s.Target.Position(), "Use a mutable variable declared with 'var'.")
			return
		}
		got := c.checkExpr(s.Value, &targetType)
		if got.Kind != TypeInvalid && !got.Equals(targetType) {
			c.addDiag("TYPE_MISMATCH", fmt.Sprintf("Cannot assign %s to target of type %s.", got.String(), targetType.String()), s.Value.Position(), "Use a value with matching type.")
		}
	case *ReadStmt:
		sym, ok := c.lookup(s.Name)
		if !ok {
			c.addDiag("TYPE_UNDECLARED_IDENTIFIER", fmt.Sprintf("Variable '%s' is not declared.", s.Name), s.Pos, "Declare the variable before reading into it.")
			return
		}
		if !sym.mutable {
			c.addDiag("TYPE_CONST_REASSIGN", fmt.Sprintf("Cannot read into constant '%s'.", s.Name), s.Pos, "Use 'var' for values that change.")
		}
	case *WriteStmt:
		for _, e := range s.Values {
			c.checkExpr(e, nil)
		}
	case *IfStmt:
		c.checkConditionBool(s.Primary.Condition)
		c.checkStmt(s.Primary.Block)
		for _, clause := range s.ElseIfs {
			c.checkConditionBool(clause.Condition)
			c.checkStmt(clause.Block)
		}
		if s.ElsePart != nil {
			c.checkStmt(s.ElsePart)
		}
	case *WhileStmt:
		c.checkConditionBool(s.Condition)
		c.checkStmt(s.Body)
	case *ForStmt:
		fromType := c.checkExpr(s.From, nil)
		untilType := c.checkExpr(s.Until, nil)
		stepType := c.checkExpr(s.Step, nil)
		if fromType.Kind != TypeInt && fromType.Kind != TypeInvalid {
			c.addDiag("TYPE_MISMATCH", "For loop 'from' expression must be int.", s.From.Position(), "Use an integer expression.")
		}
		if untilType.Kind != TypeInt && untilType.Kind != TypeInvalid {
			c.addDiag("TYPE_MISMATCH", "For loop 'until' expression must be int.", s.Until.Position(), "Use an integer expression.")
		}
		if stepType.Kind != TypeInt && stepType.Kind != TypeInvalid {
			c.addDiag("TYPE_MISMATCH", "For loop 'step' expression must be int.", s.Step.Position(), "Use an integer expression.")
		}
		if lit, ok := s.Step.(*LiteralExpr); ok {
			if v, isInt := lit.Value.(int64); isInt && v == 0 {
				c.addDiag("TYPE_MISMATCH", "For loop step cannot be zero.", s.Step.Position(), "Use a non-zero step value.")
			}
		}

		c.pushScope()
		c.scopes[len(c.scopes)-1][s.VarName] = symbol{typ: Type{Kind: TypeInt}, mutable: true}
		c.checkStmt(s.Body)
		c.popScope()
	case *BlockStmt:
		c.pushScope()
		for _, nested := range s.Statements {
			c.checkStmt(nested)
		}
		c.popScope()
	}
}

func (c *checker) checkExpr(expr Expr, expected *Type) Type {
	switch e := expr.(type) {
	case *LiteralExpr:
		switch v := e.Value.(type) {
		case int64:
			return Type{Kind: TypeInt}
		case float64:
			return Type{Kind: TypeFloat}
		case bool:
			return Type{Kind: TypeBool}
		case string:
			_ = v
			return Type{Kind: TypeString}
		case nil:
			return invalidType()
		default:
			return invalidType()
		}
	case *IdentifierExpr:
		sym, ok := c.lookup(e.Name)
		if !ok {
			c.addDiag("TYPE_UNDECLARED_IDENTIFIER", fmt.Sprintf("Variable '%s' is not declared.", e.Name), e.Pos, "Declare the variable before using it.")
			return invalidType()
		}
		return sym.typ
	case *UnaryExpr:
		right := c.checkExpr(e.Right, nil)
		switch e.Operator {
		case TokenMinus:
			if right.Kind != TypeInt && right.Kind != TypeFloat && right.Kind != TypeInvalid {
				c.addDiag("TYPE_MISMATCH", "Unary '-' expects int or float.", e.Pos, "Apply '-' to a numeric value.")
				return invalidType()
			}
			return right
		case TokenBang:
			if right.Kind != TypeBool && right.Kind != TypeInvalid {
				c.addDiag("TYPE_MISMATCH", "Unary '!' expects bool.", e.Pos, "Apply '!' to a boolean expression.")
				return invalidType()
			}
			return Type{Kind: TypeBool}
		default:
			c.addDiag("TYPE_MISMATCH", "Unsupported unary operator.", e.Pos, "Use supported unary operators.")
			return invalidType()
		}
	case *BinaryExpr:
		left := c.checkExpr(e.Left, nil)
		right := c.checkExpr(e.Right, nil)
		if left.Kind == TypeInvalid || right.Kind == TypeInvalid {
			return invalidType()
		}
		switch e.Operator {
		case TokenPlus:
			if left.Kind == TypeString && right.Kind == TypeString {
				return Type{Kind: TypeString}
			}
			if left.Kind == TypeInt && right.Kind == TypeInt {
				return Type{Kind: TypeInt}
			}
			if left.Kind == TypeFloat && right.Kind == TypeFloat {
				return Type{Kind: TypeFloat}
			}
			c.addDiag("TYPE_MISMATCH", "Operator '+' requires both operands to have the same numeric type or both strings.", e.Pos, "Do not mix types implicitly.")
			return invalidType()
		case TokenMinus, TokenStar, TokenSlash:
			if left.Kind == TypeInt && right.Kind == TypeInt {
				return Type{Kind: TypeInt}
			}
			if left.Kind == TypeFloat && right.Kind == TypeFloat {
				return Type{Kind: TypeFloat}
			}
			c.addDiag("TYPE_MISMATCH", "Arithmetic operators require matching numeric types.", e.Pos, "Use int with int or float with float.")
			return invalidType()
		case TokenPercent:
			if left.Kind == TypeInt && right.Kind == TypeInt {
				return Type{Kind: TypeInt}
			}
			c.addDiag("TYPE_MISMATCH", "Operator '%' requires int operands.", e.Pos, "Use integer operands with modulo.")
			return invalidType()
		case TokenGreater, TokenGreaterEq, TokenLess, TokenLessEq:
			if (left.Kind == TypeInt && right.Kind == TypeInt) || (left.Kind == TypeFloat && right.Kind == TypeFloat) {
				return Type{Kind: TypeBool}
			}
			c.addDiag("TYPE_MISMATCH", "Comparison operators require matching numeric types.", e.Pos, "Compare int with int or float with float.")
			return invalidType()
		case TokenEqual, TokenNotEqual:
			if !left.Equals(right) {
				c.addDiag("TYPE_MISMATCH", "Equality operators require both operands to have the same type.", e.Pos, "Compare values of identical type.")
				return invalidType()
			}
			return Type{Kind: TypeBool}
		default:
			c.addDiag("TYPE_MISMATCH", "Unsupported binary operator.", e.Pos, "Use supported operators.")
			return invalidType()
		}
	case *GroupExpr:
		return c.checkExpr(e.Inner, expected)
	case *ArrayLiteralExpr:
		if len(e.Elements) == 0 {
			if expected != nil && expected.Kind == TypeArray {
				return *expected
			}
			c.addDiag("TYPE_MISMATCH", "Cannot infer type for empty array literal.", e.Pos, "Provide a non-empty literal or assign where array type is known.")
			return invalidType()
		}

		var elemType Type
		if expected != nil && expected.Kind == TypeArray && expected.Elem != nil {
			elemType = *expected.Elem
		} else {
			elemType = invalidType()
		}

		for idx, el := range e.Elements {
			var target *Type
			if elemType.Kind != TypeInvalid {
				target = &elemType
			}
			t := c.checkExpr(el, target)
			if t.Kind == TypeInvalid {
				continue
			}
			if elemType.Kind == TypeInvalid {
				elemType = t
				continue
			}
			if !t.Equals(elemType) {
				c.addDiag("TYPE_MISMATCH", fmt.Sprintf("Array element %d has type %s but expected %s.", idx, t.String(), elemType.String()), el.Position(), "Use homogeneous array element types.")
			}
		}

		if elemType.Kind == TypeInvalid {
			return invalidType()
		}
		return Type{Kind: TypeArray, Elem: &elemType}
	case *IndexExpr:
		coll := c.checkExpr(e.Collection, nil)
		if coll.Kind != TypeArray {
			if coll.Kind != TypeInvalid {
				c.addDiag("TYPE_INVALID_INDEX", "Index operation requires an array target.", e.Pos, "Use indexing only on array values.")
			}
			return invalidType()
		}
		idxType := c.checkExpr(e.Index, nil)
		if idxType.Kind != TypeInt && idxType.Kind != TypeInvalid {
			c.addDiag("TYPE_INVALID_INDEX", "Array index must be int.", e.Index.Position(), "Use an integer index.")
		}
		if coll.Elem == nil {
			return invalidType()
		}
		return *coll.Elem
	default:
		return invalidType()
	}
}

func (c *checker) checkConditionBool(expr Expr) {
	t := c.checkExpr(expr, &Type{Kind: TypeBool})
	if t.Kind != TypeBool && t.Kind != TypeInvalid {
		c.addDiag("TYPE_MISMATCH", "Condition must have type bool.", expr.Position(), "Use a boolean expression in if/while conditions.")
	}
}

func (c *checker) isMutableTarget(expr Expr) bool {
	switch e := expr.(type) {
	case *IdentifierExpr:
		sym, ok := c.lookup(e.Name)
		if !ok {
			c.addDiag("TYPE_UNDECLARED_IDENTIFIER", fmt.Sprintf("Variable '%s' is not declared.", e.Name), e.Pos, "Declare the variable before assignment.")
			return false
		}
		return sym.mutable
	case *IndexExpr:
		return c.isMutableTarget(e.Collection)
	default:
		return false
	}
}

func (c *checker) lookup(name string) (symbol, bool) {
	for i := len(c.scopes) - 1; i >= 0; i-- {
		if sym, ok := c.scopes[i][name]; ok {
			return sym, true
		}
	}
	return symbol{}, false
}

func (c *checker) pushScope() {
	c.scopes = append(c.scopes, make(map[string]symbol))
}

func (c *checker) popScope() {
	if len(c.scopes) == 0 {
		return
	}
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *checker) addDiag(code, message string, pos Position, hint string) {
	c.diags = append(c.diags, newDiagnostic(code, "type", message, pos, hint))
}
