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
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo"
)

// customer HTTPErrorHandler
func customHTTPErrorHandler(err error, c echo.Context) {

	code := http.StatusInternalServerError
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}

	if *flagDebug > 0 {
		tmpPath := c.Path()
		if tmpPath == "" {
			tmpPath = "/"
		}
		log.Printf("%s %s %d: %v", c.Request().Method, tmpPath, code, err)
	}

	c.JSON(code, map[string]interface{}{
		"code":    code,
		"message": err.Error(),
	})

}

func getMacroBindParams(macro *Macro, method string) []map[string]interface{} {

	bindInput := map[string]string{}

	if len(macro.Bind) > 0 {
		for k := range macro.Bind {
			bindInput[k] = "query"
		}
	} else {
		bindInput = make(map[string]string)
	}

	pathIds := map[string]bool{}

	for _, k := range strings.Split(macro.Path, "/") {
		if strings.HasPrefix(k, ":") {
			pathID := k[1:]
			pathIds[pathID] = true
			bindInput[pathID] = "path"
		}
	}

	ret := []map[string]interface{}{}

	switch method {
	case "get", "delete":
		for k, v := range bindInput {
			ret = append(ret, map[string]interface{}{
				"name":        k,
				"in":          v,
				"description": k,
				"required":    false,
				"type":        "string",
			})
		}
	default:
		for k := range pathIds {
			ret = append(ret, map[string]interface{}{
				"name":        k,
				"in":          "path",
				"description": k,
				"required":    false,
				"type":        "string",
			})
		}

		ret = append(ret, map[string]interface{}{
			"name":        "body",
			"in":          "body",
			"description": "body",
			"required":    false,
			"schema": map[string]interface{}{
				"$ref": "#/definitions/" + macro.name + ".input",
			},
		})

	}

	return ret
}

