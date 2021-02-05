package goembed

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

func BytesToList(data []byte) string {
	var ar []string
	for _, v := range data {
		ar = append(ar, fmt.Sprintf("%d", v))
	}
	return strings.Join(ar, ",")
}

func BytesToHex(data []byte) string {
	var buf bytes.Buffer
	WriteToHex(data, &buf)
	return buf.String()
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
