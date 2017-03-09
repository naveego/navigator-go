package jsonrpc

import (
	"fmt"
	"io"
)

func NewResponseWriter(writer io.Writer) *ResponseWriter {
	return &ResponseWriter{
		writer: writer,
	}
}

type ResponseWriter struct {
	writer io.Writer
}

func (r *ResponseWriter) WriteResponse(res Response) error {

	content, err := SerializeResponse(res)
	if err != nil {
		return err
	}

	headers := fmt.Sprintf("Content-Length: %d\r\n\r\n", len([]byte(content)))
	resBytes := append([]byte(headers), []byte(content)...)

	_, err = r.writer.Write(resBytes)
	if err != nil {
		return err
	}

	return nil
}

func (r *ResponseWriter) WriteRequest(req Request) error {
	content, err := SerializeRequest(req)
	if err != nil {
		return err
	}

	headers := fmt.Sprintf("Content-Length: %d\r\n\r\n", len([]byte(content)))
	resBytes := append([]byte(headers), []byte(content)...)

	_, err = r.writer.Write(resBytes)
	if err != nil {
		return err
	}

	return nil
}
