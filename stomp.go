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
	"time"

	"github.com/go-stomp/stomp"
)

// STOMPMessageQueueProvider - STOP协议提供
type STOMPMessageQueueProvider struct {
	conn *stomp.Conn
}

var stompConn *stomp.Conn

// NewSTOMP - 创建STOMP协议提供器
func NewSTOMP(config *MessageQueueConfig) (MessageQueueProvider, error) {
	u, err := url.Parse(config.URI)
	if err != nil {
		return nil, err
	}

	c := &STOMPMessageQueueProvider{
		conn: nil,
	}

	connTimeout := config.Timeout
	if connTimeout <= 0 {
		connTimeout = 3
	}

	userPassword, _ := u.User.Password()

	c.conn, err = stomp.Dial(u.Scheme, u.Host,
		stomp.ConnOpt.Login(u.User.Username(), userPassword),
		stomp.ConnOpt.HeartBeat(time.Duration(connTimeout)*time.Second, time.Duration(connTimeout)*time.Second),
		stomp.ConnOpt.HeartBeatError(time.Duration(connTimeout*3)*time.Second))

	if err != nil {
		return nil, err
	}

	stompConn = c.conn

	return c, nil
}

// Consume - 消费数据
func (c *STOMPMessageQueueProvider) Consume(m *Macro) error {
	if c.conn == nil {
		return fmt.Errorf("STOMP not connect")
	}

	args := m.Consume

	queueName := args["queue"]
	if queueName == "" {
		queueName = args["topic"]
		if queueName == "" {
			queueName = args["name"]

			if queueName == "" {
				return fmt.Errorf("consume args must be set queue or topic or name")
			}
		}
	}

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
		return err
	}

	if *flagDebug > 0 {
		log.Printf("STOMP %s subscribed", queueName)
	}

	go stompHandleMessage(sub.C, m)

	return nil
}

// Shutdown - 停止监听
func (c *STOMPMessageQueueProvider) Shutdown() error {
	if c.conn == nil {
		return nil
	}
	if err := c.conn.Disconnect(); err != nil {
		return err
	}
	if *flagDebug > 2 {
		defer log.Printf("STOMP shutdown OK")
	}
	return nil
}

func stompHandleMessage(message <-chan *stomp.Message, m *Macro) {
	for d := range message {
		if d.Err != nil {
			if *flagDebug > 0 {
				log.Printf("STOMP subscribe error: %v", d.Err)
				log.Printf("STOMP well be resubscribe....")
			}
			go macrosManager.ConsumeMessage(m)
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
			err := m.Consume(msgdata)
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
