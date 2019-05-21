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
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/dop251/goja"
	"gopkg.in/resty.v1"
)

// initJSVM - creates a new javascript virtual machine
func initJSVM(ctx map[string]interface{}) *goja.Runtime {
	vm := goja.New()
	for k, v := range ctx {
		vm.Set(k, v)
	}
	vm.Set("fetch", jsFetchfunc)
	vm.Set("call_api", jsJWTFetchfunc)
	vm.Set("jwt_token", jsJWTTokenfunc)
	vm.Set("exec_sql", jsExecSQLFunc)
	vm.Set("exec_cmd", jsExecCommandFunc)
	vm.Set("log", log.Println)
	vm.Set("emit_msg", jsExecEmitMessage)

	vm.Set("ws_broacast", jsExecWebsocketBroacastMessage)
	vm.Set("ws_send", jsExecWebsocketSendMessage)
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

	var req *resty.Request

	if nil != option["body"] {
		body, _ = option["body"]
	}

	switch {
	case option["format"] == "form":
		req = resty.R().SetHeaders(headers).SetFormData(convertBodyToMap(body))
	case option["format"] == "query":
		req = resty.R().SetHeaders(headers).SetQueryParams(convertBodyToMap(body))
	default:
		req = resty.R().SetHeaders(headers).SetBody(body)
	}

	resp, err := req.Execute(method, url)

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

func convertBodyToMap(body interface{}) map[string]string {
	switch body.(type) {
	case map[string]interface{}:
		out := map[string]string{}
		for k, v := range body.(map[string]interface{}) {
			out[k] = fmt.Sprintf("%v", v)
		}
		return out
	default:
		return map[string]string{}
	}
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

	requestToken, err := macrosManager.JwtIdentityConfig().CreateRequestToken()

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

	var req *resty.Request

	if nil != option["body"] {
		body, _ = option["body"]
	}

	switch {
	case option["format"] == "form":
		req = resty.R().SetHeaders(headers).SetFormData(convertBodyToMap(body))
	case option["format"] == "query":
		req = resty.R().SetHeaders(headers).SetQueryParams(convertBodyToMap(body))
	default:
		req = resty.R().SetHeaders(headers).SetBody(body)
	}

	resp, err := req.Execute(method, url)

	if err != nil {
		return map[string]interface{}{
			"code":    5000,
			"message": err.Error(),
		}, nil
	}

	rspHdrs := resp.Header()
	rspHdrsNormalized := map[string]string{}
	for k, v := range rspHdrs {
		rspHdrsNormalized[strings.ToLower(k)] = v[0]
	}

	var respData map[string]interface{}
	respCode := resp.StatusCode()

	if respCode >= 200 && respCode < 400 {
		respCode = 0
	}

	if nil != json.Unmarshal(resp.Body(), &respData) {

		return map[string]interface{}{
			"code":       respCode,
			"message":    "操作成功！",
			"data":       string(resp.Body()),
			"__header__": rspHdrsNormalized,
		}, nil

	}

	respData["__header__"] = rspHdrsNormalized

	return respData, nil

}

func jsJWTTokenfunc() (string, error) {

	if !macrosManager.JwtIdentityConfig().IsEnabled() {
		return "", fmt.Errorf("jwt not enabled")
	}

	requestToken, err := macrosManager.JwtIdentityConfig().CreateRequestToken()

	if err != nil {
		return "", err
	}

	return requestToken, nil

}

//jsExecSQL - 执行SQL
func jsExecSQLFunc(sql string, args ...map[string]interface{}) (interface{}, error) {
	var arg map[string]interface{}

	if len(args) > 0 {
		arg = args[0]
	}

	ret, err := jsExecSQLQuery(strings.Split(strings.TrimSpace(sql), "---"), arg)

	if err != nil {
		return nil, err
	}

	if *flagDebug > 2 {
		log.Printf("jsExecSQL return %v\n", ret)
	}

	return ret, nil
}

// execSQLQuery - execute the specified sql query
func jsExecSQLQuery(sqls []string, args map[string]interface{}) (interface{}, error) {

	if !macrosManager.DatabaseConfig().IsDatabaseEnabled() {
		return nil, fmt.Errorf("Database forget enable??")
	}

	conn, err := sqlx.Open(macrosManager.DatabaseConfig().Driver, macrosManager.DatabaseConfig().Dsn)
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("jsExecSQL open database error: %v\n", err)
		}
		return nil, err
	}
	defer conn.Close()

	for i, sql := range sqls {
		if strings.TrimSpace(sql) == "" {
			sqls = append(sqls[0:i], sqls[i+1:]...)
		}
	}

	//先执行前面的SQL
	for i, sql := range sqls[0 : len(sqls)-1] {
		sql = strings.TrimSpace(sql)
		if "" == sql {
			continue
		}

		if *flagDebug > 2 {
			log.Printf("jsExecSQL exec sql%d:\n==sql==\n%s\n==args==\n%v\n", i, sql, args)
		}

		if _, err := conn.NamedExec(sql, args); err != nil {
			if *flagDebug > 0 {
				log.Printf("jsExecSQL exec sql%d error: %v\n==sql==\n%s\n", i, err, sql)
			}
			return nil, fmt.Errorf("jsExecSQL exec sql%d error: %v", i, err)
		}
	}

	if *flagDebug > 2 {
		log.Printf("jsExecSQL exec sql%d:\n==sql==\n%s\n==args==\n%v\n", len(sqls)-1, sqls[len(sqls)-1], args)
	}

	//最后一个用于返回数据
	rows, err := conn.NamedQuery(sqls[len(sqls)-1], args)
	if err != nil {
		if *flagDebug > 1 {
			log.Printf("jsExecSQL exec sql%d error: %v\n==sql==\n%s\n", len(sqls)-1, err, sqls[len(sqls)-1])
		}
		return nil, fmt.Errorf("jsExecSQL exec sql%d error: %v", len(sqls)-1, err)
	}
	defer rows.Close()

	ret := []map[string]interface{}{}

	for rows.Next() {
		row, err := jsScanSQLRow(rows)
		if err != nil {
			if *flagDebug > 1 {
				log.Printf("jsExecSQL sql%d fetch rows error:\n%v\n==sql==\n%s\n==rows==\n%v\n",
					len(sqls)-1, err, sqls[len(sqls)-1], rows)
			}
			continue
		}
		ret = append(ret, row)
	}

	return interface{}(ret), nil
}

