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
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

// AMQPEmitMessage - 发送消息结构
type AMQPEmitMessage struct {
	dest  string                 //目标队列
	msg   string                 //消息
	opts  map[string]interface{} //参数
	tries int                    //重发次数
}

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

// AMQPMessageSendProvider - send
type AMQPMessageSendProvider struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	args     map[string]string
	timeout  int
	uri      string
	tries    int
	message  chan *AMQPEmitMessage
	failmsg  *AMQPEmitMessage
	shutdown bool
	clock    *sync.RWMutex
}

// NewAMQP - 创建AMQP1.1x协议的Provider
func NewAMQP(m *Macro, config *MessageQueueConfig) (MessageQueueProvider, error) {

	p := &AMQPMessageQueueProvider{
		conn:     nil,
		channel:  nil,
		macro:    m,
		uri:      config.URL,
		tries:    0,
		clock:    &sync.RWMutex{},
		tag:      config.Tag,
		shutdown: false,
		failover: m.Consume["failover"] == "" ||
			strings.ToLower(m.Consume["failover"]) == "true" ||
			strings.ToLower(m.Consume["failover"]) == "on",
	}

	p.tag = config.Tag

	return p, nil
}

// ExchangeName - 消息队列名称
func (c *AMQPMessageQueueProvider) ExchangeName() string {
	exchangeName := c.macro.Consume["exchange"]
	if exchangeName != "" {
		return exchangeName
	}
	return "default"
}

// BindKey - 消息队列名称
func (c *AMQPMessageQueueProvider) BindKey() string {
	bindKey := c.macro.Consume["key"]
	if bindKey != "" {
		return bindKey
	}
	return c.QueueName()
}

