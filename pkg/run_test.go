package pkg

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	fn, err := Load("TestLoad", `
writeline("What is your name?")
name = readline()
writeline("Hello " + name + "!")
`)
	if err != nil {
		t.Error(err)
		return
	}

	buffer := new(bytes.Buffer)
	if err := fn(buffer, strings.NewReader("Test\r\n")); err != nil {
		t.Error(err)
	}

	assert.Equal(t, "What is your name?\r\nTest\r\r\nHello Test!\r\n", buffer.String(), "unexpected script output")
}
