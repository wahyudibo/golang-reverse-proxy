package encoding

import (
	"compress/gzip"
	"io"
)

func Reader(contentEncoding string, body io.ReadCloser) (io.ReadCloser, error) {
	switch contentEncoding {
	case "gzip":
		return gzip.NewReader(body)
	default:
		return body, nil
	}
}

func Writer(contentEncoding string, data []byte, buf io.Writer) (func() error, error) {
	switch contentEncoding {
	case "gzip":
		writer := gzip.NewWriter(buf)
		_, err := writer.Write(data)
		return writer.Close, err
	default:
		_, err := buf.Write(data)
		return func() error { return nil }, err
	}
}
