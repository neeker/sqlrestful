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
	"runtime"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo"
)

// routeHealth
func routeHealth(c echo.Context) error {
	retcode := 0
	errmsg := "运行正常！"

	retdata := make(map[string]interface{})

	if macrosManager.DatabaseConfig().IsDatabaseEnabled() {
		tstconn, err := sqlx.Connect(macrosManager.DatabaseConfig().Driver, macrosManager.DatabaseConfig().Dsn)
		if err != nil {
			retcode = 500
			errmsg = err.Error()
			retdata["database"] = "down"
		} else {
			retdata["database"] = "up"
			tstconn.Close()
		}
	} else {
		retdata["database"] = "disabled"
	}

	retdata["runtime"] = map[string]interface{}{
		"arch": runtime.GOARCH,
		"os":   runtime.GOOS,
	}

	retdata["provider"] = map[string]interface{} {
		"name": "SQLRestful",
		"version": serverVersion,
		"supports": supportDatabases,
	}

	if macrosManager.DatabaseConfig().IsRedisEnabled() {
		err := macrosManager.DatabaseConfig().redisClient.Ping().Err()
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

	if macrosManager.JwtIdentityConfig().IsEnabled() {
		retdata["jwt"] = "enabled"
	} else {
		retdata["jwt"] = "disabled"
	}

	respcode := 200
	if retcode != 0 {
		respcode = retcode
	}

	return c.JSON(respcode, map[string]interface{}{
		"code":    retcode,
		"message": errmsg,
		"data":    retdata,
	})
}
