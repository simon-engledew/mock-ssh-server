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
	newSize := t.buffer.Len() - 1
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

// dropCR drops a terminal \r from the data.
// see bufio/scan.go
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func processCharacters(w io.Writer, r io.Reader, predicate stopPredicate) (string, error) {
	buffer := getBuffer()
	defer putBuffer(buffer)

	peek := getPeek()
	defer putPeek(peek)

	text := &Text{
		buffer: buffer,
		writer: w,
	}

	var err error
	var stop bool

	for {
		_, err = r.Read(peek)
		if err != nil {
			break
		}

		// handle CTRL-D
		if peek[0] == '\x04' {
			err = io.EOF
			break
		}

		// handle backspace
		if peek[0] == '\x7f' {
			if err = text.Backspace(); err != nil {
				break
			}
			continue
		}

		if stop, err = predicate(text, peek[0]); stop || err != nil {
			break
		}

		if err := buffer.WriteByte(peek[0]); err != nil {
			return "", err
		}
	}

	return string(dropCR(buffer.Bytes())), err
}

type stopPredicate func(text *Text, next byte) (bool, error)

func isNewline(text *Text, next byte) (bool, error) {
	return next == 13, nil
}

func isPattern(expr *regexp.Regexp) stopPredicate {
	return func(text *Text, next byte) (bool, error) {
		return expr.Match(append(text.Bytes(), next)), nil
	}
}

func clearLine(done bool) stopPredicate {
	return func(text *Text, next byte) (bool, error) {
		return done, text.Clear()
	}
}

func whileStop(predicates ...stopPredicate) stopPredicate {
	return func(text *Text, next byte) (bool, error) {
		for _, predicate := range predicates {
			stop, err := predicate(text, next)
			if err != nil {
				return stop, err
			}
			if !stop {
				return false, nil
			}
		}
		return true, nil
	}
}

func untilStop(predicates ...stopPredicate) stopPredicate {
	return func(text *Text, next byte) (bool, error) {
		for _, predicate := range predicates {
			stop, err := predicate(text, next)
			if err != nil {
				return stop, err
			}
			if stop {
				return true, nil
			}
		}
		return false, nil
	}
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

	r := thread.Local("reader").(io.Reader)
	w := thread.Local("writer").(io.Writer)

	found, err := processCharacters(w, r, whileStop(isNewline, untilStop(isPattern(pattern), clearLine(false))))
	if err != nil {
		return nil, err
	}

	matches := pattern.FindStringSubmatch(found)

	output := make([]starlark.Value, len(matches))

	for n, match := range matches {
		output[n] = starlark.String(match)
	}

	_, err = fmt.Fprint(w, "\r\n")

	return starlark.NewList(output), err
}

func readline(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	r := thread.Local("reader").(io.Reader)
	w := thread.Local("writer").(io.Writer)

	found, err := processCharacters(w, r, isNewline)
	if err != nil {
		return nil, err
	}

	_, err = fmt.Fprint(w, "\r\n")

	return starlark.String(found), err
}

func init() {
	register(starlark.NewBuiltin("readline", readline))
	register(starlark.NewBuiltin("matchline", matchline))
}
