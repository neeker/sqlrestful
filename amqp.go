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

	"github.com/streadway/amqp"
)

// AMQPMessageQueueProvider - amqp
type AMQPMessageQueueProvider struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	tag      string
	macro    *Macro
	uri      string
	shutdown bool
	clock    *sync.RWMutex
}

// NewAMQP - 创建AMQP1.1x协议的Provider
func NewAMQP(m *Macro, config *MessageQueueConfig) (MessageQueueProvider, error) {

	p := &AMQPMessageQueueProvider{
		conn:     nil,
		channel:  nil,
		macro:    m,
		uri:      config.URI,
		tag:      config.Tag,
		shutdown: false,
		clock:    &sync.RWMutex{},
	}

	p.tag = config.Tag

	return p, nil
}

// IsShutdown - 是否已停止
func (c *AMQPMessageQueueProvider) IsShutdown() bool {
	c.clock.RLock()
	defer c.clock.RUnlock()
	return c.shutdown
}

// connect - 连接到AMQP
func (c *AMQPMessageQueueProvider) connect() (err error) {
	c.conn, err = amqp.Dial(c.uri)

	if err != nil {
		if *flagDebug > 0 {
			log.Printf("AMQP provider error: %v", err)
		}
		return err
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		return fmt.Errorf("AMQP channel error: %v", err)
	}
	return nil
}

// Consume - 消费数据
func (c *AMQPMessageQueueProvider) Consume() error {

	if err := c.connect(); err != nil {
		return err
	}

	args := c.macro.Consume

	exchangeName := args["name"]
	queueName := args["queue"]
	topicName := args["topic"]
	queueType := args["kind"]
	bindKey := args["key"]

	durable := args["durable"] == "" || strings.ToLower(args["durable"]) == "true"

	autoDelete := args["delete"] != "" && strings.ToLower(args["delete"]) == "auto"

	noWait := args["wait"] != "" && strings.ToLower(args["wait"]) == "true"

	noACK := args["ack"] != "" && strings.ToLower(args["ack"]) == "no"

	if topicName != "" {
		queueType = "topic"
		queueName = topicName
	} else if queueType == "" {
		queueType = "direct"
	}

	if queueName == "" {
		return fmt.Errorf("consume args must be set queue or topic or name")
	}

	if err := c.channel.ExchangeDeclare(
		exchangeName, // 名称
		queueType,    // 类型
		durable,      // durable
		autoDelete,   // delete when complete
		false,        // internal
		noWait,       // noWait
		nil,          // arguments
	); err != nil {
		return fmt.Errorf("AMQP Exchange Declare: %s", err)
	}

	queue, err := c.channel.QueueDeclare(
		queueName,  // name of the queue
		durable,    // durable
		autoDelete, // delete when unused
		false,      // exclusive
		noWait,     // noWait
		nil,        // arguments
	)

	if err != nil {
		return fmt.Errorf("AMQP Queue Declare error: %v", err)
	}

	if err = c.channel.QueueBind(
		queue.Name,   // name of the queue
		bindKey,      // bindingKey
		exchangeName, // sourceExchange
		noWait,       // noWait
		nil,          // arguments
	); err != nil {
		return fmt.Errorf("AMQP Queue Bind error: %s", err)
	}

	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		noACK,      // noAck
		false,      // exclusive
		false,      // noLocal
		noWait,     // noWait
		nil,        // arguments
	)

	go amqpHandleMessage(deliveries, c.macro)

	return nil
}

// Shutdown - 停止监听
func (c *AMQPMessageQueueProvider) Shutdown() error {
	if c.conn == nil {
		return nil
	}

	if c.channel != nil {
		if err := c.channel.Cancel(c.tag, true); err != nil {
			return err
		}
	}

	if err := c.conn.Close(); err != nil {
		return err
	}

	if *flagDebug > 2 {
		defer log.Printf("AMQP shutdown OK")
	}
	return nil
}

func amqpHandleMessage(deliveries <-chan amqp.Delivery, m *Macro) {
	for d := range deliveries {
		if *flagDebug > 2 {
			log.Printf(
				"AMQP got %dB delivery: [%v] %q",
				len(d.Body),
				d.DeliveryTag,
				d.Body,
			)
		}

		d.Ack(false)
	}

	if *flagDebug > 2 {
		log.Printf("AMQP deliveries channel closed")
	}

}
