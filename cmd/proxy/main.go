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

	"github.com/gin-gonic/gin"
	"github.com/quic-go/quic-go/http3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"

	"github.com/lyp256/proxy/pkg/auth"
	"github.com/lyp256/proxy/pkg/config"
	"github.com/lyp256/proxy/version"
)

func main() {
	var (
		configPath              string
		versionDump, configDump bool
	)
	pflag.StringVarP(&configPath, "config", "c", "config.yaml", "configure file")
	pflag.BoolVarP(&versionDump, "version", "v", false, "print version")
	pflag.BoolVarP(&configDump, "dump", "d", false, "print config")

	pflag.Parse()
	if versionDump {
		version.Print(os.Stdout)
		return
	}

	conf := config.Default()

	err := conf.LoadFile(configPath)
	if err != nil {
		logrus.Fatal(err)
	}

	if configDump {
		os.Stdout.Write(conf.Marshal())
		return
	}

	if conf.HTTP == "" && conf.HTTPS == "" {
		pflag.Usage()
		return
	}
	if conf.HTTPS != "" {
		switch {
		case conf.TLS.Cert == "":
			logrus.Fatal("cert file must be specified")
		case conf.TLS.Key == "":
			logrus.Fatal("key file must be specified")
		}
	}

	if conf.LOG != "" {
		level, err := logrus.ParseLevel(conf.LOG)
		if err != nil {
			logrus.Fatal(err)
		}
		logrus.SetLevel(level)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	auth.Load(conf.Users...)

	en := newEngine(conf)

	var (
		h3Server    *http3.Server
		httpsServer *http.Server
		httpServer  *http.Server
	)
	if conf.EnableHTTP3 && conf.HTTPS != "" {
		h3Server = &http3.Server{
			Addr: conf.HTTPS,
		}
		en.Use(func(c *gin.Context) {
			err := h3Server.SetQUICHeaders(c.Writer.Header())
			if err != nil {
				logrus.Errorf("setQuicHeaders:%s", err)
			}
		})
	}

	wg := sync.WaitGroup{}

	if conf.HTTPS != "" {
		l, err := net.Listen("tcp", conf.HTTPS)
		if err != nil {
			logrus.Fatal(err)
		}
		httpsServer = &http.Server{Handler: en}
		if !conf.EnableHTTP2 {
			httpsServer.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = httpsServer.ServeTLS(l, conf.TLS.Cert, conf.TLS.Key)
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
		h3Server.Handler = en
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := h3Server.ListenAndServeTLS(conf.TLS.Cert, conf.TLS.Key)
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

	if conf.HTTP != "" {
		l, err := net.Listen("tcp", conf.HTTP)
		if err != nil {
			logrus.Fatal(err)
		}
		httpServer = &http.Server{
			Handler: en,
		}
		if !conf.Insecure && conf.HTTPS != "" {
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
