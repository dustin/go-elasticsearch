package elasticsearch

import (
	"io"
)

type bytesource struct {
	ch      <-chan []byte
	current []byte
}

func (b *bytesource) Read(out []byte) (int, error) {
	if len(b.current) == 0 {
		var ok bool
		b.current, ok = <-b.ch
		if !ok {
			return 0, io.EOF
		}
	}
	copied := copy(out, b.current)
	b.current = b.current[copied:]
	return copied, nil
}

func (b *bytesource) Close() {
	// Can't really do anything here.
}

func NewByteSource(from <-chan []byte) *bytesource {
	return &bytesource{ch: from}
}
