# 监听消息队列与发送消息

## 应用场景

有时可能需要接收来自消息消息队列中的消息入库或执行某个操作（如：流式数据计算时），或者把`Restful`请求转入消息队列时就需要使用消息处理机制。

在`SQLRestful`中我们也是使用配置式的方式使用消息队列能力。

## 使用消息队列

## 监听消息

```hcl

//服务配置
_meta {
  //消息驱动配置
  mq {
    //消息驱动
    driver = "stomp"
    //连接地址
    url = "tcp://stomp_host:port/"
  }
}

_msg_handler {

  //监听队列
  consume {
    //消息队列名称（或主题名）
    name = "queue_name"
    //ACK标记：auto（默认）、client、each
    ack = "auto"
    //...
  }

  //SQL参数绑定：$input表示来自消息队列中的参数，如$input.xxx
  bind {
    // ...
  }

  //执行代码
  exec = <<SQL
    // ...
  SQL

  //结果应答队列
  reply {
    //队列名称（或主题名称）
    name = "replay_name"
    //ACK标记：auto（默认）、client、each
    ack = "auto"
    //应答头
    header = {
      // ...
    }
  }

  //...

}

```

### 发送消息

```js
(function(){
  emit_msg("queue_name", "msg object", {
    //消息头
  })
})()
```

