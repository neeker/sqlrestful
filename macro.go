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
	"fmt"
	"strings"
	"strconv"
	"log"
	"bytes"
	"os/exec"

	"github.com/jmoiron/sqlx"
)

// Cache 缓存配置
type Cache struct {
	Put  []string
	Evit []string
	Idle uint32
	Live uint32
}

// Macro - a macro configuration
type Macro struct {
	Methods     []string
	Include     []string
	Validators  map[string]string
	Authorizer  string
	Bind        map[string]string
	Impl        string
	Ret         string
	Exec        string
	Provider    string
	Aggregate   []string
	Transformer string
	Tags        []string
	Model      map[string]map[string]string

	Summary string

	Path string

	Total string
	Result  string

	Get    *Macro
	Post   *Macro
	Put    *Macro
	Patch  *Macro
	Delete *Macro
	Cache  *Cache

	name         string
	manager      *Manager
	methodMacros map[string]*Macro
}

func getCacheKey(input map[string]interface{}) string {
	if len(input) == 0 {
		return ""
	}
	ret, _ := json.Marshal(input)
	return string(ret)
}

// Put cache data
func putCacheData(cacheNames []string, cacheKey string, val interface{}) (bool, error) {
	var (
		ret bool
		err error
	)
	ret = false
	for _, k := range cacheNames {
		jsonData, _ := json.Marshal(val)
		ret, err = redisDb.HSet(k, cacheKey, string(jsonData)).Result()
		if err != nil {
			if *flagDebug > 1 {
				log.Printf("putcache(key=%s in %s) error: %v\n", cacheKey, k, err)
			}
			return false, err
		}

		if *flagDebug > 2 {
			log.Printf("putcache(key=%s in %s) data: %s\n", cacheKey, k, jsonData)
		}

	}
	return ret, nil
}

// Get cache data
func getCacheData(cacheNames []string, cacheKey string) (interface{}, error) {
	for _, k := range cacheNames {
		if redisDb.HExists(k, cacheKey).Val() {
			jsonData, _ := redisDb.HGet(k, cacheKey).Result()
			var outData interface{}
			err := json.Unmarshal([]byte(jsonData), &outData)
			if err != nil {
				if *flagDebug > 1 {
					log.Printf("getcache(key=%s in %s) error: %v\n", cacheKey, k, err)
				}
				return nil, err
			}

			if *flagDebug > 2 {
				log.Printf("getcache((key=%s in %s)) data: %s\n", cacheKey, k, jsonData)
			}

			return outData, nil
		}
	}
	return nil, nil
}

