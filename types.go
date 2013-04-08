package s

import (
	"bytes"
	"fmt"
	"io"
)

type (
	Expression interface {
		Scan(...interface{}) error
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
func (this List) Head() (Expression, error) {
	if len(this) == 0 {
		return nil, fmt.Errorf("Empty List")
	}
	return this[0], nil
}
func (this List) Tail() (List, error) {
	if len(this) == 0 {
		return nil, fmt.Errorf("Empty list")
	}
	return List(this[1:]), nil
}
func (this List) Prepend(exp Expression) List {
	return List(append([]Expression{exp}, this...))
}
func (this List) Append(exp Expression) List {
	return List(append(this, exp))
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
