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

// STOMPEmitMessage - STOMP 发送消息结构
type STOMPEmitMessage struct {
	dest  string                 //目标队列
	msg   string                 //消息
	opts  map[string]interface{} //参数
	tries int                    //重发次数
}

// STOMPMessageQueueProvider - STOMP协议消息接收器
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

// STOMPMessageSendProvider - STOMP协议消息发送器
type STOMPMessageSendProvider struct {
	conn     *stomp.Conn
	timeout  int
	uri      *url.URL
	tries    int
	message  chan *STOMPEmitMessage
	failmsg  *STOMPEmitMessage
	shutdown bool
	clock    *sync.RWMutex
}

var stompConn *stomp.Conn

// NewSTOMP - 创建STOMP协议提供器
func NewSTOMP(macro *Macro, config *MessageQueueConfig) (MessageQueueProvider, error) {
	u, err := url.Parse(config.URL)
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

// NewSTOMPSender - 创建发送器
func NewSTOMPSender(config *MessageQueueConfig) (MessageSendProvider, error) {
	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	connTimeout := config.Timeout
	if connTimeout <= 0 {
		connTimeout = 3
	}

	sender := &STOMPMessageSendProvider{
		conn:     nil,
		timeout:  connTimeout,
		uri:      u,
		tries:    0,
		message:  make(chan *STOMPEmitMessage),
		failmsg:  nil,
		shutdown: false,
		clock:    &sync.RWMutex{},
	}

	go sender.runLoop()

	return sender, nil
}

// IsShutdown - 是否已停止
func (c *STOMPMessageQueueProvider) IsShutdown() bool {
	c.clock.RLock()
	defer c.clock.RUnlock()
	return c.shutdown
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

// DestType - 目标类型
func (c *STOMPMessageQueueProvider) DestType() string {
	destType := "queue"
	if strings.HasPrefix(c.QueueName(), "/topic/") {
		destType = "topic"
	}
	return destType
}

func (c *STOMPMessageQueueProvider) connect() (err error) {
	if c.conn != nil {
		if err := c.conn.Disconnect(); err != nil {
			log.Printf("%s consume %s(%s) close existed connect error: %+v", c.macro.name, c.DestType(), c.QueueName(), err)
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
		if *flagDebug > 0 {
			log.Printf("%s consume %s(%s) open connect error: %+v", c.macro.name, c.DestType(), c.QueueName(), err)
		}
		return err
	}

	return nil
}

// Consume - 消费数据
func (c *STOMPMessageQueueProvider) Consume() error {

	if c.IsShutdown() {
		return nil
	}

	queueName := c.QueueName()
	queueType := c.DestType()

	if queueName == "" {
		return fmt.Errorf("%s consume args must be set queue or topic", c.macro.name)
	}

	c.tries++

	if err := c.connect(); err != nil {
		if strings.Contains(err.Error(), "password is invalid") {
			return err
		}

		if c.failover {
			log.Printf("%s consume %s(%s) connect error: %+v", c.macro.name, queueType, queueName, err)
			go func() {
				if *flagDebug > 0 {
					log.Printf("%s wait 500 ms retry consume %s(%s) connect", c.macro.name, queueType, queueName)
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
			log.Printf("%s consume %s(%s) subscribe error: %+v", c.macro.name, queueType, queueName, err)
			go func() {
				if *flagDebug > 0 {
					log.Printf("%s wait 500 ms retry consume %s(%s) connect", c.macro.name, queueType, queueName)
				}
				time.Sleep(500 * time.Millisecond)
				c.Consume()
			}()
			return nil
		}
		return err
	}

	if *flagDebug > 0 {
		log.Printf("%s consume %s(%s) subscribed", c.macro.name, queueType, queueName)
	}

	go stompHandleMessage(sub.C, c)

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
		log.Printf("%s shutdown consume %s(%s) at cancel error: %+v", c.macro.name, c.DestType(), c.QueueName(), err)
		return err
	}
	if *flagDebug > 2 {
		defer log.Printf("%s shutdown consume %s(%s) successed", c.macro.name, c.DestType(), c.QueueName())
	}
	return nil
}

// stompHandleMessage - 处理消息
func stompHandleMessage(message <-chan *stomp.Message, c *STOMPMessageQueueProvider) {
	for d := range message {
		if d.Err != nil {
			if *flagDebug > 0 && !c.IsShutdown() {
				log.Printf("%s consume %s(%s) subscribe error message", c.macro.name, c.DestType(), c.QueueName())
			}
			if !c.IsShutdown() {
				go func() {
					if *flagDebug > 0 {
						log.Printf("wait 500 ms retry %s consume %s(%s) connect", c.macro.name, c.DestType(), c.QueueName())
					}
					time.Sleep(500 * time.Millisecond)
					c.Consume()
				}()
			}
			return
		}
		if *flagDebug > 2 {
			log.Printf(
				"%s consume %s(%s) got %dB delivery: %v",
				c.macro.name,
				c.DestType(),
				c.QueueName(),
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

		go func(d *stomp.Message, c *STOMPMessageQueueProvider) {
			_, err := c.macro.Call(msgdata, nil)
			if err != nil {
				if *flagDebug > 0 {
					log.Printf("%s consume message error: %+v\n===message===\n%s\n", c.macro.name, err, string(d.Body))
				}
				return
			}
			if d.ShouldAck() {
				stompConn.Ack(d)
			}
		}(d, c)

	}

	if *flagDebug > 2 {
		log.Printf("%s consume %s(%s) deliveries channel closed", c.macro.name, c.DestType(), c.QueueName())
	}

}

//connect - 连接到消息服务器
func (c *STOMPMessageSendProvider) connect() (err error) {
	if c.conn != nil {
		if err := c.conn.Disconnect(); err != nil {
			log.Printf("message sender close existed connect error: %+v", err)
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
		if *flagDebug > 0 {
			log.Printf("message sender open connect error: %+v", err)
		}
		return err
	}

	return nil
}

// runLoop - 循环发送消息
func (c *STOMPMessageSendProvider) runLoop() {
reconn:
	if c.IsShutdown() {
		return
	}
	c.tries++
	if err := c.connect(); err != nil {
		log.Printf("message sender connect error: %v", err)
		time.Sleep(500 * time.Millisecond)
		goto reconn
	}

	if c.failmsg != nil {
		msg := c.failmsg
		msg.tries++
		if err := c.conn.Send(msg.dest, "", []byte(msg.msg)); err != nil {
			log.Printf("message sender emit error: %v", err)
			time.Sleep(500 * time.Millisecond)
			goto reconn
		}
		c.failmsg = nil
	}

	for {
		msg := <-c.message
		if msg == nil {
			if c.IsShutdown() {
				log.Print("message sender shutdown singal")
				return
			}
			time.Sleep(500 * time.Millisecond)
			goto reconn
		}
		msg.tries++
		if err := c.conn.Send(msg.dest, "", []byte(msg.msg)); err != nil {
			log.Printf("message sender emit error: %v", err)
			c.failmsg = msg
			time.Sleep(500 * time.Millisecond)
			goto reconn
		}
	}

}

//IsShutdown - 是否停止
func (c *STOMPMessageSendProvider) IsShutdown() bool {
	c.clock.RLock()
	defer c.clock.RUnlock()
	return c.shutdown
}

//Shutdown - 停止发送
func (c *STOMPMessageSendProvider) Shutdown() error {
	if c.conn == nil || c.IsShutdown() {
		return nil
	}

	c.clock.Lock()
	c.shutdown = true
	c.clock.Unlock()

	close(c.message)

	if err := c.conn.Disconnect(); err != nil {
		log.Printf("shutdown message sender close connect error: %+v", err)
		return err
	}
	if *flagDebug > 2 {
		defer log.Printf("shutdown message sender successed")
	}
	return nil
}

// EmitMessage - 发送消息
func (c *STOMPMessageSendProvider) EmitMessage(dest string, msg string, opts map[string]interface{}) error {
	defer func() {
		if err := recover(); err != nil {
			if *flagDebug > 0 {
				log.Printf("emit to %s message error: %v\n===msg===\n%s\n\n", dest, err, msg)
			}
		}
	}()

	pmsg := &STOMPEmitMessage{
		dest:  dest,
		msg:   msg,
		opts:  opts,
		tries: 0,
	}
	c.message <- pmsg
	return nil
}