// Call - executes the macro
func (m *Macro) Call(input map[string]interface{}, inputKey map[string]interface{}) (interface{}, error) {
	ok, err := m.authorize(input)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errAccessDenyError
	}

	invalid, err := m.validate(input)
	if err != nil {
		return nil, err
	} else if len(invalid) > 0 {
		return nil, errValidationError
	}

	if m.Result == "page" {
		if input["offset"] == nil || input["offset"] == ""{
			input["offset"] = "0"
		}
		if input["limit"] == nil || input["limit"] == "" {
			input["limit"] = "0"
		} else {
			_, err = strconv.Atoi(input["limit"].(string))
			if err != nil {
				input["limit"] = "0"
			}
		}
	}

	var (
		out      interface{}
		cacheKey string
	)

	//获取缓存
	if redisDb != nil && m.Cache != nil && (len(m.Cache.Put) > 0 || len(m.Cache.Evit) > 0) {
		cacheKey = getCacheKey(inputKey)
		if cacheKey != "" && len(m.Cache.Put) > 0 {
			out, err = getCacheData(m.Cache.Put, cacheKey)
			if err != nil {
				return nil, err
			}
			if out != nil {
				if *flagDebug > 1 {
					log.Printf("%s getted data in cache(%s): %v\n", m.name, cacheKey, out)
				}
				return out, nil
			}
		}
	}

	if err := m.runIncludes(input, inputKey); err != nil {
		return nil, err
	}

	if len(m.Aggregate) > 0 {
		out, err = m.aggregate(input, inputKey)
		if err != nil {
			return nil, err
		}
	}
	
	pageTotal  := m.Total
	execScript := m.Exec
	scriptImpl := m.Impl
	
	if m.Provider != "" {
		resolvedVar, err := m.resolveExecScript(m.Provider, input);

		if err != nil {
			return nil, err
		}

		switch resolvedVar.(type) {
		case string:
			execScript = resolvedVar.(string)
			
			if *flagDebug > 1 {
				log.Printf("%s resolved exec script:\n\n%s\n\n", m.name, execScript)
			}

		case []string:
			for _, v := range resolvedVar.([]string) {
				execScript = execScript + "\n" + v
			}
			if *flagDebug > 1 {
				log.Printf("%s resolved exec sql:\n\n%s\n\n", m.name, execScript)
			}
		case map[string]interface{}:
			pageTotal = resolvedVar.(map[string]interface{})["total"].(string)
			execScript = resolvedVar.(map[string]interface{})["exec"].(string)
			if resolvedVar.(map[string]interface{})["impl"] != nil &&
				resolvedVar.(map[string]interface{})["impl"].(string) != "" {
				scriptImpl = resolvedVar.(map[string]interface{})["impl"].(string)
			}

			if *flagDebug > 1 {
				log.Printf("%s resolved exec %s:\n\n%s\n\n", m.name, scriptImpl, execScript)
				if len(pageTotal) > 0 {
					log.Printf("%s resolved total %s:\n\n%s\n\n", m.name, scriptImpl, execScript)
				}
			}

		default:
			v, _ := json.Marshal(resolvedVar)
			return nil, fmt.Errorf("%s provider return error type: %s", m.name, string(v))
		}
	}
	
	if len(pageTotal) > 0 {
		var resultLimit int
		if input["offset"] == nil || input["offset"] == "" {
			input["offset"] = "0"
		}
		if input["limit"] == nil || input["limit"] == "" {
			input["limit"] = "0"
		} else {
			resultLimit, err = strconv.Atoi(input["limit"].(string))
			if err != nil {
				resultLimit = 0
				input["limit"] = "0"
			}
		}

		var total int64

		switch {
		case scriptImpl == "js":
			total, err = m.execJavaScriptTotal(pageTotal, input)
		case scriptImpl == "cmd":
			total, err = m.execCommandTotal(pageTotal, input)
		default:
			total, err = m.execSQLTotal(strings.Split(strings.TrimSpace(pageTotal), "---"), input)
		}

		if err != nil {
			if *flagDebug > 0 {
				log.Printf("%s total error: %v\n", m.name, err)
			}
			return nil, err
		}

		if *flagDebug > 1 {
			log.Printf("%s total result: %d\n", m.name, total)
		}

		pageRet := make(map[string]interface{})

		pageRet["offset"] = input["offset"]
		pageRet["total"] = total

		if resultLimit > 0 && total > 0 {
			switch {
			case scriptImpl == "js":
				out, err = m.execJavaScript(execScript, input)
			case scriptImpl == "cmd":
				out, err = m.execCommand(execScript, input)
			default:
				out, err = m.execSQLQuery(strings.Split(strings.TrimSpace(execScript), "---"), input)
			}

			if err != nil {
				return nil, err
			}

			if *flagDebug > 1 {
				log.Printf("%s exec result: %v\n", m.name, out)
			}

			out, err = m.transform(out)
			if err != nil {
				if *flagDebug > 0 {
					log.Printf("%s transformer error: %v\n", m.name, err)
				}
				return nil, err
			}

			if *flagDebug > 1 {
				log.Printf("%s transformer result: %v\n", m.name, out)
			}

			pageRet["data"] = out
		}

		//设置缓存
		if redisDb != nil && m.Cache != nil && len(m.Cache.Put) > 0 {
			if cacheKey != "" && len(m.Cache.Put) > 0 {
				putCacheData(m.Cache.Put, cacheKey, pageRet)
			}
		}

		return pageRet, nil
	} 

	switch {
	case scriptImpl == "js":
		out, err = m.execJavaScript(execScript, input)
	case scriptImpl == "cmd":
		out, err = m.execCommand(execScript, input)
	default:
		out, err = m.execSQLQuery(strings.Split(strings.TrimSpace(execScript), "---"), input)
	}

	if err != nil {
		if *flagDebug > 0 {
			log.Printf("%s exec error: %v\n", m.name, err)
		}
		return nil, err
	}

	
	if *flagDebug > 1 {
		log.Printf("%s exec result: %v\n", m.name, out)
	}

	if m.Result == "object" && scriptImpl == "sql" {
		switch out.(type) {
		case []map[string]interface{}:
			if *flagDebug > 1 {
				log.Printf("%s exec origin result was list: %v\n", m.name, out)
			}			
			tmp := out.([]map[string]interface{})
			if len(tmp) < 1 {
				return nil, errObjNotFound
			}
			out = tmp[0]
		default:
			if *flagDebug > 0 {
				log.Printf("%s exec origin result was not list: %v\n", m.name, out)
			}			
		}
	}

	out, err = m.transform(out)
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("%s transformer error: %v\n", m.name, err)
		}			
		return nil, err
	}

	if *flagDebug > 1 {
		log.Printf("%s exec transformer result: %v\n", m.name, out)
	}			


	//设置缓存
	if redisDb != nil && m.Cache != nil && (len(m.Cache.Put) > 0 || len(m.Cache.Evit) > 0) {
		if cacheKey != "" && len(m.Cache.Put) > 0 {
			putCacheData(m.Cache.Put, cacheKey, out)
		}

		if len(m.Cache.Evit) > 0 {
			for _, k := range m.Cache.Evit {
				redisDb.Del(k)
			}
		}
	}

	return out, nil
}

