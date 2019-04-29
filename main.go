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
	"fmt"
	"strconv"
)

func main() {
	fmt.Println(serverBrand)
	fmt.Printf("  工具版本: %s \n", serverVersion)

	if *flagDBDriver == "" || *flagDBDSN == "" {
		fmt.Printf("  SQL 驱动: %s \n", "<未配置>")
		fmt.Printf("  连接字串: %s \n", "<未配置>")
	} else {
		fmt.Printf("  SQL 驱动: %s \n", *flagDBDriver)
		fmt.Printf("  连接字串: %s \n", *flagDBDSN)
	}

	fmt.Printf("  工作线程: %s \n", strconv.Itoa(*flagWorkers))
	fmt.Printf("  监听端口: %s \n", *flagRESTListenAddr)

	if *flagRedisURL == "" {
		fmt.Printf("  Redis 缓存: %s \n", "<未配置>")
	} else {
		fmt.Printf("  Redis 缓存: %s \n", *flagRedisURL)
	}

	if *flagRSAPrivkey == "" || *flagJWTSecret == "" {
		fmt.Printf("  JWT RSA私钥: %s \n", "<未配置>")
		fmt.Printf("  JWT 安全令牌: %s \n", "<未配置>")
		fmt.Printf("  JWT 令牌期限: %s \n", "<未生效>")
	} else {
		fmt.Printf("  JWT RSA私钥: %s \n", *flagRSAPrivkey)
		fmt.Printf("  JWT 安全令牌: %s \n", *flagJWTSecret)
		fmt.Printf("  JWT 令牌期限: %s \n", strconv.Itoa(*flagJWTExpires)+"秒")
	}

	fmt.Printf("         \n")
	fmt.Printf("  服务地址: %s\n", *flagBasePath)
	fmt.Printf("  服务名称: %s\n", *flagName)
	fmt.Printf("  实现脚本: %s\n", *flagAPIFile)
	fmt.Printf("  功能描述: %s\n", *flagDescription)
	fmt.Printf("  实现版本: %s\n", *flagVersion)
	fmt.Printf("  维护人员: %s\n", *flagAuthor)
	fmt.Printf("  联系邮箱: %s\n", *flagEmail)

	if *flagSwagger {
		fmt.Printf("  SwaggerUI: %s\n", "http://"+*flagRESTListenAddr+"/swagger-ui.html")
	} else {
		fmt.Printf("  SwaggerUI: %s\n", "<未配置>")
	}

	if *flagDebug > 0 {
		fmt.Printf("  输出日志: %s\n", "已开启"+strconv.Itoa(*flagDebug)+"级日志")
	}

	fmt.Println("")
	fmt.Println("")

	err := make(chan error)

	go (func() {
		err <- initRestfulServer()
	})()

	if err := <-err; err != nil {
		fmt.Printf("%s", err.Error())
	}
}
