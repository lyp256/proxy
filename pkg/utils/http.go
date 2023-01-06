package utils

import (
	"io"
	"net/http"
)

func Hijack(writer http.ResponseWriter, request *http.Request) (io.ReadWriteCloser, error) {
	if request.ProtoMajor < 2 {
		if hj, ok := writer.(http.Hijacker); ok {
			conn, _, err := hj.Hijack()
			return conn, err
		}
	}

	f, ok := writer.(http.Flusher)
	if ok {
		f.Flush()
	}
	return &h2Stream{writer: writer, request: request, flusher: f}, nil
}

func H2Hijack(writer http.ResponseWriter, request *http.Request) (io.ReadWriteCloser, error) {
	writer.WriteHeader(http.StatusContinue)
	f, ok := writer.(http.Flusher)
	if ok {
		f.Flush()
	}
	return &h2Stream{writer: writer, request: request, flusher: f}, nil
}

type h2Stream struct {
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
	if h.flusher != nil {
		h.flusher.Flush()
	}
	return
}

func (h *h2Stream) Close() error {
	return nil
}
