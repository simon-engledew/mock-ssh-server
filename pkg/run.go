package pkg

import (
	"go.starlark.net/starlark"
	"io"
)

var predeclared = make(map[string]starlark.Value)

func register(builtin *starlark.Builtin) {
	predeclared[builtin.Name()] = builtin
}

func RunScript(script string, w io.Writer, r io.Reader) error {
	echoReader := io.TeeReader(r, w)

	thread := &starlark.Thread{Name: "SSH Session"}
	thread.SetLocal("reader", echoReader)
	thread.SetLocal("writer", w)

	_, err := starlark.ExecFile(thread, script, nil, predeclared)
	return err
}