// execSQLTotal - execute the specified sql query
func (m *Macro) execSQLTotal(sqls []string, input map[string]interface{}) (int64, error) {
	args, err := m.buildBind(input)
	if err != nil {
		return 0, err
	}

	conn, err := sqlx.Open(*flagDBDriver, *flagDBDSN)
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("run %s total(sql) open database error: %v\n", m.name, err)
		}
		return 0, err
	}
	defer conn.Close()

	for i, sql := range sqls {
		if strings.TrimSpace(sql) == "" {
			sqls = append(sqls[0:i], sqls[i+1:]...)
		}
	}

	for i, sql := range sqls[0 : len(sqls)-1] {
		sql = strings.TrimSpace(sql)
		if "" == sql {
			continue
		}
		
		if *flagDebug > 2 {
			log.Printf("run %s total sql%d:\n==sql==%s\n==args==\n%v\n", m.name, i, sql, args)
		}

		if _, err := conn.NamedExec(sql, args); err != nil {
			if *flagDebug > 0 {
				log.Printf("run %s total sql%d error: %v\n==sql==\n%v\n", m.name, i, err, sql)
			}
			return 0, fmt.Errorf("run %s total(sql) error: %v", m.name, err)
		}
	}


	if *flagDebug > 2 {
		log.Printf("run %s total sql%d:\n==sql==\n%s\n==args==\n%v\n", m.name, len(sqls)-1, sqls[len(sqls)-1], args)
	}

	rows, err := conn.NamedQuery(sqls[len(sqls)-1], args)
	if err != nil {
		if *flagDebug > 1 {
			log.Printf("run %s total sql%d error: %v\n==sql==\n%s\n", m.name, len(sqls)-1, err, sqls[len(sqls)-1])
		}
		return 0, fmt.Errorf("run %s total(sql) error: %v", m.name, err)
	}
	defer rows.Close()

	for rows.Next() {
		row, err := m.scanSQLRow(rows)
		if err != nil {
			continue
		}
		for _, v := range row {
			return int64(v.(float64)), nil
		}
	}

	return 0, nil
}

// execSQLQuery - execute the specified sql query
func (m *Macro) execSQLQuery(sqls []string, input map[string]interface{}) (interface{}, error) {
	args, err := m.buildBind(input)
	if err != nil {
		return nil, err
	}

	conn, err := sqlx.Open(*flagDBDriver, *flagDBDSN)
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("run %s exec(sql) open database error: %v\n", m.name, err)
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
			log.Printf("run %s exec sql%d:\n==sql==\n%s\n==args==\n%v\n", m.name, i, sql, args)
		}
	
		if _, err := conn.NamedExec(sql, args); err != nil {
			if *flagDebug > 0 {
				log.Printf("run %s exec sql%d error: %v\n==sql==\n%s\n", m.name, i, err, sql)
			}
			return nil, fmt.Errorf("run %s exec sql%d error: %v", m.name, i, err)
		}
	}

	
	if *flagDebug > 2 {
		log.Printf("run %s exec sql%d:\n==sql==\n%s\n==args==\n%v\n", m.name, len(sqls)-1, sqls[len(sqls)-1], args)
	}

	//最后一个用于返回数据
	rows, err := conn.NamedQuery(sqls[len(sqls)-1], args)
	if err != nil {
		if *flagDebug > 1 {
			log.Printf("run %s exec sql%d error: %v\n==sql==\n%s\n", m.name, len(sqls)-1, err, sqls[len(sqls)-1])
		}
		return nil, fmt.Errorf("run %s exec sql%d error: %v", m.name, len(sqls)-1, err)
	}
	defer rows.Close()

	ret := []map[string]interface{}{}

	for rows.Next() {
		row, err := m.scanSQLRow(rows)
		if err != nil {
			if *flagDebug > 1 {
			log.Printf("%s %s exec sql%d fetch rows error:\n%v\n==sql==\n%s\n==rows==\n%v\n", 
				m.name, len(sqls)-1, err, sqls[len(sqls)-1], rows)
			}
			continue
		}
		ret = append(ret, row)
	}

	return interface{}(ret), nil
}

