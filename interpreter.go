package simpl

import (
	"fmt"
	"strings"
	"time"
)

type runtimeVar struct {
	value   Value
	mutable bool
}

type interpreter struct {
	scopes     []map[string]runtimeVar
	diags      []Diagnostic
	stdout     strings.Builder
	input      []string
	inputIndex int
	steps      int64
	maxSteps   int64
	deadline   time.Time
	timedOut   bool
	halted     bool
}

func newInterpreter(stdin string, opts RunOptions) *interpreter {
	return &interpreter{
		scopes:   make([]map[string]runtimeVar, 0, 8),
		diags:    make([]Diagnostic, 0, 8),
		input:    strings.Fields(stdin),
		maxSteps: opts.MaxSteps,
		deadline: time.Now().Add(opts.Timeout),
	}
}

func (i *interpreter) run(prog *Program) RunResult {
	i.pushScope()
	i.execStatements(prog.Statements)
	i.popScope()

	return RunResult{
		Stdout:      i.stdout.String(),
		Diagnostics: i.diags,
		StepsUsed:   i.steps,
		TimedOut:    i.timedOut,
	}
}

func (i *interpreter) execStatements(stmts []Stmt) {
	for _, stmt := range stmts {
		if i.halted {
			return
		}
		if !i.tick(stmt.Position()) {
			return
		}
		i.execStmt(stmt)
	}
}

func (i *interpreter) execStmt(stmt Stmt) {
	switch s := stmt.(type) {
	case *DeclStmt:
		val := zeroValue(s.DeclaredTyp)
		if s.Initializer != nil {
			evaluated, ok := i.evalExpr(s.Initializer)
			if !ok {
				return
			}
			val = cloneValue(evaluated)
		}
		i.scopes[len(i.scopes)-1][s.Name] = runtimeVar{value: val, mutable: !s.Const}
	case *AssignStmt:
		v, ok := i.evalExpr(s.Value)
		if !ok {
			return
		}
		i.assignTarget(s.Target, v)
	case *ReadStmt:
		i.execRead(s)
	case *WriteStmt:
		for _, expr := range s.Values {
			v, ok := i.evalExpr(expr)
			if !ok {
				return
			}
			i.stdout.WriteString(valueToString(v))
		}
	case *IfStmt:
		if i.evalCondition(s.Primary.Condition) {
			i.execStmt(s.Primary.Block)
			return
		}
		for _, clause := range s.ElseIfs {
			if i.evalCondition(clause.Condition) {
				i.execStmt(clause.Block)
				return
			}
		}
		if s.ElsePart != nil {
			i.execStmt(s.ElsePart)
		}
	case *WhileStmt:
		for {
			if i.halted {
				return
			}
			if !i.tick(s.Pos) {
				return
			}
			if !i.evalCondition(s.Condition) {
				return
			}
			i.execStmt(s.Body)
		}
	case *ForStmt:
		fromV, ok := i.evalExpr(s.From)
		if !ok {
			return
		}
		untilV, ok := i.evalExpr(s.Until)
		if !ok {
			return
		}
		stepV, ok := i.evalExpr(s.Step)
		if !ok {
			return
		}
		if fromV.Type.Kind != TypeInt || untilV.Type.Kind != TypeInt || stepV.Type.Kind != TypeInt {
			i.addRuntimeDiag("RUNTIME_TYPE", "For loop bounds ('from'/'until') and step must be int at runtime.", s.Pos, "Ensure for-loop expressions are integers.")
			return
		}
		if stepV.Int == 0 {
			i.addRuntimeDiag("RUNTIME_INVALID_STEP", "For loop step cannot be zero.", s.Step.Position(), "Use a positive or negative non-zero step.")
			return
		}

		i.pushScope()
		i.scopes[len(i.scopes)-1][s.VarName] = runtimeVar{value: Value{Type: Type{Kind: TypeInt}}, mutable: true}
		cur := fromV.Int
		limit := untilV.Int
		step := stepV.Int
		for {
			if i.halted {
				i.popScope()
				return
			}
			if !i.tick(s.Pos) {
				i.popScope()
				return
			}

			if step > 0 {
				if cur >= limit {
					break
				}
			} else {
				if cur <= limit {
					break
				}
			}

			i.scopes[len(i.scopes)-1][s.VarName] = runtimeVar{value: Value{Type: Type{Kind: TypeInt}, Int: cur}, mutable: true}
			i.execStmt(s.Body)
			cur += step
		}
		i.popScope()
	case *BlockStmt:
		i.pushScope()
		i.execStatements(s.Statements)
		i.popScope()
	}
}

