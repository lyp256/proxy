package vless

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/lyp256/proxy/pkg/utils"
)

type Handler interface {
	Handle(ctx context.Context, connect io.ReadWriteCloser) error
}

type handler struct {
	preHandle  func(ctx context.Context, request Requester) error
	postHandle func(ctx context.Context, request Requester, upBytes, downBytes int64, err error)
	dialer     net.Dialer
}

func (h *handler) Handle(ctx context.Context, connect io.ReadWriteCloser) error {
	// vless handshake
	request := requestPool.Get().(*requestInfo)
	defer requestPool.Put(request)
	err := request.FromReader(connect)
	logrus.Debug(request)
	if err != nil {
		return fmt.Errorf("handshake:%w", err)
	}
	// verify
	if h.preHandle != nil {
		err = h.preHandle(ctx, request)
		if err != nil {
			return fmt.Errorf("preVerify:%w", err)
		}
	}
	var network string
	switch request.Command() {
	case TCP:
		network = "tcp"
	case UDP:
		network = "udp"
	default:
		return fmt.Errorf("unknown command:%d", request.Command())
	}

	// replay
	resp := Response{
		Version: request.Version(),
		Addons:  nil,
	}
	err = resp.ToWriter(connect)
	if err != nil {
		return fmt.Errorf("reply :%w", err)
	}

	// dial
	upStream, err := h.dialer.DialContext(ctx, network, request.DestAddr())
	if err != nil {
		return fmt.Errorf("dial %s %s fail:%s", network, request.DestAddr(), err)
	}
	defer upStream.Close()

	up, down, err := utils.Transport(upStream, connect)
	if h.postHandle != nil {
		h.postHandle(ctx, request, up, down, err)
	}
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("traffic forward:%w", err)
	}
	return nil

}

type Option func(*handler)

func WithPreHandle(preHand func(ctx context.Context, request Requester) error) Option {
	return func(h *handler) {
		h.preHandle = preHand
	}
}

func WithPostHandle(postHandle func(ctx context.Context, request Requester, upBytes, downBytes int64, err error)) Option {
	return func(h *handler) {
		h.postHandle = postHandle
	}
}

func WithDial(dialer net.Dialer) Option {
	return func(h *handler) {
		h.dialer = dialer
	}
}

func NewHandler(opts ...Option) Handler {
	h := &handler{
		dialer: net.Dialer{
			Timeout: time.Second * 10,
		},
	}
	for i := range opts {
		if opts[i] != nil {
			opts[i](h)
		}
	}
	return h
}

type httpHandler struct {
	Handler
}

func (h httpHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	connect, err := utils.H2Hijack(writer, request)
	if err != nil {
		return
	}
	defer connect.Close()
	err = h.Handle(request.Context(), connect)
	if err != nil {
		logrus.Info(err)
		http.Error(writer, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}

// NewHTTPHandler create handle over http
func NewHTTPHandler(opts ...Option) http.Handler {
	return httpHandler{
		Handler: NewHandler(opts...),
	}
}
