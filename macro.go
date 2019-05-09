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
	"strconv"
	"strings"
	"time"

	"os/exec"

	"github.com/jmoiron/sqlx"
)

// Cache - 缓存配置
type Cache struct {
	Put  []string //设置缓存列表
	Evit []string //移除缓存列表
}

// Author 作者
type Author struct {
	Name  string //名称
	Email string //邮件
	URL   string //URL
}

// Macro - a macro configuration
type Macro struct {
	Brand        string                       //起始标记
	Base         string                       // 起始地址
	Name         string                       //名称
	Desc         string                       //描述
	Version      string                       //版本
	Author       *Author                      //作者
	Swagger      bool                         //是否启用SWAGGER-UI
	Debug        int                          //调试级别
	Const        map[string]string            //常量
	Methods      []string                     //请求方法
	Include      []string                     //引用宏列表
	Validators   map[string]string            //参数校验
	Authorizer   string                       //身份校验
	Security     *SecurityConfig              //安全验证配置
	Jwt          *JwtConfig                   //JWT身份配置
	Mq           *MessageQueueConfig          //消息队列配置
	Consume      map[string]string            //消费消息配置
	Db           *DatabaseConfig              //数据库配置
	Bind         map[string]string            //绑定表达式
	Impl         string                       //实现语言：js、sql、cmd
	Format       string                       //应答格式：enclosed（封装）、origin（原样）
	Exec         string                       //执行实现
	Provider     string                       //实现提供器
	Aggregate    []string                     //组合实现
	Transformer  string                       //转换器
	Tags         []string                     //定义标签
	Model        map[string]map[string]string //应答模型
	Proxy        []string                     //前置代理
	Summary      string                       //接口概述 Desc优先
	Path         string                       //实现路径
	Total        string                       //分页查询的记录总数实现
	Result       string                       //结果类型：page、list、object、nil
	Get          *Macro                       //GET实现
	Post         *Macro                       //POSTs实现
	Put          *Macro                       //PUT实现
	Patch        *Macro                       //PATCH实现
	Delete       *Macro                       //DELETE实现
	Cache        *Cache                       //缓存配置
	file         string                       //实现文件
	name         string                       //宏名称
	rolesMap     map[string]bool              //要求角色
	usersMap     map[string]bool              //排除角色
	manager      *Manager                     //管理器
	methodMacros map[string]*Macro            //内置的方法宏
	consts       map[string]interface{}       //常量表
	mqp          MessageQueueProvider         //提供器实现
}

