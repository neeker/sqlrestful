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
	"crypto/rsa"
	"errors"
	"flag"
	"runtime"

	"github.com/go-redis/redis"
)

var (
	flagDBDriver       = flag.String("driver", "", "SQL类型")
	flagDBDSN          = flag.String("dsn", "", "SQL数据源配置")
	flagRedisURL       = flag.String("redis", "", "Redis连接：redis://:password@<redis host>:6379/0")
	flagRESTListenAddr = flag.String("port", ":80", "HTTP监听端口")
	flagWorkers        = flag.Int("workers", runtime.NumCPU(), "工作线程数量")
	flagRSAPrivkey     = flag.String("jwt-keyfile", "", "RSA私钥文件（PEM格式）")
	flagJWTSecret      = flag.String("jwt-secret", "", "JWT安全令牌")
	flagJWTExpires     = flag.Int("jwt-expires", 1800, "JWT安全令牌")
	flagName           = flag.String("name", "SQLRestful", "服务名称")
	flagBasePath       = flag.String("base", "/", "服务地址")
	flagDescription    = flag.String("desc", "SQLRestful，您的云原生应用生产力工具！", "微服务接口功能描述")
	flagVersion        = flag.String("ver", "1.0", "实现版本")
	flagAuthor         = flag.String("author", "痞子飞猪", "维护人员")
	flagEmail          = flag.String("email", "13317312768@qq.com", "联系邮箱")
	flagDebug          = flag.Int("debug", 0, "调试模式级别：0关闭、1普通、2，详细")
	flagSwagger        = flag.Bool("swagger", false, "是否开启内置SwaggerUI文档")
	flagUserAPI        = flag.String("uumapi", "", "统一用户权限管理服务地址")
	flagUserScope      = flag.String("userscope", "", "统一用户服务组织域代码")
	flagUserIDType     = flag.String("useridtype", "", "请求头中的用户标识类型")
	flagAPIFile        = flag.String("config", "./*.hcl", "缺省的配置文件路径（多个逗号分隔）")
)

var (
	errNoMacroFound       = errors.New("未知宏定义")
	errObjNotFound        = errors.New("对象不存在")
	errValidationError    = errors.New("校验出错了")
	errAuthorizationError = errors.New("用户未登录")
	errAccessDenyError    = errors.New("无权访问")
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
	macrosManager *Manager
	redisDb       *redis.Client
	jwtRSAPrivkey *rsa.PrivateKey
	jwtSecret     string
	jwtExpires    int
)

const (
	serverVersion = "v0.8ex"
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
