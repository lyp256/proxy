package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/lyp256/proxy/pkg/auth"
	"github.com/lyp256/proxy/pkg/config"
	"github.com/lyp256/proxy/pkg/httpproxy"
	"github.com/lyp256/proxy/pkg/registry"
	"github.com/lyp256/proxy/pkg/vless"
)

func newEngine(
	conf *config.Config) *gin.Engine {
	e := gin.New()

	if conf.Component.Proxy.Enable {
		proxyConf := conf.Component.Proxy
		h := httpproxy.NewProxyHandler()
		e.Use(func(context *gin.Context) {
			if httpproxy.ProxyRequest(context.Request) {
				switch proxyConf.Auth {
				case auth.GlobalUser, "":
					username, passwd, ok := httpproxy.GetProxyBaseAuth(context.Request.Header)
					if !ok {
						if proxyConf.Active {
							context.Status(http.StatusProxyAuthRequired)
							context.Abort()
							return
						}
						return
					}
					if auth.Auth(username, passwd) {
						h.ServeHTTP(context.Writer, context.Request)
						context.Abort()
					}
					return
				case "none":
				default:
					return
				}
			}
		})
	}

	if conf.Component.Vless.Enable {
		vlessConf := conf.Component.Vless
		// uuid auth
		h := vless.NewHTTPHandler(vless.WithPreHandle(func(ctx context.Context, request vless.Requester) error {
			user := getUser(ctx)
			if user != nil {
				return nil
			}
			uid := request.UUID()
			user = auth.GetByUUID(uid.String())
			if user == nil {
				return fmt.Errorf("not get user：%s", uid)
			}
			return nil
		}))

		e.Any(conf.Component.Vless.Path, func(c *gin.Context) {
			switch vlessConf.Auth {
			case auth.GlobalUser, "":
				username := c.Param("username")
				password := c.Param("password")
				if user := auth.GetByUsername(username); user != nil && user.Password == password {
					c.Request = c.Request.WithContext(withUser(c, user))
				}
				h.ServeHTTP(c.Writer, c.Request)
			case "none":
				c.Request = c.Request.WithContext(withUser(c, &auth.User{Username: "anonymous"}))
				h.ServeHTTP(c.Writer, c.Request)
			default:
				c.AbortWithStatus(http.StatusNotFound)
			}
			c.Abort()
		})
	}

	if conf.Component.Registry.Enable {
		h := gin.WrapF(func(writer http.ResponseWriter, request *http.Request) {
			rh := registry.NewHTTPHandler(&conf.Component.Registry.Config)
			rh.ServeHTTP(writer, request)
		})
		e.Any("/v2/*all", h)
	}
	if conf.Component.Static.Enable {
		e.NoRoute(gin.WrapH(http.FileServer(http.Dir(conf.Component.Static.Root))))
	}
	return e
}

func RedirectHTTPS(writer http.ResponseWriter, request *http.Request) {
	u := *request.URL
	u.Host = request.Host
	u.Scheme = "https"
	http.Redirect(writer, request, u.String(), http.StatusPermanentRedirect)
	return
}
