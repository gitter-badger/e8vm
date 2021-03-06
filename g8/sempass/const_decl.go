package sempass

import (
	"e8vm.io/e8vm/g8/ast"
	"e8vm.io/e8vm/g8/tast"
	"e8vm.io/e8vm/g8/types"
	"e8vm.io/e8vm/lex8"
	"e8vm.io/e8vm/sym8"
)

func declareConst(b *Builder, tok *lex8.Token, t types.T) *sym8.Symbol {
	name := tok.Lit
	s := sym8.Make(b.path, name, tast.SymConst, nil, t, tok.Pos)
	conflict := b.scope.Declare(s)
	if conflict != nil {
		b.Errorf(tok.Pos, "%q already decalred as a %s",
			name, tast.SymStr(conflict.Type),
		)
		return nil
	}
	return s
}

func buildConstDecl(b *Builder, d *ast.ConstDecl) *tast.Define {
	if d.Type != nil {
		b.Errorf(ast.ExprPos(d.Type), "typed const not implemented yet")
		return nil
	}

	right := buildConstExprList(b, d.Exprs)
	if right == nil {
		return nil
	}

	nright := right.R().Len()
	idents := d.Idents.Idents
	nleft := len(idents)
	if nleft != nright {
		b.Errorf(d.Eq.Pos, "%d values for %d identifiers",
			nright, nleft,
		)
		return nil
	}

	var syms []*sym8.Symbol
	for i, ident := range idents {
		t := right.R().At(i).Type()
		if !types.IsConst(t) {
			b.Errorf(ast.ExprPos(d.Exprs.Exprs[i]), "not a const")
			return nil
		}

		sym := declareConst(b, ident, t)
		if sym == nil {
			return nil
		}
		syms = append(syms, sym)
	}

	return &tast.Define{syms, right}
}

func buildConstDecls(b *Builder, decls *ast.ConstDecls) tast.Stmt {
	if len(decls.Decls) == 0 {
		return nil
	}

	var ret []*tast.Define
	for _, d := range decls.Decls {
		d := buildConstDecl(b, d)
		if d != nil {
			ret = append(ret, d)
		}
	}
	return &tast.ConstDecls{ret}
}
