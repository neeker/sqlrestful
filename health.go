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
	"github.com/labstack/echo"
	
	"github.com/jmoiron/sqlx"
)

// routeHealth
func routeHealth(c echo.Context) error {
	retcode := 0
	errmsg := "运行正常！"

	retdata := make(map[string]interface{})

	if *flagDBDriver != "" && *flagDBDSN != "" {
		tstconn, err := sqlx.Connect(*flagDBDriver, *flagDBDSN)
		if err != nil {
			retcode = 500
			errmsg = err.Error()
			retdata["database"] = "down"
		} else {
			retdata["database"] = "up"
		}
		
		tstconn.Close()
	} else {
		retdata["database"] = "disabled"
	}

	if redisDb != nil {
		err := redisDb.Ping().Err()
		if err != nil {
			retcode = 500
			errmsg = err.Error()
			retdata["redis"] = "down"
		} else {
			retdata["redis"] = "up"
		}
	} else {
		retdata["redis"] = "disabled"
	}

	if *flagRSAPrivkey != "" {
		retdata["jwt"] = "enabled"
	} else {
		retdata["jwt"] = "disabled"
	}

	retdata["routes"] = echoServer.Routes()

	retdata["headers"] = c.Request().Header

	respcode := 200
	if (retcode != 0) {
		respcode = retcode
	}

	return c.JSON(respcode, map[string]interface{}{
		"code": retcode,
		"message": errmsg,
		"data": retdata,
	})
}
