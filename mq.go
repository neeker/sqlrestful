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
	"strings"
	"sync"
)

// MessageQueueProvider - 消息实现提供器
type MessageQueueProvider interface {
	//消费
	Consume() error
	//停止
	Shutdown() error
	//是否停止
	IsShutdown() bool
	//消息队列
	QueueName() string
	//消息l类型
	DestType() string
}

// MessageSendProvider - 消息发送提供器
type MessageSendProvider interface {
	//发送消息
	EmitMessage(string, string, map[string]interface{}) error
	//停止
	Shutdown() error
	//是否停止
	IsShutdown() bool
}

// MessageQueueConfig - 消息队列
type MessageQueueConfig struct {
	Driver    string              //驱动，如amqp
	URL       string              //连接地址
	Tag       string              //标签
	Timeout   int                 // 超时
	sender    MessageSendProvider //消息发送提供器
	sendMutex *sync.RWMutex       //发送锁
}

// IsMessageQueueEnabled - 判断是否启用了消息队列
func (c *MessageQueueConfig) IsMessageQueueEnabled() bool {
	return c.Driver != "" && c.URL != ""
}

// NewMessageQueueProvider - 初始化消息队列提供器
func (c *MessageQueueConfig) NewMessageQueueProvider(m *Macro) (err error) {

	if !c.IsMessageQueueEnabled() {
		return fmt.Errorf("message queue is disabled")
	}

	switch {
	case strings.ToLower(c.Driver) == "amqp":
		m.mqp, err = NewAMQP(m, c)
	case strings.ToLower(c.Driver) == "stomp":
		m.mqp, err = NewSTOMP(m, c)
	default:
		return fmt.Errorf("not found message queue driver(%s)", c.Driver)
	}

	if err != nil {
		if *flagDebug > 0 {
			log.Printf("create message provider error: %v", err)
		}
		return err
	}

	return nil
}

// HasMessageSendProvider - 是否有消息发送器
func (c *MessageQueueConfig) HasMessageSendProvider() bool {
	c.sendMutex.RLock()
	defer c.sendMutex.RUnlock()
	return c.sender != nil
}

// MessageSendProvider - 消息队列提供器
func (c *MessageQueueConfig) MessageSendProvider() (MessageSendProvider, error) {
	if !c.IsMessageQueueEnabled() {
		return nil, fmt.Errorf("message queue is disabled")
	}

	c.sendMutex.RLock()
	if c.sender != nil {
		c.sendMutex.RUnlock()
		return c.sender, nil
	}
	c.sendMutex.RUnlock()

	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()

	var err error

	switch {
	//case strings.ToLower(c.Driver) == "amqp":
	//c.sender, err = NewAMQPSender(c)
	case strings.ToLower(c.Driver) == "stomp":
		c.sender, err = NewSTOMPSender(c)
	default:
		err = fmt.Errorf("not found message queue driver(%s)", c.Driver)
	}

	return c.sender, err
}

// startMacrosConsumeMessage - 启动消息监听服务
func startMacrosConsumeMessage() error {
	if macrosManager.IsMessageQueueEnabled() {
		for _, n := range macrosManager.Names() {
			m := macrosManager.Get(n)
			if m.IsMessageConsumeEnabled() {
				if err := m.ConsumeMessage(); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// stopMacrosConsumeMessage - 停止消息监听服务
func stopMacrosConsumeMessage() {
	if macrosManager.IsMessageQueueEnabled() {
		for _, n := range macrosManager.Names() {
			m := macrosManager.Get(n)
			if m.IsMessageConsumeEnabled() {
				if err := m.ShutdownConsume(); err != nil {
					log.Printf("shutdown %s consume error: %+v", m.name, err)
				}
			}
		}
	}
}

// stopMessageSendProvider - 停止发送
func stopMessageSendProvider() {
	if !macrosManager.IsMessageQueueEnabled() {
		return
	}
	if !macrosManager.HasMessageSendProvider() {
		if *flagDebug > 0 {
			log.Printf("no message sender created")
		}
		return
	}
	sender, err := macrosManager.MessageSendProvider()
	if err != nil {
		return
	}
	sender.Shutdown()
	if *flagDebug > 0 {
		log.Print("message sender shutdown")
	}
}
