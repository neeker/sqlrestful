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

// SecurityConfig - 配置
type SecurityConfig struct {
	API       string   //统一安全服务API地址
	From      string   //用户标识来自请求头名称
	Idtype    string   //用户标识类型
	Anonymous bool     //是否允许匿名
	Scope     string   //用户组织范围
	Roles     []string //可访问的角色
	Users     []string //可访问的用户
	Policy    string   //判定策略，include表示包含、exclude表示排除
}

// IsEnabled - return is enable security config
func (c *SecurityConfig) IsEnabled() bool {
	return c.API != ""
}
