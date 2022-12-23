package main

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/lyp256/proxy/protocol/hproxy"
	"github.com/lyp256/proxy/protocol/vless"
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
		proxy:     hproxy.NewProxyHandler(),
		vless:     vless.NewHandler(),
		vlessPath: vlessPath,
		users:     authUsers,
		setHeader: setHeader,
	}
}

type handle struct {
	static    http.Handler
	proxy     http.Handler
	vless     http.Handler
	vlessPath string
	users     auth
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
	if h.vlessPath != "" && request.URL.Path == h.vlessPath {
		h.vless.ServeHTTP(writer, request)
		return
	}
	if h.proxyAuth(request.Header) {
		h.proxy.ServeHTTP(writer, request)
		return
	}
	if h.static != nil {
		h.static.ServeHTTP(writer, request)
		return
	}
	http.NotFound(writer, request)
}

func (h handle) proxyAuth(header http.Header) bool {
	if len(h.users) == 0 {
		return false
	}
	user, passwd, ok := hproxy.GetProxyBaseAuth(header)
	if !ok {
		return false
	}
	return h.users.Auth(user, passwd)
}

func RedirectHTTPS(writer http.ResponseWriter, request *http.Request) {
	u := *request.URL
	u.Host = request.Host
	u.Scheme = "https"
	http.Redirect(writer, request, u.String(), http.StatusPermanentRedirect)
	return
}
