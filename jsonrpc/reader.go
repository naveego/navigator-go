package jsonrpc

import (
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
)

var (
	cr = []byte("\r")[0]
	lf = []byte("\n")[0]
)

type RequestReader interface {
	ReadRequest() (Request, bool, error)
	ReadResponse() (Response, bool, error)
}

func NewRequestBuffer(data []byte) *RequestBuffer {
	return &RequestBuffer{
		buffer: data,
		index:  len(data),
	}
}

type RequestBuffer struct {
	buffer []byte
	index  int
}

func (rb *RequestBuffer) Append(data []byte) {
	for _, d := range data {
		rb.buffer = append(rb.buffer, d)
	}
	rb.index = len(rb.buffer)
}

func (rb *RequestBuffer) ReadHeaders() (map[string]string, bool, error) {
	headers := map[string]string{}

	c := 0
	b := rb.buffer

	for {
		if c+3 < rb.index && (b[c] != cr || b[c+1] != lf || b[c+2] != cr || b[c+3] != lf) {
			c++
			continue
		}
		break
	}

	if c+3 > rb.index {
		return headers, false, nil
	}

	hs := strings.Split(string(b[:c]), "\r\n")
	for _, h := range hs {
		i := strings.Index(h, ":")
		if i == -1 {
			return headers, false, errors.New("invalid message header format")
		}

		key := h[:i]
		val := strings.TrimSpace(h[i+1:])
		headers[key] = val
	}

	nextStart := c + 4
	rb.buffer = b[nextStart:]
	rb.index = rb.index - nextStart

	return headers, true, nil
}

func (rb *RequestBuffer) ReadContent(length int) ([]byte, bool) {
	if rb.index < length {
		return nil, false
	}

	s := rb.buffer[:length]
	rb.buffer = rb.buffer[length:]
	rb.index = rb.index - length
	return s, true
}

func NewRequestReader(reader io.Reader) RequestReader {
	return &ioRequestReader{
		reader: reader,
		buf:    &RequestBuffer{},
	}
}

type ioRequestReader struct {
	reader io.Reader
	buf    *RequestBuffer
}

func (i *ioRequestReader) ReadRequest() (Request, bool, error) {
	var req Request
	var headers map[string]string

	contentLength := 0
	hasReadHeaders := false

	for {
		dataBuf := make([]byte, 512)
		bc, err := i.reader.Read(dataBuf)
		if err != nil {
			return Request{}, false, err
		}

		logrus.Debugf("Appending data: %d", bc)
		i.buf.Append(dataBuf[:bc])

		if !hasReadHeaders {
			var exist bool
			headers, exist, err = i.buf.ReadHeaders()
			if err != nil {
				return Request{}, false, err
			}
			if !exist {
				continue
			}

			contentLengthStr, ok := headers["Content-Length"]
			if !ok {
				return Request{}, false, errors.New("no content length was provided")
			}
			contentLength, err = strconv.Atoi(contentLengthStr)
			if err != nil {
				return Request{}, false, errors.New("invalid content length")
			}

		}
		hasReadHeaders = true

		if i.buf.index < contentLength {
			continue
		}

		content, success := i.buf.ReadContent(contentLength)
		if !success {
			return Request{}, false, errors.New("could not read content")
		}

		req = Request{}
		err = json.Unmarshal(content, &req)

		if err != nil {
			logrus.Debugf("content: %s - %s", string(content), err.Error())
			return Request{}, false, errors.New("content was not valid json")
		}
		break
	}

	return req, true, nil
}

func (i *ioRequestReader) ReadResponse() (Response, bool, error) {
	var req Response
	var headers map[string]string

	contentLength := 0
	hasReadHeaders := false

	for {
		dataBuf := make([]byte, 512)
		bc, err := i.reader.Read(dataBuf)
		if err != nil {
			return Response{}, false, err
		}

		logrus.Debugf("Appending data: %d", bc)
		i.buf.Append(dataBuf[:bc])

		if !hasReadHeaders {
			var exist bool
			headers, exist, err = i.buf.ReadHeaders()
			if err != nil {
				return Response{}, false, err
			}
			if !exist {
				continue
			}

			contentLengthStr, ok := headers["Content-Length"]
			if !ok {
				return Response{}, false, errors.New("no content length was provided")
			}
			contentLength, err = strconv.Atoi(contentLengthStr)
			if err != nil {
				return Response{}, false, errors.New("invalid content length")
			}

		}
		hasReadHeaders = true

		if i.buf.index < contentLength {
			continue
		}

		content, success := i.buf.ReadContent(contentLength)
		if !success {
			return Response{}, false, errors.New("could not read content")
		}

		req = Response{}
		err = json.Unmarshal(content, &req)

		if err != nil {
			logrus.Debugf("content: %s - %s", string(content), err.Error())
			return Response{}, false, errors.New("content was not valid json")
		}
		break
	}

	return req, true, nil
}
