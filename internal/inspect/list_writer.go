package inspect

import (
	"fmt"

	"github.com/jedib0t/go-pretty/v6/list"
)

type listWriter struct {
	list.Writer
}

func newListWriter() listWriter {
	l :=  listWriter{Writer: list.NewWriter()}
	style := list.StyleConnectedLight
	style.CharItemSingle = style.CharItemBottom
	style.CharItemTop = style.CharItemFirst
	l.SetStyle(style)
	return l
}

func (lw listWriter) Append(v any) {
	lw.Writer.AppendItem(v)
}

func (lw listWriter) Appendf(format string, a ...any) {
	lw.Writer.AppendItem(fmt.Sprintf(format, a...))
}

func (lw listWriter) IndentFunc(fn func(listWriter)) {
	lw.Writer.Indent()
	fn(lw)
	lw.Writer.UnIndent()
}
