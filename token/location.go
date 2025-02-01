package token

import "fmt"

type SrcLoc struct {
	File string
	Line uint
	Col  uint
}

func (loc SrcLoc) String() string {
	return fmt.Sprintf("%s:%d:%d:", loc.File, loc.Line, loc.Col)
}
