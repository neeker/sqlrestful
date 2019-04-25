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

func getMacroBindParams(macro *Macro, method string) []map[string]interface{} {

	bindInput := map[string]string{}

	if len(macro.Bind) > 0 {
		for k := range macro.Bind {
			bindInput[k] = "query"
		}
	} else {
		bindInput = make(map[string]string)
	}

	for _, k := range strings.Split(macro.Path, "/") {
		if strings.HasPrefix(k, ":") {
			bindInput[k[1:]] = "path"
		}
	}

	ret := []map[string]interface{}{}
	if len(bindInput) > 0 {
		for k, v := range bindInput {
			ret = append(ret, map[string]interface{}{
				"name":        k,
				"in":          v,
				"description": k,
				"required":    false,
				"type":        "string",
			})
		}
	} else if method == "post" || method == "put" || method == "patch" {
		ret = append(ret, map[string]interface{}{
			"name":        "body",
			"in":          "body",
			"description": "body",
			"required":    false,
			"schema": map[string]interface{}{
				"$ref": "#/definitions/jsonbody",
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
								"$ref": "#/definitions/result",
							},
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
							"schema": map[string]interface{}{
								"$ref": "#/definitions/result",
							},
						},
						"403": map[string]interface{}{
							"description": "Forbidden",
							"schema": map[string]interface{}{
								"$ref": "#/definitions/result",
							},
						},
					},
				}
			}
		} else {
			for k, childm := range macro.methodMacros {
				methodName := strings.ToLower(k)
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
								"$ref": "#/definitions/result",
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

	inline_tag := "z.内置接口"

	tagsMap[inline_tag] = "inline implements"

	pathsMap["/health"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{inline_tag},
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

	pathsMap["/clear_caches"] = map[string]interface{}{
		"post": map[string]interface{}{
			"tags":        []string{inline_tag},
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

	pathsMap["/v2/api-docs"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{inline_tag},
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

	pathsMap["/webjars"] = map[string]interface{}{
		"get": map[string]interface{}{
			"tags":        []string{inline_tag},
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
			"tags":        []string{inline_tag},
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
			"tags":        []string{inline_tag},
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
			"tags":        []string{inline_tag},
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
			"description": *flagDescription,
			"version":     *flagVersion,
			"title":       *flagName,
			"contact": map[string]interface{}{
				"name":  *flagAuthor,
				"email": *flagEmail,
			},
		},
		"host":     "",
		"basePath": *flagBasePath,
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

	return c.JSON(200, retdata)
}

//routeApiDocs
func routeClearCaches(c echo.Context) error {

	for _, macroName := range macrosManager.Names() {

		macro := macrosManager.Get(macroName)

		if macro.Cache != nil && macro.Cache.Put != nil && len(macro.Cache.Put) > 0 {
			for _, k := range macro.Cache.Put {
				redisDb.Del(k)
			}
		}

		for _, childm := range macro.methodMacros {
			if childm.Cache != nil && childm.Cache.Put != nil && len(childm.Cache.Put) > 0 {
				for _, k := range childm.Cache.Put {
					redisDb.Del(k)
				}
			}
		}

	}

	return c.JSON(200, map[string]interface{}{
		"code":    0,
		"message": "操作成功！",
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

	if macro.Ret == "origin" {
		return c.JSON(200, out)
	} else if out != nil && macro.Ret != "null" {
		return c.JSON(200, map[string]interface{}{
			"code":    0,
			"message": "操作成功！",
			"data":    out,
		})
	} else {
		return c.JSON(200, map[string]interface{}{
			"code":    0,
			"message": "操作成功！",
		})
	}
}
