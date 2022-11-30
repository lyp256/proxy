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
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/lyp256/proxy/version"
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
	pflag.StringVar(&httpAddr, "http", ":80", "listen http port. example :80")
	pflag.StringVar(&httpsAddr, "https", ":443", "listen https port. example :443")
	pflag.StringVar(&certFile, "cert", "", "tls cert file")
	pflag.StringVar(&keyFile, "key", "", "tls key file")
	pflag.StringVar(&StaticRoot, "static", "", "static resource root. example /srv/www")
	pflag.StringSliceVarP(&authUsers, "user", "u", nil, "proxy auth user. example admin:123456")
	pflag.StringVar(&logLevel, "log", "warning", "log level. one of panic,fatal,error,warning,info,debug,trace")
	pflag.StringVar(&vless, "vless", "/vless", "vless path. example /vless")
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
		setHeader func(handlerFunc http.Header) error
		h3Server  *http3.Server
	)

	if h3 && httpsAddr != "" {
		h3Server = &http3.Server{
			Addr: httpsAddr,
		}
		setHeader = h3Server.SetQuicHeaders
	}

	proxyHandle := proxyHandler(StaticRoot, vless, parseAuthUsers(authUsers), setHeader, insecure)

	wg := sync.WaitGroup{}

	if httpsAddr != "" {
		l, err := net.Listen("tcp", httpsAddr)
		if err != nil {
			logrus.Fatal(err)
		}
		httpsServer := &http.Server{Handler: proxyHandle}
		if !h2 {
			httpsServer.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = httpsServer.ServeTLS(l, certFile, keyFile)
			if err != nil && err != http.ErrServerClosed {
				logrus.Fatal(err)
			}
			cancel()
		}()
		go func() {
			<-ctx.Done()
			_ = httpsServer.Shutdown(context.TODO())
		}()
	}

	if h3Server != nil {
		h3Server.Handler = proxyHandle
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := h3Server.ListenAndServeTLS(certFile, keyFile)
			if err != nil {
				logrus.Fatal(err)
			}
			cancel()
		}()
		go func() {
			<-ctx.Done()
			_ = h3Server.Close()
		}()
	}

	if httpAddr != "" {
		l, err := net.Listen("tcp", httpAddr)
		if err != nil {
			logrus.Fatal(err)
		}
		httpServer := &http.Server{
			Handler: proxyHandle,
		}
		if !insecure {
			httpServer.Handler = http.HandlerFunc(RedirectHTTPS)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = httpServer.Serve(l)
			if err != nil && err != http.ErrServerClosed {
				logrus.Fatal(err)
			}
			cancel()
		}()
		go func() {
			<-ctx.Done()
			_ = httpServer.Shutdown(context.TODO())
		}()
	}
	wg.Wait()
}

func init() {
	logrus.StandardLogger().SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		TimestampFormat: "15:04:05",
	})
	logrus.StandardLogger().SetReportCaller(true)
}
