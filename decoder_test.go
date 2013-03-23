package s

import (
	"bytes"
	"reflect"
	"testing"
)

type (
	basicTest struct {
		input string
		typ reflect.Type
		output interface{}
		value interface{}
	}
)

var (
	basicTests = []basicTest {
		// positive numbers
		{
			`1`,
			reflect.TypeOf(Number{}),
			new(uint8),
			uint8(1),
		},
		{
			`99999999999`,
			reflect.TypeOf(Number{}),
			new(uint8),
			uint8(255),
		},
		{
			`99999999999`,
			reflect.TypeOf(Number{}),
			new(uint16),
			uint16(65535),
		},
		{
			`99999999999`,
			reflect.TypeOf(Number{}),
			new(uint32),
			uint32(4294967295),
		},
		// negative numbers
		{
			`-1`,
			reflect.TypeOf(Number{}),
			new(uint8),
			uint8(0),
		},
		{
			`-1`,
			reflect.TypeOf(Number{}),
			new(int8),
			int8(-1),
		},
		{
			`-999999999999`,
			reflect.TypeOf(Number{}),
			new(int8),
			int8(-128),
		},
		{
			`-999999999999`,
			reflect.TypeOf(Number{}),
			new(int16),
			int16(-32768),
		},
		{
			`-999999999999`,
			reflect.TypeOf(Number{}),
			new(int32),
			int32(-2147483648),
		},
		{
			`-999999999999`,
			reflect.TypeOf(Number{}),
			new(int64),
			int64(-999999999999),
		},
		// floating
		{
			`3.14`,
			reflect.TypeOf(Number{}),
			new(uint8),
			uint8(3),
		},
		{
			`3.14`,
			reflect.TypeOf(Number{}),
			new(float32),
			float32(3.14),
		},
		{
			`-3.14`,
			reflect.TypeOf(Number{}),
			new(float64),
			float64(-3.14),
		},
		// true
		{
			`#t`,
			reflect.TypeOf(True{}),
			new(bool),
			true,
		},
		// false
		{
			`#f`,
			reflect.TypeOf(False{}),
			new(bool),
			false,
		},
		// identifier
		{
			`abcd`,
			reflect.TypeOf(Identifier{}),
			new(string),
			"abcd",
		},
		{
			`!$%&*+-./:<=>?@^_~`,
			reflect.TypeOf(Identifier{}),
			new([]byte),
			[]byte(`!$%&*+-./:<=>?@^_~`),
		},
		// binary
		{
			`#baGVsbG8gd29ybGQ=`,
			reflect.TypeOf(Binary{}),
			new([]byte),
			[]byte(`hello world`),
		},
		// string
		{
			`"test\""`,
			reflect.TypeOf(String{}),
			new(string),
			`test"`,
		},
	}
)

func TestDecoder(t *testing.T) {
	var buf bytes.Buffer
	for _, test := range basicTests {
		buf.WriteString(test.input)
		node, err := Decode(&buf)
		if err != nil {
			t.Errorf("Expected no error got: %v", err)
			continue
		}
		if test.typ != reflect.TypeOf(node) {
			t.Errorf("Expected %v got %T", test.typ, node)
		}
		err = node.Scan(test.output)
		if err != nil {
			t.Errorf("Expected no error got: %v", err)
			continue
		}
		o := reflect.ValueOf(test.output).Elem().Interface()

		if !reflect.DeepEqual(o, test.value) {
			t.Errorf("Expected %v to equal %v", o, test.value)
			continue
		}
	}
}
