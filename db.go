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
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"log"
)

// DatabaseConfig - 数据库配置
type DatabaseConfig struct {
	Driver      string        // 驱动名
	Dsn         string        //连接方式
	Redis       string        //Redis连接
	redisClient *redis.Client //Redis客户端连接
}

// ConnectToRedis - 连接到Redis服务器
func (c *DatabaseConfig) ConnectToRedis() error {
	redisOpts, err := redis.ParseURL(c.Redis)
	if err != nil {
		return fmt.Errorf("invalid redis url %s", c.Redis)
	}
	c.redisClient = redis.NewClient(redisOpts)
	return nil
}

// IsDatabaseEnabled - 数据库是否已启用
func (c *DatabaseConfig) IsDatabaseEnabled() bool {
	return c.Driver != "" && c.Dsn != ""
}

// IsRedisEnabled - 是否已启用Redis缓存
func (c *DatabaseConfig) IsRedisEnabled() bool {
	return c.Redis != "" && c.redisClient != nil
}

// BuildCacheKey - 获取缓存键值
func (c *DatabaseConfig) BuildCacheKey(input map[string]interface{}) string {
	if len(input) == 0 {
		return ""
	}
	return fmt.Sprintf("%v", input)
}

// PutCacheData - Put cache data
func (c *DatabaseConfig) PutCacheData(cacheNames []string, cacheKey string, val interface{}) (bool, error) {
	if !c.IsRedisEnabled() {
		if *flagDebug > 0 {
			fmt.Printf("put %s=%v to cache %v error: redis is disabled\n", cacheKey, val, cacheNames)
		}
		return false, nil
	}

	if len(cacheNames) == 0 || len(cacheKey) == 0 {
		return false, nil
	}

	var (
		ret bool
		err error
	)
	ret = false
	for _, k := range cacheNames {
		jsonData, _ := json.Marshal(val)
		ret, err = c.redisClient.HSet(k, cacheKey, string(jsonData)).Result()
		if err != nil {
			if *flagDebug > 1 {
				log.Printf("put %s to cache %s error: %v\n", cacheKey, k, err)
			}
			return false, err
		}

		if *flagDebug > 2 {
			log.Printf("put %s to cache %s: %s\n", cacheKey, k, jsonData)
		}

	}
	return ret, nil
}

// GetCacheData Get cache data
func (c *DatabaseConfig) GetCacheData(cacheNames []string, cacheKey string) (interface{}, error) {
	if !c.IsRedisEnabled() {
		if *flagDebug > 0 {
			fmt.Printf("get %s from cache %v error: redis is disabled\n", cacheKey, cacheNames)
		}
		return nil, nil
	}

	if len(cacheNames) == 0 || len(cacheKey) == 0 {
		return nil, nil
	}

	for _, k := range cacheNames {
		if c.redisClient.HExists(k, cacheKey).Val() {
			jsonData, _ := c.redisClient.HGet(k, cacheKey).Result()
			var outData interface{}
			err := json.Unmarshal([]byte(jsonData), &outData)
			if err != nil {
				if *flagDebug > 1 {
					log.Printf("get %s from cache %s error: %v\n", cacheKey, k, err)
				}
				return nil, err
			}

			if *flagDebug > 2 {
				log.Printf("get %s from %s cache data: %s\n", cacheKey, k, jsonData)
			}

			return outData, nil
		}
	}
	return nil, nil
}

// ClearCaches - clear caches
func (c *DatabaseConfig) ClearCaches(cacheNames []string) {
	if len(cacheNames) > 0 {
		for _, k := range cacheNames {
			c.redisClient.Del(k)
		}
	}
}
