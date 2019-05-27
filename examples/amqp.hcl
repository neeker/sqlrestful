_meta {
  mq {

    driver = "amqp"

    url = "amqp://guest:guest@192.168.1.201:5672/"

  }
}

_consumer {

  consume {

    queue = "app.sqlrestful.test.amqp"

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
    var msg = {
      create_time: new Date(),
      message: $input.message
    }
    emit_msg('app.sqlrestful.test.amqp', msg)
  })()

  JS

}