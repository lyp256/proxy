package proxy

import (
	"encoding/base64"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

type HTTPOptions func(handler *httpHandler)

func WithLog(entry logrus.Ext1FieldLogger) HTTPOptions {
	return func(handler *httpHandler) {
		if entry != nil {
			handler.log = entry
		}
	}
}

func NewHTTPProxyHandler(options ...HTTPOptions) http.Handler {
	h := &httpHandler{
		log: logrus.StandardLogger(),
	}
	for _, opt := range options {
		opt(h)
	}
	return h
}

type httpHandler struct {
	log logrus.Ext1FieldLogger
}

func (p *httpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	p.log.Infof("%s %s %s %s", request.RemoteAddr, request.Proto, request.Method, request.Host)
	if request.Method == http.MethodConnect {
		err := p.tunnel(writer, request)
		if err != nil {
			p.log.Error(err)
		}
		return
	}
	err := p.transfer(writer, request)
	if err != nil {
		p.log.Error(err)
		return
	}
}

// 隧道代理
func (p *httpHandler) tunnel(writer http.ResponseWriter, request *http.Request) error {
	d := net.Dialer{}
	// 服务端建立连接
	upstream, err := d.DialContext(request.Context(), "tcp", request.URL.Host)
	if err != nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
		return err
	}
	defer upstream.Close()

	// 劫持客户端连接
	var downstream io.ReadWriteCloser
	writer.WriteHeader(http.StatusOK)
	if flush, ok := writer.(http.Flusher); ok {
		flush.Flush()
	}
	downstream, err = hijack(writer, request)
	if err != nil {
		return err
	}
	defer downstream.Close()
	_, _, err = transport(upstream, downstream)
	return err
}

// 中间人代理
func (p *httpHandler) transfer(writer http.ResponseWriter, request *http.Request) error {
	newRequest, err := fromRequest(request)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(newRequest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	copyHeader(writer.Header(), resp.Header, true)
	writer.WriteHeader(resp.StatusCode)

	_, _ = io.Copy(writer, resp.Body)
	if f, ok := writer.(http.Flusher); ok {
		f.Flush()
	}
	return nil
}

func copyHeader(dst, src http.Header, withProxy bool) {
	for key, values := range src {
		if !withProxy && strings.HasPrefix(strings.ToLower(key), "proxy-") {
			continue
		}
		for _, val := range values {
			dst.Add(key, val)
		}
	}
}

func fromRequest(r *http.Request) (*http.Request, error) {
	u := url.URL{
		Scheme:     r.URL.Scheme,
		User:       r.URL.User,
		Host:       r.URL.Host,
		Path:       r.URL.Path,
		RawPath:    r.URL.RawPath,
		ForceQuery: false,
		RawQuery:   r.URL.RawQuery,
	}
	if u.Host == "" {
		u.Host = r.Host
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}

	newReq, err := http.NewRequestWithContext(r.Context(), r.Method, u.String(), r.Body)
	if err != nil {
		return nil, err
	}
	copyHeader(newReq.Header, r.Header, false)
	return newReq, nil
}

func GetProxyBaseAuth(header http.Header) (username, password string, ok bool) {
	auth := header.Get("Proxy-Authorization")
	if auth == "" {
		return
	}

	const prefix = "basic "
	if len(auth) < len(prefix) || strings.ToLower(auth[:len(prefix)]) != prefix {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}

func hijack(writer http.ResponseWriter, request *http.Request) (io.ReadWriteCloser, error) {
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
