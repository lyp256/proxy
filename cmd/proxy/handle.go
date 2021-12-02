package main

import (
	"net/http"

	"github.com/lyp256/proxy"
	"github.com/sirupsen/logrus"
)

func proxyHandler(
	staticRoot,
	vlessPath string,
	authUsers auth,
	setHeader func(header http.Header) error,
	insecure bool) *handle {
	var staticHandle http.Handler
	if staticRoot != "" {
		staticHandle = http.FileServer(http.Dir(staticRoot))
	}
	return &handle{
		static:    staticHandle,
		proxy:     proxy.NewHTTPProxyHandler(),
		vless:     proxy.NewVLESSHandler(),
		vlessPath: vlessPath,
		users:     authUsers,
		insecure:  insecure,
		setHeader: setHeader,
	}
}

type handle struct {
	static    http.Handler
	proxy     http.Handler
	vless     http.Handler
	vlessPath string
	users     auth
	insecure  bool
	setHeader func(header http.Header) error
}

func (h handle) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if h.setHeader != nil {
		err := h.setHeader(writer.Header())
		if err != nil {
			logrus.Error(err)
		}
	}
	h.serveHTTP(writer, request)
}

func (h handle) serveHTTP(writer http.ResponseWriter, request *http.Request) {
	if h.insecure || request.TLS != nil {
		if h.vlessPath != "" && request.URL.Path == h.vlessPath {
			h.vless.ServeHTTP(writer, request)
			return
		}
		if h.proxyAuth(request.Header) {
			h.proxy.ServeHTTP(writer, request)
			return
		}
	}
	if h.static != nil {
		h.static.ServeHTTP(writer, request)
		return
	}
	http.NotFound(writer, request)
}

func (h handle) proxyAuth(header http.Header) bool {
	if len(h.users) == 0 {
		return true
	}
	user, passwd, ok := proxy.GetProxyBaseAuth(header)
	if !ok {
		return false
	}
	return h.users.Auth(user, passwd)
}
