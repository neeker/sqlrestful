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
	"crypto/rsa"
	"github.com/dgrijalva/jwt-go"
	"time"
)

// JwtConfig - JWT定义
type JwtConfig struct {
	Rsa     string
	Secret  string
	Expires int
	privkey *rsa.PrivateKey
	file    string
}

// IsEnabled - 是否启用了JWT配置
func (j *JwtConfig) IsEnabled() bool {
	return j.privkey != nil && j.Secret != ""
}

// CreateRequestToken - 创建新的请求令牌
func (j *JwtConfig) CreateRequestToken() (string, error) {
	claims := &jwt.StandardClaims{
		ExpiresAt: time.Now().Unix() + int64(j.Expires),
		Issuer:    j.Secret,
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	ret, err := jwtToken.SignedString(j.privkey)
	return ret, err
}