//getTagsAndRestfulPaths
func getTagsAndRestfulPaths() ([]map[string]interface{}, map[string]interface{}) {
	tagsMap := make(map[string]interface{})
	pathsMap := make(map[string]interface{})

	for _, macro := range macrosManager.List() {
		for _, tag := range macro.Tags {
			if tagsMap[tag] == nil {
				tagsMap[tag] = macro.name
			} else {
				tagsMap[tag] = tagsMap[tag].(string) + " " + macro.name
			}
		}
		apiPathMethods := make(map[string]interface{})
		if len(macro.Exec) > 0 {
			definedMethods := []string{}
			copy(definedMethods, macro.Methods)
			if len(definedMethods) == 0 {
				definedMethods = append(definedMethods, "GET")
			}

			schemaRef := "#/definitions/result"

			if macro.Model != nil {
				schemaRef = macro.name + ".result"
			}

			for _, k := range definedMethods {
				methodName := strings.ToLower(k)
				apiPathMethods[methodName] = map[string]interface{}{
					"tags":        macro.Tags,
					"summary":     macro.Summary,
					"operationId": macro.name,
					"consumes":    []string{"application/json"},
					"produces":    []string{"application/json;charset=UTF-8"},
					"parameters":  getMacroBindParams(macro, methodName),
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "OK",
							"schema": map[string]interface{}{
								"$ref": schemaRef,
							},
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
							"schema": map[string]interface{}{
								"$ref": "#/definitions/error",
							},
						},
						"403": map[string]interface{}{
							"description": "Forbidden",
							"schema": map[string]interface{}{
								"$ref": "#/definitions/error",
							},
						},
					},
				}
			}
		} else {
			for k, childm := range macro.methodMacros {
				methodName := strings.ToLower(k)

				schemaRef := "#/definitions/result"

				if childm.Model != nil {
					schemaRef = childm.name + ".result"
				}

				apiPathMethods[methodName] = map[string]interface{}{
					"tags":        childm.Tags,
					"summary":     childm.Summary,
					"operationId": childm.name,
					"consumes":    []string{"application/json"},
					"produces":    []string{"application/json;charset=UTF-8"},
					"parameters":  getMacroBindParams(childm, methodName),
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "OK",
							"schema": map[string]interface{}{
								"$ref": schemaRef,
							},
						},
						"401": map[string]interface{}{
							"description": "用户未登录",
							"schema": map[string]interface{}{
								"$ref": "#/definitions/error",
							},
						},
						"403": map[string]interface{}{
							"description": "无权访问",
							"schema": map[string]interface{}{
								"$ref": "#/definitions/error",
							},
						},
					},
				}
			}
		}

		apd := false
		tmpPaths := ""
		for _, k := range strings.Split(macro.Path, "/") {
			if apd {
				tmpPaths = tmpPaths + "/"
			} else {
				apd = true
			}
			if strings.HasPrefix(k, ":") {
				tmpPaths = tmpPaths + "{" + k[1:] + "}"
			} else {
				tmpPaths = tmpPaths + k
			}
		}

		pathsMap[tmpPaths] = apiPathMethods

	}

	inlineTag := "z.内置接口"

	tagsMap[inlineTag] = "inline implements"

	pathsMap["/health"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{inlineTag},
			"summary":     "心跳检测",
			"operationId": "inline_helath",
			"consumes":    []string{"application/json"},
			"produces":    []string{"application/json;charset=UTF-8"},
			"parameters":  map[string]interface{}{},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "OK",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/result",
					},
				},
			},
		},
	}

	if macrosManager.DatabaseConfig().IsRedisEnabled() {
		pathsMap["/clear_caches"] = map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{inlineTag},
				"summary":     "清理所有缓存数据",
				"operationId": "inline_clear_caches",
				"consumes":    []string{"application/json"},
				"produces":    []string{"application/json;charset=UTF-8"},
				"parameters":  map[string]interface{}{},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "OK",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/result",
						},
					},
					"403": map[string]interface{}{
						"description": "无权访问",
						"schema": map[string]interface{}{
							"$ref": "#/definitions/error",
						},
					},
				},
			},
		}
	}

	pathsMap["/v2/api-docs"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{inlineTag},
			"summary":     "Swagger文档接口",
			"operationId": "inline_v2_api_docs",
			"consumes":    []string{"application/json"},
			"produces":    []string{"application/json;charset=UTF-8"},
			"parameters":  map[string]interface{}{},
			"responses": map[string]interface{}{
				"200": map[string]interface{}{
					"description": "OK",
					"schema": map[string]interface{}{
						"$ref": "#/definitions/result",
					},
				},
			},
		},
	}

	if macrosManager.IsSwaggerEnabled() {
		pathsMap["/webjars"] = map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{inlineTag},
				"summary":     "SwaggerUI资源",
				"operationId": "inline_webjars",
				"consumes":    []string{"text/html"},
				"produces":    []string{"text/html;charset=UTF-8"},
				"parameters":  map[string]interface{}{},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "OK",
					},
				},
			},
		}

		pathsMap["/webjars"] = map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{inlineTag},
				"summary":     "SwaggerUI静态资源",
				"operationId": "inline_webjars",
				"consumes":    []string{"text/html"},
				"produces":    []string{"text/html;charset=UTF-8"},
				"parameters":  map[string]interface{}{},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "OK",
					},
				},
			},
		}

		pathsMap["/swagger-resources"] = map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{inlineTag},
				"summary":     "SwaggerUI配置资源",
				"operationId": "inline_swagger_resources",
				"consumes":    []string{"text/html"},
				"produces":    []string{"text/html;charset=UTF-8"},
				"parameters":  map[string]interface{}{},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "OK",
					},
				},
			},
		}

		pathsMap["/swagger-ui.html"] = map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{inlineTag},
				"summary":     "SwaggerUI界面",
				"operationId": "inline_swagger_ui",
				"consumes":    []string{"text/html"},
				"produces":    []string{"text/html;charset=UTF-8"},
				"parameters":  map[string]interface{}{},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "OK",
					},
				},
			},
		}
	}

	retmap := []map[string]interface{}{}

	for k, v := range tagsMap {
		retmap = append(retmap, map[string]interface{}{
			"name":        k,
			"description": v,
		})
	}

	return retmap, pathsMap
}

