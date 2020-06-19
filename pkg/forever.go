package pkg

import (
	"fmt"
	"go.starlark.net/starlark"
)

type foreverIterator struct {
}
type foreverIterable struct {
}

func (f foreverIterable) String() string {
	return "forever"
}
func (f foreverIterable) Type() string {
	return "forever"
}
func (f foreverIterable) Freeze()                    {} // immutable
func (f foreverIterable) Truth() starlark.Bool       { return starlark.True }
func (f foreverIterable) Hash() (uint32, error)      { return 0, fmt.Errorf("unhashable: %s", f.Type()) }
func (f foreverIterable) Iterate() starlark.Iterator { return &foreverIterator{} }

func (i *foreverIterator) Next(p *starlark.Value) bool {
	return true
}
func (i *foreverIterator) Done() {
}

func forever(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return foreverIterable{}, nil
}

func init() {
	register(starlark.NewBuiltin("forever", forever))
}
