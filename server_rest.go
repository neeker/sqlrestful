/*********************************************
                   _ooOoo_
                  o8888888o
                  88" . "88
                  (| -_- |)
                  O\  =  /O
               ____/`---'\____
             .'  \\|     |//  `.
            /  \\|||  :  |||//  \
           /  _||||| -:- |||||-  \
           |   | \\\  -  /// |   |
           | \_|  ''\---/''  |   |
           \  .-\__  `-`  ___/-. /
         ___`. .'  /--.--\  `. . __
      ."" '<  `.___\_<|>_/___.'  >'"".
     | | :  `- \`.;`\ _ /`;.`/ - ` : | |
     \  \ `-.   \_ __\ /__ _/   .-` /  /
======`-.____`-.___\_____/___.-`____.-'======
                   `=---='

^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
           佛祖保佑       永无BUG
           心外无法       法外无心
           三宝弟子       飞猪宏愿
*********************************************/

package main

import (
	"log"
	"net/url"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// startRestfulServer - stop RESTful server
func stopRestfulServer() {
	echoServer.Close()
}

// startRestfulServer - start RESTful server
func startRestfulServer() error {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.CORS())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 9}))
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = customHTTPErrorHandler

	routeBase := macrosManager.ServiceBasePath()

	if routeBase == "/" {
		routeBase = ""
	}

	if macrosManager.IsSwaggerEnabled() {
		//启用swagger-ui的webJar
		e.Static(routeBase+"/webjars", "/swagger2/webjars")
		e.File(routeBase+"/doc.html", "/swagger2/doc.html")
		e.File(routeBase+"/swagger-resources/configuration/ui", "/swagger2/ui.json")
		e.File(routeBase+"/swagger-resources", "/swagger2/swagger-resources.json")
		e.File(routeBase+"/swagger-resources/configuration/security", "/swagger2/security.json")
	}

	e.GET(routeBase+"/v2/api-docs", routeAPIDocs)

	//添加默认路由
	e.GET(routeBase+"/health", routeHealth)

	if macrosManager.DatabaseConfig().IsRedisEnabled() {
		e.POST(routeBase+"clear_caches", routeClearCaches)
	}

	//添加微服务接口路由
	for _, macro := range macrosManager.List() {
		if macro.IsDir() {
			e.Static(routeBase+macro.Path, macro.Dir)
			continue
		}

		if macro.IsFile() {
			e.File(routeBase+macro.Path, macro.File)
			continue
		}

		if macro.IsProxy() {
			proxyTargets := []*middleware.ProxyTarget{}
			for _, proxyAddress := range macro.Proxy {
				u, err := url.Parse(proxyAddress)
				if err != nil {
					log.Printf("%s defined proxy(%s) was not URL!", macro.name, proxyAddress)
					continue
				}
				proxyTargets = append(proxyTargets, &middleware.ProxyTarget{
					URL: u,
				})
			}
			g := e.Group(routeBase + macro.Path)
			g.Use(middleware.Proxy(middleware.NewRoundRobinBalancer(proxyTargets)))
			continue
		}

		if len(macro.Exec) > 0 {
			if macro.IsWebsocket() {
				macro.websocket = NewWSClientRegistry(macro.name, macro.Websocket.Keepalive)
			}
			if len(macro.Methods) > 0 {
				for _, method := range macro.Methods {
					method = strings.ToUpper(method)
					switch {
					case method == "GET":
						e.GET(routeBase+macro.Path, routeExecMacro, getMiddlewareAuthorizeFunc(macro))
					case method == "POST":
						e.POST(routeBase+macro.Path, routeExecMacro, getMiddlewareAuthorizeFunc(macro))
					case method == "PUT":
						e.PUT(routeBase+macro.Path, routeExecMacro, getMiddlewareAuthorizeFunc(macro))
					case method == "PATCH":
						e.PATCH(routeBase+macro.Path, routeExecMacro, getMiddlewareAuthorizeFunc(macro))
					case method == "DELETE":
						e.DELETE(routeBase+macro.Path, routeExecMacro, getMiddlewareAuthorizeFunc(macro))
					}
				}
				continue
			}
			e.GET(routeBase+macro.Path, routeExecMacro, getMiddlewareAuthorizeFunc(macro))
			continue
		}

		for method, childMacro := range macro.methodMacros {
			if macro.IsWebsocket() {
				childMacro.websocket = NewWSClientRegistry(childMacro.name, macro.Websocket.Keepalive)
			}
			switch {
			case method == "GET":
				e.GET(routeBase+macro.Path, routeExecMacro, getMiddlewareAuthorizeFunc(childMacro))
			case method == "POST":
				e.POST(routeBase+macro.Path, routeExecMacro, getMiddlewareAuthorizeFunc(childMacro))
			case method == "PUT":
				e.PUT(routeBase+macro.Path, routeExecMacro, getMiddlewareAuthorizeFunc(childMacro))
			case method == "PATCH":
				e.PATCH(routeBase+macro.Path, routeExecMacro, getMiddlewareAuthorizeFunc(childMacro))
			case method == "DELETE":
				e.DELETE(routeBase+macro.Path, routeExecMacro, getMiddlewareAuthorizeFunc(childMacro))
			}
		}
	}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))

	echoServer = e
	return e.Start(*flagRESTListenAddr)
}

// getMiddlewareAuthorizeFunc
func getMiddlewareAuthorizeFunc(macro *Macro) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return middlewareAuthorize(macro, next)
	}
}

// middlewareAuthorize - the authorizer middleware
func middlewareAuthorize(macro *Macro, next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if len(macro.Methods) < 1 {
			macro.Methods = []string{c.Request().Method}
		}

		methodIsAllowed := false
		for _, method := range macro.Methods {
			method = strings.ToUpper(method)
			if c.Request().Method == method {
				methodIsAllowed = true
				break
			}
		}

		if !methodIsAllowed {
			return c.JSON(405, map[string]interface{}{
				"code":    405,
				"message": "方法不允许！",
			})
		}

		c.Set("macro", macro)

		return next(c)
	}
}

var (
	echoServer *echo.Echo
)
