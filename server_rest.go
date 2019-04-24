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
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// initialize RESTful server
func initRestfulServer() error {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.CORS())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 9}))
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = customHTTPErrorHandler

	routeBase := *flagBasePath

	if routeBase == "/" {
		routeBase = ""
	}

	//启用swagger-ui的webJar
	e.Static(routeBase + "/webjars", "/swagger2/webjars");
	e.File(routeBase + "/swagger-ui.html", "/swagger2/swagger-ui.html")
	e.File(routeBase + "/swagger-resources/configuration/ui", "/swagger2/ui.json")
	e.File(routeBase + "/swagger-resources", "/swagger2/swagger-resources.json")
	e.File(routeBase + "/swagger-resources/configuration/security", "/swagger2/security.json")

	e.GET(routeBase + "/v2/api-docs", routeAPIDocs)

	//添加默认路由
	e.GET(routeBase + "/", routeIndex)
	e.GET(routeBase + "/health", routeHealth)
	e.POST(routeBase + "clear_caches", routeClearCaches)

	//添加微服务接口路由
	for _, macro := range macrosManager.List() {
		if (len(macro.Exec) > 0) {
			if (len(macro.Methods) > 0) {
				for _, method := range macro.Methods {
					method = strings.ToUpper(method)
					switch {
					case method == "GET" :
						e.GET(routeBase + macro.Path, routeExecMacro, func(next echo.HandlerFunc) echo.HandlerFunc { 
							return middlewareAuthorize(macro, next)
						})
					case method == "POST" :
						e.POST(routeBase + macro.Path, routeExecMacro, func(next echo.HandlerFunc) echo.HandlerFunc { 
							return middlewareAuthorize(macro, next)
						})
					case method == "PUT" :
						e.POST(routeBase + macro.Path, routeExecMacro, func(next echo.HandlerFunc) echo.HandlerFunc { 
							return middlewareAuthorize(macro, next)
						})
					case method == "PATCH" :
						e.PATCH(routeBase + macro.Path, routeExecMacro, func(next echo.HandlerFunc) echo.HandlerFunc { 
							return middlewareAuthorize(macro, next)
						})
					case method == "DELETE" :
						e.DELETE(routeBase + macro.Path, routeExecMacro, func(next echo.HandlerFunc) echo.HandlerFunc { 
							return middlewareAuthorize(macro, next)
						})
					}
				}
			} else {
				e.GET(routeBase + macro.Path, routeExecMacro, func(next echo.HandlerFunc) echo.HandlerFunc { 
					return middlewareAuthorize(macro, next)
				})
			}
		} else {
			for method, childMacro := range macro.methodMacros {
				switch {
				case method == "GET" :
					e.GET(routeBase + macro.Path, routeExecMacro, func(next echo.HandlerFunc) echo.HandlerFunc { 
						return middlewareAuthorize(childMacro, next)
					})
				case method == "POST" :
					e.POST(routeBase + macro.Path, routeExecMacro, func(next echo.HandlerFunc) echo.HandlerFunc { 
						return middlewareAuthorize(childMacro, next)
					})
				case method == "PUT" :
					e.POST(routeBase + macro.Path, routeExecMacro, func(next echo.HandlerFunc) echo.HandlerFunc { 
						return middlewareAuthorize(childMacro, next)
					})
				case method == "PATCH" :
					e.PATCH(routeBase + macro.Path, routeExecMacro, func(next echo.HandlerFunc) echo.HandlerFunc { 
						return middlewareAuthorize(childMacro, next)
					})
				case method == "DELETE" :
					e.DELETE(routeBase + macro.Path, routeExecMacro, func(next echo.HandlerFunc) echo.HandlerFunc { 
						return middlewareAuthorize(childMacro, next)
					})
				}
			}
		}
	}
	
	echoServer = e

	return e.Start(*flagRESTListenAddr)
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
				"code": 405,
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
