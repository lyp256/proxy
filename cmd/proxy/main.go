package main

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/lucas-clemente/quic-go/http3"
	"github.com/lyp256/proxy/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

func main() {
	var (
		httpAddr   string
		httpsAddr  string
		certFile   string
		keyFile    string
		StaticRoot string
		authUsers  []string
		logLevel   string
		h2         bool
		h3         bool
		vless      string
		insecure   bool
		v          bool
	)
	pflag.StringVar(&httpAddr, "http", "", "listen http port. example :80")
	pflag.StringVar(&httpsAddr, "https", "", "listen https port. example :443")
	pflag.StringVar(&certFile, "cert", "", "tls cert file")
	pflag.StringVar(&keyFile, "key", "", "tls key file")
	pflag.StringVar(&StaticRoot, "static", "", "static resource root. example /srv/www")
	pflag.StringSliceVarP(&authUsers, "user", "u", nil, "proxy auth user. example admin:123456")
	pflag.StringVar(&logLevel, "log", "", "log level. one of panic,fatal,error,warning,info,debug,trace")
	pflag.StringVar(&vless, "vless", "", "vless path. example /vless")
	pflag.BoolVar(&h2, "http2", true, "enable http2")
	pflag.BoolVar(&h3, "http3", true, "enable http3")
	pflag.BoolVar(&insecure, "insecure", false, "enable proxy protocols on insecure connections")
	pflag.BoolVarP(&v, "version", "v", false, "print version")
	pflag.Parse()
	if v {
		version.Print(os.Stdout)
		return
	}

	if httpAddr == "" && httpsAddr == "" {
		pflag.Usage()
		return
	}
	if httpsAddr != "" {
		switch {
		case certFile == "":
			logrus.Fatal("cert file must be specified")
		case keyFile == "":
			logrus.Fatal("key file must be specified")
		}
	}

	if logLevel != "" {
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.SetLevel(level)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	var (
		setHeader  func(handlerFunc http.Header) error
		quicServer *http3.Server
	)

	if h3 && httpsAddr != "" {
		quicServer = &http3.Server{}
		setHeader = quicServer.SetQuicHeaders
	}

	h := proxyHandler(StaticRoot, vless, parseAuthUsers(authUsers), setHeader, insecure)

	wg := sync.WaitGroup{}

	if httpAddr != "" {
		l, err := net.Listen("tcp", httpAddr)
		if err != nil {
			logrus.Fatal(err)
		}
		s := http.Server{}
		if insecure || httpsAddr == "" {
			s.Handler = h
		} else {
			s.Handler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				u := request.URL
				u.Scheme = "https"
				http.Redirect(writer, request, u.String(), http.StatusTemporaryRedirect)
			})
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = s.Serve(l)
			if err != nil && err != http.ErrServerClosed {
				logrus.Fatal(err)
			}
			cancel()
		}()
		go func() {
			<-ctx.Done()
			_ = s.Shutdown(context.TODO())
		}()
	}

	if quicServer != nil {
		quicServer.Server = &http.Server{
			Addr:    httpsAddr,
			Handler: h,
		}
		wg.Add(1)
		go func() {

			defer wg.Done()
			err := quicServer.ListenAndServeTLS(certFile, keyFile)
			if err != nil {
				logrus.Fatal(err)
			}
			cancel()
		}()
		go func() {
			<-ctx.Done()
			_ = quicServer.Shutdown(context.TODO())
		}()
	}

	if httpsAddr != "" {
		l, err := net.Listen("tcp", httpsAddr)
		if err != nil {
			logrus.Fatal(err)
		}
		s := &http.Server{Handler: h}
		if !h2 {
			s.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			s := http.Server{Handler: h}
			err = s.ServeTLS(l, certFile, keyFile)
			if err != nil && err != http.ErrServerClosed {
				logrus.Fatal(err)
			}
			cancel()
		}()
		go func() {
			<-ctx.Done()
			_ = s.Shutdown(context.TODO())
		}()
	}
	wg.Wait()
}