func (i *interpreter) execRead(stmt *ReadStmt) {
	v, scopeIdx, ok := i.lookupVar(stmt.Name)
	if !ok {
		i.addRuntimeDiag("RUNTIME_UNDECLARED", fmt.Sprintf("Variable '%s' is not declared.", stmt.Name), stmt.Pos, "Declare it before reading.")
		return
	}
	if !v.mutable {
		i.addRuntimeDiag("RUNTIME_CONST_REASSIGN", fmt.Sprintf("Cannot read into constant '%s'.", stmt.Name), stmt.Pos, "Use a mutable variable declared with var.")
		return
	}
	if i.inputIndex >= len(i.input) {
		i.addRuntimeDiag("RUNTIME_READ_EOF", fmt.Sprintf("No more input tokens available for '%s'.", stmt.Name), stmt.Pos, "Provide enough input tokens for all read statements.")
		return
	}
	token := i.input[i.inputIndex]
	i.inputIndex++
	parsed, err := parseInputToken(token, v.value.Type)
	if err != nil {
		i.addRuntimeDiag("RUNTIME_READ_PARSE", fmt.Sprintf("Input token '%s' cannot be parsed as %s.", token, v.value.Type.String()), stmt.Pos, "Ensure input tokens match variable types.")
		return
	}
	v.value = parsed
	i.scopes[scopeIdx][stmt.Name] = v
}

func (i *interpreter) evalCondition(expr Expr) bool {
	v, ok := i.evalExpr(expr)
	if !ok {
		return false
	}
	if v.Type.Kind != TypeBool {
		i.addRuntimeDiag("RUNTIME_TYPE", "Condition did not evaluate to bool.", expr.Position(), "Use boolean expressions in conditions.")
		return false
	}
	return v.Bool
}

func (i *interpreter) evalExpr(expr Expr) (Value, bool) {
	switch e := expr.(type) {
	case *LiteralExpr:
		switch v := e.Value.(type) {
		case int64:
			return Value{Type: Type{Kind: TypeInt}, Int: v}, true
		case float64:
			return Value{Type: Type{Kind: TypeFloat}, Float: v}, true
		case bool:
			return Value{Type: Type{Kind: TypeBool}, Bool: v}, true
		case string:
			return Value{Type: Type{Kind: TypeString}, String: v}, true
		default:
			i.addRuntimeDiag("RUNTIME_TYPE", "Invalid literal value.", e.Pos, "Use supported literal types.")
			return Value{}, false
		}
	case *IdentifierExpr:
		v, _, ok := i.lookupVar(e.Name)
		if !ok {
			i.addRuntimeDiag("RUNTIME_UNDECLARED", fmt.Sprintf("Variable '%s' is not declared.", e.Name), e.Pos, "Declare it before use.")
			return Value{}, false
		}
		return cloneValue(v.value), true
	case *UnaryExpr:
		right, ok := i.evalExpr(e.Right)
		if !ok {
			return Value{}, false
		}
		switch e.Operator {
		case TokenMinus:
			if right.Type.Kind == TypeInt {
				return Value{Type: right.Type, Int: -right.Int}, true
			}
			if right.Type.Kind == TypeFloat {
				return Value{Type: right.Type, Float: -right.Float}, true
			}
			i.addRuntimeDiag("RUNTIME_TYPE", "Unary '-' requires int or float.", e.Pos, "Apply '-' to numeric values.")
			return Value{}, false
		case TokenBang:
			if right.Type.Kind == TypeBool {
				return Value{Type: Type{Kind: TypeBool}, Bool: !right.Bool}, true
			}
			i.addRuntimeDiag("RUNTIME_TYPE", "Unary '!' requires bool.", e.Pos, "Apply '!' to a boolean value.")
			return Value{}, false
		default:
			i.addRuntimeDiag("RUNTIME_TYPE", "Unsupported unary operator.", e.Pos, "Use supported unary operators.")
			return Value{}, false
		}
	case *BinaryExpr:
		left, ok := i.evalExpr(e.Left)
		if !ok {
			return Value{}, false
		}
		right, ok := i.evalExpr(e.Right)
		if !ok {
			return Value{}, false
		}
		return i.evalBinary(e, left, right)
	case *GroupExpr:
		return i.evalExpr(e.Inner)
	case *ArrayLiteralExpr:
		arr := make([]Value, len(e.Elements))
		var elemType *Type
		for idx := range e.Elements {
			v, ok := i.evalExpr(e.Elements[idx])
			if !ok {
				return Value{}, false
			}
			if elemType == nil {
				t := v.Type
				elemType = &t
			} else if !v.Type.Equals(*elemType) {
				i.addRuntimeDiag("RUNTIME_TYPE", "Array literal contains mixed element types.", e.Elements[idx].Position(), "Use homogeneous array element types.")
				return Value{}, false
			}
			arr[idx] = cloneValue(v)
		}
		if elemType == nil {
			unknown := invalidType()
			elemType = &unknown
		}
		return Value{Type: Type{Kind: TypeArray, Elem: elemType}, Array: arr}, true
	case *IndexExpr:
		collection, ok := i.evalExpr(e.Collection)
		if !ok {
			return Value{}, false
		}
		if collection.Type.Kind != TypeArray {
			i.addRuntimeDiag("RUNTIME_INDEX_OOB", "Index target is not an array.", e.Pos, "Use indexing only on arrays.")
			return Value{}, false
		}
		idxValue, ok := i.evalExpr(e.Index)
		if !ok {
			return Value{}, false
		}
		if idxValue.Type.Kind != TypeInt {
			i.addRuntimeDiag("RUNTIME_TYPE", "Array index must be int.", e.Index.Position(), "Use an integer index.")
			return Value{}, false
		}
		if idxValue.Int < 0 || idxValue.Int >= int64(len(collection.Array)) {
			i.addRuntimeDiag("RUNTIME_INDEX_OOB", "Array index is out of range.", e.Index.Position(), "Use an index within array bounds.")
			return Value{}, false
		}
		return cloneValue(collection.Array[idxValue.Int]), true
	default:
		i.addRuntimeDiag("RUNTIME_TYPE", "Unsupported expression node.", expr.Position(), "Use a supported expression form.")
		return Value{}, false
	}
}

