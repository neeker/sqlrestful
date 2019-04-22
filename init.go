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

	_ "github.com/SAP/go-hdb/driver"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/kshvakov/clickhouse"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"

	"github.com/alash3al/go-color"
	"github.com/jmoiron/sqlx"
)

func init() {
	usage := flag.Usage
	flag.Usage = func() {
		fmt.Println(color.MagentaString(serverBrand))
		usage()
	}

	flag.Parse()
	runtime.GOMAXPROCS(*flagWorkers)

	{
		tstconn, err := sqlx.Connect(*flagDBDriver, *flagDBDSN)
		if err != nil {
			fmt.Println(color.RedString("[%s] %s - connection error - (%s)", *flagDBDriver, *flagDBDSN, err.Error()))
			os.Exit(0)
		}
		tstconn.Close()
	}

	{
		manager, err := NewManager(*flagAPIFile)
		if err != nil {
			fmt.Println(color.RedString("(%s)", err.Error()))
			os.Exit(0)
		}
		macrosManager = manager
	}
}
