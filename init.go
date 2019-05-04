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
	"io/ioutil"
	"os"
	"runtime"

	_ "github.com/SAP/go-hdb/driver"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/kshvakov/clickhouse"
	_ "github.com/lib/pq"

	"github.com/go-redis/redis"

	"github.com/dgrijalva/jwt-go"
)

func init() {

	supportDatabases = []string {
		"postgres(PostgresQL)",
		"mysql(MySQL)",
		"mssql(SQLServer)",
		"hdb(SAP HANA)",
		"clickhouse(Yandex ClickHouse)",
	}

	preLoadSQLite3()
	preLoadOci8()

	usage := flag.Usage
	flag.Usage = func() {
		fmt.Println(serverBrand)
		usage()
	}

	flag.Parse()
	runtime.GOMAXPROCS(*flagWorkers)

	//https://github.com/go-redis/redis
	if *flagRedisURL != "" {
		_, err := redis.ParseURL(*flagRedisURL)
		if err != nil {
			fmt.Println(fmt.Sprintf("redis url(%s) error：%v", *flagRedisURL, err))
			os.Exit(0)
		}
	}

	if *flagRSAPrivkey != "" {
		if *flagJWTSecret == "" {
			fmt.Println("jwt secret is empty!")
			os.Exit(0)
		}

		rsaKeyData, err := ioutil.ReadFile(*flagRSAPrivkey)
		if err != nil {
			fmt.Println("load rsa private key error:", err.Error())
			os.Exit(0)
		}

		_, err = jwt.ParseRSAPrivateKeyFromPEM(rsaKeyData)
		if err != nil {
			fmt.Println("jwt rsa private key error:", err.Error())
			os.Exit(0)
		}

		if *flagJWTExpires < int(10) {
			*flagJWTExpires = 1800
		}

	}

	if *flagAPIFile == "" {
		flag.Usage()
		os.Exit(0)
	}

	{
		manager, err := NewManager(*flagAPIFile)
		if err != nil {
			fmt.Printf("Run SQLRestful macro error: %v!", err)
			os.Exit(0)
		}
		macrosManager = manager
	}
}
