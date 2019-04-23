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

	"github.com/go-redis/redis"
)

var (
	flagDBDriver       = flag.String("driver", "postgres", "SQL类型")
	flagDBDSN          = flag.String("dsn", "user=postgres password= dbname=postgres sslmode=disable connect_timeout=3", "SQL数据源配置")
	flagAPIFile        = flag.String("config", "./*.hcl", "缺省的配置文件路径（多个文件使用逗号分隔）")
	flagRedisURL       = flag.String("redis", "", "Redis连接：redis://:password@<redis host>:6379/0")
	flagRESTListenAddr = flag.String("port", ":80", "HTTP监听端口")
	flagWorkers        = flag.Int("workers", runtime.NumCPU(), "工作线程数量")
	flagSQLSeparator   = flag.String("sep", `---\\--`, "SQL分隔符")
)

var (
	errNoMacroFound       = errors.New("资源不存在！")
	errValidationError    = errors.New("校验出错了！")
	errAuthorizationError = errors.New("验证失败了！")
)

var (
	errStatusCodeMap = map[error]int{
		errNoMacroFound:       404,
		errValidationError:    422,
		errAuthorizationError: 401,
	}
)

var (
	macrosManager *Manager
	redisDb       *redis.Client
)

const (
	serverVersion = "v0.3"
	serverBrand   = `
	
   ____   ___  _     ____           _    __       _
  / ___| / _ \| |   |  _ \ ___  ___| |_ / _|_   _| |
  \___ \| | | | |   | |_) / _ \/ __| __| |_| | | | |
   ___) | |_| | |___|  _ <  __/\__ \ |_|  _| |_| | |
  |____/ \__\_\_____|_| \_\___||___/\__|_|  \__,_|_|
										
  嘿，谁用谁知道爽~               特别感谢SQLer
                 --痞子飞猪(13317312768@qq.com)   
  
	`
)
