package s

import (
	"bytes"
	"io"
)

type (
	Expression interface {
		Scan(interface{}) error
		Write(io.Writer) error
	}
	List       []Expression
	Binary     []byte
	String     string
	Number     string
	Identifier string
	True       struct{}
	False      struct{}
)

func NewList(expressions ...Expression) List {
	return List(expressions)
}

func (this Binary) String() string {
	var buf bytes.Buffer
	this.Write(&buf)
	return buf.String()
}
func (this Identifier) String() string {
	var buf bytes.Buffer
	this.Write(&buf)
	return buf.String()
}
func (this False) String() string {
	var buf bytes.Buffer
	this.Write(&buf)
	return buf.String()
}
func (this List) String() string {
	var buf bytes.Buffer
	this.Write(&buf)
	return buf.String()
}
func (this String) String() string {
	var buf bytes.Buffer
	this.Write(&buf)
	return buf.String()
}
func (this True) String() string {
	var buf bytes.Buffer
	this.Write(&buf)
	return buf.String()
}
