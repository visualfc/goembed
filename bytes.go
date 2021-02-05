package goembed

import (
	"bytes"
	"io"
)

func BytesToHex(data []byte) (string, error) {
	var buf bytes.Buffer
	_, err := WriteToHex(data, &buf)
	return buf.String(), err
}

const hex = "0123456789abcdef"

func WriteToHex(data []byte, w io.Writer) (n int, err error) {
	buf := []byte{'\\', 'x', 0, 0}
	for _, d := range data {
		buf[2], buf[3] = hex[d/16], hex[d%16]
		_, err = w.Write(buf)
		if err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}