func (i *interpreter) evalBinary(expr *BinaryExpr, left, right Value) (Value, bool) {
	switch expr.Operator {
	case TokenPlus:
		if left.Type.Kind == TypeString && right.Type.Kind == TypeString {
			return Value{Type: Type{Kind: TypeString}, String: left.String + right.String}, true
		}
		if left.Type.Kind == TypeInt && right.Type.Kind == TypeInt {
			return Value{Type: Type{Kind: TypeInt}, Int: left.Int + right.Int}, true
		}
		if left.Type.Kind == TypeFloat && right.Type.Kind == TypeFloat {
			return Value{Type: Type{Kind: TypeFloat}, Float: left.Float + right.Float}, true
		}
	case TokenMinus:
		if left.Type.Kind == TypeInt && right.Type.Kind == TypeInt {
			return Value{Type: Type{Kind: TypeInt}, Int: left.Int - right.Int}, true
		}
		if left.Type.Kind == TypeFloat && right.Type.Kind == TypeFloat {
			return Value{Type: Type{Kind: TypeFloat}, Float: left.Float - right.Float}, true
		}
	case TokenStar:
		if left.Type.Kind == TypeInt && right.Type.Kind == TypeInt {
			return Value{Type: Type{Kind: TypeInt}, Int: left.Int * right.Int}, true
		}
		if left.Type.Kind == TypeFloat && right.Type.Kind == TypeFloat {
			return Value{Type: Type{Kind: TypeFloat}, Float: left.Float * right.Float}, true
		}
	case TokenSlash:
		if left.Type.Kind == TypeInt && right.Type.Kind == TypeInt {
			if right.Int == 0 {
				i.addRuntimeDiag("RUNTIME_DIV_ZERO", "Division by zero is not allowed.", expr.Pos, "Ensure divisor is non-zero.")
				return Value{}, false
			}
			return Value{Type: Type{Kind: TypeInt}, Int: left.Int / right.Int}, true
		}
		if left.Type.Kind == TypeFloat && right.Type.Kind == TypeFloat {
			if right.Float == 0 {
				i.addRuntimeDiag("RUNTIME_DIV_ZERO", "Division by zero is not allowed.", expr.Pos, "Ensure divisor is non-zero.")
				return Value{}, false
			}
			return Value{Type: Type{Kind: TypeFloat}, Float: left.Float / right.Float}, true
		}
	case TokenPercent:
		if left.Type.Kind == TypeInt && right.Type.Kind == TypeInt {
			if right.Int == 0 {
				i.addRuntimeDiag("RUNTIME_DIV_ZERO", "Modulo by zero is not allowed.", expr.Pos, "Ensure divisor is non-zero.")
				return Value{}, false
			}
			return Value{Type: Type{Kind: TypeInt}, Int: left.Int % right.Int}, true
		}
	case TokenGreater:
		if left.Type.Kind == TypeInt && right.Type.Kind == TypeInt {
			return Value{Type: Type{Kind: TypeBool}, Bool: left.Int > right.Int}, true
		}
		if left.Type.Kind == TypeFloat && right.Type.Kind == TypeFloat {
			return Value{Type: Type{Kind: TypeBool}, Bool: left.Float > right.Float}, true
		}
	case TokenGreaterEq:
		if left.Type.Kind == TypeInt && right.Type.Kind == TypeInt {
			return Value{Type: Type{Kind: TypeBool}, Bool: left.Int >= right.Int}, true
		}
		if left.Type.Kind == TypeFloat && right.Type.Kind == TypeFloat {
			return Value{Type: Type{Kind: TypeBool}, Bool: left.Float >= right.Float}, true
		}
	case TokenLess:
		if left.Type.Kind == TypeInt && right.Type.Kind == TypeInt {
			return Value{Type: Type{Kind: TypeBool}, Bool: left.Int < right.Int}, true
		}
		if left.Type.Kind == TypeFloat && right.Type.Kind == TypeFloat {
			return Value{Type: Type{Kind: TypeBool}, Bool: left.Float < right.Float}, true
		}
	case TokenLessEq:
		if left.Type.Kind == TypeInt && right.Type.Kind == TypeInt {
			return Value{Type: Type{Kind: TypeBool}, Bool: left.Int <= right.Int}, true
		}
		if left.Type.Kind == TypeFloat && right.Type.Kind == TypeFloat {
			return Value{Type: Type{Kind: TypeBool}, Bool: left.Float <= right.Float}, true
		}
	case TokenEqual:
		if left.Type.Equals(right.Type) {
			return Value{Type: Type{Kind: TypeBool}, Bool: valuesEqual(left, right)}, true
		}
	case TokenNotEqual:
		if left.Type.Equals(right.Type) {
			return Value{Type: Type{Kind: TypeBool}, Bool: !valuesEqual(left, right)}, true
		}
	}
	i.addRuntimeDiag("RUNTIME_TYPE", "Invalid operand types for operator.", expr.Pos, "Use matching operand types for this operator.")
	return Value{}, false
}

