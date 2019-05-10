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
	"time"

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
	failover bool
	tries    int
	clock    *sync.RWMutex
}

// NewAMQP - 创建AMQP1.1x协议的Provider
func NewAMQP(m *Macro, config *MessageQueueConfig) (MessageQueueProvider, error) {

	p := &AMQPMessageQueueProvider{
		conn:     nil,
		channel:  nil,
		macro:    m,
		uri:      config.URL,
		tag:      config.Tag,
		shutdown: false,
		failover: false,
		tries:    0,
		clock:    &sync.RWMutex{},
	}

	p.tag = config.Tag

	return p, nil
}

// QueueName - 消息队列名称
func (c *AMQPMessageQueueProvider) QueueName() string {
	topicName := c.macro.Consume["topic"]
	queueName := c.macro.Consume["queue"]
	if topicName != "" {
		queueName = topicName
	}
	return queueName
}

// DestType - 目标类型
func (c *AMQPMessageQueueProvider) DestType() string {
	topicName := c.macro.Consume["topic"]
	queueType := strings.ToLower(c.macro.Consume["kind"])
	if topicName != "" {
		queueType = "topic"
	} else if queueType == "" {
		queueType = "direct"
	}
	if queueType == "" {
		queueType = "queue"
	}
	return queueType
}

// IsShutdown - 是否已停止
func (c *AMQPMessageQueueProvider) IsShutdown() bool {
	c.clock.RLock()
	defer c.clock.RUnlock()
	return c.shutdown
}

// connect - 连接到AMQP
func (c *AMQPMessageQueueProvider) connect() (err error) {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("%s consume %s(%s) at close old connect error : %+v", c.macro.name, c.DestType(), c.QueueName(), err)
		}
	}

	c.conn, err = amqp.Dial(c.uri)

	if err != nil {
		if *flagDebug > 0 {
			log.Printf("%s consume %s(%s) open connect error: %+v", c.macro.name, c.DestType(), c.QueueName(), err)
		}
		return err
	}

	c.channel, err = c.conn.Channel()

	if err != nil {
		c.conn.Close()
		c.conn = nil
		return fmt.Errorf("%s consume %s(%s) open channel error: %+v", c.macro.name, c.DestType(), c.QueueName(), err)
	}

	return nil
}

// Consume - 消费数据
func (c *AMQPMessageQueueProvider) Consume() error {
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
		if c.failover {
			log.Printf("%s consume %s(%s) close existed connect error: %+v", c.macro.name, queueType, queueName, err)
			go func() {
				if *flagDebug > 0 {
					log.Printf("%s wait 500 ms retry consume %s(%s)", c.macro.name, queueType, queueName)
				}
				time.Sleep(500 * time.Millisecond)
				c.Consume()
			}()
			return nil
		}
		return err
	}

	args := c.macro.Consume

	exchangeName := args["name"]
	bindKey := args["key"]
	durable := args["durable"] == "" || strings.ToLower(args["durable"]) == "true"
	autoDelete := args["delete"] != "" && strings.ToLower(args["delete"]) == "auto"
	noWait := args["wait"] != "" && strings.ToLower(args["wait"]) == "true"
	noACK := args["ack"] != "" && strings.ToLower(args["ack"]) == "no"

	if err := c.channel.ExchangeDeclare(
		exchangeName, // 名称
		queueType,    // 类型
		durable,      // durable
		autoDelete,   // delete when complete
		false,        // internal
		noWait,       // noWait
		nil,          // arguments
	); err != nil {
		if c.failover {
			log.Printf("%s consume %s(%s) at exchange error: %+v", c.macro.name, queueType, queueName, err)
			go func() {
				if *flagDebug > 0 {
					log.Printf(" %s wait 500 ms retry consume %s(%s)", c.macro.name, queueType, queueName)
				}
				time.Sleep(500 * time.Millisecond)
				c.Consume()
			}()
			return nil
		}
		return err
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
		if c.failover {
			log.Printf("%s consume %s(%s) at queue error: %+v", c.macro.name, queueType, queueName, err)
			go func() {
				if *flagDebug > 0 {
					log.Printf(" %s wait 500 ms retry consume %s(%s)", c.macro.name, queueType, queueName)
				}
				time.Sleep(500 * time.Millisecond)
				c.Consume()
			}()
			return nil
		}
		return err
	}

	if err = c.channel.QueueBind(
		queue.Name,   // name of the queue
		bindKey,      // bindingKey
		exchangeName, // sourceExchange
		noWait,       // noWait
		nil,          // arguments
	); err != nil {
		if c.failover {
			log.Printf("%s consume %s(%s) at bind error: %+v", c.macro.name, queueType, queueName, err)
			go func() {
				if *flagDebug > 0 {
					log.Printf(" %s wait 500 ms retry consume %s(%s)", c.macro.name, queueType, queueName)
				}
				time.Sleep(500 * time.Millisecond)
				c.Consume()
			}()
			return nil
		}
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

	if err != nil {
		if c.failover {
			log.Printf("%s consume %s(%s) at subsctribe error: %+v", c.macro.name, queueType, queueName, err)
			go func() {
				if *flagDebug > 0 {
					log.Printf("%s wait 500 ms retry consume %s(%s)", c.macro.name, queueType, queueName)
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

	go amqpHandleMessage(deliveries, c)

	return nil
}

// Shutdown - 停止监听
func (c *AMQPMessageQueueProvider) Shutdown() error {
	if c.conn == nil || c.IsShutdown() {
		return nil
	}

	c.clock.Lock()
	c.shutdown = true
	c.clock.Unlock()

	if c.channel != nil {
		if err := c.channel.Cancel(c.tag, true); err != nil && *flagDebug > 0 {
			log.Printf("%s shutdown consume %s(%s) at cancel error: %+v", c.macro.name, c.DestType(), c.QueueName(), err)
		}
	}

	if err := c.conn.Close(); err != nil && *flagDebug > 0 {
		log.Printf("%s shutdown consume %s(%s) at close error: %+v", c.macro.name, c.DestType(), c.QueueName(), err)
		return err
	}

	if *flagDebug > 2 {
		defer log.Printf("%s shutdown consume %s(%s) successed", c.macro.name, c.DestType(), c.QueueName())
	}

	return nil
}

func amqpHandleMessage(deliveries <-chan amqp.Delivery, c *AMQPMessageQueueProvider) {
	for d := range deliveries {
		if *flagDebug > 2 {
			log.Printf(
				"%s consume %s(%s) got %dB delivery: [%v] %v",
				c.macro.name,
				c.DestType(),
				c.QueueName(),
				len(d.Body),
				d.DeliveryTag,
				d.Body,
			)
		}
		d.Ack(false)
	}

	if *flagDebug > 2 {
		log.Printf("%s consume %s(%s) deliveries channel closed", c.macro.name, c.DestType(), c.QueueName())
	}

}
