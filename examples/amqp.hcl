_meta {
  mq {

    driver = "amqp"

    url = "amqp://guest:guest@localhost:5672/"

  }
}

_consumer {

  consume {

    queue = "app.sqlrestful.test.amqp"
  
    exchange = "test"

  }

  impl = "js"

  exec = <<JS
  (function(){
    log(JSON.stringify($input))
  })()
  JS

  format = "nil"

}

sender {

  path = "/send"

  bind {
    message = "输入的消息"
  }

  impl = "js"

  exec = <<JS

  (function(){
    emit_msg('app.sqlrestful.test.amqp', $input.message, {
      exchange: 'test'
    })
  })()

  JS

}