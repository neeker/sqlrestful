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
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
)

func main() {

	listenPortStart := strings.LastIndex(*flagRESTListenAddr, ":")
	tmpHostIP := (*flagRESTListenAddr)[0:listenPortStart]
	tmpPort := (*flagRESTListenAddr)[listenPortStart:]
	if tmpHostIP == "" {
		localAddrs, err := net.InterfaceAddrs()
		if err != nil || len(localAddrs) == 0 {
			tmpHostIP = "127.0.0.1"
		} else {
			for _, a := range localAddrs {
				if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
					tmpHostIP = ipnet.IP.String()
					break
				}
			}
		}
	}

	fmt.Println(macrosManager.ServiceBrand())
	fmt.Printf("  SQLRestful  : %s \n", serverVersion)
	fmt.Printf("  Listen Port : %s%s \n", tmpHostIP, tmpPort)
	fmt.Printf("  HCL Files   : %s\n", *flagAPIFile)
	fmt.Printf("  API Path    : %s\n", macrosManager.ServiceBasePath())
	fmt.Printf("  Service Name: %s\n", macrosManager.ServiceName())

	if macrosManager.DatabaseConfig().IsDatabaseEnabled() {
		fmt.Printf("  Database    : %s\n", macrosManager.DatabaseConfig().Driver)
	} else {
		fmt.Printf("  Database    : %s\n", "<Disabled>")
	}

	if macrosManager.DatabaseConfig().IsRedisEnabled() {
		fmt.Printf("  Redis Cache : %s\n", "<Enabled>")
	} else {
		fmt.Printf("  Redis Cache : %s\n", "<Disabled>")
	}

	if len(macrosManager.TrustedProxyList()) > 0 {
		fmt.Printf("  Trust Proxy : %v\n", macrosManager.TrustedProxyList())
	} else {
		fmt.Printf("  Trust Proxy : %v\n", "<All>")
	}

	if macrosManager.JwtIdentityConfig().IsEnabled() {
		fmt.Printf("  JWT Identity: <RS256>\n")
	} else {
		fmt.Printf("  JWT Identity: <Disabled>\n")
	}

	if macrosManager.SecurityConfig().IsEnabled() {
		fmt.Printf("  UUM Security: %s\n", macrosManager.SecurityConfig().API)
	} else {
		fmt.Printf("  UUM Security: <Disabled>\n")
	}

	if tmpPort == ":80" {
		tmpPort = ""
	}

	if macrosManager.IsSwaggerEnabled() {
		fmt.Printf("  Swagger UI  : %s\n", "http://"+tmpHostIP+tmpPort+"/swagger-ui.html")
	} else {
		fmt.Printf("  Swagger UI  : %s\n", "<Disabled>")
	}

	if *flagDebug > 0 {
		fmt.Printf("  Log Level   : %s\n", ""+strconv.Itoa(*flagDebug)+"")
	}

	fmt.Println("")
	fmt.Println("")

	err := make(chan error)

	go (func() {
		err <- startRestfulServer()
	})()

	go (func() {
		if err := startMacrosConsumeMessage(); err != nil {
			fmt.Printf("start consume message error: %s\n", err.Error())
			stopRestfulServer()
		}
	})()

	c := make(chan os.Signal)
	signal.Notify(c, os.Kill, os.Interrupt)
	go func() {
		<-c
		stopMacrosConsumeMessage()
		stopRestfulServer()
	}()

	rerr := <-err
	if rerr != nil {
		fmt.Printf("%s\n", rerr.Error())
	}

}
