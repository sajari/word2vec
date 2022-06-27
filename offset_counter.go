package word2vec

import "bufio"

type offsetCounter struct {
	offset int64
	reader *bufio.Reader
}

func (o *offsetCounter) Discard(n int) (discarded int, err error) {
	discarded, err = o.reader.Discard(n)
	o.offset += int64(discarded)
	return
}

func (o *offsetCounter) ReadByte() (byte, error) {
	b, err := o.reader.ReadByte()
	if err == nil {
		o.offset++
	}
	return b, err
}

func (o *offsetCounter) ReadRune() (r rune, size int, err error) {
	r, size, err = o.reader.ReadRune()
	o.offset += int64(size)
	return
}

func (o *offsetCounter) ReadSlice(delim byte) (line []byte, err error) {
	line, err = o.reader.ReadSlice(delim)
	o.offset += int64(len(line))
	return
}

func (o *offsetCounter) ReadBytes(delim byte) ([]byte, error) {
	b, err := o.reader.ReadBytes(delim)
	o.offset += int64(len(b))
	return b, err
}

func (o *offsetCounter) ReadString(delim byte) (string, error) {
	s, err := o.reader.ReadString(delim)
	o.offset += int64(len(s))
	return s, err
}

func (o *offsetCounter) Read(p []byte) (n int, err error) {
	n, err = o.reader.Read(p)
	o.offset += int64(n)
	return
}

func (o *offsetCounter) UnreadByte() error {
	err := o.reader.UnreadByte()
	if err == nil {
		o.offset--
	}
	return err
}
