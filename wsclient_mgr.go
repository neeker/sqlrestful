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
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebsocketClientHolder - WebSocket客户端
type WebsocketClientHolder struct {
	id           string          //客户端ID
	start        time.Time       //开始时间
	ws           *websocket.Conn //websocket连接
	sendMux      *sync.Mutex     //发送消息锁
	msgTimestamp int64           //消息时间
}

// WebsocketClientRegistry - Websocket客户端注册表
type WebsocketClientRegistry struct {
	Endpoint      string                            //端点
	keepalive     int                               //保持keepalive
	clients       map[string]*WebsocketClientHolder //客户端
	wsClientMutex *sync.RWMutex                     //websocket客户锁
}

var (
	endpointRegistry map[string]*WebsocketClientRegistry
	endpointMutex    *sync.RWMutex
)

// InitEndpointRegistry - InitEndpointRegistry
func InitEndpointRegistry() {
	if endpointRegistry != nil {
		return
	}
	endpointMutex = &sync.RWMutex{}
	endpointRegistry = map[string]*WebsocketClientRegistry{}
}

// NewWSClientRegistry - 创建
func NewWSClientRegistry(endpoint string, keepalive int) *WebsocketClientRegistry {
	endpointMutex.RLock()
	if endpointRegistry[endpoint] != nil {
		endpointMutex.RUnlock()
		return endpointRegistry[endpoint]
	}
	endpointMutex.RUnlock()

	endpointMutex.Lock()
	r := &WebsocketClientRegistry{
		Endpoint:      endpoint,
		clients:       map[string]*WebsocketClientHolder{},
		wsClientMutex: &sync.RWMutex{},
	}

	if keepalive > 0 {
		r.keepalive = keepalive
	}

	go r.keepaliveWebsocketClients()

	endpointRegistry[endpoint] = r
	endpointMutex.Unlock()
	return r
}

// GetWSClientRegistry - 获取
func GetWSClientRegistry(endpoint string) *WebsocketClientRegistry {
	endpointMutex.RLock()
	defer endpointMutex.RUnlock()
	return endpointRegistry[endpoint]
}

// AddWebsocketClient - 添加
func (r *WebsocketClientRegistry) AddWebsocketClient(id string, ws *websocket.Conn) *WebsocketClientHolder {
	r.wsClientMutex.Lock()
	defer r.wsClientMutex.Unlock()
	wsHolder := &WebsocketClientHolder{
		id:      id,
		ws:      ws,
		start:   time.Now(),
		sendMux: &sync.Mutex{},
	}
	r.clients[id] = wsHolder
	return wsHolder
}

// RemoveWebsocketClient - 移除
func (r *WebsocketClientRegistry) RemoveWebsocketClient(id string) {
	r.wsClientMutex.Lock()
	defer r.wsClientMutex.Unlock()
	delete(r.clients, id)
}

// SendWebsocketMessage - 发送
func (r *WebsocketClientRegistry) SendWebsocketMessage(id string, v interface{}) error {
	r.wsClientMutex.RLock()
	wsHolder := r.clients[id]

	if wsHolder == nil {
		return fmt.Errorf("%s websocket client(%s) not found", r.Endpoint, id)
	}

	ws := wsHolder.ws
	lc := wsHolder.sendMux

	r.wsClientMutex.RUnlock()
	lc.Lock()
	defer lc.Unlock()

	wsHolder.msgTimestamp = time.Now().Unix()

	return ws.WriteJSON(v)
}

// BroacastWebsocketMessage - 广播
func (r *WebsocketClientRegistry) BroacastWebsocketMessage(v interface{}) error {
	r.wsClientMutex.RLock()
	defer r.wsClientMutex.RUnlock()

	for k, c := range r.clients {

		lc := c.sendMux
		lc.Lock()
		ws := c.ws

		c.msgTimestamp = time.Now().Unix()
		if err := ws.WriteJSON(v); err != nil {
			log.Printf("%s websocket send client %s error: %v", r.Endpoint, k, err)
		}

		lc.Unlock()
	}

	return nil
}

func (r *WebsocketClientRegistry) keepaliveWebsocketClients() {

	if r.keepalive <= 0 {
		return
	}

	c := time.Tick(time.Duration(r.keepalive) * time.Second)

	for {
		<-c
		r.wsClientMutex.RLock()
		nu := time.Now().Unix()
		for k, c := range r.clients {
			if (nu - c.msgTimestamp) >= int64(r.keepalive) {
				if err := c.ws.PingHandler()("PING"); err != nil {
					log.Printf("%s websocket client(%s) ping error: %v", k, c.id, err)
				}
			}
		}
		r.wsClientMutex.RUnlock()

	}

}