func (i *interpreter) assignTarget(target Expr, value Value) {
	switch t := target.(type) {
	case *IdentifierExpr:
		cur, scopeIdx, ok := i.lookupVar(t.Name)
		if !ok {
			i.addRuntimeDiag("RUNTIME_UNDECLARED", fmt.Sprintf("Variable '%s' is not declared.", t.Name), t.Pos, "Declare it before assignment.")
			return
		}
		if !cur.mutable {
			i.addRuntimeDiag("RUNTIME_CONST_REASSIGN", fmt.Sprintf("Cannot assign to constant '%s'.", t.Name), t.Pos, "Use a variable declared with var.")
			return
		}
		if !cur.value.Type.Equals(value.Type) {
			i.addRuntimeDiag("RUNTIME_TYPE", fmt.Sprintf("Cannot assign %s to %s.", value.Type.String(), cur.value.Type.String()), t.Pos, "Assign values with matching types.")
			return
		}
		cur.value = cloneValue(value)
		i.scopes[scopeIdx][t.Name] = cur
	case *IndexExpr:
		i.assignIndexedTarget(t, value)
	default:
		i.addRuntimeDiag("RUNTIME_TYPE", "Invalid assignment target.", target.Position(), "Assign to a variable or indexed array element.")
	}
}

func (i *interpreter) assignIndexedTarget(expr *IndexExpr, value Value) {
	name, idxExprs, ok := flattenIndexTarget(expr)
	if !ok {
		i.addRuntimeDiag("RUNTIME_TYPE", "Invalid indexed assignment target.", expr.Pos, "Use a variable followed by one or more indexes.")
		return
	}

	varIndexes := make([]int64, len(idxExprs))
	for idx := range idxExprs {
		v, ok := i.evalExpr(idxExprs[idx])
		if !ok {
			return
		}
		if v.Type.Kind != TypeInt {
			i.addRuntimeDiag("RUNTIME_TYPE", "Array index must be int.", idxExprs[idx].Position(), "Use integer indexes.")
			return
		}
		varIndexes[idx] = v.Int
	}

	root, scopeIdx, ok := i.lookupVar(name)
	if !ok {
		i.addRuntimeDiag("RUNTIME_UNDECLARED", fmt.Sprintf("Variable '%s' is not declared.", name), expr.Pos, "Declare it before assignment.")
		return
	}
	if !root.mutable {
		i.addRuntimeDiag("RUNTIME_CONST_REASSIGN", fmt.Sprintf("Cannot assign into constant '%s'.", name), expr.Pos, "Use var for mutable arrays.")
		return
	}

	mutable := cloneValue(root.value)
	if !setValueAtIndexes(&mutable, varIndexes, value, expr.Pos, i) {
		return
	}
	root.value = mutable
	i.scopes[scopeIdx][name] = root
}

