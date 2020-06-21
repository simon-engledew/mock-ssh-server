package pkg

import (
	"go.starlark.net/starlark"
	"io"
)

var predeclared starlark.StringDict = make(map[string]starlark.Value)

func register(builtin *starlark.Builtin) {
	predeclared[builtin.Name()] = builtin
}

func Load(filename string, src interface{}) (func(w io.Writer, r io.Reader) error, error) {
	_, mod, err := starlark.SourceProgram(filename, src, predeclared.Has)
	if err != nil {
		return nil, err
	}

	return func(w io.Writer, r io.Reader) error {
		echoReader := io.TeeReader(r, w)

		thread := &starlark.Thread{Name: filename}
		thread.SetLocal("reader", echoReader)
		thread.SetLocal("writer", w)

		_, err := mod.Init(thread, predeclared)
		return err
	}, nil
}