// resolveExecScript - run the javascript function
func (m *Macro) resolveExecScript(javascript string, input map[string]interface{}) (interface{}, error) {
	vm := initJSVM(map[string]interface{}{
		"$input": input,
		"$name": m.name,
		"$macro": "provider",
	})

	if *flagDebug > 2 {
		log.Printf("run %s provider(js):\n==js==\n%s\n", m.name, javascript)
	}

	val, err := vm.RunString(javascript)
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("run %s provider(js) error: %v\n", m.name, err)
		}
		return nil, fmt.Errorf("run %s provider(js) error: %v", m.name, err)
	}

	return val.Export(), nil
}


// execJavaScript - run the javascript function
func (m *Macro) execJavaScript(javascript string, input map[string]interface{}) (interface{}, error) {

	vm := initJSVM(map[string]interface{}{
		"$input": input,
		"$name": m.name,
		"$macro": "exec",
	})

	if *flagDebug > 2 {
		log.Printf("run %s total js:\n==js==\n%s\n==input==\n%v\n", m.name, javascript, input)
	}

	val, err := vm.RunString(javascript)
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("run %s exec(js) error: %v\n", m.name, err)
		}
		return nil, fmt.Errorf("run %s exec(js) error: %v", m.name, err)
	}

	return val.Export(), nil
}

// execJavaScriptTotal - run the javascript total function
func (m *Macro) execJavaScriptTotal(javascript string, input map[string]interface{}) (int64, error) {
	vm := initJSVM(map[string]interface{}{
		"$input": input,
		"$name": m.name,
		"$macro": "total",
	})

	if *flagDebug > 2 {
		log.Printf("run %s total js:\n==js==\n%s\n==input==\n%v\n", m.name, javascript, input)
	}

	val, err := vm.RunString(javascript)
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("run %s total(js) error: %v\n", m.name, err)
		}
		return 0, fmt.Errorf("run %s total(js) error: %v", m.name, err)
	}

	return val.ToInteger(), nil
}

// scanSQLRow - scan a row from the specified rows
func (m *Macro) scanSQLRow(rows *sqlx.Rows) (map[string]interface{}, error) {
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

// transform - run the transformer function
func (m *Macro) transform(data interface{}) (interface{}, error) {
	transformer := strings.TrimSpace(m.Transformer)
	if transformer == "" {
		return data, nil
	}

	vm := initJSVM(map[string]interface{}{
		"$result": data,
		"$name": m.name,
		"$macro": "transformer",
	})

	if *flagDebug > 2 {
		log.Printf("run %s transformer js:\n==js==\n%s\n==data==\n%v\n", m.name, transformer, data)
	}

	v, err := vm.RunString(transformer)
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("run %s transformer error: %v\n", m.name, err)
		}
		return nil, fmt.Errorf("run %s transformer error: %v", m.name, err)
	}

	return v.Export(), nil
}

// authorize - run the authorizer function
func (m *Macro) authorize(input map[string]interface{}) (bool, error) {
	authorizer := strings.TrimSpace(m.Authorizer)
	if authorizer == "" {
		return true, nil
	}

	vm := initJSVM(map[string]interface{}{
		"$input": input,
		"$name": m.name,
		"$macro": "authorizer",
	})

	if *flagDebug > 2 {
		log.Printf("run %s authorizer js:\n==js==\n%s\n==input==\n%v\n", m.name, authorizer, input)
	}

	val, err := vm.RunString(m.Authorizer)
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("run %s authorize error:\n%v\n", m.name, err)
		}
		return false, fmt.Errorf("run %s authorize error: %v", m.name, err)
	}

	return val.ToBoolean(), nil
}

// aggregate - run the aggregators
func (m *Macro) aggregate(input map[string]interface{}, inputKey map[string]interface{}) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	for _, k := range m.Aggregate {
		macro := m.manager.Get(k)
		if nil == macro {
			if *flagDebug > 1 {
				log.Printf("%s aggregate not existed macro(%s)\n", m.name, k)
			}
			err := fmt.Errorf("%s aggregate not existed macro(%s)", m.name, k)
			return nil, err
		}

		if *flagDebug > 0 {
			log.Printf("run %s aggregate: entry %s\n", m.name, macro.name)
		}

		out, err := macro.Call(input, inputKey)
		if err != nil {
			return nil, err
		}
		ret[k] = out
	}
	return ret, nil
}

