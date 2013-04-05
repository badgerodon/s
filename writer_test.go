package s

import (
	"bytes"
	"testing"
)

func TestWriters(t *testing.T) {
	type testCase struct {
		Expression
		Result string
	}
	cases := []testCase{
		{Binary([]byte("abcd")), `#bYWJjZA==`},
		{False{}, `#f`},
		{Identifier("test"), `test`},
		{NewList(String("test")), `("test")`},
		{NewList(True{}, False{}, String("a")), `(#t #f "a")`},
		{Number("1.23"), `1.23`},
		{String(`"a"`), `"\"a\""`},
		{String("\n"), `"\n"`},
		{String("\r"), `"\r"`},
		{True{}, `#t`},
	}

	for _, c := range cases {
		var buf bytes.Buffer
		err := c.Expression.Write(&buf)
		if err != nil {
			t.Errorf("Expected err to be nil got %v in %v", err, c.Expression)
		}
		str := buf.String()
		if str != c.Result {
			t.Errorf("Expected %v got %v in %v", c.Result, str, c.Expression)
		}
	}
}
