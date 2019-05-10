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
)

// MessageQueueConfig - 消息队列
type MessageQueueConfig struct {
	Driver  string //驱动，如amqp
	URL     string //连接地址
	Tag     string //标签
	Timeout int    // 超时
}

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