// Call - executes the macro
func (m *Macro) Call(input map[string]interface{}, inputKey map[string]interface{}) (interface{}, error) {

	ok, err := m.filterSecurity(input)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errAccessDenyError
	}

	ok, err = m.authorize(input)
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
		if input["offset"] == nil || input["offset"] == "" {
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
	if m.manager.DatabaseConfig().IsRedisEnabled() && m.Cache != nil && (len(m.Cache.Put) > 0 || len(m.Cache.Evit) > 0) {
		cacheKey = m.manager.DatabaseConfig().BuildCacheKey(inputKey)

		if cacheKey != "" && len(m.Cache.Put) > 0 {
			out, err = m.manager.DatabaseConfig().GetCacheData(m.Cache.Put, cacheKey)
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

	pageTotal := m.Total
	execScript := m.Exec
	scriptImpl := m.Impl

	if m.Provider != "" {
		resolvedVar, err := m.resolveExecScript(m.Provider, input)

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
		if m.manager.DatabaseConfig().IsRedisEnabled() && m.Cache != nil && len(m.Cache.Put) > 0 {
			m.manager.DatabaseConfig().PutCacheData(m.Cache.Put, cacheKey, pageRet)
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
	if m.manager.DatabaseConfig().IsRedisEnabled() && m.Cache != nil &&
		(len(m.Cache.Put) > 0 || len(m.Cache.Evit) > 0) {
		m.manager.DatabaseConfig().PutCacheData(m.Cache.Put, cacheKey, out)
		m.manager.DatabaseConfig().ClearCaches(m.Cache.Evit)
	}

	return out, nil
}

// execSQLTotal - execute the specified sql query
func (m *Macro) execSQLTotal(sqls []string, input map[string]interface{}) (int64, error) {

	if !m.manager.DatabaseConfig().IsDatabaseEnabled() {
		return 0, fmt.Errorf("Database forget enable??")
	}

	args, err := m.buildBind(input)
	if err != nil {
		return 0, err
	}

	conn, err := sqlx.Open(m.manager.DatabaseConfig().Driver, m.manager.DatabaseConfig().Dsn)
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

	if !m.manager.DatabaseConfig().IsDatabaseEnabled() {
		return nil, fmt.Errorf("Database forget enable??")
	}

	args, err := m.buildBind(input)
	if err != nil {
		return nil, err
	}

	conn, err := sqlx.Open(m.manager.DatabaseConfig().Driver, m.manager.DatabaseConfig().Dsn)
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
				log.Printf("%s exec sql%d fetch rows error:\n%v\n==sql==\n%s\n==rows==\n%v\n\n",
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
		"$const": m.consts,
		"$input": input,
		"$name":  m.name,
		"$stage": "provider",
	})

	if *flagDebug > 2 {
		log.Printf("run %s provider(js):\n==js==\n%s\n\n", m.name, javascript)
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
		"$const": m.consts,
		"$input": input,
		"$name":  m.name,
		"$stage": "exec",
	})

	if *flagDebug > 2 {
		log.Printf("run %s exec js:\n==js==\n%s\n==input==\n%v\n\n", m.name, javascript, input)
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
		"$const": m.consts,
		"$input": input,
		"$name":  m.name,
		"$stage": "total",
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
		"$const":  m.consts,
		"$result": data,
		"$name":   m.name,
		"$stage":  "transformer",
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
		"$const": m.consts,
		"$input": input,
		"$name":  m.name,
		"$stage": "authorizer",
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
		"$const": m.consts,
		"$input": input,
		"$name":  m.name,
		"$stage": "validators",
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

func (m *Macro) buildConst() (map[string]interface{}, error) {
	if len(m.Const) == 0 {
		return nil, nil
	}

	vm := initJSVM(map[string]interface{}{
		"$name":  m.name,
		"$stage": "const",
	})
	ret := map[string]interface{}{}
	for k, src := range m.Const {

		if *flagDebug > 2 {
			log.Printf("run %s const(%s): %s\n", m.name, k, src)
		}

		val, err := vm.RunString(src)
		if err != nil {
			if *flagDebug > 0 {
				log.Printf("run %s const(%s=\"%s\") error: %v\n", m.name, k, src, err)
			}
			return nil, fmt.Errorf("run %s const(%s=\"%s\") error: %v", m.name, k, src, err)
		}
		ret[k] = val.Export()

	}

	return ret, nil
}

// buildBind - build the bind vars
func (m *Macro) buildBind(input map[string]interface{}) (map[string]interface{}, error) {
	if len(m.Bind) == 0 {
		return nil, nil
	}

	vm := initJSVM(map[string]interface{}{
		"$const": m.consts,
		"$input": input,
		"$name":  m.name,
		"$stage": "bind",
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
func (m *Macro) execCommandTotal(cmdline string, input map[string]interface{}) (int64, error) {
	args, err := m.buildBind(input)
	if err != nil {
		return 0, err
	}

	cmdExecute, inputArgs := getCommandDefines(cmdline)

	for k, v := range args {
		inputArgs = append(inputArgs, k)
		switch v.(type) {
		case string:
			if v.(string) != "" {
				inputArgs = append(inputArgs, v.(string))
			}
		}
	}

	cmd := exec.Command(cmdExecute, inputArgs[0:]...)

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

// execCommand - execute the command line
func (m *Macro) execCommand(cmdline string, input map[string]interface{}) (interface{}, error) {
	args, err := m.buildBind(input)
	if err != nil {
		return 0, err
	}

	cmdExecute, inputArgs := getCommandDefines(cmdline)

	for k, v := range args {
		inputArgs = append(inputArgs, k)
		switch v.(type) {
		case string:
			if v.(string) != "" {
				inputArgs = append(inputArgs, v.(string))
			}
		}
	}

	cmd := exec.Command(cmdExecute, inputArgs[0:]...)

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

func (m *Macro) isAnonymousAllow() bool {
	return m.Security == nil || m.Security.Anonymous
}

// filterSecurity - run the filterSecurity config
func (m *Macro) filterSecurity(input map[string]interface{}) (bool, error) {
	var (
		userid  string
		idtype  string
		scope   string
		options map[string]interface{}
	)

	//如果允许匿名则直接返回
	if m.isAnonymousAllow() {
		return true, nil
	}

	//获取用户ID
	if m.manager.meta.Security.From != "" {
		userid, _ = input[m.manager.meta.Security.From].(string)
		if m.manager.meta.Security.Idtype != "" {
			idtype = m.manager.meta.Security.Idtype
		} else {
			idtype = "uname"
		}
	} else {
		userid, _ = input["http_x_credential_userid"].(string)
		idtype = "id"

		if userid == "" {
			//获取用户名
			userid, _ = input["http_x_credential_username"].(string)
			if userid == "" {
				//获取TAM兼容用户
				userid, _ = input["http_iv_user"].(string)
			}
			//从命令行配置中获取ID类型
			if m.manager.meta.Security.Idtype != "" {
				idtype = m.manager.meta.Security.Idtype
			} else {
				idtype = "uname"
			}
		}

	}

	//用户未登录则直接退出
	if userid == "" {
		if *flagDebug > 2 {
			log.Printf("%s run security deny: user(%s) not found\n", m.name, userid)
		}
		return false, nil
	}

	//获取用户组织域
	if m.manager.meta.Security.Scope != "" {
		scope, _ = input["http_x_user_scope"].(string)
		if scope != m.manager.meta.Security.Scope {
			if *flagDebug > 2 {
				log.Printf("%s run security deny:  %s not equals %s\n", m.name, scope, m.manager.meta.Security.Scope)
			}
			return false, nil
		}
	}

	//判断用户是否有权访问
	if m.usersMap != nil && len(m.usersMap) > 0 {
		_, inUsers := m.usersMap[userid]
		if m.Security.Policy == "deny" {
			if inUsers {
				if *flagDebug > 2 {
					log.Printf("%s run security user deny: %s in %v\n", m.name, userid, m.usersMap)
				}
				return false, nil
			}
		} else if !inUsers {
			if *flagDebug > 2 {
				log.Printf("%s run security user allow: %s not in %v\n", m.name, userid, m.usersMap)
			}
			return false, nil
		}
	}

	options = make(map[string]interface{})

	//prepare jsJWTFetchfunc options params
	options["method"] = "GET"

	scopeValue := ""
	if scope != "" {
		scopeValue = "&scope=" + scope
	}

	//帐号获取接口地址
	accAPIURL := fmt.Sprintf("%s/get_user_account?userid=%s&idtype=%s%s&contain_roles=true&timestamp=%d",
		m.manager.meta.Security.API, userid, idtype, scopeValue, time.Now().UnixNano())

	out, err := jsJWTFetchfunc(accAPIURL, options)

	if err != nil {
		if *flagDebug > 0 {
			log.Printf("%s run security fetch (%s) error: %v\n",
				m.name, accAPIURL, err)
		}
		return false, err
	}

	resultCode, codeFound := out["code"].(int)

	if codeFound {
		if *flagDebug > 0 {
			log.Printf("%s run security call (%s) error: %v\n",
				m.name, accAPIURL, out)
		}
		return false, fmt.Errorf("%s run security call (%s) error: %v",
			m.name, accAPIURL, out)
	}

	if resultCode == 404 {
		if *flagDebug > 0 {
			log.Printf("%s run security call (%s) user not found: %v\n",
				m.name, accAPIURL, out["message"])
		}
		return false, nil
	}

	if resultCode != 0 {
		if *flagDebug > 0 {
			log.Printf("%s run security call (%s) error: %v\n",
				m.name, accAPIURL, out["message"])
		}
		return false, fmt.Errorf("%v", out["message"])
	}

	userItem, _ := out["data"].(map[string]interface{})

	if userItem == nil {
		if *flagDebug > 0 {
			log.Printf("%s run security call (%s) error data: %v\n", m.name, accAPIURL, out)
		}
		return false, fmt.Errorf("%s run security call (%s) error data: %v", m.name, accAPIURL, out)
	}

	userRoles, _ := userItem["roles"].([]interface{})

	if m.rolesMap != nil && len(m.rolesMap) > 0 {

		if m.Security.Policy == "deny" {
			for _, r := range userRoles {
				rn, cv := r.(string)
				if !cv {
					continue
				}
				if _, inRoles := m.rolesMap[rn]; inRoles {
					if *flagDebug > 2 {
						log.Printf("%s run security roles deny: user(%s) role(%s) in %v\n",
							m.name, userid, rn, m.rolesMap)
					}
					return false, nil
				}
			}
		} else {
			userRolesMap := map[string]bool{}
			for _, r := range userRoles {
				rn, cv := r.(string)
				if !cv {
					continue
				}
				userRolesMap[rn] = true
			}
			for k := range m.rolesMap {
				if _, inRoles := userRolesMap[k]; !inRoles {
					if *flagDebug > 2 {
						log.Printf("%s run security roles allow: user(%s) role(%s) not in %v\n",
							m.name, userid, k, m.rolesMap)
					}
					return false, nil
				}
			}
		}

		return true, nil
	}

	return true, nil
}

// IsMessageConsumeEnabled - 是否启用了消费消息
func (m *Macro) IsMessageConsumeEnabled() bool {
	return m.Consume != nil && (m.Consume["name"] != "" ||
		m.Consume["topic"] != "" ||
		m.Consume["queue"] != "")
}

// ConsumeMessage - 消费消息
func (m *Macro) ConsumeMessage() error {
	return m.mqp.Consume()
}

// ShutdownConsume - 停止消费
func (m *Macro) ShutdownConsume() error {
	return m.mqp.Shutdown()
}
