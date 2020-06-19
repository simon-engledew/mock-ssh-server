package pkg

import (
	"bytes"
	"fmt"
	"go.starlark.net/starlark"
	"io"
	"regexp"
	"sync"
)

var buffers = sync.Pool{
	New: func() interface{} {
		buffer := new(bytes.Buffer)
		buffer.Grow(16 * 1024)
		return buffer
	},
}

func getBuffer() *bytes.Buffer {
	return buffers.Get().(*bytes.Buffer)
}

func putBuffer(buffer *bytes.Buffer) {
	buffer.Reset()
	buffers.Put(buffer)
}

var peeks = sync.Pool{
	New: func() interface{} {
		peek := make([]byte, 1)
		return peek
	},
}

func getPeek() []byte {
	return peeks.Get().([]byte)
}

func putPeek(peek []byte) {
	peek[0] = 0
	peeks.Put(peek)
}

type Text struct {
	buffer *bytes.Buffer
	writer io.Writer
}

func (t *Text) Bytes() []byte {
	return t.buffer.Bytes()
}

func (t *Text) Backspace() error {
	if _, err := io.WriteString(t.writer, "\b \b"); err != nil {
		return err
	}
	newSize := t.buffer.Len() - 2
	if newSize < 0 {
		newSize = 0
	}
	t.buffer.Truncate(newSize)
	return nil
}

func (t *Text) Clear() error {
	if _, err := io.WriteString(t.writer, "\u001b[2K"); err != nil {
		return err
	}
	t.buffer.Truncate(0)
	return nil
}

func processCharacters(thread *starlark.Thread, stopPredicate func(text *Text, next byte) bool) (string, error) {
	buffer := getBuffer()
	defer putBuffer(buffer)

	peek := getPeek()
	defer putPeek(peek)

	r := thread.Local("reader").(io.Reader)
	w := thread.Local("writer").(io.Writer)

	text := &Text{
		buffer: buffer,
		writer: w,
	}

	var err error

	for {
		_, err = r.Read(peek)
		if err != nil {
			break
		}

		// handle backspace
		if peek[0] == '\x7f' {
			text.Backspace()
		}

		if err := buffer.WriteByte(peek[0]); err != nil {
			return "", err
		}

		if stopPredicate(text, peek[0]) {
			break
		}
	}

	return buffer.String(), err
}

func matchline(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var expr string

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "s", &expr); err != nil {
		return nil, err
	}

	pattern, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	found, err := processCharacters(thread, func(text *Text, next byte) bool {
		if next == 13 {
			if pattern.Match(text.Bytes()) {
				return true
			} else {
				text.Clear()
			}
		}
		return false
	})
	if err != nil {
		return nil, err
	}

	matches := pattern.FindStringSubmatch(found)

	output := make([]starlark.Value, len(matches))

	for n, match := range matches {
		output[n] = starlark.String(match)
	}

	w := thread.Local("writer").(io.Writer)

	_, err = fmt.Fprint(w, "\r\n")

	return starlark.NewList(output), err
}

func readline(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	found, err := processCharacters(thread, func(text *Text, next byte) bool {
		return next == 13
	})
	if err != nil {
		return nil, err
	}

	w := thread.Local("writer").(io.Writer)

	_, err = fmt.Fprint(w, "\r\n")

	return starlark.String(found), err
}

func init() {
	register(starlark.NewBuiltin("readline", readline))
	register(starlark.NewBuiltin("matchline", matchline))
}