//routeApiDocs
func routeAPIDocs(c echo.Context) error {
	apiTags, apiPaths := getTagsAndRestfulPaths()
	retdata := map[string]interface{}{
		"swagger": "2.0",
		"info": map[string]interface{}{
			"description": macrosManager.ServiceDesc(),
			"version":     macrosManager.ServiceVersion(),
			"title":       macrosManager.ServiceName(),
			"contact": map[string]interface{}{
				"name":  macrosManager.ServiceAuthor().Name,
				"email": macrosManager.ServiceAuthor().Email,
				"url":   macrosManager.ServiceAuthor().Url,
			},
		},
		"host":     "",
		"basePath": macrosManager.ServiceBasePath(),
		"tags":     apiTags,
		"paths":    apiPaths,
		"definitions": map[string]interface{}{
			"result": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":   "integer",
						"format": "int32",
					},
					"message": map[string]interface{}{
						"type": "string",
					},
					"data": map[string]interface{}{
						"type": "object",
					},
				},
			},
			"error": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":   "integer",
						"format": "int32",
					},
					"message": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"jsonbody": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}

	definitionsMap := retdata["definitions"].(map[string]interface{})

	for _, macro := range macrosManager.List() {
		buildResultDefinitionMap(macro, definitionsMap)

		for _, childm := range macro.methodMacros {
			buildResultDefinitionMap(childm, definitionsMap)
		}
	}

	return c.JSON(200, retdata)
}

//buildResultDefinitionMap
func buildResultDefinitionMap(macro *Macro, definitionsMap map[string]interface{}) {
	definitionName := macro.name + ".input"
	if macro.Bind != nil {
		bindMap := map[string]interface{}{}
		for k := range macro.Bind {
			bindMap[k] = map[string]interface{}{
				"type": "string",
			}
		}
		definitionsMap[definitionName] = map[string]interface{}{
			"properties": bindMap,
			"type":       "object",
		}
	} else {
		definitionsMap[definitionName] = map[string]interface{}{
			"properties": map[string]interface{}{},
			"type":       "object",
		}
	}

	if *flagDebug > 2 {
		log.Printf("%s input model %v", definitionName, definitionsMap[definitionName])
	}

	definitionName = macro.name + ".result"
	if macro.Model != nil {
		switch macro.Format {
		case "origin":
			definitionsMap[definitionName] = map[string]interface{}{
				"properties": macro.Model,
				"type":       "object",
			}
		case "nil":
		default:
			definitionsMap[definitionName] = map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":   "integer",
						"format": "int32",
					},
					"message": map[string]interface{}{
						"type": "string",
					},
					"data": map[string]interface{}{
						"properties": macro.Model,
						"type":       "object",
					},
				},
			}
		}
	} else {
		definitionsMap[definitionName] = map[string]interface{}{
			"properties": map[string]interface{}{},
			"type":       "object",
		}
	}

	if *flagDebug > 2 {
		log.Printf("%s bind model %v", definitionName, definitionsMap[definitionName])
	}

}

//routeApiDocs
func routeClearCaches(c echo.Context) error {

	for _, macroName := range macrosManager.Names() {

		macro := macrosManager.Get(macroName)

		if macro.Cache != nil && macro.Cache.Put != nil && len(macro.Cache.Put) > 0 {
			macrosManager.DatabaseConfig().ClearCaches(macro.Cache.Put)
		}

		for _, childm := range macro.methodMacros {
			if childm.Cache != nil && childm.Cache.Put != nil && len(childm.Cache.Put) > 0 {
				macrosManager.DatabaseConfig().ClearCaches(childm.Cache.Put)
			}
		}

	}

	return c.JSON(200, map[string]interface{}{
		"code":    0,
		"message": "操作成功！",
	})
}

