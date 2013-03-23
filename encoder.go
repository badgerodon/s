package s

import (
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"unicode/utf8"
)

type (
	Encoder interface {
		EncodeS(io.Writer) error
	}
)

var hex = "0123456789abcdef"

func encodeString(dst io.Writer, s string) error {
	dst.Write([]byte{'"'})
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if 0x20 <= b && b != '\\' && b != '"' {
				i++
				continue
			}
			if start < i {
				io.WriteString(dst, s[start:i])
			}
			switch b {
			case '\\', '"':
				dst.Write([]byte{'\\',b})
			case '\n':
				dst.Write([]byte{'\\','n'})
			case '\r':
				dst.Write([]byte{'\\','r'})
			default:
				// This encodes bytes < 0x20 except for \n and \r,
				// as well as < and >. The latter are escaped because they
				// can lead to security holes when user-controlled strings
				// are rendered into JSON and served to some browsers.
				io.WriteString(dst, `\u00`)
				dst.Write([]byte{hex[b>>4], hex[b&0xF]})
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			return fmt.Errorf("Invalid UTF8 %v", c)
		}
		i += size
	}
	if start < len(s) {
		io.WriteString(dst, s[start:])
	}
	dst.Write([]byte{'"'})
	return nil
}

func encodeValue(dst io.Writer, src reflect.Value) error {
	var err error

	if src.CanInterface() {
		encoder, ok := src.Interface().(Encoder)
		if !ok {
			if src.Kind() != reflect.Ptr && src.CanAddr() {
				encoder, ok = src.Addr().Interface().(Encoder)
				if ok {
					src = src.Addr()
				}
			}
		}
		if ok && (src.Kind() != reflect.Ptr || !src.IsNil()) {
			return encoder.EncodeS(dst)
		}
	}

	switch src.Kind() {
	case reflect.Bool:
		if src.Bool() {
			_, err = io.WriteString(dst, "#t")
		} else {
			_, err = io.WriteString(dst, "#f")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, err = io.WriteString(dst, strconv.FormatInt(src.Int(), 10))
  case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		_, err = io.WriteString(dst, strconv.FormatUint(src.Uint(), 10))
  case reflect.Float32, reflect.Float64:
  	f := src.Float()
  	if math.IsInf(f, 0) {
  		err = fmt.Errorf("Infinity is not supported")
  	} else if math.IsNaN(f) {
  		err = fmt.Errorf("NaN is not supported")
  	} else {
			_, err = io.WriteString(dst, strconv.FormatFloat(f, 'f', -1, 64))
  	}
  case reflect.String:
  	err = encodeString(dst, src.String())
	case reflect.Struct:
		_, err = dst.Write([]byte{'('})
		if err != nil {
			return err
		}
		for i := 0; i < src.NumField(); i++ {
			if i > 0 {
				_, err = dst.Write([]byte{' '})
				if err != nil {
					return err
				}
			}
			err = encodeValue(dst, src.Field(i))
			if err != nil {
				return err
			}
		}
		_, err = dst.Write([]byte{')'})
	case reflect.Map:
		if src.IsNil() {
			_, err = io.WriteString(dst, "()")
			break
		}
		_, err = dst.Write([]byte{'('})
		if err != nil {
			return err
		}
		for i, k := range src.MapKeys() {
			if i > 0 {
				_, err = dst.Write([]byte{' '})
				if err != nil {
					return err
				}
			}
			_, err = dst.Write([]byte{'('})
			if err != nil {
				return err
			}
			err = encodeValue(dst, k)
			if err != nil {
				return err
			}
			_, err = dst.Write([]byte{' '})
			if err != nil {
				return err
			}
			err = encodeValue(dst, src.MapIndex(k))
			if err != nil {
				return err
			}
			_, err = dst.Write([]byte{')'})
			if err != nil {
				return err
			}
		}
		_, err = dst.Write([]byte{')'})
	case reflect.Slice:
		if src.IsNil() {
			_, err = io.WriteString(dst, "()")
			break
		}
		fallthrough
	case reflect.Array:
		_, err = dst.Write([]byte{'('})
		if err != nil {
			return err
		}
		for i := 0; i < src.Len(); i++ {
			if i > 0 {
				_, err = dst.Write([]byte{' '})
				if err != nil {
					return err
				}
			}
			err = encodeValue(dst, src.Index(i))
			if err != nil {
				return err
			}
		}
		_, err = dst.Write([]byte{')'})
	case reflect.Interface, reflect.Ptr:
		if src.IsNil() {
			_, err = io.WriteString(dst, "()")
			break
		}
		err = encodeValue(dst, src.Elem())
	default:
		err = fmt.Errorf("Unable to convert `%v` of type `%v` into s expression", src, src.Kind())
	}
	return err
}

func EncodeList(dst io.Writer, values ... interface{}) error {
	return encodeValue(dst, reflect.ValueOf(values))
}

func Encode(dst io.Writer, src interface{}) error {
	return encodeValue(dst, reflect.ValueOf(src))
}
