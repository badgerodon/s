package s

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
)

type (
	Reader  struct{ *bufio.Reader }
	nothing struct{}
)

var (
	MAX_IDENTIFIER_LENGTH = 256
	MAX_NUMBER_LENGTH     = 64
	extended              map[rune]nothing
)

func init() {
	extended = map[rune]nothing{}
	for _, c := range []rune{'!', '$', '%', '&', '*', '+', '-', '.', '/', ':', '<', '=', '>', '?', '@', '^', '_', '~'} {
		extended[c] = nothing{}
	}
}

func isWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}
func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}
func isExtended(ch rune) bool {
	_, ok := extended[ch]
	return ok
}

func (this Reader) skipWhitespace() error {
	for {
		r, _, err := this.ReadRune()
		if err != nil {
			return err
		}
		if !isWhitespace(r) {
			return this.UnreadRune()
		}
	}
	return nil
}

func (this Reader) readString() (String, error) {
	var buf bytes.Buffer
	var err error
	var n int
	tmp := make([]byte, 1)
	escape := false

outer:
	for {
		n, err = this.Read(tmp)
		if err != nil {
			break outer
		}
		if n == 0 {
			err = io.EOF
			break outer
		}

		if escape {
			switch tmp[0] {
			case 'n':
				err = buf.WriteByte('\n')
			case 'r':
				err = buf.WriteByte('\r')
			default:
				err = buf.WriteByte(tmp[0])
			}
			escape = false
		} else {
			switch tmp[0] {
			case '"':
				break outer
			case '\\':
				escape = true
			default:
				err = buf.WriteByte(tmp[0])
			}
		}

		if err != nil {
			break outer
		}
	}

	return String(buf.String()), err
}
func (this Reader) readIdentifier(initial rune) (Identifier, error) {
	var r rune
	var err error

	buf := bytes.NewBuffer(make([]byte, 0, MAX_IDENTIFIER_LENGTH))
	buf.WriteRune(initial)
	for buf.Len() < MAX_IDENTIFIER_LENGTH {
		r, _, err = this.ReadRune()
		if !(isDigit(r) || isLetter(r) || isExtended(r)) {
			this.UnreadRune()
			break
		}
		buf.WriteRune(r)
		if err != nil {
			break
		}
	}

	return Identifier(buf.String()), err
}
func (this Reader) readList() (List, error) {
	var err error
	var lst List
	var exp Expression
	var b byte
	exps := []Expression{}

	for {
		// Skip opening whitespace
		err = this.skipWhitespace()
		if err != nil {
			return lst, err
		}

		b, err = this.ReadByte()
		if err != nil {
			return lst, err
		}
		if b == ')' {
			break
		}
		err = this.UnreadByte()
		if err != nil {
			return lst, err
		}

		exp, err = this.readExpression()
		if err != nil {
			return lst, err
		}
		exps = append(exps, exp)
	}

	lst = List(exps)
	return lst, nil
}
func (this Reader) readBinary() (Binary, error) {
	rdr := &readerTill{
		reader: this.Reader,
		atEnd: func(b byte) bool {
			if (b >= 'a' && b <= 'z') ||
				(b >= 'A' && b <= 'Z') ||
				(b >= '0' && b <= '9') ||
				(b == '+') || (b == '/') || (b == '=') {
				return false
			}

			return true
		},
	}

	base64Decoder := base64.NewDecoder(base64.StdEncoding, rdr)
	bs, err := ioutil.ReadAll(base64Decoder)
	return Binary(bs), err
}
func (this Reader) readNumber(initial rune) (Number, error) {
	var r rune
	var err error

	seenPeriod := false
	seenNegative := false

	buf := bytes.NewBuffer(make([]byte, 0, MAX_NUMBER_LENGTH))
	buf.WriteRune(initial)
	for buf.Len() < MAX_NUMBER_LENGTH {
		r, _, err = this.ReadRune()
		if r == '-' && !seenNegative {
		} else if r == '.' && !seenPeriod {
			seenPeriod = true
		} else if !isDigit(r) {
			this.UnreadRune()
			break
		}
		buf.WriteRune(r)
		if err != nil {
			break
		}
		// only allowed at the start
		seenNegative = true
	}

	return Number(buf.String()), err
}
func (this Reader) readExpression() (Expression, error) {
	var exp Expression
	var err error

	// Skip opening whitespace
	err = this.skipWhitespace()
	if err != nil {
		return exp, err
	}

	// Read the next character
	r, _, err := this.ReadRune()
	if err != nil {
		return exp, err
	}

	switch {
	// Numbers
	case isDigit(r) || r == '-':
		exp, err = this.readNumber(r)
	// Identifiers
	case isLetter(r) || isExtended(r):
		exp, err = this.readIdentifier(r)
	default:
		switch r {
		case '(':
			exp, err = this.readList()
		case '#':
			r, _, err = this.ReadRune()
			if err != nil {
				return exp, err
			}
			switch r {
			case 't':
				exp = True{}
			case 'f':
				exp = False{}
			case 'b':
				exp, err = this.readBinary()
			}
		case '"':
			exp, err = this.readString()
		}
	}

	if err == io.EOF {
		err = nil
	}

	if exp != nil {
		return exp, err
	}

	return nil, fmt.Errorf("Unknown token %v", r)
}

func Read(reader io.Reader) (Expression, error) {
	return Reader{bufio.NewReader(reader)}.readExpression()
}
