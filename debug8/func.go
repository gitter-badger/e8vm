package debug8

import (
	"bytes"
	"fmt"

	"e8vm.io/e8vm/lex8"
)

func funcString(name string, f *Func) string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%8x +%4d: ", f.Start, f.Size)
	fmt.Fprintf(buf, "%s", name)
	if f.Pos != nil {
		fmt.Fprintf(buf, "  // %s", f.Pos)
	}
	if f.Frame > 0 {
		fmt.Fprintf(buf, " (frame=%d)", f.Frame)
	}
	return buf.String()
}

// Func saves the debug information of a function
type Func struct {
	Start uint32
	Size  uint32
	Frame uint32

	Pos *lex8.Pos
}
