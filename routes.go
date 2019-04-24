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
	"encoding/json"
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

func getMacroBindParams(macro *Macro) []map[string]interface{} {
	ret := []map[string]interface{}{}
	if len(macro.Bind) > 0 {
		for k, v := range ret {
			var desc []byte
			desc, _ = json.Marshal(v)
			ret = append(ret, map[string]interface{}{
				"name":        k,
				"in":          "query",
				"description": string(desc),
				"required":    false,
				"type":        "string",
			})
		}
	} else {
		ret = append(ret, map[string]interface{}{
			"name":        "parameters",
			"in":          "query",
			"description": "parameters",
			"required":    false,
			"type":        "object",
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
			tagsMap[tag] = macro.name
		}
		apiPathMethods := make(map[string]interface{})
		if len(macro.Exec) > 0 {
			definedMethods := macro.Methods
			if len(definedMethods) == 0 {
				definedMethods = append(definedMethods, "GET")
			}
			for _, k := range definedMethods {
				apiPathMethods[strings.ToLower(k)] = map[string]interface{}{
					"tags":        macro.Tags,
					"summary":     macro.Summary,
					"operationId": macro.name,
					"consumes":    []string{"application/json"},
					"produces":    []string{"application/json;charset=UTF-8"},
					"parameters":  getMacroBindParams(macro),
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
					},
				}
			}
		} else {
			for k, childm := range macro.methodMacros {
				apiPathMethods[strings.ToLower(k)] = map[string]interface{}{
					"tags":        childm.Tags,
					"summary":     childm.Summary,
					"operationId": childm.name,
					"consumes":    []string{"application/json"},
					"produces":    []string{"application/json;charset=UTF-8"},
					"parameters":  getMacroBindParams(childm),
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
					},
				}
			}
		}

		pathsMap[macro.Path] = apiPathMethods

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
	} else {
		return c.JSON(200, map[string]interface{}{
			"code":    0,
			"message": "操作成功！",
			"data":    out,
		})
	}
}
