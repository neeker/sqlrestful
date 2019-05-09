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
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-stomp/stomp"
)

// STOMPMessageQueueProvider - STOP协议提供
type STOMPMessageQueueProvider struct {
	conn     *stomp.Conn
	macro    *Macro
	timeout  int
	uri      *url.URL
	tries    int
	clock    *sync.RWMutex
	shutdown bool
	failover bool
}

var stompConn *stomp.Conn

// NewSTOMP - 创建STOMP协议提供器
func NewSTOMP(macro *Macro, config *MessageQueueConfig) (MessageQueueProvider, error) {
	u, err := url.Parse(config.URI)
	if err != nil {
		return nil, err
	}

	connTimeout := config.Timeout
	if connTimeout <= 0 {
		connTimeout = 3
	}

	return &STOMPMessageQueueProvider{
		conn:     nil,
		macro:    macro,
		timeout:  connTimeout,
		uri:      u,
		tries:    0,
		clock:    &sync.RWMutex{},
		shutdown: false,
		failover: macro.Consume["failover"] == "" ||
			strings.ToLower(macro.Consume["failover"]) == "true" ||
			strings.ToLower(macro.Consume["failover"]) == "on",
	}, nil

}

// IsShutdown - 是否已停止
func (c *STOMPMessageQueueProvider) IsShutdown() bool {
	c.clock.RLock()
	defer c.clock.RUnlock()
	return c.shutdown
}

func (c *STOMPMessageQueueProvider) connect() (err error) {
	if c.conn != nil {
		if err := c.conn.Disconnect(); err != nil {
			log.Printf("%s mq(%s) dis connect error: %+v", c.macro.name, c.uri.String(), err)
		}
	}

	u := c.uri
	connTimeout := c.timeout
	userPassword, _ := u.User.Password()

	c.conn, err = stomp.Dial(u.Scheme, u.Host,
		stomp.ConnOpt.Login(u.User.Username(), userPassword),
		stomp.ConnOpt.HeartBeat(time.Duration(connTimeout)*time.Second, time.Duration(connTimeout)*time.Second),
		stomp.ConnOpt.HeartBeatError(time.Duration(connTimeout*3)*time.Second))

	if err != nil {
		return err
	}

	return nil
}

// QueueName - 获取队列名称
func (c *STOMPMessageQueueProvider) QueueName() string {
	args := c.macro.Consume
	queueName := args["queue"]
	if queueName == "" {
		queueName = args["topic"]
		if queueName == "" {
			queueName = args["name"]
		}
	}
	return queueName
}

// Consume - 消费数据
func (c *STOMPMessageQueueProvider) Consume() error {

	if c.IsShutdown() {
		return nil
	}

	queueName := c.QueueName()
	if queueName == "" {
		return fmt.Errorf("consume args must be set queue or topic or name")
	}

	c.tries++

	if err := c.connect(); err != nil {
		if strings.Contains(err.Error(), "password is invalid") {
			return err
		}

		if c.failover {
			log.Printf("%s mq(%s) consume %s connect error: %+v", c.macro.name, c.uri.String(), queueName, err)
			go func() {
				if *flagDebug > 0 {
					log.Printf("wait 500 ms retry %s mq(%s) consume %s connect", c.macro.name, c.uri.String(), queueName)
				}
				time.Sleep(500 * time.Millisecond)
				c.Consume()
			}()
			return nil
		}
		return err
	}

	args := c.macro.Consume
	ackFlag := strings.ToLower(args["ack"])
	ackFlagVal := stomp.AckAuto

	switch ackFlag {
	case "client":
		ackFlagVal = stomp.AckClient
	case "each":
		ackFlagVal = stomp.AckClientIndividual
	}

	sub, err := c.conn.Subscribe(queueName, ackFlagVal)

	if err != nil {
		if c.failover {
			log.Printf("%s mq(%s) consume %s error: %+v", c.macro.name, c.uri.String(), queueName, err)
			go c.Consume()
			return nil
		}
		return err
	}

	if *flagDebug > 0 {
		log.Printf("STOMP %s subscribed", queueName)
	}

	go stompHandleMessage(sub.C, c.macro)

	return nil
}

// Shutdown - 停止监听
func (c *STOMPMessageQueueProvider) Shutdown() error {
	if c.conn == nil || c.IsShutdown() {
		return nil
	}

	c.clock.Lock()
	c.shutdown = true
	c.clock.Unlock()

	if err := c.conn.Disconnect(); err != nil {
		return err
	}
	if *flagDebug > 2 {
		defer log.Printf("STOMP shutdown OK")
	}
	return nil
}

// stompHandleMessage - 处理消息
func stompHandleMessage(message <-chan *stomp.Message, m *Macro) {
	for d := range message {
		if d.Err != nil {
			if *flagDebug > 0 && !m.mqp.IsShutdown() {
				log.Printf("STOMP subscribe error: %v", d.Err)
			}
			if !m.mqp.IsShutdown() {
				go m.mqp.Consume()
			}
			return
		}
		if *flagDebug > 2 {
			log.Printf(
				"STOMP got %dB delivery: %q",
				len(d.Body),
				d.Body,
			)
		}

		var msgdata map[string]interface{}
		if err := json.Unmarshal(d.Body, &msgdata); err != nil {
			msgdata = map[string]interface{}{
				"message": string(d.Body),
			}
		}

		go func(d *stomp.Message, m *Macro) {
			_, err := m.Call(msgdata, nil)
			if err != nil {
				if *flagDebug > 0 {
					log.Printf("%s consume message error: %+v\n===message===\n%s\n", m.name, err, string(d.Body))
				}
				return
			}
			if d.ShouldAck() {
				stompConn.Ack(d)
			}
		}(d, m)

	}

	if *flagDebug > 2 {
		log.Printf("STOMP message channel closed")
	}

}