// routeExecMacro - execute the requested macro
func routeExecMacro(c echo.Context) (err error) {
	macro := c.Get("macro").(*Macro)

	input := make(map[string]interface{})
	body := make(map[string]interface{})

	keyInput := make(map[string]interface{})

	tmpPath := c.Request().URL.Path
	if tmpPath == "" {
		tmpPath = "/"
	}

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

	{
		formParams, err := c.FormParams()
		if err != nil {
			if *flagDebug > 0 {
				log.Printf("%s %s route to %s prepare form error: %v\n",
					c.Request().Method, tmpPath, macro.name, err)
			}
		} else {
			for k := range formParams {
				input[k] = c.FormValue(k)
				keyInput[k] = c.FormValue(k)
			}
		}
	}

	headers := c.Request().Header
	for k, v := range headers {
		input["http_"+strings.Replace(strings.ToLower(k), "-", "_", -1)] = v[0]
	}

	for _, cookie := range c.Cookies() {
		input["cookie_"+strings.Replace(strings.ToLower(cookie.Name), "-", "_", -1)] = cookie.Value
	}

	input["http_method"] = c.Request().Method
	input["http_path"] = tmpPath
	input["http_url"] = c.Request().URL.String()
	input["http_uri"] = c.Request().RequestURI
	input["http_remote_addr"] = c.Request().RemoteAddr
	input["http_real_ip"] = c.RealIP()
	input["http_scheme"] = c.Scheme()

	if *flagDebug > 2 {
		log.Printf("%s %s route to %s:\n==input==\n%v\n==path_vars==\n%v\n",
			c.Request().Method, tmpPath, macro.name, input, keyInput)
	}

	var trustedProxy []string

	if macro.Proxy != nil {
		trustedProxy = macro.Proxy
	} else {
		trustedProxy = macro.manager.TrustedProxyList()
	}

	if trustedProxy != nil && len(trustedProxy) > 0 {
		requestAllow := false
		portIndex := strings.LastIndex(c.Request().RemoteAddr, ":")
		var clientIP string
		if portIndex > 0 {
			clientIP = c.Request().RemoteAddr[0:portIndex]
		} else {
			clientIP = c.Request().RemoteAddr
		}

		for _, proxyIP := range trustedProxy {
			ipMatched, err := regexp.Match(proxyIP, []byte(clientIP))
			if err != nil && *flagDebug > 0 {
				log.Printf("%s request %s, but regex (%s) match error: %v",
					clientIP, macro.name, proxyIP, err)
			}
			if ipMatched {
				requestAllow = true
				break
			}
		}

		if !requestAllow {
			return c.JSON(403, map[string]interface{}{
				"code":    403,
				"message": "不允许访问！",
			})
		}

	}

	out, err := macro.Call(input, keyInput)

	if err != nil {
		code := errStatusCodeMap[err]
		if code < 1 {
			code = 500
		}
		if *flagDebug > 0 {
			log.Printf("call %s error: %v\n", macro.name, err)
		}
		return c.JSON(code, map[string]interface{}{
			"code":    code,
			"message": err.Error(),
		})
	}

	if macro.Format == "origin" {
		if *flagDebug > 2 {
			log.Printf("%s ret is origin\n", macro.name)
		}

		return c.JSON(200, out)
	}

	if macro.Format == "redirect" {
		if *flagDebug > 2 {
			log.Printf("%s ret is redirect\n", macro.name)
		}

		switch out.(type) {
		case string:
			return c.Redirect(302, out.(string))
		default:

			if *flagDebug > 0 {
				log.Printf("%s result is not url: %v\n", macro.name, out)
			}

			return c.JSON(200, map[string]interface{}{
				"code":    500,
				"message": "返回值不是有效的网址！",
				"data":    out,
			})
		}
	}

	if macro.Format == "nil" {

		if *flagDebug > 2 {
			log.Printf("%s ret is defined nil\n", macro.name)
		}

		return c.JSON(200, map[string]interface{}{
			"code":    0,
			"message": "操作成功！",
		})

	}

	if *flagDebug > 2 {
		log.Printf("%s ret is normal\n", macro.name)
	}

	if out != nil {
		if *flagDebug > 2 {
			log.Printf("%s result: %s\n", macro.name, out)
		}
		return c.JSON(200, map[string]interface{}{
			"code":    0,
			"message": "操作成功！",
			"data":    out,
		})
	}

	if *flagDebug > 2 {
		log.Printf("%s result is nil\n", macro.name)
	}

	return c.JSON(200, map[string]interface{}{
		"code":    0,
		"message": "操作成功！",
	})

}
