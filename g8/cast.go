package g8

import (
	"e8vm.io/e8vm/arch8"
	"e8vm.io/e8vm/g8/ir"
	"e8vm.io/e8vm/g8/types"
)

func isPointerType(t types.T) bool {
	return types.IsPointer(t) || types.IsFuncPointer(t)
}

func regSizeCastable(to, from types.T) bool {
	if types.IsPointer(to) && types.IsPointer(from) {
		return true
	}
	if isPointerType(to) && types.IsBasic(from, types.Uint) {
		return true
	}
	if types.IsBasic(to, types.Uint) && isPointerType(from) {
		return true
	}
	return false
}

func buildCast(b *builder, from *ref, t types.T) *ref {
	srcType := from.Type()
	ret := b.newTemp(t)

	if types.IsNil(srcType) {
		size := t.Size()
		if size == arch8.RegSize {
			return newRef(t, ir.Num(0))
		}
		if _, ok := t.(*types.Slice); !ok {
			panic("bug")
		}
		ret := b.newTemp(t)
		b.b.Zero(ret.IR())
		return ret
	}

	if c, ok := srcType.(*types.Const); ok {
		if v, ok := types.NumConst(srcType); ok && types.IsInteger(t) {
			return newRef(t, constNumIr(v, t))
		}
		// TODO: we do not support typed const right?
		// so why need this line?
		srcType = c.Type // using the underlying type
	}

	if types.IsInteger(t) && types.IsInteger(srcType) {
		b.b.Arith(ret.IR(), nil, "cast", from.IR())
		return ret
	}
	if regSizeCastable(t, srcType) {
		b.b.Arith(ret.IR(), nil, "", from.IR())
		return ret
	}
	panic("bug")
}
