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
	"errors"
	"flag"
	"runtime"
)

var (
	flagDBDriver       = flag.String("driver", "", "name of SQL driver plugin")
	flagDBDSN          = flag.String("dsn", "", "connection dsn of database")
	flagRedisURL       = flag.String("redis", "", "redis url, eg：redis://:password@<redis host>:6379/0")
	flagRESTListenAddr = flag.String("port", ":80", "listen address of http")
	flagWorkers        = flag.Int("workers", runtime.NumCPU(), "worker count, default is cpus")
	flagRSAPrivkey     = flag.String("jwt-keyfile", "", "JWT rs256 alg rsa private key")
	flagJWTSecret      = flag.String("jwt-secret", "", "JWT secret")
	flagJWTExpires     = flag.Int("jwt-expires", 1800, "JWT expires seconds")
	flagName           = flag.String("name", "SQLRestful", "name of micro-service")
	flagBasePath       = flag.String("base", "/", "root path of service")
	flagDescription    = flag.String("desc", "Use SQL & Javascript to develop MicroServices of CloudNative", "description of service")
	flagVersion        = flag.String("ver", "1.0", "version of service")
	flagAuthor         = flag.String("author", "neeker", "name of author")
	flagEmail          = flag.String("email", "13317312768@qq.com", "email of author")
	flagDebug          = flag.Int("debug", 0, "debug level：0,1,2")
	flagSwagger        = flag.Bool("swagger", false, "enable swagger-ui")
	flagTrustedProxy   = flag.String("trusted-proxy", "", "address list of trusted proxy")
	flagUserAPI        = flag.String("xeai-url", "", "api address of xeai")
	flagUserScope      = flag.String("xeai-userscope", "", "scope of org")
	flagUserIDType     = flag.String("xeai-useridtype", "", "header name of user identity")
	flagAPIFile        = flag.String("config", "", "hcl file of SQLRestful")
	flagMQDriver       = flag.String("mq-driver", "", "message queue driver")
	flagMQURL          = flag.String("mq-url", "", "url of message queue")
)

var (
	errNoMacroFound       = errors.New("未知宏定义")
	errObjNotFound        = errors.New("对象不存在")
	errValidationError    = errors.New("请求参数不正确")
	errAuthorizationError = errors.New("用户未登录")
	errAccessDenyError    = errors.New("无权访问")
	errHandlerError       = errors.New("请求")
)

var (
	errStatusCodeMap = map[error]int{
		errNoMacroFound:       404,
		errObjNotFound:        404,
		errValidationError:    422,
		errAuthorizationError: 401,
		errAccessDenyError:    403,
	}
)

var (
	macrosManager    *Manager
	supportDatabases []string
)

const (
	serverVersion = "v0.13ex"
	serverBrand   = `
	
   ____   ___  _     ____           _    __       _
  / ___| / _ \| |   |  _ \ ___  ___| |_ / _|_   _| |
  \___ \| | | | |   | |_) / _ \/ __| __| |_| | | | |
   ___) | |_| | |___|  _ <  __/\__ \ |_|  _| |_| | |
  |____/ \__\_\_____|_| \_\___||___/\__|_|  \__,_|_|
										
   CloundNative                        Thanks SQLer
   Author:                neeker(13317312768@qq.com)   

	`
)
