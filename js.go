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
	"strings"
	"time"

	"encoding/json"

	"github.com/dop251/goja"
	"gopkg.in/resty.v1"
	"github.com/dgrijalva/jwt-go"

)


// genJWTRequestToken
func genJWTRequestToken() (string, error) {
	claims := &jwt.StandardClaims {
    ExpiresAt: time.Now().Unix() + int64(jwtExpires),
    Issuer:   jwtSecret,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ret, err := jwtToken.SignedString(jwtRSAPrivkey)
	return ret, err
}

// initJSVM - creates a new javascript virtual machine
func initJSVM(ctx map[string]interface{}) *goja.Runtime {
	vm := goja.New()
	for k, v := range ctx {
		vm.Set(k, v)
	}
	vm.Set("fetch", jsFetchfunc)
	vm.Set("call_api", jsJWTFetchfunc)
	vm.Set("log", log.Println)
	return vm
}

// jsFetchfunc - the fetch function used inside the js vm
func jsFetchfunc(url string, options ...map[string]interface{}) (map[string]interface{}, error) {
	var option map[string]interface{}
	var method string
	var headers map[string]string
	var body interface{}

	if len(options) > 0 {
		option = options[0]
	}

	if nil != option["method"] {
		method, _ = option["method"].(string)
	}

	if nil != option["headers"] {
		hdrs, _ := option["headers"].(map[string]interface{})
		headers = make(map[string]string)
		for k, v := range hdrs {
			headers[k], _ = v.(string)
		}
	}

	if nil != option["body"] {
		body, _ = option["body"]
	}

	resp, err := resty.R().SetHeaders(headers).SetBody(body).Execute(method, url)
	if err != nil {
		return nil, err
	}

	rspHdrs := resp.Header()
	rspHdrsNormalized := map[string]string{}
	for k, v := range rspHdrs {
		rspHdrsNormalized[strings.ToLower(k)] = v[0]
	}

	return map[string]interface{}{
		"status":     resp.Status(),
		"statusCode": resp.StatusCode(),
		"headers":    rspHdrsNormalized,
		"body":       string(resp.Body()),
	}, nil
}

func jsJWTFetchfunc(url string, options ...map[string]interface{}) (map[string]interface{}, error) {
	var option map[string]interface{}
	var method string
	var headers map[string]string
	var body interface{}

	if len(options) > 0 {
		option = options[0]
	}

	if nil != option["method"] {
		method, _ = option["method"].(string)
	}

	if nil != option["headers"] {
		hdrs, _ := option["headers"].(map[string]interface{})
		headers = make(map[string]string)
		for k, v := range hdrs {
			headers[k], _ = v.(string)
		}
	}

	requestToken, err := genJWTRequestToken()

	if err == nil {
		if headers == nil {
			headers = make(map[string]string)
		}
		headers["Authorization"] = "Bearer " + requestToken
	}

	if headers == nil {
		headers = make(map[string]string)
	}

	if headers["Content-Type"] == "" {
		headers["Content-Type"] = "application/json; charset=UTF-8"
	}
	
	if nil != option["body"] {
		body, _ = option["body"]
	}

	resp, err := resty.R().SetHeaders(headers).SetBody(body).Execute(method, url)
	if err != nil {
		return map[string]interface{}{
			"code": 5000,
			"message": err.Error(),
		}, nil
	}

	var respData map[string]interface{}
	respCode := resp.StatusCode()

	if respCode >= 200 &&  respCode < 400 {
		respCode = 0
	}

	if nil != json.Unmarshal(resp.Body(), &respData) || respData == nil || len(respData) == 0 {

		rspHdrs := resp.Header()
		rspHdrsNormalized := map[string]string{}
		for k, v := range rspHdrs {
			rspHdrsNormalized[strings.ToLower(k)] = v[0]
		}
	
		return map[string]interface{}{
			"status":     resp.Status(),
			"statusCode": resp.StatusCode(),
			"headers":    rspHdrsNormalized,
			"code": 			respCode,
			"body":       string(resp.Body()),
		}, nil
	} else {
		return respData, nil
	}

}

