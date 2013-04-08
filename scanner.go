package s

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"reflect"
)

func (this List) Scan(dst ... interface{}) error {
	// do nothing if the list is empty
	if len(this) == 0 {
		return nil
	}

	// one argument
	if len(dst) == 1 {
		val := reflect.ValueOf(dst[0])
		// pointer to a struct
		if val.Kind() == reflect.Ptr {
			if e := val.Elem(); e.Kind() == reflect.Struct {
				n := len(this)
				if n > e.NumField() {
					n = e.NumField()
				}
				for i := 0; i < n; i++ {
					f := e.Field(i).Addr().Interface()
					err := this[i].Scan(f)
					if err != nil {
						return err
					}
				}
				return nil
			}
		}
	}

	n := len(this)
	if n > len(dst) {
		n = len(dst)
	}
	for i := 0; i < n; i++ {
		err := this[i].Scan(dst[i])
		if err != nil {
			return err
		}
	}
	return nil
}
func (this String) Scan(dsts ... interface{}) error {
	if len(dsts) == 0 {
		return fmt.Errorf("Expected at least one argument")
	}
	dst := dsts[0]

	var err error
	switch t := dst.(type) {
	case *interface{}:
		*t = string(this)
	case io.Writer:
		io.WriteString(t, string(this))
	case *[]byte:
		*t = []byte(this)
	case *string:
		*t = string(this)
	default:
		err = fmt.Errorf("Cannot convert string into %T", dst)
	}
	return err
}
func (this Binary) Scan(dsts ... interface{}) error {
	if len(dsts) == 0 {
		return fmt.Errorf("Expected at least one argument")
	}
	dst := dsts[0]

	var err error
	switch t := dst.(type) {
	case *interface{}:
		*t = []byte(this)
	case io.Writer:
		_, err = t.Write([]byte(this))
	case *[]byte:
		*t = []byte(this)
	default:
		err = fmt.Errorf("Cannot convert binary into %T", dst)
	}
	return err
}
func (this True) Scan(dsts ... interface{}) error {
	if len(dsts) == 0 {
		return fmt.Errorf("Expected at least one argument")
	}
	dst := dsts[0]

	switch t := dst.(type) {
	case *interface{}:
		*t = true
	case *bool:
		*t = true
	default:
		return fmt.Errorf("Cannot convert true to %T", dst)
	}
	return nil
}
func (this False) Scan(dsts ... interface{}) error {
	if len(dsts) == 0 {
		return fmt.Errorf("Expected at least one argument")
	}
	dst := dsts[0]

	switch t := dst.(type) {
	case *interface{}:
		*t = false
	case *bool:
		*t = false
	default:
		return fmt.Errorf("Cannot convert false to %T", dst)
	}
	return nil
}
func (this Number) Scan(dsts ... interface{}) error {
	if len(dsts) == 0 {
		return fmt.Errorf("Expected at least one argument")
	}
	dst := dsts[0]

	var vi int64
	var vui uint64
	var vf float64

	str := string(this)

	p := strings.IndexRune(str, '.')
	if p == -1 {
		vi, _ = strconv.ParseInt(str, 10, 64)
		vui, _ = strconv.ParseUint(str, 10, 64)
	} else {
		vi, _ = strconv.ParseInt(str[:p], 10, 64)
		vui, _ = strconv.ParseUint(str[:p], 10, 64)
	}
	vf, _ = strconv.ParseFloat(str, 64)

	switch t := dst.(type) {
	case *interface{}:
		if p == -1 {
			*t = vi
		} else {
			*t = vf
		}
	case *uint8:
		var max uint8 = 1<<8 - 1
		if vui > uint64(max) {
			*t = max
		} else {
			*t = uint8(vui)
		}
	case *uint16:
		var max uint16 = 1<<16 - 1
		if vui > uint64(max) {
			*t = max
		} else {
			*t = uint16(vui)
		}
	case *uint32:
		var max uint32 = 1<<32 - 1
		if vui > uint64(max) {
			*t = max
		} else {
			*t = uint32(vui)
		}
	case *uint64:
		*t = vui
	case *int8:
		var max int8 = 1<<7 - 1
		var min = -max - 1
		if vi > int64(max) {
			*t = max
		} else if vi < int64(min) {
			*t = min
		} else {
			*t = int8(vi)
		}
	case *int16:
		var max int16 = 1<<15 - 1
		var min = -max - 1
		if vi > int64(max) {
			*t = max
		} else if vi < int64(min) {
			*t = min
		} else {
			*t = int16(vi)
		}
	case *int32:
		var max int32 = 1<<31 - 1
		var min = -max - 1
		if vi > int64(max) {
			*t = max
		} else if vi < int64(min) {
			*t = min
		} else {
			*t = int32(vi)
		}
	case *int64:
		*t = vi
	case *int:
		*t = int(vi)
	case *float32:
		*t = float32(vf)
	case *float64:
		*t = vf
	default:
		return fmt.Errorf("Cannot convert number to %T", dst)
	}
	return nil
}
func (this Identifier) Scan(dsts ... interface{}) error {
	if len(dsts) == 0 {
		return fmt.Errorf("Expected at least one argument")
	}
	dst := dsts[0]
	switch t := dst.(type) {
	case *interface{}:
		*t = string(this)
	case *string:
		*t = string(this)
	case *[]byte:
		*t = []byte(this)
	default:
		return fmt.Errorf("Cannot convert identifier into %T", dst)
	}
	return nil
}
