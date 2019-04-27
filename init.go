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
	"flag"
	"fmt"
	"os"
	"runtime"
	"io/ioutil"

	_ "github.com/SAP/go-hdb/driver"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/kshvakov/clickhouse"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-oci8"

	"github.com/alash3al/go-color"
	"github.com/jmoiron/sqlx"

	"github.com/go-redis/redis"

	"github.com/dgrijalva/jwt-go"
)

func init() {
	usage := flag.Usage
	flag.Usage = func() {
		fmt.Println(color.MagentaString(serverBrand))
		usage()
	}

	flag.Parse()
	runtime.GOMAXPROCS(*flagWorkers)

	if *flagDBDriver != "" && *flagDBDSN != "" {
		tstconn, err := sqlx.Connect(*flagDBDriver, *flagDBDSN)
		if err != nil {
			fmt.Println(color.RedString("[%s] %s 连接出错：%s", *flagDBDriver, *flagDBDSN, err.Error()))
			os.Exit(0)
		}
		tstconn.Close()
	}

	//https://github.com/go-redis/redis
	if *flagRedisURL != "" {
		redisOpts, err := redis.ParseURL(*flagRedisURL)
		if err != nil {
			fmt.Println(color.RedString("[redis] %s 不正确：%s", *flagRedisURL, err.Error()))
			os.Exit(0)
		}

		redisClient := redis.NewClient(redisOpts);
		err = redisClient.Ping().Err()
		
		if err != nil {
			fmt.Println(color.RedString("[redis] %s 连接出错：%s", *flagRedisURL, err.Error()))
			os.Exit(0)
		}
		redisDb = redisClient
	}

	if *flagRSAPrivkey != "" {
		if *flagJWTSecret == "" {
			fmt.Println(color.RedString("[jwt] JWT 安全令牌不能为空！"))
			os.Exit(0)
		}

		rsaKeyData, err := ioutil.ReadFile(*flagRSAPrivkey)
		if err != nil {
			fmt.Println(color.RedString("[jwt] 加载 JWT RSA 私钥文件出错：%s", err.Error()))
			os.Exit(0)
		}

		tmpPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(rsaKeyData)
		if err != nil {
			fmt.Println(color.RedString("[jwt] JWT RSA 私钥格式错误：%s", err.Error()))
			os.Exit(0)
		}

		if *flagJWTExpires < int(10) {
			fmt.Println(color.RedString("[jwt] JWT 令牌有效期必须大于10秒！"))
			os.Exit(0)
		} 

		jwtRSAPrivkey = tmpPrivateKey
		jwtSecret = *flagJWTSecret
		jwtExpires = *flagJWTExpires

	} 

	{
		manager, err := NewManager(*flagAPIFile)
		if err != nil {
			fmt.Println(color.RedString("HCL配置错误: %s", err.Error()))
			os.Exit(0)
		}
		macrosManager = manager
	}
}
