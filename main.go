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

	"github.com/alash3al/go-color"
)

func main() {
	fmt.Println(color.MagentaString(serverBrand))
	fmt.Printf("  版本: %s \n", color.GreenString(serverVersion))
	fmt.Printf("  驱动: %s \n", color.GreenString(*flagDBDriver))
	fmt.Printf("  连接: %s \n", color.GreenString(*flagDBDSN))
	fmt.Printf("  线程: %s \n", color.GreenString(strconv.Itoa(*flagWorkers)))
	fmt.Printf("  监听: %s \n", color.GreenString(*flagRESTListenAddr))
	if *flagRedisURL == "" {
		fmt.Printf("  缓存: %s \n", color.RedString("未配置Redis连接"))
	} else {
		fmt.Printf("  缓存: %s \n", color.GreenString(*flagRedisURL))
	}

	err := make(chan error)

	go (func() {
		err <- initRestfulServer()
	})()

	if err := <-err; err != nil {
		color.Red(err.Error())
	}
}
