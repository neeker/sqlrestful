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
	Exec        string
	Aggregate   []string
	Transformer string

	Path string

	Total string
	Type  string

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
			fmt.Printf("putCacheData, k = %s, f = %s, value = %s, err: %s", k, cacheKey, string(jsonData), err.Error())
		} else {
			fmt.Printf("putCacheData, k = %s, f = %s, value = %s", k, cacheKey, string(jsonData))
		}
		if err != nil {
			return false, err
		}
	}
	return ret, err
}

// Get cache data
func getCacheData(cacheNames []string, cacheKey string) (interface{}, error) {
	for _, k := range cacheNames {
		if redisDb.HExists(k, cacheKey).Val() {
			jsonData, _ := redisDb.HGet(k, cacheKey).Result()
			var outData interface{}
			err := json.Unmarshal([]byte(jsonData), &outData)

			if err != nil {
				fmt.Printf("getCacheData, k = %s, f = %s, value = %s, err: %s", k, cacheKey, jsonData, err.Error())
			} else {
				fmt.Printf("getCacheData, k = %s, f = %s, value = %s", k, cacheKey, jsonData)
			}

			if err != nil {
				return nil, err
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
		return err.Error(), err
	}

	if !ok {
		return errAuthorizationError.Error(), errAuthorizationError
	}

	invalid, err := m.validate(input)
	if err != nil {
		return err.Error(), err
	} else if len(invalid) > 0 {
		return invalid, errValidationError
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
			return err.Error(), err
		}
	} else if len(m.Total) > 0 {
		if input["offset"] == nil {
			input["offset"] = uint32(0)
		}
		if input["limit"] == nil {
			input["limit"] = uint32(0)
		}

		var total uint32

		total, err = m.execSQLTotal(strings.Split(strings.TrimSpace(m.Total), *flagSQLSeparator), input)
		if err != nil {
			return err.Error(), err
		}

		pageRet := make(map[string]interface{})

		pageRet["offset"] = input["offset"]
		pageRet["total"] = total

		out, err = m.execSQLQuery(strings.Split(strings.TrimSpace(m.Exec), *flagSQLSeparator), input)
		if err != nil {
			return err.Error(), err
		}

		out, err = m.transform(out)
		if err != nil {
			return err.Error(), err
		}

		pageRet["data"] = out

		//设置缓存
		if redisDb != nil && m.Cache != nil && len(m.Cache.Put) > 0 {
			if cacheKey != "" && len(m.Cache.Put) > 0 {
				putCacheData(m.Cache.Put, cacheKey, pageRet)
			}
		}

		return pageRet, nil
	} else {
		out, err = m.execSQLQuery(strings.Split(strings.TrimSpace(m.Exec), *flagSQLSeparator), input)
		if err != nil {
			return err.Error(), err
		}

		if m.Type == "object" {
			out = out.([]map[string]interface{})[0]
		}
	}

	out, err = m.transform(out)
	if err != nil {
		return err.Error(), err
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

// execSQLQuery - execute the specified sql query
func (m *Macro) execSQLTotal(sqls []string, input map[string]interface{}) (uint32, error) {
	args, err := m.buildBind(input)
	if err != nil {
		return 0, err
	}

	conn, err := sqlx.Open(*flagDBDriver, *flagDBDSN)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	for i, sql := range sqls {
		if strings.TrimSpace(sql) == "" {
			sqls = append(sqls[0:i], sqls[i+1:]...)
		}
	}

	for _, sql := range sqls[0 : len(sqls)-1] {
		sql = strings.TrimSpace(sql)
		if "" == sql {
			continue
		}
		if _, err := conn.NamedExec(sql, args); err != nil {
			return 0, err
		}
	}

	rows, err := conn.NamedQuery(sqls[len(sqls)-1], args)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		row, err := m.scanSQLRow(rows)
		if err != nil {
			continue
		}
		for _, v := range row {
			return uint32(v.(float64)), nil
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
		return nil, err
	}
	defer conn.Close()

	for i, sql := range sqls {
		if strings.TrimSpace(sql) == "" {
			sqls = append(sqls[0:i], sqls[i+1:]...)
		}
	}

	for _, sql := range sqls[0 : len(sqls)-1] {
		sql = strings.TrimSpace(sql)
		if "" == sql {
			continue
		}
		if _, err := conn.NamedExec(sql, args); err != nil {
			return nil, err
		}
	}

	rows, err := conn.NamedQuery(sqls[len(sqls)-1], args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := []map[string]interface{}{}

	for rows.Next() {
		row, err := m.scanSQLRow(rows)
		if err != nil {
			continue
		}
		ret = append(ret, row)
	}

	return interface{}(ret), nil
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

	vm := initJSVM(map[string]interface{}{"$result": data})

	v, err := vm.RunString(transformer)
	if err != nil {
		return nil, err
	}

	return v.Export(), nil
}

// authorize - run the authorizer function
func (m *Macro) authorize(input map[string]interface{}) (bool, error) {
	authorizer := strings.TrimSpace(m.Authorizer)
	if authorizer == "" {
		return true, nil
	}

	var execError error

	vm := initJSVM(map[string]interface{}{"$input": input})

	val, err := vm.RunString(m.Authorizer)
	if err != nil {
		return false, err
	}

	if execError != nil {
		return false, execError
	}

	return val.ToBoolean(), nil
}

// aggregate - run the aggregators
func (m *Macro) aggregate(input map[string]interface{}, inputKey map[string]interface{}) (map[string]interface{}, error) {
	ret := map[string]interface{}{}
	for _, k := range m.Aggregate {
		macro := m.manager.Get(k)
		if nil == macro {
			err := fmt.Errorf("不存在的宏： %s", k)
			return nil, err
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

	vm := initJSVM(map[string]interface{}{"$input": input})

	for k, src := range m.Validators {
		val, err := vm.RunString(src)
		if err != nil {
			return nil, err
		}

		if !val.ToBoolean() {
			ret = append(ret, k)
		}
	}

	return ret, err
}

// buildBind - build the bind vars
func (m *Macro) buildBind(input map[string]interface{}) (map[string]interface{}, error) {
	if len(m.Bind) < 1 {
		return nil, nil
	}

	vm := initJSVM(map[string]interface{}{"$input": input})
	ret := map[string]interface{}{}

	for k, src := range m.Bind {
		val, err := vm.RunString(src)
		if err != nil {
			return nil, err
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
			return fmt.Errorf("宏%s不存在！", name)
		}
		_, err := macro.Call(input, inputKey)
		if err != nil {
			return err
		}
	}
	return nil
}
