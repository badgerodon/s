package s

import (
	"bytes"
	"testing"
)

type (
	encodeTestCase struct {
		src interface{}
		val string
	}
)

var (
	encodeTestCases = []encodeTestCase{
		{ true, "#t" },
		{ false, "#f" },
		{ uint8(1), "1" },
		{ uint16(1), "1" },
		{ uint32(1), "1" },
		{ uint64(1), "1" },
		{ int8(-1), "-1" },
		{ int16(-1), "-1" },
		{ int32(-1), "-1" },
		{ int64(-1), "-1" },
		{ float64(-1.1), "-1.1" },
		{ [4]int{1,2,3,4}, "(1 2 3 4)"},
		{ []float64{1,2,3,4}, "(1 2 3 4)"},
		{ "test", `"test"` },
		{ struct{x,y int}{25, 34}, "(25 34)" },
		{ new(int), "0" },
	}
)

func TestEncoder(t *testing.T) {
	for _, tc := range encodeTestCases {
		var buf bytes.Buffer
		err := Encode(&buf, tc.src)
		if err != nil {
			t.Errorf("Expected no error got: %v", err)
		}
		val := buf.String()
		if val != tc.val {
			t.Errorf("Expected `%v` got `%v`", tc.val, val)
		}
	}
}
