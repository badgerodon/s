package s

import (
  "bufio"
  "bytes"
  "encoding/base64"
  "fmt"
  "io"
  "io/ioutil"
  "strconv"
  "strings"
  "unicode"
)

type (
  stringReader struct {
    reader *bufio.Reader
    done bool
  }
  oneTimeReader struct {
    reader *bufio.Reader
    done bool
    atEnd func(byte)bool
    consumeLastByte bool
    includeLastByte bool
  }

  Node interface {
    Scan(interface{}) error
  }
  List struct {
    reader *bufio.Reader
    current Node
  }
  Binary struct {
    reader io.Reader
  }
  String struct {
    reader io.Reader
  }
  Number struct {
    value string
  }
  Identifier struct {
    value string
  }
  True struct {}
  False struct {}

  nothing struct{}
)

var (
  MAX_IDENTIFIER_LENGTH = 256
  MAX_NUMBER_LENGTH = 64
  extended map[rune]nothing
)

func init() {
  extended = map[rune]nothing{}
  for _, c := range []rune{'!','$','%','&','*','+','-','.','/',':','<','=','>','?','@','^','_','~'} {
    extended[c] = nothing{}
  }
}

func (this *oneTimeReader) Read(p []byte) (int, error) {
  if this.done {
    return 0, io.EOF
  }

  var err error
  read := 0
  var tmp []byte

  for i := 0; i < len(p); i++ {
    tmp, err = this.reader.Peek(1)
    if len(tmp) > 0 {
      if this.atEnd(tmp[0]) {
        if this.includeLastByte {
          p[i] = tmp[0]
        }
        if this.consumeLastByte {
          this.reader.Read(tmp)
        }
        this.done = true
        break
      }
      p[i] = tmp[0]
      read++
      this.reader.Read(tmp)
    }
    if err != nil {
      break
    }
  }

  return read, err
}

func (this *stringReader) Read(p []byte) (int, error) {
  if this.done {
    return 0, io.EOF
  }

  var err error
  var read int
  var tmp []byte

  for i := 0; i < len(p); i++ {
    tmp, err = this.reader.Peek(1)
    if len(tmp) > 0 {
      switch tmp[0] {
      case '"':
        this.reader.Read(tmp)
        this.done = true
        return read, io.EOF
      case '\\':
        this.reader.Read(tmp)
        tmp, err = this.reader.Peek(1)
        if len(tmp) > 0 {
          switch tmp[0] {
          case 'n':
            p[i] = '\n'
          case 'r':
            p[i] = '\r'
          default:
            p[i] = tmp[0]
          }
        }
      default:
        p[i] = tmp[0]
      }
    }
    if err != nil {
      break
    }
    read++
    this.reader.Read(tmp)
  }

  return read, err
}

func newList(reader *bufio.Reader) *List {
  return &List{
    bufio.NewReader(&oneTimeReader{
      reader: reader,
      done: false,
      atEnd: func(b byte) bool {
        return b == ')'
      },
      consumeLastByte: true,
      includeLastByte: false,
    }),
    nil,
  }
}

func newString(reader *bufio.Reader) String {
  return String{
    &stringReader{
      reader: reader,
      done: false,
    },
  }
}

func newBinary(reader *bufio.Reader) Binary {
  limitedReader := bufio.NewReader(&oneTimeReader{
    reader: reader,
    done: false,
    atEnd: func(b byte) bool {
      if (b >= 'a' && b <= 'z') ||
        (b >= 'A' && b <= 'Z') ||
        (b >= '0' && b <= '9') ||
        (b == '+') || (b == '/') || (b == '=') {
        return false
      }

      return true
    },
    consumeLastByte: true,
    includeLastByte: false,
  })
  base64Decoder := base64.NewDecoder(base64.StdEncoding, limitedReader)
  return Binary{base64Decoder}
}

