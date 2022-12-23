package utils

import (
	"io"
	"net/http"
	"sync"
)

func Hijack(writer http.ResponseWriter, request *http.Request) (io.ReadWriteCloser, error) {
	if hj, ok := writer.(http.Hijacker); ok {
		conn, _, err := hj.Hijack()
		return conn, err
	}
	return &h2Stream{writer: writer, request: request}, nil
}

type h2Stream struct {
	init    sync.Once
	flusher http.Flusher
	writer  http.ResponseWriter
	request *http.Request
}

func (h *h2Stream) Read(p []byte) (n int, err error) {
	return h.request.Body.Read(p)
}

func (h *h2Stream) Write(p []byte) (n int, err error) {
	n, err = h.writer.Write(p)
	if err != nil {
		return
	}
	h.init.Do(func() {
		h.flusher, _ = h.writer.(http.Flusher)
	})
	if h.flusher != nil {
		h.flusher.Flush()
	}
	return
}

func (h *h2Stream) Close() error {
	return nil
}
