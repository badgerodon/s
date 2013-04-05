package s

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"unicode/utf8"
)

var hex = "0123456789abcdef"

func (this Binary) Write(dst io.Writer) (err error) {
	_, err = io.WriteString(dst, "#b")
	if err != nil {
		return
	}
	enc := base64.NewEncoder(base64.StdEncoding, dst)
	_, err = enc.Write([]byte(this))
	if err != nil {
		return
	}
	err = enc.Close()
	return
}
func (this False) Write(dst io.Writer) (err error) {
	_, err = io.WriteString(dst, "#f")
	return
}
func (this Identifier) Write(dst io.Writer) (err error) {
	_, err = io.WriteString(dst, string(this))
	return
}
func (this List) Write(dst io.Writer) (err error) {
	_, err = io.WriteString(dst, "(")
	if err != nil {
		return
	}
	for i, exp := range this {
		if i > 0 {
			_, err = io.WriteString(dst, " ")
			if err != nil {
				return
			}
		}
		err = exp.Write(dst)
		if err != nil {
			return
		}
	}
	_, err = io.WriteString(dst, ")")
	return
}
func (this Number) Write(dst io.Writer) (err error) {
	_, err = io.WriteString(dst, string(this))
	return
}
func (this String) Write(dst io.Writer) (err error) {
	w := bufio.NewWriter(dst)
	err = w.WriteByte('"')
	if err != nil {
		return
	}
	s := string(this)
	start := 0
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			if 0x20 <= b && b != '\\' && b != '"' {
				i++
				continue
			}
			if start < i {
				_, err = io.WriteString(w, s[start:i])
				if err != nil {
					return
				}
			}
			switch b {
			case '\\', '"':
				_, err = w.Write([]byte{'\\', b})
			case '\n':
				_, err = w.Write([]byte{'\\', 'n'})
			case '\r':
				_, err = w.Write([]byte{'\\', 'r'})
			default:
				_, err = w.Write([]byte{'\\', 'u', '0', '0', hex[b>>4], hex[b&0xF]})
			}
			if err != nil {
				return
			}
			i++
			start = i
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			err = fmt.Errorf("Invalid UTF8 %v", c)
			return
		}
		i += size
	}
	if start < len(s) {
		_, err = io.WriteString(w, s[start:])
		if err != nil {
			return
		}
	}
	err = w.WriteByte('"')
	if err != nil {
		return
	}
	err = w.Flush()
	return
}
func (this True) Write(dst io.Writer) (err error) {
	_, err = io.WriteString(dst, "#t")
	return
}
