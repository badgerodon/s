package s

import (
	"bufio"
)

type (
	readerTill struct {
		reader *bufio.Reader
		atEnd  func(byte) bool
	}
)

func (this *readerTill) Read(p []byte) (int, error) {
	var err error
	read := 0
	var tmp []byte

	for i := 0; i < len(p); i++ {
		tmp, err = this.reader.Peek(1)
		if len(tmp) > 0 {
			if this.atEnd(tmp[0]) {
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
