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

	"github.com/gorilla/websocket"
)

// WebsocketClientRegistry - Websocket客户端注册表
type WebsocketClientRegistry struct {
	Endpoint      string                     //端点
	wsConns       map[string]*websocket.Conn //websocket连接
	sendMuxs      map[string]*sync.Mutex     //send锁
	wsClientMutex *sync.RWMutex              //websocket客户锁
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
func NewWSClientRegistry(endpoint string) *WebsocketClientRegistry {
	endpointMutex.RLock()
	if endpointRegistry[endpoint] != nil {
		endpointMutex.RUnlock()
		return endpointRegistry[endpoint]
	} else {
		endpointMutex.RUnlock()
	}

	endpointMutex.Lock()
	r := &WebsocketClientRegistry{
		Endpoint:      endpoint,
		wsConns:       map[string]*websocket.Conn{},
		sendMuxs:      map[string]*sync.Mutex{},
		wsClientMutex: &sync.RWMutex{},
	}
	endpointRegistry[endpoint] = r
	endpointMutex.Unlock()
	return r
}

// AddWebsocketClient - 添加
func (r *WebsocketClientRegistry) AddWebsocketClient(id string, ws *websocket.Conn) {
	r.wsClientMutex.Lock()
	defer r.wsClientMutex.Unlock()
	r.wsConns[id] = ws
	r.sendMuxs[id] = &sync.Mutex{}
}

// RemoveWebsocketClient - 移除
func (r *WebsocketClientRegistry) RemoveWebsocketClient(id string) {
	r.wsClientMutex.Lock()
	defer r.wsClientMutex.Unlock()
	delete(r.wsConns, id)
	delete(r.sendMuxs, id)
}

// SendWebsocketMessage - 发送
func (r *WebsocketClientRegistry) SendWebsocketMessage(id string, v interface{}) error {
	r.wsClientMutex.RLock()
	ws := r.wsConns[id]
	lc := r.sendMuxs[id]
	r.wsClientMutex.RUnlock()

	if ws == nil {
		return fmt.Errorf("%s not found client %s", r.Endpoint, id)
	}

	lc.Lock()
	defer lc.Unlock()

	return ws.WriteJSON(v)
}

// BroacastWebsocketMessage - 广播
func (r *WebsocketClientRegistry) BroacastWebsocketMessage(v interface{}) error {
	r.wsClientMutex.RLock()
	defer r.wsClientMutex.RUnlock()

	for k, ws := range r.wsConns {
		lc := r.sendMuxs[k]
		lc.Lock()

		if err := ws.WriteJSON(v); err != nil {
			log.Printf("%s websocket send client %s error: %v", r.Endpoint, k, err)
		}

		lc.Unlock()
	}

	return nil
}
