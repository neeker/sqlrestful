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
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

// routeIndex - the index route
func routeIndex(c echo.Context) error {
	return c.Redirect(301, "health")
}

// customer HTTPErrorHandler
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	c.Logger().Error(c.JSON(code, map[string]interface{}{
		"code":    code,
		"message": err.Error(),
	}))
}

//routeApiDocs
func routeAPIDocs(c echo.Context) error {
	return c.JSON(200, map[string]interface{}{
		"code":    0,
		"message": "操作成功！",
		"data":    echoServer.Routes(),
	})
}

// routeExecMacro - execute the requested macro
func routeExecMacro(c echo.Context) error {
	macro := c.Get("macro").(*Macro)

	input := make(map[string]interface{})
	body := make(map[string]interface{})

	keyInput := make(map[string]interface{})

	c.Bind(&body)

	for k := range c.QueryParams() {
		input[k] = c.QueryParam(k)
		keyInput[k] = c.QueryParam(k)
	}

	for k, v := range body {
		input[k] = v
	}

	for _, k := range c.ParamNames() {
		input[k] = c.Param(k)
		keyInput[k] = c.Param(k)
	}

	headers := c.Request().Header
	for k, v := range headers {
		input["http_"+strings.Replace(strings.ToLower(k), "-", "_", -1)] = v[0]
	}

	out, err := macro.Call(input, keyInput)

	if err != nil {
		code := errStatusCodeMap[err]
		if code < 1 {
			code = 500
		}
		return c.JSON(code, map[string]interface{}{
			"code":    code,
			"message": err.Error(),
		})
	}

	return c.JSON(200, map[string]interface{}{
		"code":    0,
		"message": "操作成功！",
		"data":    out,
	})
}