// QueueName - 消息队列名称
func (c *AMQPMessageQueueProvider) QueueName() string {
	queueName := c.macro.Consume["queue"]
	if queueName == "" {
		queueName = c.macro.Consume["topic"]
		if queueName == "" {
			queueName = c.macro.Consume["name"]
		}
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
	if c.conn != nil && !c.conn.IsClosed() {
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
		if c.failover && c.tries > 1 {
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

	exchangeName := c.ExchangeName()

	bindKey := c.BindKey()

	durable := args["durable"] == "" || strings.ToLower(args["durable"]) == "true"
	autoDelete := args["delete"] != "" && strings.ToLower(args["delete"]) == "auto"
	noWait := args["wait"] != "" && strings.ToLower(args["wait"]) == "true"
	noACK := args["ack"] != "" && strings.ToLower(args["ack"]) == "no"
	var inputArgs map[string]interface{}

	if args["args"] != "" {
		if err := json.Unmarshal([]byte(args["args"]), &inputArgs); err != nil {
			log.Printf("%s consumer args error: %v\n", c.macro.name, args["args"])
		}
	}

	if err := c.channel.ExchangeDeclare(
		exchangeName, // 名称
		queueType,    // 类型
		durable,      // durable
		autoDelete,   // delete when complete
		false,        // internal
		noWait,       // noWait
		inputArgs,    // arguments
	); err != nil {
		if c.failover && c.tries > 1 {
			log.Printf("%s consume %s(%s) at exchange error: %+v", c.macro.name, queueName, queueType, err)
			go func() {
				if *flagDebug > 0 {
					log.Printf(" %s wait 500 ms retry consume %s(%s)", c.macro.name, queueName, queueType)
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
		if c.failover && c.tries > 1 {
			log.Printf("%s consume %s(%s) at queue error: %+v", c.macro.name, queueName, queueType, err)
			go func() {
				if *flagDebug > 0 {
					log.Printf(" %s wait 500 ms retry consume %s(%s)", c.macro.name, queueName, queueType)
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
		if c.failover && c.tries > 1 {
			log.Printf("%s consume %s(%s) at bind error: %+v", c.macro.name, queueName, queueType, err)
			go func() {
				if *flagDebug > 0 {
					log.Printf(" %s wait 500 ms retry consume %s(%s)", c.macro.name, queueName, queueType)
				}
				time.Sleep(500 * time.Millisecond)
				c.Consume()
			}()
			return nil
		}
		return err
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
		if c.failover && c.tries > 1 {
			log.Printf("%s consume %s(%s) at subsctribe error: %+v", c.macro.name, queueName, queueType, err)
			go func() {
				if *flagDebug > 0 {
					log.Printf("%s wait 500 ms retry consume %s(%s)", c.macro.name, queueName, queueType)
				}
				time.Sleep(500 * time.Millisecond)
				c.Consume()
			}()
			return nil
		}
		return err
	}

	if *flagDebug > 0 {
		log.Printf("%s consume %s(%s) subscribed", c.macro.name, queueName, queueType)
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

		var msgdata map[string]interface{}
		if err := json.Unmarshal(d.Body, &msgdata); err != nil {
			msgdata = map[string]interface{}{
				"data": string(d.Body),
			}
		}

		msgHeaders := map[string]interface{}{}

		for k, v := range d.Headers {
			msgHeaders[k] = v
		}

		msgHeaders["app_id"] = d.AppId
		msgHeaders["user_id"] = d.UserId
		msgHeaders["mime"] = d.ContentType
		msgHeaders["encoding"] = d.ContentEncoding
		msgHeaders["priority"] = d.Priority
		msgHeaders["delivery_mode"] = d.DeliveryMode
		msgHeaders["delivery_tag"] = d.DeliveryTag
		msgHeaders["reply_to"] = d.ReplyTo
		msgHeaders["correlation_id"] = d.CorrelationId
		msgHeaders["expiration"] = d.Expiration
		msgHeaders["message_id"] = d.MessageId
		msgHeaders["timestamp"] = d.Timestamp
		msgHeaders["type"] = d.Type
		msgHeaders["consumer_tag"] = d.ConsumerTag
		msgHeaders["exchange"] = d.Exchange
		msgHeaders["routing_key"] = d.RoutingKey
		msgHeaders["message_count"] = d.MessageCount

		msgdata["__header__"] = msgHeaders

		go func(d amqp.Delivery) {
			hasReply, outMsg, outHeader, err := c.macro.MsgCall(msgdata)

			if err != nil {
				log.Printf("%s consume message error: %+v\n===message===\n%s\n", c.macro.name, err, string(d.Body))
				return
			}

			if hasReply {
				sender, err := macrosManager.MessageSendProvider()

				if err != nil {
					log.Printf("%s consume message reply error: %v", c.macro.name, err)
					return
				}

				jsonData, _ := json.Marshal(outMsg)
				sender.EmitMessage(c.macro.ReplyDestName(), string(jsonData), outHeader)

				if err != nil {
					log.Printf("%s consume message reply error: %v", c.macro.name, err)
					return
				}
			}

			d.Ack(false)

		}(d)

	}

	if !c.IsShutdown() {
		go func() {
			if *flagDebug > 0 {
				log.Printf("wait 500 ms retry %s consume %s(%s) connect", c.macro.name, c.DestType(), c.QueueName())
			}
			time.Sleep(500 * time.Millisecond)
			c.Consume()
		}()
	} else if *flagDebug > 2 {
		log.Printf("%s consume %s(%s) deliveries channel closed", c.macro.name, c.DestType(), c.QueueName())
	}

}

// NewAMQPSender - 创建发送器
func NewAMQPSender(config *MessageQueueConfig) (MessageSendProvider, error) {
	connTimeout := config.Timeout
	if connTimeout <= 0 {
		connTimeout = 3
	}

	sender := &AMQPMessageSendProvider{
		conn:     nil,
		channel:  nil,
		timeout:  connTimeout,
		uri:      config.URL,
		tries:    0,
		message:  make(chan *AMQPEmitMessage),
		failmsg:  nil,
		shutdown: false,
		clock:    &sync.RWMutex{},
	}

	go sender.runLoop()

	return sender, nil
}

//connect - 连接到消息服务器
func (c *AMQPMessageSendProvider) connect() (err error) {
	if c.conn != nil && !c.conn.IsClosed() {
		if err := c.conn.Close(); err != nil {
			log.Printf("message sender close existed connect error: %+v", err)
		}
	}

	c.conn, err = amqp.Dial(c.uri)

	if err != nil {
		if *flagDebug > 0 {
			log.Printf("message sender open connect error: %+v", err)
		}
		return err
	}

	c.channel, err = c.conn.Channel()

	if err != nil {
		c.conn.Close()
		c.conn = nil
		log.Printf("message sender open channel error: %+v", err)
	}

	return nil
}

func (c *AMQPMessageSendProvider) publishMessage(msg *AMQPEmitMessage) error {

	args := msg.opts

	queueName := msg.dest

	var queueType string
	if args["kind"] != nil {
		queueType = strings.ToLower(fmt.Sprintf("%s", args["kind"]))
	}

	if queueType == "" {
		queueType = "direct"
	}

	var exchangeName string

	if args["exchange"] != nil {
		exchangeName = fmt.Sprintf("%s", args["exchange"])
	}

	if exchangeName == "" {
		exchangeName = "default"
	}

	durable := args["durable"] == nil || args["durable"] != nil && args["durable"] != "" && strings.ToLower(fmt.Sprintf("%s", args["durable"])) == "true"
	autoDelete := args["delete"] != nil && args["delete"] != "" && strings.ToLower(fmt.Sprintf("%s", args["delete"])) == "auto"
	noWait := args["wait"] != nil && args["wait"] != "" && strings.ToLower(fmt.Sprintf("%s", args["wait"])) == "true"
	mandatory := args["mandatory"] != nil && args["mandatory"] != "" && strings.ToLower(fmt.Sprintf("%s", args["mandatory"])) == "true"
	immediate := args["immediate"] != nil && args["immediate"] != "" && strings.ToLower(fmt.Sprintf("%s", args["immediate"])) == "true"

	persistentFlag := amqp.Transient
	if args["persistent"] != nil && args["persistent"] != "" && strings.ToLower(fmt.Sprintf("%s", args["persistent"])) == "true" {
		persistentFlag = amqp.Persistent
	}

	var priority uint8

	if args["priority"] != nil {
		switch args["priority"].(type) {
		case int:
			priority = args["priority"].(uint8)
		default:
			priority = 0
		}
	} else {
		priority = 0
	}

	var inputArgs map[string]interface{}

	if args["args"] != nil {
		switch args["args"].(type) {
		case map[string]interface{}:
			inputArgs = args["args"].(map[string]interface{})
		}
	}

	if err := c.channel.ExchangeDeclare(
		exchangeName, // 名称
		queueType,    // 类型
		durable,      // durable
		autoDelete,   // delete when complete
		false,        // internal
		noWait,       // noWait
		inputArgs,    // arguments
	); err != nil {
		return err
	}

	return c.channel.Publish(exchangeName, queueName, mandatory, immediate,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "UTF-8",
			Body:            []byte(msg.msg),
			DeliveryMode:    persistentFlag, // 1=non-persistent, 2=persistent
			Priority:        priority,       // 0-9
			// a bunch of application/implementation-specific fields
		})

}

// runLoop - 循环发送消息
func (c *AMQPMessageSendProvider) runLoop() {
reconn:

	if err := c.connect(); err != nil {
		if strings.Contains(err.Error(), "username or password not allowed") {
			log.Printf("message sender connect error: %v", err)
			time.Sleep(500 * time.Millisecond)
			goto reconn
		}
	}

	if c.failmsg != nil {
		msg := c.failmsg
		msg.tries++
		if err := c.publishMessage(msg); err != nil {
			log.Printf("message sender emit(%d) error: %v", msg.tries, err)
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

		if err := c.publishMessage(msg); err != nil {
			log.Printf("message sender emit error: %v", err)
			c.failmsg = msg
			time.Sleep(500 * time.Millisecond)
			goto reconn
		}

	}

}

//IsShutdown - 是否停止
func (c *AMQPMessageSendProvider) IsShutdown() bool {
	c.clock.RLock()
	defer c.clock.RUnlock()
	return c.shutdown
}

//Shutdown - 停止发送
func (c *AMQPMessageSendProvider) Shutdown() error {
	if c.conn == nil || c.IsShutdown() {
		return nil
	}

	c.clock.Lock()
	c.shutdown = true
	c.clock.Unlock()

	close(c.message)

	if err := c.conn.Close(); err != nil {
		log.Printf("shutdown message sender close connect error: %+v", err)
		return err
	}
	if *flagDebug > 2 {
		defer log.Printf("shutdown message sender successed")
	}
	return nil
}

// EmitMessage - 发送消息
func (c *AMQPMessageSendProvider) EmitMessage(dest string, msg string, opts map[string]interface{}) error {
	defer func() {
		if err := recover(); err != nil {
			if *flagDebug > 0 {
				log.Printf("emit to %s message error: %v\n===msg===\n%s\n\n", dest, err, msg)
			}
		}
	}()

	pmsg := &AMQPEmitMessage{
		dest:  dest,
		msg:   msg,
		opts:  opts,
		tries: 0,
	}
	c.message <- pmsg
	return nil
}
