package simpl

// Program is the root of a Simpl source file.
type Program struct {
	Statements []Stmt
}

type Stmt interface {
	stmtNode()
	Position() Position
}

type Expr interface {
	exprNode()
	Position() Position
}

type DeclStmt struct {
	Pos         Position
	Name        string
	DeclaredTyp Type
	Initializer Expr
	Const       bool
}

func (s *DeclStmt) stmtNode()          {}
func (s *DeclStmt) Position() Position { return s.Pos }

type AssignStmt struct {
	Pos    Position
	Target Expr // IdentifierExpr or IndexExpr
	Value  Expr
}

func (s *AssignStmt) stmtNode()          {}
func (s *AssignStmt) Position() Position { return s.Pos }

type ReadStmt struct {
	Pos  Position
	Name string
}

func (s *ReadStmt) stmtNode()          {}
func (s *ReadStmt) Position() Position { return s.Pos }

type WriteStmt struct {
	Pos    Position
	Values []Expr
}

func (s *WriteStmt) stmtNode()          {}
func (s *WriteStmt) Position() Position { return s.Pos }

type PopStmt struct {
	Pos    Position
	Target Expr // IdentifierExpr or IndexExpr
}

func (s *PopStmt) stmtNode()          {}
func (s *PopStmt) Position() Position { return s.Pos }

type PushStmt struct {
	Pos    Position
	Target Expr // IdentifierExpr or IndexExpr
	Values []Expr
}

func (s *PushStmt) stmtNode()          {}
func (s *PushStmt) Position() Position { return s.Pos }

type IfClause struct {
	Pos       Position
	Condition Expr
	Block     *BlockStmt
}

type IfStmt struct {
	Pos      Position
	Primary  IfClause
	ElseIfs  []IfClause
	ElsePart *BlockStmt
}

func (s *IfStmt) stmtNode()          {}
func (s *IfStmt) Position() Position { return s.Pos }

type WhileStmt struct {
	Pos       Position
	Condition Expr
	Body      *BlockStmt
}

func (s *WhileStmt) stmtNode()          {}
func (s *WhileStmt) Position() Position { return s.Pos }

type ForStmt struct {
	Pos     Position
	VarName string
	From    Expr
	Until   Expr
	Step    Expr
	Body    *BlockStmt
}

func (s *ForStmt) stmtNode()          {}
func (s *ForStmt) Position() Position { return s.Pos }

type BlockStmt struct {
	Pos        Position
	Statements []Stmt
}

func (s *BlockStmt) stmtNode()          {}
func (s *BlockStmt) Position() Position { return s.Pos }

type LiteralExpr struct {
	Pos   Position
	Value any
}

func (e *LiteralExpr) exprNode()          {}
func (e *LiteralExpr) Position() Position { return e.Pos }

type IdentifierExpr struct {
	Pos  Position
	Name string
}

func (e *IdentifierExpr) exprNode()          {}
func (e *IdentifierExpr) Position() Position { return e.Pos }

type UnaryExpr struct {
	Pos      Position
	Operator TokenType
	Right    Expr
}

func (e *UnaryExpr) exprNode()          {}
func (e *UnaryExpr) Position() Position { return e.Pos }

type SizeExpr struct {
	Pos   Position
	Value Expr
}

func (e *SizeExpr) exprNode()          {}
func (e *SizeExpr) Position() Position { return e.Pos }

type BinaryExpr struct {
	Pos      Position
	Left     Expr
	Operator TokenType
	Right    Expr
}

func (e *BinaryExpr) exprNode()          {}
func (e *BinaryExpr) Position() Position { return e.Pos }

type GroupExpr struct {
	Pos   Position
	Inner Expr
}

func (e *GroupExpr) exprNode()          {}
func (e *GroupExpr) Position() Position { return e.Pos }

type ArrayLiteralExpr struct {
	Pos      Position
	Elements []Expr
}

func (e *ArrayLiteralExpr) exprNode()          {}
func (e *ArrayLiteralExpr) Position() Position { return e.Pos }

type IndexExpr struct {
	Pos        Position
	Collection Expr
	Index      Expr
}

func (e *IndexExpr) exprNode()          {}
func (e *IndexExpr) Position() Position { return e.Pos }