func flattenIndexTarget(expr Expr) (string, []Expr, bool) {
	switch e := expr.(type) {
	case *IdentifierExpr:
		return e.Name, []Expr{}, true
	case *IndexExpr:
		name, idxs, ok := flattenIndexTarget(e.Collection)
		if !ok {
			return "", nil, false
		}
		idxs = append(idxs, e.Index)
		return name, idxs, true
	default:
		return "", nil, false
	}
}

func setValueAtIndexes(root *Value, indexes []int64, newValue Value, pos Position, interp *interpreter) bool {
	if len(indexes) == 0 {
		if !root.Type.Equals(newValue.Type) {
			interp.addRuntimeDiag("RUNTIME_TYPE", fmt.Sprintf("Cannot assign %s to %s.", newValue.Type.String(), root.Type.String()), pos, "Assign values with matching element type.")
			return false
		}
		*root = cloneValue(newValue)
		return true
	}
	if root.Type.Kind != TypeArray {
		interp.addRuntimeDiag("RUNTIME_TYPE", "Indexed assignment target is not an array.", pos, "Use indexing only on arrays.")
		return false
	}
	idx := indexes[0]
	if idx < 0 || idx >= int64(len(root.Array)) {
		interp.addRuntimeDiag("RUNTIME_INDEX_OOB", "Array index is out of range.", pos, "Use an index within array bounds.")
		return false
	}
	child := root.Array[idx]
	if !setValueAtIndexes(&child, indexes[1:], newValue, pos, interp) {
		return false
	}
	root.Array[idx] = child
	return true
}

func (i *interpreter) lookupVar(name string) (runtimeVar, int, bool) {
	for idx := len(i.scopes) - 1; idx >= 0; idx-- {
		if v, ok := i.scopes[idx][name]; ok {
			return v, idx, true
		}
	}
	return runtimeVar{}, -1, false
}

func (i *interpreter) pushScope() {
	i.scopes = append(i.scopes, make(map[string]runtimeVar))
}

func (i *interpreter) popScope() {
	if len(i.scopes) == 0 {
		return
	}
	i.scopes = i.scopes[:len(i.scopes)-1]
}

func (i *interpreter) tick(pos Position) bool {
	if i.halted {
		return false
	}
	if !i.deadline.IsZero() && time.Now().After(i.deadline) {
		i.timedOut = true
		i.halted = true
		i.diags = append(i.diags, newDiagnostic("LIMIT_TIMEOUT", "limit", "Execution timed out.", pos, "Optimize the loop or increase timeout for this exercise."))
		return false
	}
	i.steps++
	if i.maxSteps > 0 && i.steps > i.maxSteps {
		i.halted = true
		i.diags = append(i.diags, newDiagnostic("LIMIT_STEPS_EXCEEDED", "limit", "Execution step limit exceeded.", pos, "Check for infinite loops or reduce algorithm steps."))
		return false
	}
	return true
}

func (i *interpreter) addRuntimeDiag(code, message string, pos Position, hint string) {
	i.halted = true
	i.diags = append(i.diags, newDiagnostic(code, "runtime", message, pos, hint))
}
