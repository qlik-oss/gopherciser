package helpers

import (
	"bytes"
	"testing"
)

func TestBuffer(t *testing.T) {
	expected := "Test1"
	buf := NewBuffer()
	buf.WriteString(expected)

	str := buf.String()
	if str != expected {
		t.Errorf("Test1: Unexpected buffer content, expexted<%s> got<%s>", expected, str)
	}

	expected = "Test2"
	buf.Reset()
	buf.WriteBytes([]byte(expected))
	str = buf.String()
	if str != expected {
		t.Errorf("Test2: Unexpected buffer content, expexted<%s> got<%s>", expected, str)
	}

	expected = "Test3"
	buf = NewBuffer()
	buf.WriteString("T")
	buf.WriteString("e")
	buf.WriteRune(rune('s'))
	buf.WriteByte(byte('t'))
	buf.WriteBytes([]byte("3"))
	str = buf.String()
	if str != expected {
		t.Errorf("Test3: Unexpected buffer content, expexted<%s> got<%s>", expected, str)
	}

	expected = "Test4"
	buf = NewBuffer()
	buf.WriteString(expected)
	buf2 := bytes.NewBuffer(nil)
	buf.WriteTo(buf2)
	str = buf2.String()
	if str != expected {
		t.Errorf("Test4: Unexpected buffer content, expexted<%s> got<%s>", expected, str)
	}
}
