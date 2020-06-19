package main // import "github.com/simon-engledew/mock-ssh-server"
import (
	"bytes"
	"fmt"
	"github.com/gliderlabs/ssh"
	"go.starlark.net/starlark"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"io"
	"log"
	"os"
	"regexp"
	"sync"
)

type Script func(w io.Writer, r io.Reader)

func Actor(script Script, enc encoding.Encoding) ssh.Handler {
	return func(s ssh.Session) {
		log.Print("client connected: ", s.RemoteAddr())
		script(enc.NewEncoder().Writer(s), enc.NewDecoder().Reader(s))
	}
}

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

const expect = "expect"

func processCharacters(thread *starlark.Thread, stopPredicate func(current *bytes.Buffer, next byte) bool) (string, error) {
	buffer := getBuffer()
	defer putBuffer(buffer)

	peek := getPeek()
	defer putPeek(peek)

	r := thread.Local("reader").(io.Reader)
	w := thread.Local("writer").(io.Writer)

	var err error

	for {
		_, err = r.Read(peek)
		if err != nil {
			break
		}

		if peek[0] == '\x7f' {
			_, err = io.WriteString(w, "\b \b")
			if err != nil {
				break
			}
			newSize := buffer.Len() - 2
			if newSize < 0 {
				newSize = 0
			}
			buffer.Truncate(newSize)
		}

		if err := buffer.WriteByte(peek[0]); err != nil {
			return "", err
		}

		if stopPredicate(buffer, peek[0]) {
			break
		}
	}

	return buffer.String(), err
}

func expectFn(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var expr string

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "s", &expr); err != nil {
		return nil, err
	}

	pattern, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	found, err := processCharacters(thread, func(current *bytes.Buffer, next byte) bool {
		return pattern.Match(current.Bytes())
	})
	if err != nil {
		return nil, err
	}

	matches := pattern.FindStringSubmatch(found)

	output := make([]starlark.Value, len(matches))

	for n, match := range matches {
		output[n] = starlark.String(match)
	}

	return starlark.NewList(output), err
}

const readline = "readline"

func readlineFn(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	found, err := processCharacters(thread, func(current *bytes.Buffer, next byte) bool {
		return next == 13
	})
	if err != nil {
		return nil, err
	}

	return starlark.String(found), err
}

const write = "write"

func writeFn(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	w := thread.Local("writer").(io.Writer)

	var s string

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "s", &s); err != nil {
		return nil, err
	}

	n, err := fmt.Fprint(w, s)

	return starlark.MakeInt(n), err
}

const writeline = "writeline"

func writelineFn(thread *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	w := thread.Local("writer").(io.Writer)

	var s string

	if err := starlark.UnpackArgs(b.Name(), args, kwargs, "s?", &s); err != nil {
		return nil, err
	}

	n, err := fmt.Fprint(w, s+"\r\n")

	return starlark.MakeInt(n), err
}

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		log.Fatal("usage: gosshim <script.star>")
	}

	script := os.Args[1]

	ssh.Handle(Actor(func(w io.Writer, r io.Reader) {
		echoReader := io.TeeReader(r, w)

		thread := &starlark.Thread{Name: "SSH Session"}
		thread.SetLocal("reader", echoReader)
		thread.SetLocal("writer", w)

		_, err := starlark.ExecFile(thread, script, nil, starlark.StringDict{
			expect:    starlark.NewBuiltin(expect, expectFn),
			readline:  starlark.NewBuiltin(readline, readlineFn),
			writeline: starlark.NewBuiltin(writeline, writelineFn),
			write:     starlark.NewBuiltin(write, writeFn),
		})
		if err != nil {
			fmt.Fprint(w, err.Error()+"\r\n")
			log.Print(err)
		}
	}, unicode.UTF8))

	log.Fatal(ssh.ListenAndServe(":2222", nil))
}
