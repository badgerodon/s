package s

import (
	"bytes"
	"testing"
)

func TestEncoder(t *testing.T) {
	type testCase struct {
		Value  interface{}
		Result string
	}
	testCases := []testCase{
		{true, "#t"},
		{false, "#f"},
		{uint8(1), "1"},
		{uint16(1), "1"},
		{uint32(1), "1"},
		{uint64(1), "1"},
		{int8(-1), "-1"},
		{int16(-1), "-1"},
		{int32(-1), "-1"},
		{int64(-1), "-1"},
		{float64(-1.1), "-1.1"},
		{[4]int{1, 2, 3, 4}, "(1 2 3 4)"},
		{[]float64{1, 2, 3, 4}, "(1 2 3 4)"},
		{"test", `"test"`},
		{struct{ x, y int }{25, 34}, "(25 34)"},
		{new(int), "0"},
	}
	for _, tc := range testCases {
		var buf bytes.Buffer
		exp, err := Encode(tc.Value)
		if err != nil {
			t.Errorf("Expected no error got %v for %v", err, tc.Value)
			continue
		}
		err = exp.Write(&buf)
		if err != nil {
			t.Errorf("Expected no error got %v for %v", err, tc.Value)
			continue
		}
		val := buf.String()
		if val != tc.Result {
			t.Errorf("Expected `%v` got `%v` for %v", tc.Result, val, tc.Value)
		}
	}
}
