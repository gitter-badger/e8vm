package sempass

import (
	"e8vm.io/e8vm/g8/ast"
	"e8vm.io/e8vm/g8/tast"
	"e8vm.io/e8vm/g8/types"
	"e8vm.io/e8vm/lex8"
	"e8vm.io/e8vm/sym8"
)

func varDeclPrepare(
	b *Builder, toks []*lex8.Token, lst *tast.ExprList, t types.T,
) *tast.ExprList {
	ret := tast.NewExprList()
	for i, tok := range toks {
		e := lst.Exprs[i]
		etype := e.Type()
		if types.IsNil(etype) {
			e = tast.NewCast(e, t)
		} else if v, ok := types.NumConst(etype); ok {
			e = constCast(b, tok.Pos, v, e, t)
			if e == nil {
				return nil
			}
		}
		ret.Append(e)
	}
	return ret
}

func declareVars(b *Builder, ids []*lex8.Token, t types.T) []*sym8.Symbol {
	var syms []*sym8.Symbol
	for _, id := range ids {
		s := declareVar(b, id, t)
		if s == nil {
			return nil
		}
		syms = append(syms, s)
	}
	return syms
}

func buildVarDecl(b *Builder, d *ast.VarDecl) *tast.Define {
	ids := d.Idents.Idents

	if d.Eq != nil {
		right := b.BuildExpr(d.Exprs)
		if right == nil {
			return nil
		}

		if d.Type == nil {
			ret := define(b, ids, right, d.Eq)
			if ret == nil {
				return nil
			}
			return ret
		}

		tdest := b.BuildType(d.Type)
		if tdest == nil {
			return nil
		}

		if !types.IsAllocable(tdest) {
			pos := ast.ExprPos(d.Type)
			b.Errorf(pos, "%s is not allocatable", tdest)
			return nil
		}

		// assignable check
		ts := right.R().TypeList()
		for _, t := range ts {
			if !types.CanAssign(tdest, t) {
				b.Errorf(d.Eq.Pos, "cannot assign %s to %s", t, tdest)
				return nil
			}
		}

		// cast literal expression lists
		// after the casting, all types should be matching to tdest
		if exprList, ok := tast.MakeExprList(right); ok {
			exprList = varDeclPrepare(b, ids, exprList, tdest)
			if exprList == nil {
				return nil
			}
			right = exprList
		}

		syms := declareVars(b, ids, tdest)
		if syms == nil {
			return nil
		}
		return &tast.Define{syms, right}
	}

	if d.Type == nil {
		panic("type missing")
	}

	t := b.BuildType(d.Type)
	if t == nil {
		return nil
	}

	syms := declareVars(b, ids, t)
	if syms == nil {
		return nil
	}
	return &tast.Define{syms, nil}
}

func buildVarDecls(b *Builder, decls *ast.VarDecls) tast.Stmt {
	if len(decls.Decls) == 0 {
		return nil
	}

	var ret []*tast.Define
	for _, d := range decls.Decls {
		d := buildVarDecl(b, d)
		if d != nil {
			ret = append(ret, d)
		}
	}
	return &tast.VarDecls{ret}
}
