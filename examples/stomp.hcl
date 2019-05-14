_meta {
  mq {

    driver = "stomp"

    url = "tcp://user:user9527@test.snz1.cn:61613/"

  }

}

_consumer {

  consume {
    // topicq前必须加/topic/，如：/topic/app.foo.bar
    name = "app.foo.bar"
  }

  impl = "js"

  exec = <<JS
  (function(){
    var msg = JSON.stringify($input)
    log("JSON MSG：" msg)
  })()
  JS

}

_reply {

  consume {
    name = "app.foo.bar.reply"
  }

  reply {
    name = "app.foo.bar"
  }

  impl = "js"

  exec = <<JS
  (function(){
    var msg = JSON.stringify($input)
    log(msg)
    return $input.message
  })()
  JS

}

start {

  path  = "/start"
  
  impl = "js"

  bind {
    message = "输入的消息"
  }

  exec = <<JS
  (function(){
    log("收到消息：", $input.message)
    var msg = {
      create_time: new Date(),
      message: $input.message
    }
    emit_msg('app.foo.bar.reply', msg)
  })()  
  JS

}