// validate - validate the input aginst the rules
func (m *Macro) validate(input map[string]interface{}) (ret []string, err error) {
	if len(m.Validators) < 1 {
		return nil, nil
	}

	vm := initJSVM(map[string]interface{}{
		"$input": input,
		"$name": m.name,
		"$macro": "validators",
	})

	for k, src := range m.Validators {

		if *flagDebug > 2 {
			log.Printf("run %s validator(%s):\n==js==\n%s\n", m.name, k, src)
		}

		val, err := vm.RunString(src)
		if err != nil {
			if *flagDebug > 0 {
				log.Printf("run %s validate(%s=\"%s\") error: %v\n", m.name, k, src, err)
			}
			return nil, fmt.Errorf("run %s validate(%s=\"%s\") error: %v", m.name, k, src, err)
		}

		if !val.ToBoolean() {
			ret = append(ret, k)
		}
	}

	return ret, nil
}

// buildBind - build the bind vars
func (m *Macro) buildBind(input map[string]interface{}) (map[string]interface{}, error) {
	if len(m.Bind) < 1 {
		return nil, nil
	}

	vm := initJSVM(map[string]interface{}{
		"$input": input,
		"$name": m.name,
		"$macro": "bind",
	})
	ret := map[string]interface{}{}

	for k, src := range m.Bind {

		if *flagDebug > 2 {
			log.Printf("run %s bind(%s): %s\n", m.name, k, src)
		}

		val, err := vm.RunString(src)
		if err != nil {
			if *flagDebug > 0 {
				log.Printf("run %s bind(%s=\"%s\") error: %v\n", m.name, k, src, err)
			}
			return nil, fmt.Errorf("run %s bind(%s=\"%s\") error: %v", m.name, k, src, err)
		}

		ret[k] = val.Export()
	}

	return ret, nil
}

// runIncludes - run the include function
func (m *Macro) runIncludes(input map[string]interface{}, inputKey map[string]interface{}) error {
	for _, name := range m.Include {
		macro := m.manager.Get(name)
		if nil == macro {
			if *flagDebug > 1 {
				log.Printf("%s include not existed macro(%s)\n", m.name, name)
			}
	
			return fmt.Errorf("%s include not existed macro(%s)", m.name, name)
		}

		if *flagDebug > 2 {
			log.Printf("run %s include: %s\n", m.name, macro.name)
		}

		_, err := macro.Call(input, inputKey)
		if err != nil {
			return err
		}
	}
	return nil
}


// execCommand - execute the command line
func (m *Macro) execCommand(cmdline string, input map[string]interface{}) (interface{}, error) {
	args, err := m.buildBind(input)
	if err != nil {
		return 0, err
	}

	inputArgs := []string{ "-c", cmdline }

	for k, v := range args {
		inputArgs = append(inputArgs, k)
		switch v.(type) {
		case string:
			if v.(string) != "" {
				inputArgs = append(inputArgs, v.(string))
			}
		}
	}

	cmd := exec.Command("/bin/bash", inputArgs[0:]...)

	var out bytes.Buffer
	cmd.Stdout = &out
	
	err = cmd.Run()
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("run %s exec(cmd) error: %v\n==cmd==\n%s", m.name, err, cmdline)
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



// execCommand - execute the command line
func (m *Macro) execCommandTotal(cmdline string, input map[string]interface{}) (int64, error) {
	args, err := m.buildBind(input)
	if err != nil {
		return 0, err
	}

	inputArgs := []string{ "-c", cmdline }

	for k, v := range args {
		inputArgs = append(inputArgs, k)
		switch v.(type) {
		case string:
			if v.(string) != "" {
				inputArgs = append(inputArgs, v.(string))
			}
		}
	}

	cmd := exec.Command("/bin/bash", inputArgs[0:]...)

	var out bytes.Buffer
	cmd.Stdout = &out
	
	err = cmd.Run()
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("run %s total(cmd) error: %v\n==cmd==\n%s", m.name, err, cmdline)
		}
		return 0, err
	}
	outStr := out.String()
	outData, err := strconv.Atoi(outStr)
	if err != nil {
		if *flagDebug > 0 {
			log.Printf("run %s total(cmd) return error: %s\n==cmd==\n%s", m.name, outStr, cmdline)
		}
		return 0, nil
	}

	return int64(outData), nil
}
