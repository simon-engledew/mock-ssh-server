package pkg

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNewLine(t *testing.T) {
	buffer := new(bytes.Buffer)
	found, err := processCharacters(buffer, strings.NewReader("Line\r\n"), isNewline)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "Line", found, "unexpected value")
	assert.Equal(t, "", buffer.String(), "Clear should have been written")
}

func TestBackspace(t *testing.T) {
	buffer := new(bytes.Buffer)
	found, err := processCharacters(buffer, strings.NewReader("Background\u007F\u007F\u007F\u007F\u007F\u007Fspace\r\n"), isNewline)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "Backspace", found, "backspace should have been applied")
	assert.Equal(t, "\b \b\b \b\b \b\b \b\b \b\b \b", buffer.String(), "backspace should have been written")
}

func TestClear(t *testing.T) {
	buffer := new(bytes.Buffer)
	found, err := processCharacters(buffer, strings.NewReader("Line\r\n"), whileStop(isNewline, clearLine(true)))
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "", found, "unexpected value")
	assert.Equal(t, "\u001B[2K", buffer.String(), "Clear should have been written")
}