// jsScanSQLRow - return values
func jsScanSQLRow(rows *sqlx.Rows) (map[string]interface{}, error) {
	row := make(map[string]interface{})
	if err := rows.MapScan(row); err != nil {
		return nil, err
	}

	for k, v := range row {
		if nil == v {
			continue
		}

		switch v.(type) {
		case []uint8:
			v = []byte(v.([]uint8))
		default:
			v, _ = json.Marshal(v)
		}

		var d interface{}
		if nil == json.Unmarshal(v.([]byte), &d) {
			row[k] = d
		} else {
			row[k] = string(v.([]byte))
		}
	}

	return row, nil
}

// jsExecCommandFunc - js execute the command line
func jsExecCommandFunc(cmdline string, args ...string) (interface{}, error) {

	cmdExecute, inputArgs := getCommandDefines(cmdline)

	for _, v := range args {
		inputArgs = append(inputArgs, v)
	}

	cmd := exec.Command(cmdExecute, inputArgs[0:]...)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("jsExecCommand error: %v\n==cmd==\n%s", err, cmdline)
		}
		return nil, err
	}

	outStr := out.String()
	var outData interface{}
	err = json.Unmarshal([]byte(outStr), &outData)

	if err != nil {
		return outStr, nil
	}

	return outData, nil

}

// jsExecEmitMessage - 发送消息
func jsExecEmitMessage(dest string, msg string, args ...map[string]interface{}) (bool, error) {
	var arg map[string]interface{}

	if len(args) > 0 {
		arg = args[0]
	} else {
		arg = map[string]interface{}{}
	}

	sender, err := macrosManager.MessageSendProvider()

	if err != nil {
		return false, err
	}

	if err := sender.EmitMessage(dest, msg, arg); err != nil {
		return false, err
	}
	return true, nil
}

// jsExecWebsocketBroacastMessage -
func jsExecWebsocketBroacastMessage(ch string, args ...interface{}) (bool, error) {
	var msg interface{}
	if len(args) == 1 {
		msg = args[0]
	} else {
		msg = args
	}
	r := GetWSClientRegistry(ch)
	if r == nil {
		return false, fmt.Errorf("not found endpoint %s", ch)
	}
	r.BroacastWebsocketMessage(msg)
	return true, nil
}

// jsExecWebsocketSendMessage -
func jsExecWebsocketSendMessage(ch string, cid string, args ...interface{}) (bool, error) {
	var msg interface{}
	if len(args) == 1 {
		msg = args[0]
	} else {
		msg = args
	}

	r := GetWSClientRegistry(ch)
	if r == nil {
		return false, nil
	}
	err := r.SendWebsocketMessage(cid, msg)
	return err == nil, err
}
