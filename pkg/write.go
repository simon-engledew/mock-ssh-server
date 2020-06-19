package pkg

import (
	"fmt"
	"go.starlark.net/starlark"
	"io"
)

func write(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	w := thread.Local("writer").(io.Writer)

	var s string

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "s", &s); err != nil {
		return nil, err
	}

	n, err := fmt.Fprint(w, s)

	return starlark.MakeInt(n), err
}

func writeline(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	w := thread.Local("writer").(io.Writer)

	var s string

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "s?", &s); err != nil {
		return nil, err
	}

	n, err := fmt.Fprint(w, s+"\r\n")

	return starlark.MakeInt(n), err
}

func init() {
	register(starlark.NewBuiltin("write", write))
	register(starlark.NewBuiltin("writeline", writeline))
}
