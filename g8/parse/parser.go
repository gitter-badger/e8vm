package parse

import (
	"fmt"
	"io"

	"e8vm.io/e8vm/g8/ast"
	"e8vm.io/e8vm/lex8"
)

type parser struct {
	f string
	x lex8.Tokener
	*lex8.Parser

	exprFunc    func(p *parser) ast.Expr
	typeFunc    func(p *parser) ast.Expr
	seeTypeFunc func(p *parser) bool
	stmtFunc    func(p *parser) ast.Stmt

	golike bool
}

func makeTokener(f string, r io.Reader, golike bool) (
	lex8.Tokener, *lex8.Recorder,
) {
	var x lex8.Tokener = newLexer(f, r)

	x = newSemiInserter(x)

	kw := lex8.NewKeyworder(x)
	kw.Ident = Ident
	kw.Keyword = Keyword
	if !golike {
		kw.Keywords = gKeywords
	} else {
		kw.Keywords = golikeKeywords
	}

	rec := lex8.NewRecorder(kw)
	return lex8.NewCommentRemover(rec), rec
}

func newParser(f string, r io.Reader, golike bool) (*parser, *lex8.Recorder) {
	ret := new(parser)
	ret.f = f
	ret.golike = golike
	x, rec := makeTokener(f, r, golike)
	ret.x = x
	ret.Parser = lex8.NewParser(ret.x, Types)
	return ret, rec
}

func (p *parser) parseType() ast.Expr {
	if p.typeFunc == nil {
		return nil
	}

	return p.typeFunc(p)
}

func (p *parser) parseExpr() ast.Expr {
	if p.exprFunc == nil {
		return nil
	}

	return p.exprFunc(p)
}

func (p *parser) parseStmt() ast.Expr {
	if p.stmtFunc == nil {
		p.ExpectSemi()
		p.skipErrStmt()
		return nil
	}
	return p.stmtFunc(p)
}

func (p *parser) SeeOp(ops ...string) bool {
	t := p.Token()
	if t.Type != Operator {
		return false
	}
	for _, op := range ops {
		if t.Lit == op {
			return true
		}
	}
	return false
}

func (p *parser) typeStr(t *lex8.Token) string {
	if t.Type == Operator {
		return fmt.Sprintf("'%s'", t.Lit)
	} else if t.Type == Semi {
		return "';'"
	}
	return TypeStr(t.Type)
}

func (p *parser) AcceptSemi() *lex8.Token {
	if p.InError() {
		return nil
	}

	t := p.Token()
	if t.Type == Operator && (t.Lit == "}" || t.Lit == ")") {
		return t // fake semicolon by operator
	}

	if t.Type != Semi {
		return nil
	}
	return p.Shift()
}

func (p *parser) SeeSemi() bool {
	t := p.Token()
	if t.Type == Semi {
		return true
	}
	if t.Type == Operator && (t.Lit == "}" || t.Lit == ")") {
		return true
	}
	return false
}

func (p *parser) ExpectSemi() *lex8.Token {
	if p.InError() {
		return nil
	}

	t := p.Token()
	if t.Type == Operator && (t.Lit == "}" || t.Lit == ")") {
		return t // fake semicolon by operator
	}

	if t.Type != Semi {
		p.ErrorfHere("expect ';', got %s", p.typeStr(t))
		return nil
	}
	return p.Shift()
}

func (p *parser) skipErrStmt() bool {
	if !p.InError() {
		return false
	}

	for {
		t := p.Token()
		if t.Type == Semi || t.Type == lex8.EOF {
			break
		} else if p.SeeOp("}") {
			break
		}
		p.Next()
	}
	if p.See(Semi) {
		p.Next()
	}

	p.BailOut()
	return true
}

func (p *parser) SeeType() bool {
	if p.seeTypeFunc == nil {
		return false
	}

	return p.seeTypeFunc(p)
}

func (p *parser) SeeKeyword(kw string) bool {
	return p.SeeLit(Keyword, kw)
}

func (p *parser) ExpectOp(op string) *lex8.Token {
	if p.InError() {
		return nil
	}
	t := p.Token()
	if t.Type != Operator || t.Lit != op {
		p.ErrorfHere("expect '%s', got %s", op, p.typeStr(t))
		return nil
	}

	return p.Shift()
}

func (p *parser) ExpectKeyword(kw string) *lex8.Token {
	if !p.SeeLit(Keyword, kw) {
		p.ErrorfHere("expect keyword '%s', got %s",
			kw, p.typeStr(p.Token()),
		)
		return nil
	}
	return p.Shift()
}
