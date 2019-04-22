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
	fmt.Printf("  Version: %s \n", color.GreenString(serverVersion))
	fmt.Printf("  Driver: %s \n", color.GreenString(*flagDBDriver))
	fmt.Printf("  DSN: %s \n", color.GreenString(*flagDBDSN))
	fmt.Printf("  Workers Count: %s \n", color.GreenString(strconv.Itoa(*flagWorkers)))
	fmt.Printf("  Listen Address: %s \n", color.GreenString(*flagRESTListenAddr))

	err := make(chan error)

	go (func() {
		err <- initRestfulServer()
	})()

	if err := <-err; err != nil {
		color.Red(err.Error())
	}
}