func isWhitespace(ch rune) bool {
  return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
func isDigit(ch rune) bool {
  return '0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch)
}
func isLetter(ch rune) bool {
  return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch >= 0x80 && unicode.IsLetter(ch)
}
func isExtended(ch rune) bool {
  _, ok := extended[ch]
  return ok
}

func skipWhitespace(reader *bufio.Reader) error {
  for {
    r, _, err := reader.ReadRune()
    if err != nil {
      return err
    }
    if !isWhitespace(r) {
      return reader.UnreadRune()
    }
  }
  return nil
}

func readNumber(initial rune, reader *bufio.Reader) (string, error) {
  var r rune
  var err error

  seenPeriod := false
  seenNegative := false

  buf := bytes.NewBuffer(make([]byte, 0, MAX_NUMBER_LENGTH))
  buf.WriteRune(initial)
  for buf.Len() < MAX_NUMBER_LENGTH {
    r, _, err = reader.ReadRune()
    if r == '-' && !seenNegative {
    } else if r == '.' && !seenPeriod {
      seenPeriod = true
    } else if !isDigit(r) {
      reader.UnreadRune()
      break
    }
    buf.WriteRune(r)
    if err != nil {
      break
    }
    // only allowed at the start
    seenNegative = true
  }

  return buf.String(), err
}

func readIdentifier(initial rune, reader *bufio.Reader) (string, error) {
  var r rune
  var err error

  buf := bytes.NewBuffer(make([]byte, 0, MAX_IDENTIFIER_LENGTH))
  buf.WriteRune(initial)
  for buf.Len() < MAX_IDENTIFIER_LENGTH {
    r, _, err = reader.ReadRune()
    if !(isDigit(r) || isLetter(r) || isExtended(r)) {
      reader.UnreadRune()
      break
    }
    buf.WriteRune(r)
    if err != nil {
      break
    }
  }

  return buf.String(), err
}

func readNode(current Node, reader *bufio.Reader) (Node, error) {
  var token Node
  var value string

  // Exhaust the current token if it hasn't been used yet
  if current != nil {
    if rdr, ok := current.(io.Reader); ok {
      buf := make([]byte, 1)
      for {
        _, err := rdr.Read(buf)
        if err == io.EOF {
          break
        }
        if err != nil {
          return token, err
        }
      }
    }
  }

  err := skipWhitespace(reader)
  if err != nil {
    return token, err
  }

  r, _, err := reader.ReadRune()
  if err != nil {
    return token, err
  }

  switch {
  case isDigit(r) || r == '-':
    value, err = readNumber(r, reader)
    token = Number{value}
  case isLetter(r) || isExtended(r):
    value, err = readIdentifier(r, reader)
    token = Identifier{value}
  default:
    switch r {
    case '(':
      token = newList(reader)
    case '#':
      r, _, err = reader.ReadRune()
      if err != nil {
        return token, err
      }
      switch r {
      case 't':
        token = True{}
      case 'f':
        token = False{}
      case 'b':
        token = newBinary(reader)
      }
    case '"':
      token = newString(reader)
    }
  }


  if token != nil {
    return token, nil
  }

  return nil, fmt.Errorf("Unknown token %v", r)
}

func Decode(reader io.Reader) (Node, error) {
  return readNode(nil, bufio.NewReader(reader))
}

func (this *List) Next() (Node, error) {
  node, err := readNode(this.current, this.reader)
  this.current = node
  return node, err
}
func (this *List) Scan(dst interface{}) error {
  return fmt.Errorf("Unimplemented")
}
func (this String) Scan(dst interface{}) error {
  var err error
  switch t := dst.(type) {
  case *interface{}:
    *t, err = ioutil.ReadAll(this.reader)
  case io.Writer:
    io.Copy(t, this.reader)
  case *[]byte:
    *t, err = ioutil.ReadAll(this.reader)
  case *string:
    var buf bytes.Buffer
    io.Copy(&buf, this.reader)
    *t = buf.String()
  default:
    err = fmt.Errorf("Cannot convert string into %T", dst)
  }
  return err
}
func (this Binary) Scan(dst interface{}) error {
  var err error
  switch t := dst.(type) {
  case *interface{}:
    *t, err = ioutil.ReadAll(this.reader)
  case io.Writer:
    io.Copy(t, this.reader)
  case *[]byte:
    *t, err = ioutil.ReadAll(this.reader)
  default:
    err = fmt.Errorf("Cannot convert binary into %T", dst)
  }
  return err
}
func (this Identifier) Scan(dst interface{}) error {
  switch t := dst.(type) {
  case *interface{}:
    *t = this.value
  case *string:
    *t = this.value
  case *[]byte:
    *t = []byte(this.value)
  default:
    return fmt.Errorf("Cannot convert identifier into %T", dst)
  }
  return nil
}
func (this Number) Scan(dst interface{}) error {
  var vi int64
  var vui uint64
  var vf float64

  p := strings.IndexRune(this.value, '.')
  if p == -1 {
    vi, _ = strconv.ParseInt(this.value, 10, 64)
    vui, _ = strconv.ParseUint(this.value, 10, 64)
  } else {
    vi, _ = strconv.ParseInt(this.value[:p], 10, 64)
    vui, _ = strconv.ParseUint(this.value[:p], 10, 64)
  }
  vf, _ = strconv.ParseFloat(this.value, 64)

  switch t := dst.(type) {
  case *interface{}:
    if p == -1 {
      *t = vi
    } else {
      *t = vf
    }
  case *uint8:
    var max uint8 = 1<<8-1
    if vui > uint64(max) {
      *t = max
    } else {
      *t = uint8(vui)
    }
  case *uint16:
    var max uint16 = 1<<16-1
    if vui > uint64(max) {
      *t = max
    } else {
      *t = uint16(vui)
    }
  case *uint32:
    var max uint32 = 1<<32-1
    if vui > uint64(max) {
      *t = max
    } else {
      *t = uint32(vui)
    }
  case *uint64:
    *t = vui
  case *int8:
    var max int8 = 1<<7-1
    var min = -max - 1
    if vi > int64(max) {
      *t = max
    } else if vi < int64(min) {
      *t = min
    } else {
      *t = int8(vi)
    }
  case *int16:
    var max int16 = 1<<15-1
    var min = -max - 1
    if vi > int64(max) {
      *t = max
    } else if vi < int64(min) {
      *t = min
    } else {
      *t = int16(vi)
    }
  case *int32:
    var max int32 = 1<<31-1
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
  case *float32:
    *t = float32(vf)
  case *float64:
    *t = vf
  default:
    return fmt.Errorf("Cannot convert number to %T", dst)
  }
  return nil
}
func (this True) Scan(dst interface{}) error {
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
func (this False) Scan(dst interface{}) error {
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
